package authenticate

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/raystack/frontier/pkg/metadata"

	"github.com/raystack/frontier/core/audit"

	"golang.org/x/exp/slices"

	"github.com/lestrrat-go/jwx/v2/jwt"

	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/internal/metrics"
	"github.com/raystack/frontier/pkg/errors"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/raystack/frontier/core/authenticate/token"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/pkg/mailer"

	"github.com/raystack/salt/log"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate/strategy"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/str"
	"github.com/robfig/cron/v3"
)

const (
	defaultFlowExp = time.Minute * 10
	maxOTPAttempt  = 3
	otpAttemptKey  = "attempt"
)

var (
	refreshTime              = "0 0 * * *" // Once a day at midnight
	ErrStrategyNotApplicable = errors.New("strategy not applicable")
	ErrUnsupportedMethod     = errors.New("unsupported authentication method")
	ErrInvalidMailOTP        = errors.New("invalid mail otp")
	ErrMissingOIDCCode       = errors.New("OIDC code is missing")
	ErrInvalidOIDCState      = errors.New("invalid auth state")
	ErrFlowInvalid           = errors.New("invalid flow or expired")
)

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	Create(context.Context, user.User) (user.User, error)
	Update(ctx context.Context, toUpdate user.User) (user.User, error)
}

type ServiceUserService interface {
	Get(ctx context.Context, id string) (serviceuser.ServiceUser, error)
	GetByJWT(ctx context.Context, token string) (serviceuser.ServiceUser, error)
	GetBySecret(ctx context.Context, clientID, clientSecret string) (serviceuser.ServiceUser, error)
}

