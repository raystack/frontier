package mailer

type Config struct {
	SMTPHost     string            `yaml:"smtp_host" mapstructure:"smtp_host"`
	SMTPPort     int               `yaml:"smtp_port" mapstructure:"smtp_port"`
	SMTPUsername string            `yaml:"smtp_username" mapstructure:"smtp_username"`
	SMTPPassword string            `yaml:"smtp_password" mapstructure:"smtp_password"`
	SMTPInsecure bool              `yaml:"smtp_insecure" mapstructure:"smtp_insecure" default:"true"`
	Headers      map[string]string `yaml:"headers" mapstructure:"headers"`
}
