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
	patErrors "github.com/raystack/frontier/core/userpat/errors"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/server/consts"
	sessionutils "github.com/raystack/frontier/pkg/session"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

// authFlowRejection carries a gate or consent error out to a client. Both auth
// RPCs answer with a connect code, because both are called from application
// JavaScript — AuthCallback from whatever page the callback URL points at — so
// neither is a redirect the browser follows on its own.
type authFlowRejection struct {
	code connect.Code
	// err is the bare sentinel; the wrapped one names the documents and belongs
	// in a log, not a response.
	err error
}

// lookupAuthFlowRejection maps the four errors both auth RPCs have to make
// legible. FailedPrecondition for consent, because it is what separates a
// consent rejection from a bad code or an expired flow, which are
// InvalidArgument. ErrInvalidMethod shares that code rather than joining the
// InvalidArgument set: the strategy name is a valid argument that would have
// worked against an account holding that credential, so what is wrong is this
// account's state and not the request. Those two are the one pair a code does
// not separate, and the message does, because it is the bare sentinel and
// nothing else. Anything missing here surfaces as a 500.
func lookupAuthFlowRejection(err error) (authFlowRejection, bool) {
	switch {
	case errors.Is(err, authenticate.ErrLoginUserNotFound):
		return authFlowRejection{
			code: connect.CodeNotFound,
			err:  authenticate.ErrLoginUserNotFound,
		}, true
	case errors.Is(err, authenticate.ErrSignupUserExists):
		return authFlowRejection{
			code: connect.CodeAlreadyExists,
			err:  authenticate.ErrSignupUserExists,
		}, true
	case errors.Is(err, authenticate.ErrConsentRequired):
		return authFlowRejection{
			code: connect.CodeFailedPrecondition,
			err:  authenticate.ErrConsentRequired,
		}, true
	case errors.Is(err, authenticate.ErrInvalidMethod):
		return authFlowRejection{
			code: connect.CodeFailedPrecondition,
			err:  authenticate.ErrInvalidMethod,
		}, true
	}
	return authFlowRejection{}, false
}

