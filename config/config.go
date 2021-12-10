package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/odpf/salt/config"
)

type Shield struct {
	// configuration version
	Version  int           `yaml:"version"`
	Proxy    ProxyConfig   `yaml:"proxy"`
	Log      LogConfig     `yaml:"log"`
	NewRelic NewRelic      `yaml:"new_relic"`
	App      Service       `yaml:"app"`
	DB       DBConfig      `yaml:"db"`
	SpiceDB  SpiceDBConfig `yaml:"spice_db"`
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

type SpiceDBConfig struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port" default:"50051"`
	PreSharedKey string `yaml:"pre_shared_key"`
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

	// Headers which will have user's email id
	IdentityProxyHeader string `yaml:"identity_proxy_header" mapstructure:"identity_proxy_header"`

	// ResourcesPath is a directory path where resources is defined
	// that this service should implement
	ResourcesConfigPath string `yaml:"resources_config_path" mapstructure:"resources_config_path"`
	// ResourcesPathSecretSecret could be a env name, file path or actual value required
	// to access ResourcesPathSecretPath files
	ResourcesConfigPathSecret string `yaml:"resources_config_path_secret" mapstructure:"resources_config_path_secret"`
}

type NewRelic struct {
	AppName string `yaml:"app_name" mapstructure:"app_name"`
	License string `yaml:"license" mapstructure:"license"`
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
}

type DBConfig struct {
	Driver          string        `yaml:"driver" mapstructure:"driver"`
	URL             string        `yaml:"url" mapstructure:"url"`
	MaxIdleConns    int           `yaml:"max_idle_conns" mapstructure:"max_idle_conns" default:"10"`
	MaxOpenConns    int           `yaml:"max_open_conns" mapstructure:"max_open_conns" default:"10"`
	ConnMaxLifeTime time.Duration `yaml:"conn_max_life_time" mapstructure:"conn_max_life_time" default:"10ms"`
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
		if errors.As(err, &config.ConfigFileNotFoundError{}) {
			fmt.Println(err)
		} else {
			panic(err)
		}
	}
	return conf
}
