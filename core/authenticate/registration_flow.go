package authenticate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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
	tokenValidity = time.Hour * 24 * 30
)

var (
	refreshTime               = "0 0 * * *" // Once a day at midnight
	ErrMissingRSADisableToken = errors.New("rsa key missing in config, generate and pass file path")
	ErrStrategyNotApplicable  = errors.New("strategy not applicable")
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

type RegistrationService struct {
	log         log.Logger
	cron        *cron.Cron
	flowRepo    FlowRepository
	userService UserService
	config      Config
}

func NewRegistrationService(logger log.Logger, flowRepo FlowRepository, userService UserService, config Config) *RegistrationService {
	return &RegistrationService{
		log:         logger,
		cron:        cron.New(),
		flowRepo:    flowRepo,
		userService: userService,
		config:      config,
	}
}

type RegistrationStartRequest struct {
	Method   string
	ReturnTo string
}

type RegistrationFinishRequest struct {
	Method string

	// used for OIDC auth strategy
	OAuthCode  string
	OAuthState string
}

type RegistrationStartResponse struct {
	Flow *Flow
}

type RegistrationFinishResponse struct {
	User user.User
	Flow *Flow
}

func (r RegistrationService) SupportedStrategies() []string {
	// add here strategies like mail link once implemented
	var strategies = []string{}
	for name := range r.config.OIDCConfig {
		strategies = append(strategies, name)
	}
	return strategies
}

func (r RegistrationService) Start(ctx context.Context, request RegistrationStartRequest) (*RegistrationStartResponse, error) {
	flow := &Flow{
		ID:        uuid.New(),
		Method:    request.Method,
		FinishURL: request.ReturnTo,
		CreatedAt: time.Now().UTC(),
	}

	if request.Method == MailAuthMethod.String() {
		// get request.email id
		// create a new flow
		// use sns service to send a mail using flow id/nonce & method type
		// TODO
		panic("unsupported")
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
		if err = r.flowRepo.Set(ctx, flow); err != nil {
			return nil, err
		}
		return &RegistrationStartResponse{
			Flow: flow,
		}, nil
	}

	return nil, errors.New("unsupported authentication method")
}

func (r RegistrationService) Finish(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	{
		response, err := r.applyMail(ctx, request)
		if err == nil {
			return response, nil
		}
		if err != nil && !errors.Is(err, ErrStrategyNotApplicable) {
			return nil, err
		}
	}

	{
		response, err := r.applyOIDC(ctx, request)
		if err == nil {
			return response, nil
		}
		if err != nil && !errors.Is(err, ErrStrategyNotApplicable) {
			return nil, err
		}
	}

	return nil, errors.New("unsupported authentication method")
}

func (r RegistrationService) applyMail(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	if request.Method == MailAuthMethod.String() {
		// TODO(kushsharma):
		// user clicked on nonce link from the email
		// user can be considered as verified
		// create a new user if required
	}
	return nil, ErrStrategyNotApplicable
}

func (r RegistrationService) applyOIDC(ctx context.Context, request RegistrationFinishRequest) (*RegistrationFinishResponse, error) {
	// flow id is added in state params
	if len(request.OAuthState) == 0 {
		return nil, errors.New("invalid auth state")
	}

	// check for oidc flow via fetching oauth state, method parameter will not be set for oauth
	flowIDFromState, err := strategy.ExtractFlowFromOIDCState(request.OAuthState)
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
	authToken, err := idp.Token(ctx, request.OAuthCode, flow.Nonce)
	if err != nil {
		return nil, err
	}
	oauthProfile, err := idp.GetUser(ctx, authToken)
	if err != nil {
		return nil, err
	}

	// create a new user based on email if it doesn't exist
	existingUser, err := r.userService.GetByEmail(ctx, oauthProfile.Email)
	if err == nil {
		// user is already registered

		// TODO(kushsharma): should we update metadata like profile picture from outside the app
		// for registered users every time the login?
		return &RegistrationFinishResponse{
			User: existingUser,
			Flow: flow,
		}, nil
	}

	// register a new user
	newUser, err := r.userService.Create(ctx, user.User{
		Name:  oauthProfile.Name,
		Email: oauthProfile.Email,
		Slug:  str.GenerateUserSlug(oauthProfile.Email),
	})
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
		orgNames = append(orgNames, o.Slug)
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

func (r RegistrationService) Close() {
	r.cron.Stop()
}
