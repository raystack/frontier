package proxy

type ServicesConfig struct {
	Services []Config `yaml:"services" mapstructure:"services"`
}

type Config struct {
	// port to listen on
	Port int `yaml:"port" mapstructure:"port" default:"8080"`
	// the network interface to listen on
	Host string `yaml:"host" mapstructure:"host" default:"127.0.0.1"`

	Name string

	// RulesPath is a directory path where ruleset is defined
	// that this service should implement
	RulesPath string `yaml:"ruleset" mapstructure:"ruleset"`
	// RulesPathSecret could be a env name, file path or actual value required
	// to access RulesPath files
	RulesPathSecret string `yaml:"ruleset_secret" mapstructure:"ruleset_secret"`
}
