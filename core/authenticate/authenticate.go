package authenticate

import (
	"time"

	"github.com/odpf/shield/pkg/metadata"

	"github.com/google/uuid"
)

type AuthMethod string

const (
	MailOTPAuthMethod AuthMethod = "mailotp"
)

func (m AuthMethod) String() string {
	return string(m)
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

func (f Flow) IsValid() bool {
	return f.ExpiresAt.Before(time.Now().UTC())
}

type Config struct {
	// OIDCCallbackHost is external host used for oidc redirect uri
	OIDCCallbackHost string `yaml:"oidc_callback_host" mapstructure:"oidc_callback_host"`

	OIDCConfig map[string]OIDCConfig `yaml:"oidc_config" mapstructure:"oidc_config"`
	Session    SessionConfig         `yaml:"session" mapstructure:"session"`
	Token      TokenConfig           `yaml:"token" mapstructure:"token"`
	MailOTP    MailOTPConfig         `yaml:"mail_otp" mapstructure:"mail_otp"`
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
	ClientID     string        `yaml:"client_id" mapstructure:"client_id"`
	ClientSecret string        `yaml:"client_secret" mapstructure:"client_secret"`
	IssuerUrl    string        `yaml:"issuer_url" mapstructure:"issuer_url"`
	Validity     time.Duration `yaml:"validity" mapstructure:"validity" default:"15m"`
}

type MailOTPConfig struct {
	Subject  string        `yaml:"subject" mapstructure:"subject" default:"Shield Login OTP"`
	Body     string        `yaml:"body" mapstructure:"body" default:"Shield Login Link"`
	Validity time.Duration `yaml:"validity" mapstructure:"validity" default:"10m"`
}
