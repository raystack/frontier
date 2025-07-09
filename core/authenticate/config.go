package authenticate

import (
	"time"

	testusers "github.com/raystack/frontier/core/authenticate/test_users"
)

type Config struct {
	// CallbackURLs is external host used for redirect uri
	// host specified at 0th index will be used as default
	CallbackURLs []string `yaml:"callback_urls" mapstructure:"callback_urls" default:"[http://localhost:7400/v1beta1/auth/callback]"`

	AuthorizedRedirectURLs []string `yaml:"authorized_redirect_urls" mapstructure:"authorized_redirect_urls" `

	OIDCConfig map[string]OIDCConfig `yaml:"oidc_config" mapstructure:"oidc_config"`
	Session    SessionConfig         `yaml:"session" mapstructure:"session"`
	Token      TokenConfig           `yaml:"token" mapstructure:"token"`
	MailOTP    MailOTPConfig         `yaml:"mail_otp" mapstructure:"mail_otp"`
	MailLink   MailLinkConfig        `yaml:"mail_link" mapstructure:"mail_link"`
	PassKey    PassKeyConfig         `yaml:"passkey" mapstructure:"passkey"`
	TestUsers  testusers.Config      `yaml:"test_users" mapstructure:"test_users"`
}

type TokenConfig struct {
	// Path to rsa key file, it can contain more than one key as a json array
	// jwt will be signed by first key, but will be tried to be decoded by all matching key ids, this helps in key rotation.
	// If not provided, access token will not be generated
	RSAPath string `yaml:"rsa_path" mapstructure:"rsa_path"`
	// RSABase64 is base64 encoded rsa key, it can contain more than one key as a json array
	RSABase64 string `yaml:"rsa_base64" mapstructure:"rsa_base64"`

	// Issuer uniquely identifies the service that issued the token
	// a good example could be fully qualified domain name
	Issuer string `yaml:"iss" mapstructure:"iss" default:"frontier"`

	// Validity is the duration for which the token is valid
	Validity time.Duration `yaml:"validity" mapstructure:"validity" default:"1h"`

	Claims TokenClaimConfig `yaml:"claims" mapstructure:"claims"`
}

type TokenClaimConfig struct {
	AddOrgIDsClaim    bool `yaml:"add_org_ids" mapstructure:"add_org_ids" default:"true"`
	AddUserEmailClaim bool `yaml:"add_user_email" mapstructure:"add_user_email" default:"true"`
}

type SessionConfig struct {
	HashSecretKey  string `mapstructure:"hash_secret_key" yaml:"hash_secret_key" default:"hash-secret-should-be-32-chars--"`
	BlockSecretKey string `mapstructure:"block_secret_key" yaml:"block_secret_key" default:"block-secret-should-be-32-chars-"`
	Domain         string `mapstructure:"domain" yaml:"domain" default:""`
	// SameSite can be set to "default", "lax", "strict" or "none"
	SameSite string `mapstructure:"same_site" yaml:"same_site" default:"lax"`
	// Validity is the duration for which the session is valid
	Validity time.Duration `mapstructure:"validity" yaml:"validity" default:"720h"`
	Secure   bool          `mapstructure:"secure" yaml:"secure" default:"false"`
}

type OIDCConfig struct {
	ClientID     string        `yaml:"client_id" mapstructure:"client_id"`
	ClientSecret string        `yaml:"client_secret" mapstructure:"client_secret"`
	IssuerUrl    string        `yaml:"issuer_url" mapstructure:"issuer_url"`
	Validity     time.Duration `yaml:"validity" mapstructure:"validity" default:"15m"`
}

type MailOTPConfig struct {
	Subject  string        `yaml:"subject" mapstructure:"subject" default:"Frontier Login - OTP"`
	Body     string        `yaml:"body" mapstructure:"body" default:"Hi {{.Email}},<br> Please copy/paste the One Time Password in login form.<h2>{{.Otp}}</h2>This code will expire in 10 minutes."`
	Validity time.Duration `yaml:"validity" mapstructure:"validity" default:"10m"`
}

type MailLinkConfig struct {
	Subject  string        `yaml:"subject" mapstructure:"subject" default:"Frontier Login - One time link"`
	Body     string        `yaml:"body" mapstructure:"body" default:"Click on the following link or copy/paste the url in browser to login.<h3><a href='{{.Link}}' target='_blank'>Login</a></h3>Address: {{.Link}} <br>This link will expire in 10 minutes."`
	Validity time.Duration `yaml:"validity" mapstructure:"validity" default:"10m"`
}

type PassKeyConfig struct {
	// RPDisplayName configures the display name for the Relying Party Server. This can be any string.
	RPDisplayName string `yaml:"rpdisplayname" mapstructure:"rpdisplayname"`
	// RPID configures the Relying Party Server ID. This should generally be the origin without a scheme and port.
	RPID string `yaml:"rpid" mapstructure:"rpid"`
	// RPOrigins configures the list of Relying Party Server Origins that are permitted. These should be fully
	// qualified origins.
	RPOrigins []string `yaml:"rporigins" mapstructure:"rporigins"`
}
