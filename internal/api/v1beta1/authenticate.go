package v1beta1

import (
	"context"
	"fmt"
	"strings"
	"time"

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

//go:generate mockery --name=RegistrationService -r --case underscore --with-expecter --structname RegistrationService --filename registration_service.go --output=./mocks
type RegistrationService interface {
	Start(ctx context.Context, request authenticate.RegistrationStartRequest) (*authenticate.RegistrationStartResponse, error)
	Finish(ctx context.Context, request authenticate.RegistrationFinishRequest) (*authenticate.RegistrationFinishResponse, error)
	Token(ctx context.Context, user user.User, metadata map[string]string) ([]byte, error)
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
	response, err := h.registrationService.Start(ctx, authenticate.RegistrationStartRequest{
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
	response, err := h.registrationService.Finish(ctx, authenticate.RegistrationFinishRequest{
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

	if err = h.setUserContextTokenInHeaders(ctx, response.User); err != nil {
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
	for _, strategy := range h.registrationService.SupportedStrategies() {
		pbstrategy = append(pbstrategy, &shieldv1beta1.AuthStrategy{
			Name:   strategy,
			Params: nil,
		})
	}
	return &shieldv1beta1.ListAuthStrategiesResponse{Strategies: pbstrategy}, nil
}

func (h Handler) getLoggedInSessionID(ctx context.Context) (uuid.UUID, error) {
	session, err := h.sessionService.ExtractFromContext(ctx)
	if err == nil && session.IsValid(time.Now().UTC()) {
		return session.ID, nil
	}
	return uuid.Nil, err
}

func (h Handler) GetLoggedInUser(ctx context.Context) (user.User, error) {
	logger := grpczap.Extract(ctx)
	u, err := h.userService.FetchCurrentUser(ctx)
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrNotExist), errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidEmail):
			return u, grpcUserNotFoundError
		case errors.Is(err, errors.ErrUnauthenticated):
			return u, grpcUnauthenticated
		default:
			return u, grpcInternalServerError
		}
	}
	return u, nil
}

// setUserContextTokenInHeaders sends a jwt token with user/org details
func (h Handler) setUserContextTokenInHeaders(ctx context.Context, user user.User) error {
	// get orgs a user belongs to
	orgs, err := h.orgService.ListByUser(ctx, user.ID)
	if err != nil {
		return err
	}

	var orgNames []string
	for _, o := range orgs {
		orgNames = append(orgNames, o.Name)
	}

	// build jwt for user context
	userToken, err := h.registrationService.Token(ctx, user, map[string]string{
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
