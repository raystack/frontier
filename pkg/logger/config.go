package logger

type Config struct {
	// log level - debug, info, warning, error, fatal
	Level string `yaml:"level" mapstructure:"level" default:"info" json:"level,omitempty"`

	// format strategy - plain, json
	Format string `yaml:"format" mapstructure:"format" default:"json" json:"format,omitempty"`

	// audit system events - none(default), stdout, db
	AuditEvents string `yaml:"audit_events" mapstructure:"audit_events" default:"none" json:"audit_events,omitempty"`

	// IgnoredAuditEvents contains list of events which should be ignored in audit logs
	IgnoredAuditEvents []string `yaml:"ignored_audit_events" mapstructure:"ignored_audit_events" json:"ignored_audit_events,omitempty"`
}