type FlowRepository interface {
	Set(ctx context.Context, flow *Flow) error
	Get(ctx context.Context, id uuid.UUID) (*Flow, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpiredFlows(ctx context.Context) error
}

type SessionService interface {
	ExtractFromContext(ctx context.Context) (*frontiersession.Session, error)
}

type TokenService interface {
	GetPublicKeySet() jwk.Set
	Build(subjectID string, metadata map[string]string) ([]byte, error)
	Parse(ctx context.Context, userToken []byte) (string, map[string]any, error)
}

type Service struct {
	log                  log.Logger
	cron                 *cron.Cron
	flowRepo             FlowRepository
	userService          UserService
	config               Config
	mailDialer           mailer.Dialer
	Now                  func() time.Time
	internalTokenService TokenService
	sessionService       SessionService
	serviceUserService   ServiceUserService
	webAuth              *webauthn.WebAuthn
}

func NewService(logger log.Logger, config Config, flowRepo FlowRepository,
	mailDialer mailer.Dialer, tokenService TokenService, sessionService SessionService,
	userService UserService, serviceUserService ServiceUserService, webAuthConfig *webauthn.WebAuthn) *Service {
	r := &Service{
		log: logger,
		cron: cron.New(cron.WithChain(
			cron.SkipIfStillRunning(cron.DefaultLogger),
			cron.Recover(cron.DefaultLogger),
		)),
		flowRepo:    flowRepo,
		userService: userService,
		config:      config,
		mailDialer:  mailDialer,
		Now: func() time.Time {
			return time.Now().UTC()
		},
		internalTokenService: tokenService,
		sessionService:       sessionService,
		serviceUserService:   serviceUserService,
		webAuth:              webAuthConfig,
	}
	return r
}

func (s Service) SupportedStrategies() []string {
	// add here strategies like mail link once implemented
	var strategies []string
	for name := range s.config.OIDCConfig {
		strategies = append(strategies, name)
	}
	if s.mailDialer != nil {
		strategies = append(strategies, MailOTPAuthMethod.String(), MailLinkAuthMethod.String())
	}
	if s.webAuth != nil {
		strategies = append(strategies, PassKeyAuthMethod.String())
	}
	return strategies
}

// SanitizeReturnToURL allows only redirect to white listed domains from config
// to avoid https://cheatsheetseries.owasp.org/cheatsheets/Unvalidated_Redirects_and_Forwards_Cheat_Sheet.html
func (s Service) SanitizeReturnToURL(url string) string {
	if len(url) == 0 {
		return ""
	}
	if len(s.config.AuthorizedRedirectURLs) == 0 {
		return ""
	}
	if slices.Contains[[]string](s.config.AuthorizedRedirectURLs, url) {
		return url
	}
	return ""
}

// SanitizeCallbackURL allows only callback host to white listed domains from config
func (s Service) SanitizeCallbackURL(url string) string {
	if len(s.config.CallbackURLs) == 0 {
		return ""
	}
	if len(url) == 0 {
		return s.config.CallbackURLs[0]
	}
	if slices.Contains[[]string](s.config.CallbackURLs, url) {
		return url
	}
	return ""
}

func (s Service) StartFlow(ctx context.Context, request RegistrationStartRequest) (*RegistrationStartResponse, error) {
	if !utils.Contains(s.SupportedStrategies(), request.Method) {
		return nil, ErrUnsupportedMethod
	}
	flow := &Flow{
		ID:        uuid.New(),
		Method:    request.Method,
		FinishURL: request.ReturnToURL,
		CreatedAt: s.Now(),
		ExpiresAt: s.Now().Add(defaultFlowExp),
		Email:     request.Email,
		Metadata: metadata.Metadata{
			"callback_url": request.CallbackUrl,
		},
	}

	if request.Method == PassKeyAuthMethod.String() {
		needRegistration := false
		loggedInUser, err := s.userService.GetByID(ctx, request.Email)
		if err != nil {
			needRegistration = true
		} else {
			storedPasskey, passKeyExists := loggedInUser.Metadata["passkey_credentials"]
			if !passKeyExists {
				needRegistration = true
			}
			if _, ok := storedPasskey.(string); !ok {
				needRegistration = true
			}
		}

		if needRegistration {
			response, err := s.startPassKeyRegisterMethod(ctx, flow)
			if err != nil {
				return nil, err
			}
			return response, nil
		} else {
			response, err := s.startPassKeyLoginMethod(ctx, loggedInUser, flow)
			if err != nil {
				return nil, err
			}
			return response, nil
		}
	}

	if request.Method == MailOTPAuthMethod.String() {
		mailLinkStrat := strategy.NewMailOTP(s.mailDialer, s.config.MailOTP.Subject, s.config.MailOTP.Body)
		nonce, err := mailLinkStrat.SendMail(request.Email, s.config.TestUsers)
		if err != nil {
			return nil, err
		}

		flow.Nonce = nonce
		if s.config.MailOTP.Validity != 0 {
			flow.ExpiresAt = flow.CreatedAt.Add(s.config.MailOTP.Validity)
		}
		flow.Email = strings.ToLower(request.Email)
		if err = s.flowRepo.Set(ctx, flow); err != nil {
			return nil, err
		}
		return &RegistrationStartResponse{
			Flow:  flow,
			State: flow.ID.String(),
		}, nil
	}

	if len(request.CallbackUrl) == 0 {
		return nil, fmt.Errorf("callback url not configured")
	}

	if request.Method == MailLinkAuthMethod.String() {
		mailLinkStrat := strategy.NewMailLink(s.mailDialer, request.CallbackUrl, s.config.MailLink.Subject, s.config.MailLink.Body)
		nonce, err := mailLinkStrat.SendMail(flow.ID.String(), request.Email, s.config.TestUsers)
		if err != nil {
			return nil, err
		}

		flow.Nonce = nonce
		if s.config.MailLink.Validity != 0 {
			flow.ExpiresAt = flow.CreatedAt.Add(s.config.MailLink.Validity)
		}
		flow.Email = strings.ToLower(request.Email)
		if err = s.flowRepo.Set(ctx, flow); err != nil {
			return nil, err
		}
		return &RegistrationStartResponse{
			Flow: flow,
		}, nil
	}

	// check for oidc flow
	if oidcConfig, ok := s.config.OIDCConfig[request.Method]; ok {
		idp, err := strategy.NewRelyingPartyOIDC(
			oidcConfig.ClientID,
			oidcConfig.ClientSecret,
			request.CallbackUrl).
			Init(ctx, oidcConfig.IssuerUrl)
		if err != nil {
			return nil, err
		}

		oidcState, err := strategy.EmbedFlowInOIDCState(flow.ID.String())
		if err != nil {
			return nil, err
		}
		endpoint, nonce, err := idp.AuthURL(oidcState)
		if err != nil {
			return nil, err
		}

		flow.StartURL = endpoint
		flow.Nonce = nonce
		if oidcConfig.Validity != 0 {
			flow.ExpiresAt = flow.CreatedAt.Add(oidcConfig.Validity)
		}
		if err = s.flowRepo.Set(ctx, flow); err != nil {
			return nil, err
		}
		return &RegistrationStartResponse{
			Flow: flow,
		}, nil
	}

	return nil, ErrUnsupportedMethod
}

func (s Service) FinishFlow(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	if request.Method == MailOTPAuthMethod.String() || request.Method == MailLinkAuthMethod.String() {
		response, err := s.applyMailOTP(ctx, request)
		if err != nil && !errors.Is(err, ErrStrategyNotApplicable) {
			return nil, err
		}
		return response, nil
	}
	if request.Method == PassKeyAuthMethod.String() {
		response, err := s.applyPasskey(ctx, request)
		if err != nil && !errors.Is(err, ErrStrategyNotApplicable) {
			return nil, err
		}
		return response, nil
	}

	// check for oidc method config
	{
		response, err := s.applyOIDC(ctx, request)
		if err == nil {
			return response, nil
		}
		if err != nil && !errors.Is(err, ErrStrategyNotApplicable) {
			return nil, err
		}
	}
	return nil, ErrUnsupportedMethod
}

// applyMailOTP actions when user submitted otp from the email
// user can be considered as verified if code is valid
// create a new user if required
func (s Service) applyMailOTP(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	if len(request.Code) == 0 {
		return nil, ErrStrategyNotApplicable
	}
	flowID, err := uuid.Parse(request.State)
	if err != nil {
		return nil, ErrStrategyNotApplicable
	}
	flow, err := s.flowRepo.Get(ctx, flowID)
	if err != nil {
		return nil, fmt.Errorf("invalid state for mail otp: %w", err)
	}
	if !flow.IsValid(s.Now()) {
		return nil, ErrFlowInvalid
	}

	if subtle.ConstantTimeCompare([]byte(flow.Nonce), []byte(request.Code)) == 0 {
		// avoid brute forcing otp
		attemptInt := 0
		if attempts, ok := flow.Metadata[otpAttemptKey]; ok {
			attemptInt, _ = attempts.(int)
		}
		if attemptInt < maxOTPAttempt {
			flow.Metadata[otpAttemptKey] = attemptInt + 1
			if err = s.flowRepo.Set(ctx, flow); err != nil {
				return nil, fmt.Errorf("failed to process flow code missmatch")
			}
		} else {
			if err = s.consumeFlow(ctx, flowID); err != nil {
				return nil, fmt.Errorf("failed to process flow code missmatch")
			}
		}
		return nil, ErrInvalidMailOTP
	}

	// consume this flow
	if err = s.consumeFlow(ctx, flow.ID); err != nil {
		return nil, fmt.Errorf("failed to successfully register via otp: %w", err)
	}

	newUser, err := s.getOrCreateUser(ctx, flow.Email, "")
	if err != nil {
		return nil, err
	}
	return &RegistrationFinishResponse{
		User: newUser,
		Flow: flow,
	}, nil
}

func (s Service) startPassKeyRegisterMethod(ctx context.Context, flow *Flow) (*RegistrationStartResponse, error) {
	newPassKeyUser := strategy.NewPassKeyUser(flow.Email)
	options, session, err := s.webAuth.BeginRegistration(newPassKeyUser)
	if err != nil {
		return nil, err
	}

	// webauthn library expects base64 encoded challenge when verifying the session
	session.Challenge = base64.RawURLEncoding.EncodeToString([]byte(session.Challenge))
	sessionInBytes, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	flow.Metadata["passkey_session"] = sessionInBytes
	flow.Metadata["passkey_type"] = strategy.PasskeyRegisterType
	if err = s.flowRepo.Set(ctx, flow); err != nil {
		return nil, err
	}
	optionsBytes, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	return &RegistrationStartResponse{
		Flow:  flow,
		State: flow.ID.String(),
		StateConfig: map[string]any{
			"options": optionsBytes,
		},
	}, nil
}

func (s Service) startPassKeyLoginMethod(ctx context.Context, loggedInUser user.User, flow *Flow) (*RegistrationStartResponse, error) {
	decodedCredBytes, err := base64.StdEncoding.DecodeString(loggedInUser.Metadata["passkey_credentials"].(string))
	if err != nil {
		return nil, err
	}

	var webAuthCredentialData []webauthn.Credential
	err = json.Unmarshal(decodedCredBytes, &webAuthCredentialData)
	if err != nil {
		return nil, err
	}
	newPassKeyUser := strategy.NewPasskeyUserWithCredentials(flow.Email, webAuthCredentialData)
	options, session, err := s.webAuth.BeginLogin(newPassKeyUser)
	if err != nil {
		return nil, err
	}

	// webauthn library expects base64 encoded challenge when verifying the session
	session.Challenge = base64.RawURLEncoding.EncodeToString([]byte(session.Challenge))
	sessionInBytes, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}
	flow.Metadata["passkey_session"] = sessionInBytes
	flow.Metadata["passkey_type"] = strategy.PasskeyLoginType
	if err = s.flowRepo.Set(ctx, flow); err != nil {
		return nil, err
	}
	optionsBytes, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	return &RegistrationStartResponse{
		Flow:  flow,
		State: flow.ID.String(),
		StateConfig: map[string]any{
			"options": optionsBytes,
		},
	}, nil
}

