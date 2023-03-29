package authenticate

import (
	"time"

	"github.com/google/uuid"
)

type AuthMethod string

const (
	MailAuthMethod AuthMethod = "mail"
)

func (m AuthMethod) String() string {
	return string(m)
}

// Flow is a temporary state used to finish login/registration flows
type Flow struct {
	ID uuid.UUID

	// authentication flow type
	Method string

	// StartURL is where flow should start from for verification
	StartURL string
	// FinishURL is where flow should end to after successful verification
	FinishURL string

	// Nonce is a once time use random string
	Nonce string

	// CreatedAt will be used to clean-up dead auth flows
	CreatedAt time.Time
}

// Session is created on successful authentication of users
type Session struct {
	ID uuid.UUID

	// UserID is a unique identifier for logged in users
	UserID string

	// AuthenticatedAt is set when a user is successfully authn
	AuthenticatedAt time.Time

	// ExpiresAt is ideally now() + lifespan of session, e.g. 7 days
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (s Session) IsValid() bool {
	if s.ExpiresAt.After(time.Now().UTC()) && !s.AuthenticatedAt.IsZero() {
		return true
	}
	return false
}

type Config struct {
	// OIDCCallbackHost is external host used for oidc redirect uri
	OIDCCallbackHost string `yaml:"oidc_callback_host" mapstructure:"oidc_callback_host"`

	OIDCConfig map[string]OIDCConfig `yaml:"oidc_config" mapstructure:"oidc_config"`
	Session    SessionConfig         `yaml:"session" mapstructure:"session"`
	Token      TokenConfig           `yaml:"token" mapstructure:"token"`
}

type TokenConfig struct {
	// path to rsa key file, it can contain more than one key as a json array
	// jwt will be signed by first key, but will be tried to be decoded by all matching key ids, this helps in key rotation
	RSAPath string `yaml:"rsa_path" mapstructure:"rsa_path"`

	// Issuer uniquely identifies the service that issued the token
	// a good example could be fully qualified domain name
	Issuer string `yaml:"iss" mapstructure:"iss" default:"shield"`
}

type SessionConfig struct {
	HashSecretKey  string `mapstructure:"hash_secret_key" yaml:"hash_secret_key" default:"hash-secret-should-be-32-chars--"`
	BlockSecretKey string `mapstructure:"block_secret_key" yaml:"block_secret_key" default:"block-secret-should-be-32-chars-"`
}

type OIDCConfig struct {
	ClientID     string `yaml:"client_id" mapstructure:"client_id"`
	ClientSecret string `yaml:"client_secret" mapstructure:"client_secret"`
	IssuerUrl    string `yaml:"issuer_url" mapstructure:"issuer_url"`
}
