package v1beta1connect

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/authenticate/token"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/server/consts"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"

	"github.com/raystack/frontier/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (h *ConnectHandler) Authenticate(ctx context.Context, request *connect.Request[frontierv1beta1.AuthenticateRequest]) (*connect.Response[frontierv1beta1.AuthenticateResponse], error) {
	returnToURL := h.authnService.SanitizeReturnToURL(request.Msg.GetReturnTo())
	callbackURL := h.authnService.SanitizeCallbackURL(request.Msg.GetCallbackUrl())

	// check if user is already logged in
	session, err := h.sessionService.ExtractFromContext(ctx)
	if err == nil && session.IsValid(time.Now().UTC()) {
		// already logged in, set location header for return to?
		resp := connect.NewResponse(&frontierv1beta1.AuthenticateResponse{})
		if len(returnToURL) != 0 {
			resp.Header().Set(consts.LocationGatewayKey, returnToURL)
		}
		return resp, nil
	} else if err != nil && !errors.Is(err, frontiersession.ErrNoSession) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if (request.Msg.GetStrategyName() == authenticate.MailLinkAuthMethod.String() || request.Msg.GetStrategyName() == authenticate.MailOTPAuthMethod.String()) && !isValidEmail(request.Msg.GetEmail()) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidEmail)
	}

	// not logged in, try registration
	response, err := h.authnService.StartFlow(ctx, authenticate.RegistrationStartRequest{
		Method:      request.Msg.GetStrategyName(),
		ReturnToURL: returnToURL,
		CallbackUrl: callbackURL,
		Email:       request.Msg.GetEmail(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// set location header for redirect to start auth?
	resp := connect.NewResponse(&frontierv1beta1.AuthenticateResponse{
		Endpoint: response.Flow.StartURL,
		State:    response.State,
	})

	if request.Msg.GetRedirectOnstart() && len(response.Flow.StartURL) > 0 {
		resp.Header().Set(consts.LocationGatewayKey, response.Flow.StartURL)
	}

	if request.Msg.GetStrategyName() == authenticate.PassKeyAuthMethod.String() {
		userCredentils := &structpb.Value{}
		if err = userCredentils.UnmarshalJSON(response.StateConfig["options"].([]byte)); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		typeValue, ok := response.Flow.Metadata["passkey_type"].(string)
		if !ok {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		stringValue := &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: typeValue,
			},
		}
		stateOptionsValue := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"options": userCredentils,
				"type":    stringValue,
			},
		}

		resp.Msg.StateOptions = stateOptionsValue
		return resp, nil
	}

	return resp, nil
}

