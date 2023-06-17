package v1beta1

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/raystack/shield/core/authenticate/token"

	"github.com/raystack/shield/pkg/server/consts"

	"github.com/google/uuid"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/shield/core/authenticate"
	shieldsession "github.com/raystack/shield/core/authenticate/session"
	"github.com/raystack/shield/core/user"
	"github.com/raystack/shield/pkg/errors"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

//go:generate mockery --name=AuthnService -r --case underscore --with-expecter --structname AuthnService --filename authn_service.go --output=./mocks
type AuthnService interface {
	StartFlow(ctx context.Context, request authenticate.RegistrationStartRequest) (*authenticate.RegistrationStartResponse, error)
	FinishFlow(ctx context.Context, request authenticate.RegistrationFinishRequest) (*authenticate.RegistrationFinishResponse, error)
	BuildToken(ctx context.Context, principalID string, metadata map[string]string) ([]byte, error)
	JWKs(ctx context.Context) jwk.Set
	GetPrincipal(ctx context.Context) (authenticate.Principal, error)
	SupportedStrategies() []string
	InitFlows(ctx context.Context) error
	Close()
}

//go:generate mockery --name=SessionService -r --case underscore --with-expecter --structname SessionService --filename session_service.go --output=./mocks
type SessionService interface {
	ExtractFromContext(ctx context.Context) (*shieldsession.Session, error)
	Create(ctx context.Context, userID string) (*shieldsession.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
	Refresh(ctx context.Context, sessionID uuid.UUID) error
	InitSessions(ctx context.Context) error
	Close()
}

func (h Handler) Authenticate(ctx context.Context, request *shieldv1beta1.AuthenticateRequest) (*shieldv1beta1.AuthenticateResponse, error) {
	logger := grpczap.Extract(ctx)

	// check if user is already logged in
	session, err := h.sessionService.ExtractFromContext(ctx)
	if err == nil && session.IsValid(time.Now().UTC()) {
		// already logged in, set location header for return to?
		if len(request.GetReturnTo()) > 0 {
			// TODO(kushsharma): only redirect to white listed domains from config
			// to avoid https://cheatsheetseries.owasp.org/cheatsheets/Unvalidated_Redirects_and_Forwards_Cheat_Sheet.html
			if err = setRedirectHeaders(ctx, request.GetReturnTo()); err != nil {
				logger.Error(err.Error())
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
		return &shieldv1beta1.AuthenticateResponse{}, nil
	} else if err != nil && !errors.Is(err, shieldsession.ErrNoSession) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// not logged in, try registration
	response, err := h.authnService.StartFlow(ctx, authenticate.RegistrationStartRequest{
		ReturnTo: request.ReturnTo,
		Method:   request.GetStrategyName(),
		Email:    request.GetEmail(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	// set location header for redirect to start auth?
	if request.GetRedirect() && len(response.Flow.StartURL) > 0 {
		if err = setRedirectHeaders(ctx, response.Flow.StartURL); err != nil {
			logger.Error(err.Error())
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &shieldv1beta1.AuthenticateResponse{
		Endpoint: response.Flow.StartURL,
		State:    response.State,
	}, nil
}

func (h Handler) AuthCallback(ctx context.Context, request *shieldv1beta1.AuthCallbackRequest) (*shieldv1beta1.AuthCallbackResponse, error) {
	logger := grpczap.Extract(ctx)

	// handle callback
	response, err := h.authnService.FinishFlow(ctx, authenticate.RegistrationFinishRequest{
		Method: request.GetStrategyName(),
		Code:   request.GetCode(),
		State:  request.GetState(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	// registration/login complete, build a session
	session, err := h.sessionService.Create(ctx, response.User.ID)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	// save in browser cookies
	if err = setCookieHeaders(ctx, session.ID.String()); err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = h.setUserContextTokenInHeaders(ctx, response.User.ID); err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	// set location header for redirect to finish auth and send client to origin
	if len(response.Flow.FinishURL) > 0 {
		if err = setRedirectHeaders(ctx, response.Flow.FinishURL); err != nil {
			logger.Error(err.Error())
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &shieldv1beta1.AuthCallbackResponse{}, nil
}

func (h Handler) AuthLogout(ctx context.Context, request *shieldv1beta1.AuthLogoutRequest) (*shieldv1beta1.AuthLogoutResponse, error) {
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
	return &shieldv1beta1.AuthLogoutResponse{}, nil
}

func (h Handler) ListAuthStrategies(ctx context.Context, request *shieldv1beta1.ListAuthStrategiesRequest) (*shieldv1beta1.ListAuthStrategiesResponse, error) {
	var pbstrategy []*shieldv1beta1.AuthStrategy
	for _, strategy := range h.authnService.SupportedStrategies() {
		pbstrategy = append(pbstrategy, &shieldv1beta1.AuthStrategy{
			Name:   strategy,
			Params: nil,
		})
	}
	return &shieldv1beta1.ListAuthStrategiesResponse{Strategies: pbstrategy}, nil
}

func (h Handler) GetJWKs(ctx context.Context, request *shieldv1beta1.GetJWKsRequest) (*shieldv1beta1.GetJWKsResponse, error) {
	logger := grpczap.Extract(ctx)
	keySet := h.authnService.JWKs(ctx)
	jwks, err := toJSONWebKey(keySet)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &shieldv1beta1.GetJWKsResponse{
		Keys: jwks.Keys,
	}, nil
}

func (h Handler) getLoggedInSessionID(ctx context.Context) (uuid.UUID, error) {
	session, err := h.sessionService.ExtractFromContext(ctx)
	if err == nil && session.IsValid(time.Now().UTC()) {
		return session.ID, nil
	}
	return uuid.Nil, err
}

func (h Handler) GetLoggedInPrincipal(ctx context.Context) (authenticate.Principal, error) {
	logger := grpczap.Extract(ctx)
	principal, err := h.authnService.GetPrincipal(ctx)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrNotExist), errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidEmail):
			return principal, grpcUserNotFoundError
		case errors.Is(err, errors.ErrUnauthenticated):
			return principal, grpcUnauthenticated
		default:
			return principal, grpcInternalServerError
		}
	}
	return principal, nil
}

// setUserContextTokenInHeaders sends a jwt token with user/org details
func (h Handler) setUserContextTokenInHeaders(ctx context.Context, principalID string) error {
	// get orgs a user belongs to
	orgs, err := h.orgService.ListByUser(ctx, principalID)
	if err != nil {
		return err
	}

	var orgNames []string
	for _, o := range orgs {
		orgNames = append(orgNames, o.Name)
	}

	// build jwt for user context
	userToken, err := h.authnService.BuildToken(ctx, principalID, map[string]string{
		"orgs": strings.Join(orgNames, ","),
	})
	if errors.Is(err, token.ErrMissingRSADisableToken) {
		// should not fail if user has not configured jwks
		return nil
	}
	if err != nil {
		return err
	}

	// pass as response headers
	if err = grpc.SetHeader(ctx, metadata.Pairs(consts.UserTokenGatewayKey, string(userToken))); err != nil {
		return fmt.Errorf("failed to set header: %w", err)
	}
	return nil
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

type JsonWebKeySet struct {
	Keys []*shieldv1beta1.JSONWebKey `json:"keys"`
}

func toJSONWebKey(keySet jwk.Set) (*JsonWebKeySet, error) {
	jwks := &JsonWebKeySet{
		Keys: []*shieldv1beta1.JSONWebKey{},
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
