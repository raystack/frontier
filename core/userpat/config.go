package userpat

import "time"

type Config struct {
	Enabled           bool     `yaml:"enabled" mapstructure:"enabled" default:"false"`
	Prefix            string   `yaml:"prefix" mapstructure:"prefix" default:"fpt"`
	MaxPerUserPerOrg  int64    `yaml:"max_per_user_per_org" mapstructure:"max_per_user_per_org" default:"50"`
	MaxLifetime       string   `yaml:"max_lifetime" mapstructure:"max_lifetime" default:"8760h"`
	DefaultLifetime   string   `yaml:"default_lifetime" mapstructure:"default_lifetime" default:"2160h"`
	DeniedPermissions []string `yaml:"denied_permissions" mapstructure:"denied_permissions"`
}

func (c Config) MaxExpiry() time.Duration {
	d, err := time.ParseDuration(c.MaxLifetime)
	if err != nil {
		return 365 * 24 * time.Hour
	}
	return d
}

// DeniedPermissionsSet returns denied permissions as a set for efficient lookups.
func (c Config) DeniedPermissionsSet() map[string]struct{} {
	m := make(map[string]struct{}, len(c.DeniedPermissions))
	for _, p := range c.DeniedPermissions {
		m[p] = struct{}{}
	}
	return m
}