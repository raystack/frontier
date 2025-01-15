package v1beta1

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/raystack/frontier/core/authenticate/token"

	"github.com/raystack/frontier/core/organization"

	"github.com/raystack/frontier/core/relation"

	"github.com/raystack/frontier/internal/bootstrap/schema"
	"go.uber.org/zap"
	structpb "google.golang.org/protobuf/types/known/structpb"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/raystack/frontier/pkg/server/consts"

	"github.com/google/uuid"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthnService interface {
	StartFlow(ctx context.Context, request authenticate.RegistrationStartRequest) (*authenticate.RegistrationStartResponse, error)
	FinishFlow(ctx context.Context, request authenticate.RegistrationFinishRequest) (*authenticate.RegistrationFinishResponse, error)
	BuildToken(ctx context.Context, principalID string, metadata map[string]string) ([]byte, error)
	JWKs(ctx context.Context) jwk.Set
	GetPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error)
	SupportedStrategies() []string
	InitFlows(ctx context.Context) error
	SanitizeReturnToURL(url string) string
	SanitizeCallbackURL(url string) string
}

type SessionService interface {
	ExtractFromContext(ctx context.Context) (*frontiersession.Session, error)
	Create(ctx context.Context, userID string) (*frontiersession.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
	Refresh(ctx context.Context, sessionID uuid.UUID) error
}

func (h Handler) Authenticate(ctx context.Context, request *frontierv1beta1.AuthenticateRequest) (*frontierv1beta1.AuthenticateResponse, error) {
	returnToURL := h.authnService.SanitizeReturnToURL(request.GetReturnTo())
	callbackURL := h.authnService.SanitizeCallbackURL(request.GetCallbackUrl())

	// check if user is already logged in
	session, err := h.sessionService.ExtractFromContext(ctx)
	if err == nil && session.IsValid(time.Now().UTC()) {
		// already logged in, set location header for return to?
		if len(returnToURL) != 0 {
			if err = setRedirectHeaders(ctx, returnToURL); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
		return &frontierv1beta1.AuthenticateResponse{}, nil
	} else if err != nil && !errors.Is(err, frontiersession.ErrNoSession) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if (request.GetStrategyName() == authenticate.MailLinkAuthMethod.String() || request.GetStrategyName() == authenticate.MailOTPAuthMethod.String()) && !isValidEmail(request.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument, "Invalid email")
	}

	// not logged in, try registration
	response, err := h.authnService.StartFlow(ctx, authenticate.RegistrationStartRequest{
		Method:      request.GetStrategyName(),
		ReturnToURL: returnToURL,
		CallbackUrl: callbackURL,
		Email:       request.GetEmail(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// set location header for redirect to start auth?
	if request.GetRedirectOnstart() && len(response.Flow.StartURL) > 0 {
		if err = setRedirectHeaders(ctx, response.Flow.StartURL); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if request.GetStrategyName() == authenticate.PassKeyAuthMethod.String() {
		userCredentils := &structpb.Value{}
		if err = userCredentils.UnmarshalJSON(response.StateConfig["options"].([]byte)); err != nil {
			return nil, err
		}
		typeValue, ok := response.Flow.Metadata["passkey_type"].(string)
		if !ok {
			return nil, err
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
		return &frontierv1beta1.AuthenticateResponse{
			Endpoint:     response.Flow.StartURL,
			State:        response.State,
			StateOptions: stateOptionsValue,
		}, nil
	}

	return &frontierv1beta1.AuthenticateResponse{
		Endpoint: response.Flow.StartURL,
		// Note(kushsharma): can we can also store the state in cookie and validate it on callback?
		State: response.State,
	}, nil
}

func (h Handler) AuthCallback(ctx context.Context, request *frontierv1beta1.AuthCallbackRequest) (*frontierv1beta1.AuthCallbackResponse, error) {
	logger := grpczap.Extract(ctx)

	// handle callback
	response, err := h.authnService.FinishFlow(ctx, authenticate.RegistrationFinishRequest{
		Method:      request.GetStrategyName(),
		Code:        request.GetCode(),
		State:       request.GetState(),
		StateConfig: request.GetStateOptions().AsMap(),
	})
	if err != nil {
		if errors.Is(err, authenticate.ErrInvalidMailOTP) || errors.Is(err, authenticate.ErrMissingOIDCCode) || errors.Is(err, authenticate.ErrInvalidOIDCState) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// registration/login complete, build a session
	session, err := h.sessionService.Create(ctx, response.User.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// save in browser cookies
	if err = setCookieHeaders(ctx, session.ID.String()); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// set location header for redirect to finish auth and send client to origin
	if len(response.Flow.FinishURL) > 0 {
		if err = setRedirectHeaders(ctx, response.Flow.FinishURL); err != nil {
			logger.Error(err.Error())
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &frontierv1beta1.AuthCallbackResponse{}, nil
}

func (h Handler) AuthLogout(ctx context.Context, request *frontierv1beta1.AuthLogoutRequest) (*frontierv1beta1.AuthLogoutResponse, error) {
	logger := grpczap.Extract(ctx)

	// delete user session if exists
	sessionID, err := h.getLoggedInSessionID(ctx)
	if err == nil {
		if err = h.sessionService.Delete(ctx, sessionID); err != nil {
			logger.Error(err.Error())
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	// delete from browser cookies
	if err := deleteCookieHeaders(ctx); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &frontierv1beta1.AuthLogoutResponse{}, nil
}

func (h Handler) ListAuthStrategies(ctx context.Context, request *frontierv1beta1.ListAuthStrategiesRequest) (*frontierv1beta1.ListAuthStrategiesResponse, error) {
	var pbstrategy []*frontierv1beta1.AuthStrategy
	for _, strategy := range h.authnService.SupportedStrategies() {
		pbstrategy = append(pbstrategy, &frontierv1beta1.AuthStrategy{
			Name:   strategy,
			Params: nil,
		})
	}
	return &frontierv1beta1.ListAuthStrategiesResponse{Strategies: pbstrategy}, nil
}

func (h Handler) GetJWKs(ctx context.Context, request *frontierv1beta1.GetJWKsRequest) (*frontierv1beta1.GetJWKsResponse, error) {
	keySet := h.authnService.JWKs(ctx)
	jwks, err := toJSONWebKey(keySet)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.GetJWKsResponse{
		Keys: jwks.Keys,
	}, nil
}

func (h Handler) AuthToken(ctx context.Context, request *frontierv1beta1.AuthTokenRequest) (*frontierv1beta1.AuthTokenResponse, error) {
	logger := grpczap.Extract(ctx)
	existingMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		existingMD = metadata.New(map[string]string{})
	}

	// if values are passed in body instead of headers, populate them in context
	switch request.GetGrantType() {
	case "client_credentials":
		if request.GetClientId() != "" && request.GetClientSecret() != "" {
			secretVal := base64.StdEncoding.EncodeToString([]byte(request.GetClientId() + ":" + request.GetClientSecret()))
			existingMD.Set(consts.UserSecretGatewayKey, secretVal)
		}
	case "urn:ietf:params:oauth:grant-type:jwt-bearer":
		if request.GetAssertion() != "" {
			existingMD.Set(consts.UserTokenGatewayKey, request.GetAssertion())
		}
	}
	ctx = metadata.NewIncomingContext(ctx, existingMD)

	// only get principal from service user assertions
	principal, err := h.GetLoggedInPrincipal(ctx,
		authenticate.SessionClientAssertion,
		authenticate.ClientCredentialsClientAssertion,
		authenticate.JWTGrantClientAssertion)
	if err != nil {
		return nil, err
	}

	token, err := h.getAccessToken(ctx, principal)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := setUserContextTokenInHeaders(ctx, string(token)); err != nil {
		logger.Error(fmt.Errorf("error setting token in context: %w", err).Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &frontierv1beta1.AuthTokenResponse{
		AccessToken: string(token),
		TokenType:   "Bearer",
	}, nil
}

func (h Handler) getLoggedInSessionID(ctx context.Context) (uuid.UUID, error) {
	session, err := h.sessionService.ExtractFromContext(ctx)
	if err == nil && session.IsValid(time.Now().UTC()) {
		return session.ID, nil
	}
	return uuid.Nil, err
}

func (h Handler) GetLoggedInPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error) {
	principal, err := h.authnService.GetPrincipal(ctx, via...)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotExist), errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidEmail):
			return principal, grpcUserNotFoundError
		case errors.Is(err, errors.ErrUnauthenticated):
			return principal, grpcUnauthenticated
		default:
			return principal, err
		}
	}
	return principal, nil
}

// getAccessToken generates a jwt access token with user/org details
func (h Handler) getAccessToken(ctx context.Context, principal authenticate.Principal) ([]byte, error) {
	logger := grpczap.Extract(ctx)
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
	if principal.Type == schema.UserPrincipal && h.authConfig.Token.Claims.AddUserEmailClaim {
		customClaims["email"] = principal.User.Email
	}

	// find selected project id
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if projectKey := md.Get(consts.ProjectRequestKey); len(projectKey) > 0 && projectKey[0] != "" {
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
	}

	// build jwt for user context
	return h.authnService.BuildToken(ctx, principal.ID, customClaims)
}

func setRedirectHeaders(ctx context.Context, url string) error {
	return grpc.SetHeader(ctx, metadata.Pairs(consts.LocationGatewayKey, url))
}

func setCookieHeaders(ctx context.Context, cookie string) error {
	return grpc.SetHeader(ctx, metadata.Pairs(consts.SessionIDGatewayKey, cookie))
}

func deleteCookieHeaders(ctx context.Context) error {
	return grpc.SetHeader(ctx, metadata.Pairs(consts.SessionDeleteGatewayKey, "true"))
}

// setUserContextTokenInHeaders sends a jwt token in headers
func setUserContextTokenInHeaders(ctx context.Context, userToken string) error {
	return grpc.SetHeader(ctx, metadata.Pairs(consts.UserTokenGatewayKey, userToken))
}

type JsonWebKeySet struct {
	Keys []*frontierv1beta1.JSONWebKey `json:"keys"`
}

func toJSONWebKey(keySet jwk.Set) (*JsonWebKeySet, error) {
	jwks := &JsonWebKeySet{
		Keys: []*frontierv1beta1.JSONWebKey{},
	}
	keySetJson, err := json.Marshal(keySet)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(keySetJson, &jwks); err != nil {
		return nil, err
	}
	return jwks, nil
}
