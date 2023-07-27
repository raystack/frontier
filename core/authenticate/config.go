package authenticate

import "time"

type Config struct {
	// OIDCCallbackHost is external host used for oidc redirect uri
	OIDCCallbackHost string `yaml:"oidc_callback_host" mapstructure:"oidc_callback_host"`

	OIDCConfig map[string]OIDCConfig `yaml:"oidc_config" mapstructure:"oidc_config"`
	Session    SessionConfig         `yaml:"session" mapstructure:"session"`
	Token      TokenConfig           `yaml:"token" mapstructure:"token"`
	MailOTP    MailOTPConfig         `yaml:"mail_otp" mapstructure:"mail_otp"`
}

type TokenConfig struct {
	// Path to rsa key file, it can contain more than one key as a json array
	// jwt will be signed by first key, but will be tried to be decoded by all matching key ids, this helps in key rotation.
	// If not provided, access token will not be generated
	RSAPath string `yaml:"rsa_path" mapstructure:"rsa_path"`

	// Issuer uniquely identifies the service that issued the token
	// a good example could be fully qualified domain name
	Issuer string `yaml:"iss" mapstructure:"iss" default:"frontier"`

	// Validity is the duration for which the token is valid
	Validity time.Duration `yaml:"validity" mapstructure:"validity" default:"1h"`
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
	Subject  string        `yaml:"subject" mapstructure:"subject" default:"Frontier Login OTP"`
	Body     string        `yaml:"body" mapstructure:"body" default:"Please copy/paste the OneTimePassword in login form.<h2>{{.Otp}}</h2>This code will expire in 10 minutes."`
	Validity time.Duration `yaml:"validity" mapstructure:"validity" default:"10m"`
}
