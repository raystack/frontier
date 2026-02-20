package userpat

import "time"

type Config struct {
	Enabled                bool     `yaml:"enabled" mapstructure:"enabled" default:"false"`
	TokenPrefix            string   `yaml:"token_prefix" mapstructure:"token_prefix" default:"fpt"`
	MaxTokensPerUserPerOrg int64    `yaml:"max_tokens_per_user_per_org" mapstructure:"max_tokens_per_user_per_org" default:"50"`
	MaxTokenLifetime       string   `yaml:"max_token_lifetime" mapstructure:"max_token_lifetime" default:"8760h"`
	DefaultTokenLifetime   string   `yaml:"default_token_lifetime" mapstructure:"default_token_lifetime" default:"2160h"`
	DeniedRoles            []string `yaml:"denied_roles" mapstructure:"denied_roles"`
}

func (c Config) MaxExpiry() time.Duration {
	d, err := time.ParseDuration(c.MaxTokenLifetime)
	if err != nil {
		return 365 * 24 * time.Hour
	}
	return d
}