// toFlowIntent maps the request enum onto the flow intent. An unknown value
// reads as unspecified, which is the behaviour clients had before intents.
func toFlowIntent(intent frontierv1beta1.FlowIntent) authenticate.FlowIntent {
	switch intent {
	case frontierv1beta1.FlowIntent_FLOW_INTENT_LOGIN:
		return authenticate.FlowIntentLogin
	case frontierv1beta1.FlowIntent_FLOW_INTENT_SIGNUP:
		return authenticate.FlowIntentSignup
	default:
		return authenticate.FlowIntentUnspecified
	}
}

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
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Authenticate: %w", err))
	}

	if (request.Msg.GetStrategyName() == authenticate.MailLinkAuthMethod.String() || request.Msg.GetStrategyName() == authenticate.MailOTPAuthMethod.String()) && !isValidEmail(request.Msg.GetEmail()) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidEmail)
	}

	intent := toFlowIntent(request.Msg.GetFlowIntent())
	acceptedDocumentIDs := request.Msg.GetAcceptedDocumentIds()
	if intent == authenticate.FlowIntentLogin && len(acceptedDocumentIDs) > 0 && h.consentEnabled() {
		// a login writes no record, so accepting these silently would leave the
		// client believing it recorded a consent that does not exist. Disabled,
		// no record is written for any intent, so the ids are ignored rather
		// than refused, as they are everywhere else
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrConsentOnLoginIntent)
	}

	// Authenticate is skip-listed, so nothing put session metadata on the context
	// and the handler extracts it itself. Only the IP is passed on.
	sessionMetadata := sessionutils.ExtractSessionMetadata(ctx, request, h.authConfig.Session.Headers)

	// not logged in, try registration
	response, err := h.authnService.StartFlow(ctx, authenticate.RegistrationStartRequest{
		Method:              request.Msg.GetStrategyName(),
		ReturnToURL:         returnToURL,
		CallbackUrl:         callbackURL,
		Email:               request.Msg.GetEmail(),
		Intent:              intent,
		AcceptedDocumentIDs: acceptedDocumentIDs,
		IPAddress:           sessionMetadata.IpAddress,
	})
	if err != nil {
		// their own codes rather than a 500; the wrapped error stays in the log
		if rejection, ok := lookupAuthFlowRejection(err); ok {
			errorLogger.LogServiceError(ctx, request, "Authenticate.StartFlow", err,
				"strategy", request.Msg.GetStrategyName(),
				"intent", intent.String())
			return nil, connect.NewError(rejection.code, rejection.err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Authenticate: strategy=%s email=%s: %w", request.Msg.GetStrategyName(), request.Msg.GetEmail(), err))
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
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Authenticate: strategy=%s: %w", authenticate.PassKeyAuthMethod.String(), err))
		}
		typeValue, ok := response.Flow.Metadata["passkey_type"].(string)
		if !ok {
			err = fmt.Errorf("passkey_type metadata is not a string")
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Authenticate: strategy=%s: %w", authenticate.PassKeyAuthMethod.String(), err))
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
		// These join the errors handled here rather than falling through to
		// Internal, keeping their own codes rather than this list's
		// InvalidArgument. The rejection is the answer and not a redirect: the
		// callback URL points at a page the application hosts, and that page is
		// what calls this RPC, so it decides where the user goes next.
		if rejection, ok := lookupAuthFlowRejection(err); ok {
			errorLogger.LogServiceError(ctx, request, "AuthCallback.FinishFlow", err,
				"strategy", request.Msg.GetStrategyName(),
				"state", request.Msg.GetState())
			return nil, connect.NewError(rejection.code, rejection.err)
		}

		// ErrUnsupportedMethod here means the strategy and state the client sent match
		// no known method (e.g. a malformed, non-base64 state). That is bad client
		// input, same class as an empty or invalid state, so return a 4xx not a 500.
		if errors.Is(err, authenticate.ErrInvalidMailOTP) || errors.Is(err, authenticate.ErrMissingOIDCCode) || errors.Is(err, authenticate.ErrInvalidOIDCState) || errors.Is(err, authenticate.ErrFlowInvalid) || errors.Is(err, authenticate.ErrOIDCTokenExchange) || errors.Is(err, authenticate.ErrUnsupportedMethod) {
			errorLogger.LogServiceError(ctx, request, "AuthCallback.FinishFlow", err,
				"strategy", request.Msg.GetStrategyName(),
				"state", request.Msg.GetState())
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("AuthCallback: strategy=%s state=%s: %w", request.Msg.GetStrategyName(), request.Msg.GetState(), err))
	}

	// Extract session metadata from request headers
	sessionMetadata := sessionutils.ExtractSessionMetadata(ctx, request, h.authConfig.Session.Headers)

	// registration/login complete, build a session
	session, err := h.sessionService.Create(ctx, response.User.ID, sessionMetadata)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("AuthCallback: user_id=%s: %w", response.User.ID, err))
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

	// restrict to credential types allowed to exchange for a token
	principal, err := h.GetLoggedInPrincipal(ctx,
		authenticate.SessionClientAssertion,
		authenticate.ClientCredentialsClientAssertion,
		authenticate.JWTGrantClientAssertion,
		authenticate.PATClientAssertion)
	if err != nil {
		return nil, err
	}

	if principal.Type == schema.ServiceUserPrincipal {
		orgId := principal.ServiceUser.OrgID
		_, err := h.orgService.Get(ctx, orgId)
		if err != nil {
			errorLogger.LogServiceError(ctx, request, "AuthToken.orgService.Get", err,
				"org_id", orgId,
				"service_user_id", principal.ServiceUser.ID)
			if errors.Is(err, organization.ErrDisabled) {
				return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
			}
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("AuthToken.orgService.Get: org_id=%s service_user_id=%s: %w", orgId, principal.ServiceUser.ID, err))
		}
	}

	token, err := h.getAccessToken(ctx, principal, request.Header().Values(consts.ProjectRequestKey), request)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("AuthToken: principal_id=%s principal_type=%s: %w", principal.ID, principal.Type, err))
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
		orgs, err := h.orgService.List(ctx, organization.Filter{
			Principal: &principal,
			State:     organization.Enabled,
		})
		if err != nil {
			return nil, err
		}
		orgIDs := make([]string, 0, len(orgs))
		for _, o := range orgs {
			orgIDs = append(orgIDs, o.ID)
		}
		customClaims[token.OrgIDsClaimKey] = strings.Join(orgIDs, ",")
	}

	// add session ID as claims for upstream
	if h.authConfig.Token.Claims.AddSessionIDClaim && principal.Type == schema.UserPrincipal {
		if sessionID, err := h.getLoggedInSessionID(ctx); err == nil {
			customClaims[token.SessionIDClaimKey] = sessionID.String()
		} else {
			errorLogger.LogUnexpectedError(ctx, request, "getAccessToken", err, "principal", principal.ID)
		}
	}

	// find selected project id
	if len(projectKey) > 0 && projectKey[0] != "" {
		// check if project exists and user has access to it
		proj, err := h.projectService.Get(ctx, projectKey[0])
		if err != nil {
			errorLogger.LogUnexpectedError(ctx, request, "getAccessToken", err, "project", projectKey[0])
		} else {
			if err := h.IsAuthorized(ctx, relation.Object{
				Namespace: schema.ProjectNamespace,
				ID:        proj.ID,
			}, schema.GetPermission, request); err == nil {
				customClaims["project_id"] = proj.ID
			} else {
				errorLogger.LogUnexpectedError(ctx, request, "getAccessToken", err, "project", proj.ID, "principal", principal.ID)
			}
		}
	}

	// build jwt for user context
	return h.authnService.BuildToken(ctx, principal, customClaims)
}

func (h *ConnectHandler) AuthLogout(ctx context.Context, request *connect.Request[frontierv1beta1.AuthLogoutRequest]) (*connect.Response[frontierv1beta1.AuthLogoutResponse], error) {
	// delete user session if exists
	sessionID, err := h.getLoggedInSessionID(ctx)
	if err == nil {
		if err = h.sessionService.Delete(ctx, sessionID); err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("AuthLogout: session_id=%s: %w", sessionID.String(), err))
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
		case errors.Is(err, errors.ErrForbidden):
			return principal, connect.NewError(connect.CodePermissionDenied, ErrUnauthorized)
		case errors.Is(err, patErrors.ErrMalformedPAT),
			errors.Is(err, patErrors.ErrNotFound),
			errors.Is(err, patErrors.ErrExpired),
			errors.Is(err, patErrors.ErrDisabled):
			return principal, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
		default:
			return principal, connect.NewError(connect.CodeInternal, fmt.Errorf("GetLoggedInPrincipal: %w", err))
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
	keySet := h.authnService.JWKs(ctx)
	jwks, err := toJSONWebKey(keySet)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetJWKs: %w", err))
	}
	return connect.NewResponse(&frontierv1beta1.GetJWKsResponse{
		Keys: jwks.Keys,
	}), nil
}
