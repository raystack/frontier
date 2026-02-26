package userpat

import "time"

type Config struct {
	Enabled          bool     `yaml:"enabled" mapstructure:"enabled" default:"false"`
	Prefix           string   `yaml:"prefix" mapstructure:"prefix" default:"fpt"`
	MaxPerUserPerOrg int64    `yaml:"max_per_user_per_org" mapstructure:"max_per_user_per_org" default:"50"`
	MaxLifetime      string   `yaml:"max_lifetime" mapstructure:"max_lifetime" default:"8760h"`
	DefaultLifetime  string   `yaml:"default_lifetime" mapstructure:"default_lifetime" default:"2160h"`
	DeniedRoles      []string `yaml:"denied_roles" mapstructure:"denied_roles"`
}

func (c Config) MaxExpiry() time.Duration {
	d, err := time.ParseDuration(c.MaxLifetime)
	if err != nil {
		return 365 * 24 * time.Hour
	}
	return d
}
