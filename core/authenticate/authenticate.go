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
	PassKeyAuthMethod  = AuthMethod(strategy.PasskeyAuthMethod)
	OIDCAuthMethod = AuthMethod(strategy.OIDCAuthMethod)
)

func (m AuthMethod) String() string {
	return string(m)
}

type ClientAssertion string

const (
	// SessionClientAssertion is used to authenticate using session cookie
	SessionClientAssertion ClientAssertion = "session"
	// AccessTokenClientAssertion is used to authenticate using access token generated
	// by the system for the user
	AccessTokenClientAssertion ClientAssertion = "access_token"
	// OpaqueTokenClientAssertion is used to authenticate using opaque token generated
	// for API clients
	OpaqueTokenClientAssertion ClientAssertion = "opaque"
	// JWTGrantClientAssertion is used to authenticate using JWT token generated
	// using public/private key pair that provides access token for the client
	JWTGrantClientAssertion ClientAssertion = "jwt_grant"
	// ClientCredentialsClientAssertion is used to authenticate using client_id and client_secret
	// that provides access token for the client
	ClientCredentialsClientAssertion ClientAssertion = "client_credentials"
	// PassthroughHeaderClientAssertion is used to authenticate using headers passed by the client
	// this is non secure way of authenticating client in test environments
	PassthroughHeaderClientAssertion ClientAssertion = "passthrough_header"
)

func (a ClientAssertion) String() string {
	return string(a)
}

var APIAssertions = []ClientAssertion{
	SessionClientAssertion,
	AccessTokenClientAssertion,
	OpaqueTokenClientAssertion,
	JWTGrantClientAssertion,
	// ClientCredentialsClientAssertion should be removed in future to avoid DDOS attacks on CPU
	// and should only be allowed to be used get access token for the client
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
	Code        string
	State       string
	StateConfig map[string]any
}

type RegistrationStartResponse struct {
	Flow        *Flow
	State       string
	StateConfig map[string]any
}

type RegistrationFinishResponse struct {
	User user.User
	Flow *Flow
}

type Principal struct {
	// ID is the unique identifier of principal
	ID string
	// Type is the namespace of principal
	// E.g. app/user, app/serviceuser
	Type string

	User        *user.User
	ServiceUser *serviceuser.ServiceUser
}