func (s Service) finishPassKeyRegisterMethod(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	passkeyOptions, ok := request.StateConfig["options"].(string)
	if !ok {
		return nil, errors.New("invalid auth state")
	}
	requestReader := bytes.NewReader([]byte(passkeyOptions))
	credentialCreationResponse, err := protocol.ParseCredentialCreationResponseBody(requestReader)
	if err != nil {
		return nil, err
	}

	flowIdString := request.State
	flowId, err := uuid.Parse(flowIdString)
	if err != nil {
		return nil, err
	}
	flow, err := s.flowRepo.Get(ctx, flowId)
	if err != nil {
		return nil, err
	}

	userFinishRegister := strategy.NewPassKeyUser(flow.Email)
	encodedPasskeySession := flow.Metadata["passkey_session"]
	var webAuthSessionData webauthn.SessionData
	sessionBytes, err := base64.StdEncoding.DecodeString(encodedPasskeySession.(string))
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(sessionBytes, &webAuthSessionData)
	if err != nil {
		return nil, err
	}
	credential, err := s.webAuth.CreateCredential(userFinishRegister, webAuthSessionData, credentialCreationResponse)
	if err != nil {
		return nil, err
	}
	newUser, err := s.getOrCreateUser(ctx, flow.Email, "")
	if err != nil {
		return nil, err
	}
	if newUser.Metadata == nil {
		newUser.Metadata = metadata.Metadata{}
	}

	webAuthCredentialData := []webauthn.Credential{*credential}
	credBytes, err := json.Marshal(webAuthCredentialData)
	if err != nil {
		return nil, err
	}
	credBase64 := base64.StdEncoding.EncodeToString(credBytes)
	newUser.Metadata["passkey_credentials"] = credBase64
	newUpdatedUser, err := s.userService.Update(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return &RegistrationFinishResponse{
		User: newUpdatedUser,
		Flow: flow,
	}, nil
}

func (s Service) finishPassKeyLoginMethod(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	passkeyOptions, ok := request.StateConfig["options"].(string)
	if !ok {
		return nil, errors.New("invalid auth state")
	}
	requestReader := bytes.NewReader([]byte(passkeyOptions))
	response, err := protocol.ParseCredentialRequestResponseBody(requestReader)
	if err != nil {
		return nil, err
	}

	flowIdString := request.State
	flowId, err := uuid.Parse(flowIdString)
	if err != nil {
		return nil, err
	}
	flow, err := s.flowRepo.Get(ctx, flowId)
	if err != nil {
		return nil, err
	}
	userFinishRegister := strategy.NewPassKeyUser(flow.Email)
	sessionInterface, ok := flow.Metadata["passkey_session"]
	if !ok {
		return nil, err
	}
	var webAuthSessionData webauthn.SessionData
	sessionBytes, err := base64.StdEncoding.DecodeString(sessionInterface.(string))
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(sessionBytes, &webAuthSessionData)
	if err != nil {
		return nil, err
	}

	existingUser, err := s.getOrCreateUser(ctx, flow.Email, "")
	if err != nil {
		return nil, err
	}
	userCred, ok := existingUser.Metadata["passkey_credentials"]
	if !ok {
		return nil, err
	}

	decodedCredBytes, err := base64.StdEncoding.DecodeString(userCred.(string))
	if err != nil {
		return nil, err
	}
	var webAuthCredentialData []webauthn.Credential
	err = json.Unmarshal(decodedCredBytes, &webAuthCredentialData)
	if err != nil {
		return nil, err
	}
	userFinishRegister.Credentials = webAuthCredentialData

	_, err = s.webAuth.ValidateLogin(userFinishRegister, webAuthSessionData, response)
	if err != nil {
		return nil, err
	}

	return &RegistrationFinishResponse{
		User: existingUser,
		Flow: flow,
	}, nil
}

func (s Service) applyPasskey(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	flowIdString := request.State
	flowId, err := uuid.Parse(flowIdString)
	if err != nil {
		return nil, err
	}
	flow, err := s.flowRepo.Get(ctx, flowId)
	if err != nil {
		return nil, err
	}
	requestType := flow.Metadata["passkey_type"]
	if request.Method == PassKeyAuthMethod.String() {
		if requestType == strategy.PasskeyRegisterType {
			response, err := s.finishPassKeyRegisterMethod(ctx, request)
			if err != nil {
				return nil, err
			}
			return response, nil
		}

		if requestType == strategy.PasskeyLoginType {
			response, err := s.finishPassKeyLoginMethod(ctx, request)
			if err != nil {
				return nil, err
			}
			return response, nil
		}
	}
	return nil, err
}

func (s Service) applyOIDC(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	// flow id is added in state params
	if len(request.State) == 0 {
		return nil, ErrInvalidOIDCState
	}

	// flow id is added in state params
	if len(request.Code) == 0 {
		return nil, ErrMissingOIDCCode
	}

	// check for oidc flow via fetching oauth state, method parameter will not be set for oauth
	flowIDFromState, err := strategy.ExtractFlowFromOIDCState(request.State)
	if err != nil {
		return nil, ErrStrategyNotApplicable
	}
	flowID, err := uuid.Parse(flowIDFromState)
	if err != nil {
		return nil, ErrStrategyNotApplicable
	}
	// fetch auth flow
	flow, err := s.flowRepo.Get(ctx, flowID)
	if err != nil {
		return nil, err
	}

	// can't find oidc config
	oidcConfig, ok := s.config.OIDCConfig[flow.Method]
	if !ok {
		return nil, ErrStrategyNotApplicable
	}

	if _, ok := flow.Metadata["callback_url"]; !ok {
		return nil, fmt.Errorf("callback url not configured")
	}
	callbackURL := flow.Metadata["callback_url"].(string)

	idp, err := strategy.NewRelyingPartyOIDC(
		oidcConfig.ClientID,
		oidcConfig.ClientSecret,
		callbackURL).
		Init(ctx, oidcConfig.IssuerUrl)
	if err != nil {
		return nil, err
	}
	authToken, err := idp.Token(ctx, request.Code, flow.Nonce)
	if err != nil {
		return nil, err
	}
	oauthProfile, err := idp.GetUser(ctx, authToken)
	if err != nil {
		return nil, err
	}

	// register a new user
	newUser, err := s.getOrCreateUser(ctx, oauthProfile.Email, oauthProfile.Name)
	if err != nil {
		return nil, err
	}

	return &RegistrationFinishResponse{
		User: newUser,
		Flow: flow,
	}, nil
}

// BuildToken creates an access token for the given subjectID
func (s Service) BuildToken(ctx context.Context, principal Principal, metadata map[string]string) ([]byte, error) {
	metadata[token.SubTypeClaimsKey] = principal.Type
	if principal.Type == schema.UserPrincipal && s.config.Token.Claims.AddUserEmailClaim {
		metadata[token.SubEmailClaimsKey] = principal.User.Email
	}
	return s.internalTokenService.Build(principal.ID, metadata)
}

// JWKs returns the public keys to verify the access token
func (s Service) JWKs(ctx context.Context) jwk.Set {
	return s.internalTokenService.GetPublicKeySet()
}

func (s Service) InitFlows(ctx context.Context) error {
	_, err := s.cron.AddFunc(refreshTime, func() {
		if err := s.flowRepo.DeleteExpiredFlows(ctx); err != nil {
			s.log.Warn("failed to delete expired sessions", "err", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to start flows cronjob: %w", err)
	}
	s.cron.Start()
	return nil
}

func (s Service) consumeFlow(ctx context.Context, id uuid.UUID) error {
	return s.flowRepo.Delete(ctx, id)
}

func (s Service) getOrCreateUser(ctx context.Context, email, title string) (user.User, error) {
	// create a new user based on email if it doesn't exist
	existingUser, err := s.userService.GetByID(ctx, email)
	if err == nil {
		// user is already registered

		// TODO(kushsharma): should we update metadata like profile picture from social logins
		// for registered users every time the login?
		return existingUser, nil
	}

	// register a new user
	newUser, err := s.userService.Create(ctx, user.User{
		Title: title,
		Email: email,
		Name:  str.GenerateUserSlug(email),
	})
	if err != nil {
		return user.User{}, err
	}
	_ = audit.GetAuditor(ctx, schema.PlatformOrgID.String()).
		LogWithAttrs(audit.UserCreatedEvent, audit.UserTarget(newUser.ID), map[string]string{
			"email":  newUser.Email,
			"name":   newUser.Name,
			"title":  newUser.Title,
			"avatar": newUser.Avatar,
		})
	return newUser, nil
}

func (s Service) GetPrincipal(ctx context.Context, assertions ...ClientAssertion) (Principal, error) {
	if metrics.ServiceOprLatency != nil {
		promCollect := metrics.ServiceOprLatency("authenticate", "GetPrincipal")
		defer promCollect()
	}

	var currentPrincipal Principal
	if len(assertions) == 0 {
		// check all assertions
		assertions = APIAssertions
	}

	// check if already enriched by auth middleware
	if val, ok := GetPrincipalFromContext(ctx); ok {
		currentPrincipal = *val
		return currentPrincipal, nil
	}

	// extract user from session if present
	if slices.Contains[[]ClientAssertion](assertions, SessionClientAssertion) {
		session, err := s.sessionService.ExtractFromContext(ctx)
		if err == nil && session.IsValid(s.Now()) && utils.IsValidUUID(session.UserID) {
			// userID is a valid uuid
			currentUser, err := s.userService.GetByID(ctx, session.UserID)
			if err != nil {
				return Principal{}, err
			}
			return Principal{
				ID:   currentUser.ID,
				Type: schema.UserPrincipal,
				User: &currentUser,
			}, nil
		}
		if err != nil && !errors.Is(err, frontiersession.ErrNoSession) {
			return Principal{}, err
		}
	}

	// check for token
	userToken, tokenOK := GetTokenFromContext(ctx)
	if tokenOK {
		if slices.Contains[[]ClientAssertion](assertions, AccessTokenClientAssertion) {
			insecureJWT, err := jwt.ParseInsecure([]byte(userToken))
			if err != nil {
				return Principal{}, errors.ErrUnauthenticated
			}
			// check type of jwt
			if genClaim, ok := insecureJWT.Get(token.GeneratedClaimKey); ok {
				// jwt generated by frontier using public key
				claimVal, ok := genClaim.(string)
				if !ok || claimVal != token.GeneratedClaimValue {
					return Principal{}, errors.ErrUnauthenticated
				}

				// extract user from token if present as its created by frontier
				userID, claims, err := s.internalTokenService.Parse(ctx, []byte(userToken))
				if err != nil || !utils.IsValidUUID(userID) {
					s.log.Debug("failed to parse as internal token ", "err", err)
					return Principal{}, errors.ErrUnauthenticated
				}

				// userID is a valid uuid
				if claims[token.SubTypeClaimsKey] == schema.ServiceUserPrincipal {
					currentUser, err := s.serviceUserService.Get(ctx, userID)
					if err != nil {
						return Principal{}, err
					}
					return Principal{
						ID:          currentUser.ID,
						Type:        schema.ServiceUserPrincipal,
						ServiceUser: &currentUser,
					}, nil
				}

				currentUser, err := s.userService.GetByID(ctx, userID)
				if err != nil {
					return Principal{}, err
				}
				return Principal{
					ID:   currentUser.ID,
					Type: schema.UserPrincipal,
					User: &currentUser,
				}, nil
			}
		}

		// extract user from token if it's a service user
		if slices.Contains[[]ClientAssertion](assertions, JWTGrantClientAssertion) {
			serviceUser, err := s.serviceUserService.GetByJWT(ctx, userToken)
			if err == nil {
				return Principal{
					ID:          serviceUser.ID,
					Type:        schema.ServiceUserPrincipal,
					ServiceUser: &serviceUser,
				}, nil
			}
			if err != nil {
				s.log.Debug("failed to parse as user token ", "err", err)
				return Principal{}, errors.ErrUnauthenticated
			}
		}
	}

	// check for client secret
	if slices.Contains[[]ClientAssertion](assertions, ClientCredentialsClientAssertion) ||
		slices.Contains[[]ClientAssertion](assertions, OpaqueTokenClientAssertion) {
		userSecretRaw, secretOK := GetSecretFromContext(ctx)
		if secretOK {
			// verify client secret
			userSecret, err := base64.StdEncoding.DecodeString(userSecretRaw)
			if err != nil {
				return Principal{}, errors.ErrUnauthenticated
			}
			userSecretParts := strings.Split(string(userSecret), ":")
			if len(userSecretParts) != 2 {
				return Principal{}, errors.ErrUnauthenticated
			}
			clientID, clientSecret := userSecretParts[0], userSecretParts[1]

			// extract user from secret if it's a service user
			serviceUser, err := s.serviceUserService.GetBySecret(ctx, clientID, clientSecret)
			if err == nil {
				return Principal{
					ID:          serviceUser.ID,
					Type:        schema.ServiceUserPrincipal,
					ServiceUser: &serviceUser,
				}, nil
			}
			if err != nil {
				s.log.Debug("failed to parse as user token ", "err", err)
				return Principal{}, errors.ErrUnauthenticated
			}
		}
	}

	if slices.Contains[[]ClientAssertion](assertions, PassthroughHeaderClientAssertion) {
		// check if header with user email is set
		// TODO(kushsharma): this should ideally be deprecated
		if val, ok := GetEmailFromContext(ctx); ok && len(val) > 0 {
			currentUser, err := s.getOrCreateUser(ctx, strings.TrimSpace(val), strings.Split(val, "@")[0])
			if err != nil {
				return Principal{}, err
			}
			return Principal{
				ID:   currentUser.ID,
				Type: schema.UserPrincipal,
				User: &currentUser,
			}, nil
		}
	}
	return Principal{}, errors.ErrUnauthenticated
}

func (s Service) Close() error {
	return s.cron.Stop().Err()
}
