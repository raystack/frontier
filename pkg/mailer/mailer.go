package mailer

import (
	"strings"

	"gopkg.in/mail.v2"
)

type Config struct {
	SMTPHost     string            `yaml:"smtp_host" mapstructure:"smtp_host"`
	SMTPPort     int               `yaml:"smtp_port" mapstructure:"smtp_port"`
	SMTPUsername string            `yaml:"smtp_username" mapstructure:"smtp_username"`
	SMTPPassword string            `yaml:"smtp_password" mapstructure:"smtp_password"`
	SMTPInsecure bool              `yaml:"smtp_insecure" mapstructure:"smtp_insecure" default:"true"`
	Headers      map[string]string `yaml:"headers" mapstructure:"headers"`

	// SMTP TLS policy to use when establishing a connection.
	// Defaults to MandatoryStartTLS.
	// Possible values are:
	// opportunistic: Use STARTTLS if the server supports it, otherwise connect without encryption.
	// mandatory: Always use STARTTLS.
	// none: Never use STARTTLS.
	SMTPTLSPolicy string `yaml:"smtp_tls_policy" mapstructure:"smtp_tls_policy" default:"mandatory"`
}

func (c Config) TLSPolicy() mail.StartTLSPolicy {
	switch strings.ToLower(c.SMTPTLSPolicy) {
	case "opportunistic":
		return mail.OpportunisticStartTLS
	case "mandatory":
		return mail.MandatoryStartTLS
	case "none":
		return mail.NoStartTLS
	}
	return mail.MandatoryStartTLS
}
