package testusers

type Config struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Domain  string `yaml:"domain" mapstructure:"domain"`
	OTP     string `yaml:"otp" mapstructure:"otp"`
}
