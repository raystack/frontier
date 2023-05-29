package authenticate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/odpf/shield/internal/bootstrap"
	"github.com/odpf/shield/pkg/mailer"

	"github.com/odpf/salt/log"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/odpf/shield/core/authenticate/strategy"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/str"
	"github.com/robfig/cron/v3"
)

const (
	// TODO(kushsharma): should we expose this in config?
	tokenValidity  = time.Hour * 24 * 14
	defaultFlowExp = time.Minute * 10
	maxOTPAttempt  = 5
	otpAttemptKey  = "attempt"
)

var (
	refreshTime               = "0 0 * * *" // Once a day at midnight
	ErrMissingRSADisableToken = errors.New("rsa key missing in config, generate and pass file path")
	ErrStrategyNotApplicable  = errors.New("strategy not applicable")
	ErrUnsupportedMethod      = errors.New("unsupported authentication method")
	ErrInvalidMailOTP         = errors.New("invalid mail otp")
	ErrFlowInvalid            = errors.New("invalid flow or expired")
)

type UserService interface {
	GetByEmail(ctx context.Context, email string) (user.User, error)
	Create(context.Context, user.User) (user.User, error)
}

type FlowRepository interface {
	Set(ctx context.Context, flow *Flow) error
	Get(ctx context.Context, id uuid.UUID) (*Flow, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteExpiredFlows(ctx context.Context) error
}

type RegistrationStartRequest struct {
	Method   string
	ReturnTo string
	Email    string
}

type RegistrationFinishRequest struct {
	Method string

	// used for OIDC & mail otp auth strategy
	Code  string
	State string
}

type RegistrationStartResponse struct {
	Flow  *Flow
	State string
}

type RegistrationFinishResponse struct {
	User user.User
	Flow *Flow
}

type RegistrationService struct {
	log         log.Logger
	cron        *cron.Cron
	flowRepo    FlowRepository
	userService UserService
	config      Config
	mailDialer  mailer.Dialer
	Now         func() time.Time
}

func NewRegistrationService(logger log.Logger, config Config, flowRepo FlowRepository,
	userService UserService, mailDialer mailer.Dialer) *RegistrationService {
	r := &RegistrationService{
		log:         logger,
		cron:        cron.New(),
		flowRepo:    flowRepo,
		userService: userService,
		config:      config,
		mailDialer:  mailDialer,
		Now:         time.Now().UTC,
	}
	return r
}

func (r RegistrationService) SupportedStrategies() []string {
	// add here strategies like mail link once implemented
	var strategies = []string{}
	for name := range r.config.OIDCConfig {
		strategies = append(strategies, name)
	}
	if r.mailDialer != nil {
		strategies = append(strategies, MailOTPAuthMethod.String())
	}
	return strategies
}

func (r RegistrationService) Start(ctx context.Context, request RegistrationStartRequest) (*RegistrationStartResponse, error) {
	if !bootstrap.Contains(r.SupportedStrategies(), request.Method) {
		return nil, ErrUnsupportedMethod
	}
	flow := &Flow{
		ID:        uuid.New(),
		Method:    request.Method,
		FinishURL: request.ReturnTo,
		CreatedAt: r.Now(),
		ExpiresAt: r.Now().Add(defaultFlowExp),
	}

	if request.Method == MailOTPAuthMethod.String() {
		mailLinkStrat := strategy.NewMailLink(r.mailDialer, r.config.MailOTP.Subject, r.config.MailOTP.Body)
		nonce, err := mailLinkStrat.SendMail(request.Email)
		if err != nil {
			return nil, err
		}

		flow.Nonce = nonce
		if r.config.MailOTP.Validity != 0 {
			flow.ExpiresAt = flow.CreatedAt.Add(r.config.MailOTP.Validity)
		}
		flow.Email = strings.ToLower(request.Email)
		if err = r.flowRepo.Set(ctx, flow); err != nil {
			return nil, err
		}
		return &RegistrationStartResponse{
			Flow:  flow,
			State: flow.ID.String(),
		}, nil
	}

	// check for oidc flow
	if oidcConfig, ok := r.config.OIDCConfig[request.Method]; ok {
		idp, err := strategy.NewRelyingPartyOIDC(
			oidcConfig.ClientID,
			oidcConfig.ClientSecret,
			r.config.OIDCCallbackHost).
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
		if err = r.flowRepo.Set(ctx, flow); err != nil {
			return nil, err
		}
		return &RegistrationStartResponse{
			Flow: flow,
		}, nil
	}

	return nil, ErrUnsupportedMethod
}

func (r RegistrationService) Finish(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	if request.Method == MailOTPAuthMethod.String() {
		response, err := r.applyMail(ctx, request)
		if err != nil && !errors.Is(err, ErrStrategyNotApplicable) {
			return nil, err
		}
		return response, nil
	}

	// check for oidc method config
	{
		response, err := r.applyOIDC(ctx, request)
		if err == nil {
			return response, nil
		}
		if err != nil && !errors.Is(err, ErrStrategyNotApplicable) {
			return nil, err
		}
	}
	return nil, ErrUnsupportedMethod
}

// applyMail actions when user submitted otp from the email
// user can be considered as verified if correct
// create a new user if required
func (r RegistrationService) applyMail(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	if len(request.Code) == 0 {
		return nil, ErrStrategyNotApplicable
	}
	flowID, err := uuid.Parse(request.State)
	if err != nil {
		return nil, ErrStrategyNotApplicable
	}
	flow, err := r.flowRepo.Get(ctx, flowID)
	if err != nil {
		return nil, fmt.Errorf("invalid state for mail otp: %w", err)
	}
	if !flow.IsValid() {
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
			if err = r.flowRepo.Set(ctx, flow); err != nil {
				return nil, fmt.Errorf("failed to process flow code missmatch")
			}
		} else {
			if err = r.consumeFlow(ctx, flowID); err != nil {
				return nil, fmt.Errorf("failed to process flow code missmatch")
			}
		}
		return nil, ErrInvalidMailOTP
	}

	// consume this flow
	if err = r.consumeFlow(ctx, flow.ID); err != nil {
		return nil, fmt.Errorf("failed to successfully register via otp: %w", err)
	}

	newUser, err := r.getOrCreateUser(ctx, flow.Email, "")
	if err != nil {
		return nil, err
	}
	return &RegistrationFinishResponse{
		User: newUser,
		Flow: flow,
	}, nil
}