func (h *ConnectHandler) AuthCallback(ctx context.Context, request *connect.Request[frontierv1beta1.AuthCallbackRequest]) (*connect.Response[frontierv1beta1.AuthCallbackResponse], error) {
	// handle callback
	response, err := h.authnService.FinishFlow(ctx, authenticate.RegistrationFinishRequest{
		Method:      request.Msg.GetStrategyName(),
		Code:        request.Msg.GetCode(),
		State:       request.Msg.GetState(),
		StateConfig: request.Msg.GetStateOptions().AsMap(),
	})
	if err != nil {
		if errors.Is(err, authenticate.ErrInvalidMailOTP) || errors.Is(err, authenticate.ErrMissingOIDCCode) || errors.Is(err, authenticate.ErrInvalidOIDCState) || errors.Is(err, authenticate.ErrFlowInvalid) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// registration/login complete, build a session
	session, err := h.sessionService.Create(ctx, response.User.ID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	// create response and set headers
	resp := connect.NewResponse(&frontierv1beta1.AuthCallbackResponse{})

	// save in browser cookies
	resp.Header().Set(consts.SessionIDGatewayKey, session.ID.String())

	// set location header for redirect to finish auth and send client to origin
	if len(response.Flow.FinishURL) > 0 {
		resp.Header().Set(consts.LocationGatewayKey, response.Flow.FinishURL)
	}
	return resp, nil
}

func (h *ConnectHandler) AuthToken(ctx context.Context, request *connect.Request[frontierv1beta1.AuthTokenRequest]) (*connect.Response[frontierv1beta1.AuthTokenResponse], error) {
	logger := ExtractLogger(ctx)
	existingMD := metadata.New(map[string]string{})
	switch request.Msg.GetGrantType() {
	case "client_credentials":
		if request.Msg.GetClientId() != "" && request.Msg.GetClientSecret() != "" {
			secretVal := base64.StdEncoding.EncodeToString([]byte(request.Msg.GetClientId() + ":" + request.Msg.GetClientSecret()))
			existingMD.Set(consts.UserSecretGatewayKey, secretVal)
		}
	case "urn:ietf:params:oauth:grant-type:jwt-bearer":
		if request.Msg.GetAssertion() != "" {
			existingMD.Set(consts.UserTokenGatewayKey, request.Msg.GetAssertion())
		}
	}
	ctx = metadata.NewIncomingContext(ctx, existingMD)

	// only get principal from service user assertions
	principal, err := h.GetLoggedInPrincipal(ctx,
		authenticate.SessionClientAssertion,
		authenticate.ClientCredentialsClientAssertion,
		authenticate.JWTGrantClientAssertion)
	if err != nil {
		logger.Debug(fmt.Sprintf("unable to get GetLoggedInPrincipal: %v", err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if principal.Type == schema.ServiceUserPrincipal {
		orgId := principal.ServiceUser.OrgID
		_, err := h.orgService.Get(ctx, orgId)
		if err != nil {
			logger.Error(fmt.Sprintf("error while fetching service user org: %v", err))
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	token, err := h.getAccessToken(ctx, principal, request.Header().Values(consts.ProjectRequestKey))
	if err != nil {
		logger.Debug(fmt.Sprintf("unable to get accessToken: %v", err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resp := connect.NewResponse(&frontierv1beta1.AuthTokenResponse{
		AccessToken: string(token),
		TokenType:   "Bearer",
	})

	resp.Header().Set(consts.UserTokenGatewayKey, string(token))
	return resp, nil
}

// getAccessToken generates a jwt access token with user/org details
func (h *ConnectHandler) getAccessToken(ctx context.Context, principal authenticate.Principal, projectKey []string) ([]byte, error) {
	logger := ExtractLogger(ctx)
	customClaims := map[string]string{}

	if h.authConfig.Token.Claims.AddOrgIDsClaim {
		// get orgs a user belongs to
		orgs, err := h.orgService.ListByUser(ctx, principal, organization.Filter{})
		if err != nil {
			return nil, err
		}

		var orgIds []string
		for _, o := range orgs {
			orgIds = append(orgIds, o.ID)
		}
		customClaims[token.OrgIDsClaimKey] = strings.Join(orgIds, ",")
	}

	// find selected project id
	if len(projectKey) > 0 && projectKey[0] != "" {
		// check if project exists and user has access to it
		proj, err := h.projectService.Get(ctx, projectKey[0])
		if err != nil {
			logger.Error("error getting project", zap.Error(err), zap.String("project", projectKey[0]))
		} else {
			if err := h.IsAuthorized(ctx, relation.Object{
				Namespace: schema.ProjectNamespace,
				ID:        proj.ID,
			}, schema.GetPermission); err == nil {
				customClaims["project_id"] = proj.ID
			} else {
				logger.Warn("error checking project access", zap.Error(err), zap.String("project", proj.ID), zap.String("principal", principal.ID))
			}
		}
	}

	// build jwt for user context
	return h.authnService.BuildToken(ctx, principal, customClaims)
}

func (h *ConnectHandler) AuthLogout(ctx context.Context, request *connect.Request[frontierv1beta1.AuthLogoutRequest]) (*connect.Response[frontierv1beta1.AuthLogoutResponse], error) {
	logger := ExtractLogger(ctx)

	// delete user session if exists
	sessionID, err := h.getLoggedInSessionID(ctx)
	if err == nil {
		if err = h.sessionService.Delete(ctx, sessionID); err != nil {
			logger.Error(err.Error())
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	resp := connect.NewResponse(&frontierv1beta1.AuthLogoutResponse{})

	// instruct interceptor to invalidate cookie
	resp.Header().Set(consts.SessionDeleteGatewayKey, "true")
	return resp, nil
}

func (h *ConnectHandler) getLoggedInSessionID(ctx context.Context) (uuid.UUID, error) {
	session, err := h.sessionService.ExtractFromContext(ctx)
	if err == nil && session.IsValid(time.Now().UTC()) {
		return session.ID, nil
	}
	return uuid.Nil, err
}

func (h *ConnectHandler) GetLoggedInPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error) {
	principal, err := h.authnService.GetPrincipal(ctx, via...)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotExist), errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidEmail):
			return principal, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
		case errors.Is(err, errors.ErrUnauthenticated):
			return principal, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
		default:
			return principal, connect.NewError(connect.CodeInternal, err)
		}
	}
	return principal, nil
}

func (h *ConnectHandler) ListAuthStrategies(ctx context.Context, request *connect.Request[frontierv1beta1.ListAuthStrategiesRequest]) (*connect.Response[frontierv1beta1.ListAuthStrategiesResponse], error) {
	var pbstrategy []*frontierv1beta1.AuthStrategy
	for _, strategy := range h.authnService.SupportedStrategies() {
		pbstrategy = append(pbstrategy, &frontierv1beta1.AuthStrategy{
			Name:   strategy,
			Params: nil,
		})
	}
	return connect.NewResponse(&frontierv1beta1.ListAuthStrategiesResponse{Strategies: pbstrategy}), nil
}
