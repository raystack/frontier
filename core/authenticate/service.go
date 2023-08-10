package authenticate

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/raystack/frontier/core/audit"

	"golang.org/x/exp/slices"

	"github.com/lestrrat-go/jwx/v2/jwt"

	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/bootstrap/schema"
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
	ErrFlowInvalid           = errors.New("invalid flow or expired")
)

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	Create(context.Context, user.User) (user.User, error)
}

type ServiceUserService interface {
	GetByToken(ctx context.Context, token string) (serviceuser.ServiceUser, error)
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

type Service struct {
	log                  log.Logger
	cron                 *cron.Cron
	flowRepo             FlowRepository
	userService          UserService
	config               Config
	mailDialer           mailer.Dialer
	Now                  func() time.Time
	internalTokenService token.Service
	sessionService       SessionService
	serviceUserService   ServiceUserService
}

func NewService(logger log.Logger, config Config, flowRepo FlowRepository,
	mailDialer mailer.Dialer, tokenService token.Service, sessionService SessionService,
	userService UserService, serviceUserService ServiceUserService) *Service {
	r := &Service{
		log:         logger,
		cron:        cron.New(),
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
	}
	return r
}

func (s Service) SupportedStrategies() []string {
	// add here strategies like mail link once implemented
	var strategies = []string{}
	for name := range s.config.OIDCConfig {
		strategies = append(strategies, name)
	}
	if s.mailDialer != nil {
		strategies = append(strategies, MailOTPAuthMethod.String(), MailLinkAuthMethod.String())
	}
	return strategies
}

func (s Service) StartFlow(ctx context.Context, request RegistrationStartRequest) (*RegistrationStartResponse, error) {
	if !utils.Contains(s.SupportedStrategies(), request.Method) {
		return nil, ErrUnsupportedMethod
	}
	flow := &Flow{
		ID:        uuid.New(),
		Method:    request.Method,
		FinishURL: request.ReturnTo,
		CreatedAt: s.Now(),
		ExpiresAt: s.Now().Add(defaultFlowExp),
	}

	if request.Method == MailOTPAuthMethod.String() {
		mailLinkStrat := strategy.NewMailOTP(s.mailDialer, s.config.MailOTP.Subject, s.config.MailOTP.Body)
		nonce, err := mailLinkStrat.SendMail(request.Email)
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
	if request.Method == MailLinkAuthMethod.String() {
		mailLinkStrat := strategy.NewMailLink(s.mailDialer, s.config.CallbackHost, s.config.MailLink.Subject, s.config.MailLink.Body)
		nonce, err := mailLinkStrat.SendMail(flow.ID.String(), request.Email)
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
			s.config.CallbackHost).
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
	if flow.Nonce != request.Code {
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

func (s Service) applyOIDC(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	// flow id is added in state params
	if len(request.State) == 0 {
		return nil, errors.New("invalid auth state")
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

	idp, err := strategy.NewRelyingPartyOIDC(
		oidcConfig.ClientID,
		oidcConfig.ClientSecret,
		s.config.CallbackHost).
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
func (s Service) BuildToken(ctx context.Context, subjectID string, metadata map[string]string) ([]byte, error) {
	return s.internalTokenService.Build(subjectID, metadata)
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
	audit.GetAuditor(ctx, schema.PlatformOrgID.String()).
		Log(audit.UserCreatedEvent, audit.UserTarget(newUser.ID))
	return newUser, nil
}

func (s Service) GetPrincipal(ctx context.Context, assertions ...ClientAssertion) (Principal, error) {
	var currentPrincipal Principal
	if len(assertions) == 0 {
		// check all assertions
		assertions = AllClientAssertions
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
		insecureJWT, err := jwt.ParseInsecure([]byte(userToken))
		if err != nil {
			return Principal{}, errors.ErrUnauthenticated
		}

		if slices.Contains[[]ClientAssertion](assertions, AccessTokenClientAssertion) {
			//check type of jwt
			if val, ok := insecureJWT.Get(token.GeneratedClaimKey); ok {
				if claimVal, ok := val.(string); ok && claimVal == token.GeneratedClaimValue {
					// extract user from token if present as its created by frontier
					userID, _, err := s.internalTokenService.Parse(ctx, []byte(userToken))
					if err == nil && utils.IsValidUUID(userID) {
						// userID is a valid uuid
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
					if err != nil {
						s.log.Debug("failed to parse as internal token ", "err", err)
						return Principal{}, errors.ErrUnauthenticated
					}
				}
			}
		}

		// extract user from token if it's a service user
		if slices.Contains[[]ClientAssertion](assertions, JWTGrantClientAssertion) {
			serviceUser, err := s.serviceUserService.GetByToken(ctx, userToken)
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
	if slices.Contains[[]ClientAssertion](assertions, ClientCredentialsClientAssertion) {
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

func (s Service) Close() {
	s.cron.Stop()
}
