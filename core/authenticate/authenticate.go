package authenticate

import (
	"time"

	"github.com/raystack/frontier/core/authenticate/strategy"

	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	pat "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/pkg/metadata"

	"github.com/google/uuid"
)

type AuthMethod string

const (
	MailOTPAuthMethod  = AuthMethod(strategy.MailOTPAuthMethod)
	MailLinkAuthMethod = AuthMethod(strategy.MailLinkAuthMethod)
	PassKeyAuthMethod  = AuthMethod(strategy.PasskeyAuthMethod)
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
	// JWTGrantClientAssertion is used to authenticate using JWT token generated
	// using public/private key pair that provides access token for the client
	JWTGrantClientAssertion ClientAssertion = "jwt_grant"
	// ClientCredentialsClientAssertion is used to authenticate using client_id and client_secret
	// that provides access token for the client
	ClientCredentialsClientAssertion ClientAssertion = "client_credentials"
	// PATClientAssertion is used to authenticate using Personal Access Token
	PATClientAssertion ClientAssertion = "pat"
	// PassthroughHeaderClientAssertion is used to authenticate using headers passed by the client
	// this is non secure way of authenticating client in test environments
	PassthroughHeaderClientAssertion ClientAssertion = "passthrough_header"
)

func (a ClientAssertion) String() string {
	return string(a)
}

var APIAssertions = []ClientAssertion{
	SessionClientAssertion,
	PATClientAssertion,
	AccessTokenClientAssertion,
	JWTGrantClientAssertion,
	// ClientCredentialsClientAssertion should be removed in future to avoid DDOS attacks on CPU
	// and should only be allowed to be used get access token for the client
	ClientCredentialsClientAssertion,
	PassthroughHeaderClientAssertion,
}

// FlowIntent says whether the caller wants to log an existing user in or create
// a new one. It rides on the flow, the only thing that survives an OIDC redirect.
type FlowIntent string

const (
	// FlowIntentUnspecified keeps the old create-or-get behaviour, so clients
	// that send no intent are unaffected.
	FlowIntentUnspecified FlowIntent = ""
	FlowIntentLogin       FlowIntent = "login"
	FlowIntentSignup      FlowIntent = "signup"
)

func (i FlowIntent) String() string {
	return string(i)
}

// keys under Flow.Metadata.
const (
	flowIntentKey  = "intent"
	flowConsentKey = "consent"

	consentDocumentIDsKey = "accepted_document_ids"
	consentIPAddressKey   = "ip_address"
	consentAtKey          = "at"
)

// FlowConsent is what the user accepted at flow start, read back after the
// redirect. The IP and timestamp are from the acceptance, not the callback.
type FlowConsent struct {
	AcceptedDocumentIDs []string
	IPAddress           string
	At                  time.Time
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

// Intent reads the flow intent from metadata. A nil flow is allowed, so callers
// with no flow need no branch, and reads as unspecified.
func (f *Flow) Intent() FlowIntent {
	if f == nil {
		return FlowIntentUnspecified
	}
	intent, ok := f.Metadata[flowIntentKey].(string)
	if !ok {
		return FlowIntentUnspecified
	}
	return FlowIntent(intent)
}

// Consent reads what the user accepted from metadata, reporting whether the flow
// carries one at all. A nil flow is allowed.
//
// Metadata is JSONB and does not return the types it was given — ids come back
// as []any and the timestamp as a string — so this parses rather than asserts,
// and treats an unparseable key as no consent.
func (f *Flow) Consent() (FlowConsent, bool) {
	if f == nil {
		return FlowConsent{}, false
	}
	raw, ok := f.Metadata[flowConsentKey].(map[string]any)
	if !ok {
		return FlowConsent{}, false
	}

	consent := FlowConsent{
		AcceptedDocumentIDs: parseStringSlice(raw[consentDocumentIDsKey]),
	}
	if ip, ok := raw[consentIPAddressKey].(string); ok {
		consent.IPAddress = ip
	}
	switch at := raw[consentAtKey].(type) {
	case time.Time:
		consent.At = at
	case string:
		if parsed, err := time.Parse(time.RFC3339Nano, at); err == nil {
			consent.At = parsed
		}
	}
	if len(consent.AcceptedDocumentIDs) == 0 {
		// a consent that names no document is not a consent
		return FlowConsent{}, false
	}
	return consent, true
}

// parseStringSlice reads a string list that may have been through a JSON round
// trip, where it comes back as []any. Non-string entries are dropped.
func parseStringSlice(value any) []string {
	switch list := value.(type) {
	case []string:
		return list
	case []any:
		parsed := make([]string, 0, len(list))
		for _, item := range list {
			if str, ok := item.(string); ok {
				parsed = append(parsed, str)
			}
		}
		return parsed
	default:
		return nil
	}
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

	// Intent says whether this is a login or a signup. Unset keeps create-or-get.
	Intent FlowIntent
	// AcceptedDocumentIDs are stored on the flow so they survive the redirect to
	// an identity provider, and are checked when the user is created.
	AcceptedDocumentIDs []string
	// IPAddress is where the acceptance came from. Authenticate is skip-listed,
	// so the handler extracts it rather than reading it off the context.
	IPAddress string
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
	// E.g. app/user, app/serviceuser, app/pat
	Type string
	// AuthVia is the credential type that authenticated this principal
	AuthVia ClientAssertion

	User        *user.User
	ServiceUser *serviceuser.ServiceUser
	PAT         *pat.PAT
}

// ResolveSubject returns the subject ID and type for authorization queries.
// For PAT principals, it resolves to the underlying user.
func (p Principal) ResolveSubject() (id string, subjectType string) {
	if p.PAT != nil {
		return p.PAT.UserID, schema.UserPrincipal
	}
	return p.ID, p.Type
}
