package authenticate

import (
	"time"

	"github.com/raystack/frontier/core/authenticate/strategy"

	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"

	"github.com/raystack/frontier/pkg/metadata"

	"github.com/google/uuid"
)

type AuthMethod string

const (
	MailOTPAuthMethod  = AuthMethod(strategy.MailOTPAuthMethod)
	MailLinkAuthMethod = AuthMethod(strategy.MailLinkAuthMethod)
)

func (m AuthMethod) String() string {
	return string(m)
}

type ClientAssertion string

const (
	SessionClientAssertion           ClientAssertion = "session"
	AccessTokenClientAssertion       ClientAssertion = "access_token"
	JWTGrantClientAssertion          ClientAssertion = "jwt_grant"
	ClientCredentialsClientAssertion ClientAssertion = "client_credentials"
	PassthroughHeaderClientAssertion ClientAssertion = "passthrough_header"
)

func (a ClientAssertion) String() string {
	return string(a)
}

var AllClientAssertions = []ClientAssertion{
	SessionClientAssertion,
	AccessTokenClientAssertion,
	JWTGrantClientAssertion,
	ClientCredentialsClientAssertion,
	PassthroughHeaderClientAssertion,
}

// Flow is a temporary state used to finish login/registration flows
type Flow struct {
	ID uuid.UUID

	// authentication flow type
	Method string

	// Email is the email of the user
	Email string

	// StartURL is where flow should start from for verification
	StartURL string
	// FinishURL is where flow should end to after successful verification
	FinishURL string

	// Nonce is a once time use random string
	Nonce string

	Metadata metadata.Metadata

	// CreatedAt will be used to clean-up dead auth flows
	CreatedAt time.Time

	// ExpiresAt is the time when the flow will expire
	ExpiresAt time.Time
}

func (f Flow) IsValid(currentTime time.Time) bool {
	return f.ExpiresAt.After(currentTime)
}

type RegistrationStartRequest struct {
	Method string
	// ReturnToURL is where flow should end to after successful verification
	ReturnToURL string
	Email       string

	// callback_url will be used by strategy as last step to finish authentication flow
	// in OIDC this host will receive "state" and "code" query params, in case of magic links
	// this will be the url where user is redirected after clicking on magic link.
	// For most cases it could be host of frontier but in case of proxies, this will be proxy public endpoint.
	// callback_url should be one of the allowed urls configured at instance level
	CallbackUrl string
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

type Principal struct {
	ID   string
	Type string

	User        *user.User
	ServiceUser *serviceuser.ServiceUser
}
