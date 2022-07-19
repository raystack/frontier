package logger

type Config struct {
	// log level - debug, info, warning, error, fatal
	Level string `yaml:"level" mapstructure:"level" default:"info"`

	// format strategy - plain, json
	Format string `yaml:"format" mapstructure:"format" default:"json"`
}
