package v1beta1connect

import (
	"context"
	"encoding/base64"
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
	sessionutils "github.com/raystack/frontier/pkg/session"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"

	"github.com/raystack/frontier/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (h *ConnectHandler) Authenticate(ctx context.Context, request *connect.Request[frontierv1beta1.AuthenticateRequest]) (*connect.Response[frontierv1beta1.AuthenticateResponse], error) {
	errorLogger := NewErrorLogger()

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
		errorLogger.LogUnexpectedError(ctx, request, "Authenticate", err)
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
		errorLogger.LogUnexpectedError(ctx, request, "Authenticate", err,
			zap.String("strategy", request.Msg.GetStrategyName()),
			zap.String("email", request.Msg.GetEmail()))
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
			errorLogger.LogUnexpectedError(ctx, request, "Authenticate", err,
				zap.String("strategy", authenticate.PassKeyAuthMethod.String()))
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		typeValue, ok := response.Flow.Metadata["passkey_type"].(string)
		if !ok {
			errorLogger.LogUnexpectedError(ctx, request, "Authenticate", err,
				zap.String("strategy", authenticate.PassKeyAuthMethod.String()))
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
	errorLogger := NewErrorLogger()

	// handle callback
	response, err := h.authnService.FinishFlow(ctx, authenticate.RegistrationFinishRequest{
		Method:      request.Msg.GetStrategyName(),
		Code:        request.Msg.GetCode(),
		State:       request.Msg.GetState(),
		StateConfig: request.Msg.GetStateOptions().AsMap(),
	})
	if err != nil {
		if errors.Is(err, authenticate.ErrInvalidMailOTP) || errors.Is(err, authenticate.ErrMissingOIDCCode) || errors.Is(err, authenticate.ErrInvalidOIDCState) || errors.Is(err, authenticate.ErrFlowInvalid) {
			errorLogger.LogServiceError(ctx, request, "AuthCallback.FinishFlow", err,
				zap.String("strategy", request.Msg.GetStrategyName()),
				zap.String("state", request.Msg.GetState()))
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		errorLogger.LogUnexpectedError(ctx, request, "AuthCallback", err,
			zap.String("strategy", request.Msg.GetStrategyName()),
			zap.String("state", request.Msg.GetState()))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Extract session metadata from request headers
	sessionMetadata := sessionutils.ExtractSessionMetadata(ctx, request, h.authConfig.Session.Headers)

	// registration/login complete, build a session
	session, err := h.sessionService.Create(ctx, response.User.ID, sessionMetadata)
	if err != nil {
		errorLogger.LogUnexpectedError(ctx, request, "AuthCallback", err,
			zap.String("user_id", response.User.ID))
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
	errorLogger := NewErrorLogger()
	// Get existing metadata from context to preserve session info from interceptor
	existingMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		existingMD = metadata.New(map[string]string{})
	} else {
		// Clone existing metadata to avoid modifying the original
		existingMD = existingMD.Copy()
	}

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
		errorLogger.LogUnexpectedError(ctx, request, "AuthToken", err,
			zap.String("grant_type", request.Msg.GetGrantType()))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	if principal.Type == schema.ServiceUserPrincipal {
		orgId := principal.ServiceUser.OrgID
		_, err := h.orgService.Get(ctx, orgId)
		if err != nil {
			errorLogger.LogUnexpectedError(ctx, request, "AuthToken", err,
				zap.String("org_id", orgId),
				zap.String("service_user_id", principal.ServiceUser.ID))
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	token, err := h.getAccessToken(ctx, principal, request.Header().Values(consts.ProjectRequestKey), request)
	if err != nil {
		errorLogger.LogUnexpectedError(ctx, request, "AuthToken", err,
			zap.String("principal_id", principal.ID),
			zap.String("principal_type", principal.Type))
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
func (h *ConnectHandler) getAccessToken(ctx context.Context, principal authenticate.Principal, projectKey []string, request connect.AnyRequest) ([]byte, error) {
	errorLogger := NewErrorLogger()
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

	// add session ID as claims for upstream
	if h.authConfig.Token.Claims.AddSessionIDClaim && principal.Type == schema.UserPrincipal {
		if sessionID, err := h.getLoggedInSessionID(ctx); err == nil {
			customClaims[token.SessionIDClaimKey] = sessionID.String()
		} else {
			errorLogger.LogUnexpectedError(ctx, request, "getAccessToken", err, zap.String("principal", principal.ID))
		}
	}

	// find selected project id
	if len(projectKey) > 0 && projectKey[0] != "" {
		// check if project exists and user has access to it
		proj, err := h.projectService.Get(ctx, projectKey[0])
		if err != nil {
			errorLogger.LogUnexpectedError(ctx, request, "getAccessToken", err, zap.String("project", projectKey[0]))
		} else {
			if err := h.IsAuthorized(ctx, relation.Object{
				Namespace: schema.ProjectNamespace,
				ID:        proj.ID,
			}, schema.GetPermission); err == nil {
				customClaims["project_id"] = proj.ID
			} else {
				errorLogger.LogUnexpectedError(ctx, request, "getAccessToken", err, zap.String("project", proj.ID), zap.String("principal", principal.ID))
			}
		}
	}

	// build jwt for user context
	return h.authnService.BuildToken(ctx, principal, customClaims)
}

func (h *ConnectHandler) AuthLogout(ctx context.Context, request *connect.Request[frontierv1beta1.AuthLogoutRequest]) (*connect.Response[frontierv1beta1.AuthLogoutResponse], error) {
	errorLogger := NewErrorLogger()

	// delete user session if exists
	sessionID, err := h.getLoggedInSessionID(ctx)
	if err == nil {
		if err = h.sessionService.Delete(ctx, sessionID); err != nil {
			errorLogger.LogUnexpectedError(ctx, request, "AuthLogout", err,
				zap.String("session_id", sessionID.String()))
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

func (h *ConnectHandler) GetJWKs(ctx context.Context, request *connect.Request[frontierv1beta1.GetJWKsRequest]) (*connect.Response[frontierv1beta1.GetJWKsResponse], error) {
	errorLogger := NewErrorLogger()

	keySet := h.authnService.JWKs(ctx)
	jwks, err := toJSONWebKey(keySet)
	if err != nil {
		errorLogger.LogUnexpectedError(ctx, request, "GetJWKs", err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.GetJWKsResponse{
		Keys: jwks.Keys,
	}), nil
}
