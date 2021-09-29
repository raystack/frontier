package config

import (
	"os"
	"path/filepath"

	"github.com/odpf/salt/config"
)

type Shield struct {
	// configuration version
	Version int         `yaml:"version"`
	Proxy   ProxyConfig `yaml:"proxy"`
	Log     LogConfig   `yaml:"log"`
}

type LogConfig struct {
	// log level - debug, info, warning, error, fatal
	Level string `yaml:"level" mapstructure:"level" default:"info"`

	// format strategy - plain, json
	Format string `yaml:"format" mapstructure:"format" default:"json"`
}

type ProxyConfig struct {
	Services []Service `yaml:"services" mapstructure:"services"`
}

type Service struct {
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

func Load() *Shield {
	conf := &Shield{}

	var options []config.LoaderOption
	options = append(options, config.WithName(".shield.yaml"))
	options = append(options, config.WithEnvKeyReplacer(".", "_"))
	options = append(options, config.WithEnvPrefix("SHIELD"))
	if p, err := os.Getwd(); err == nil {
		options = append(options, config.WithPath(p))
	}
	if execPath, err := os.Executable(); err == nil {
		options = append(options, config.WithPath(filepath.Dir(execPath)))
	}
	if currentHomeDir, err := os.UserHomeDir(); err == nil {
		options = append(options, config.WithPath(currentHomeDir))
		options = append(options, config.WithPath(filepath.Join(currentHomeDir, ".config")))
	}

	l := config.NewLoader(options...)
	if err := l.Load(conf); err != nil {
		panic(err)
	}
	return conf
}