func (r RegistrationService) applyOIDC(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
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
	flow, err := r.flowRepo.Get(ctx, flowID)
	if err != nil {
		return nil, err
	}

	// can't find oidc config
	oidcConfig, ok := r.config.OIDCConfig[flow.Method]
	if !ok {
		return nil, ErrStrategyNotApplicable
	}

	idp, err := strategy.NewRelyingPartyOIDC(
		oidcConfig.ClientID,
		oidcConfig.ClientSecret,
		r.config.OIDCCallbackHost).
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
	newUser, err := r.getOrCreateUser(ctx, oauthProfile.Email, oauthProfile.Name)
	if err != nil {
		return nil, err
	}

	return &RegistrationFinishResponse{
		User: newUser,
		Flow: flow,
	}, nil
}

func (r RegistrationService) Token(user user.User, orgs []organization.Organization) ([]byte, error) {
	if len(r.config.Token.RSAPath) == 0 {
		return nil, ErrMissingRSADisableToken
	}
	keySet, err := jwk.ReadFile(r.config.Token.RSAPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rsa key: %w", err)
	}
	// use first key to sign token
	rsaKey, ok := keySet.Key(0)
	if !ok {
		return nil, errors.New("missing rsa key to generate token")
	}

	var orgNames []string
	for _, o := range orgs {
		orgNames = append(orgNames, o.Name)
	}

	tok, err := jwt.NewBuilder().
		Issuer(r.config.Token.Issuer).
		IssuedAt(time.Now().UTC()).
		NotBefore(time.Now().UTC()).
		Expiration(time.Now().UTC().Add(tokenValidity)).
		JwtID(uuid.New().String()).
		Subject(user.ID).
		Claim("org", strings.Join(orgNames, ",")).
		Build()
	if err != nil {
		return nil, err
	}

	return jwt.Sign(tok, jwt.WithKey(jwa.RS256, rsaKey))
}

func (r RegistrationService) InitFlows(ctx context.Context) error {
	_, err := r.cron.AddFunc(refreshTime, func() {
		if err := r.flowRepo.DeleteExpiredFlows(ctx); err != nil {
			r.log.Warn("failed to delete expired sessions", "err", err)
		}
	})
	if err != nil {
		return fmt.Errorf("failed to start flows cronjob: %w", err)
	}
	r.cron.Start()
	return nil
}

func (r RegistrationService) consumeFlow(ctx context.Context, id uuid.UUID) error {
	return r.flowRepo.Delete(ctx, id)
}

func (r RegistrationService) getOrCreateUser(ctx context.Context, email, title string) (user.User, error) {
	// create a new user based on email if it doesn't exist
	existingUser, err := r.userService.GetByEmail(ctx, email)
	if err == nil {
		// user is already registered

		// TODO(kushsharma): should we update metadata like profile picture from social logins
		// for registered users every time the login?
		return existingUser, nil
	}

	// register a new user
	return r.userService.Create(ctx, user.User{
		Title: title,
		Email: email,
		Name:  str.GenerateUserSlug(email),
	})
}

func (r RegistrationService) Close() {
	r.cron.Stop()
}
