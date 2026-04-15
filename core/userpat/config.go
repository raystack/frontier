package userpat

import "time"

type Config struct {
	Enabled           bool        `yaml:"enabled" mapstructure:"enabled" default:"false"`
	Prefix            string      `yaml:"prefix" mapstructure:"prefix" default:"fpt"`
	MaxPerUserPerOrg  int64       `yaml:"max_per_user_per_org" mapstructure:"max_per_user_per_org" default:"50"`
	MaxLifetime       string      `yaml:"max_lifetime" mapstructure:"max_lifetime" default:"8760h"`
	DefaultLifetime   string      `yaml:"default_lifetime" mapstructure:"default_lifetime" default:"2160h"`
	DeniedPermissions []string    `yaml:"denied_permissions" mapstructure:"denied_permissions"`
	Alert             AlertConfig `yaml:"alert" mapstructure:"alert"`
}

type AlertConfig struct {
	Enabled               bool   `yaml:"enabled" mapstructure:"enabled" default:"false"`
	Schedule              string `yaml:"schedule" mapstructure:"schedule" default:"@every 1h"`
	DaysBefore            int    `yaml:"days_before" mapstructure:"days_before" default:"3"`
	ExpiryReminderSubject string `yaml:"expiry_reminder_subject" mapstructure:"expiry_reminder_subject"`
	ExpiryReminderBody    string `yaml:"expiry_reminder_body" mapstructure:"expiry_reminder_body"`
	ExpiredNoticeSubject  string `yaml:"expired_notice_subject" mapstructure:"expired_notice_subject"`
	ExpiredNoticeBody     string `yaml:"expired_notice_body" mapstructure:"expired_notice_body"`
	PATPageURL            string `yaml:"pat_page_url" mapstructure:"pat_page_url"`
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
