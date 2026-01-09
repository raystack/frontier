package config

import (
	"fmt"

	"github.com/raystack/frontier/billing"

	"github.com/raystack/frontier/pkg/server"

	"github.com/raystack/frontier/internal/store/spicedb"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/salt/config"
)

type Frontier struct {
	// configuration version
	Version  int             `yaml:"version" mapstructure:"version"`
	Log      logger.Config   `yaml:"log" mapstructure:"log"`
	NewRelic NewRelic        `yaml:"new_relic" mapstructure:"new_relic"`
	App      server.Config   `yaml:"app" mapstructure:"app"`
	DB       db.Config       `yaml:"db" mapstructure:"db"`
	UI       server.UIConfig `yaml:"ui" mapstructure:"ui"`
	SpiceDB  spicedb.Config  `yaml:"spicedb" mapstructure:"spicedb"`
	Billing  billing.Config  `yaml:"billing" mapstructure:"billing"`
}

type NewRelic struct {
	AppName string `yaml:"app_name" mapstructure:"app_name"`
	License string `yaml:"license" mapstructure:"license"`
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
}

func Load(serverConfigFileFromFlag string) (*Frontier, error) {
	conf := &Frontier{}

	var options []config.Option
	options = append(options, config.WithEnvPrefix("FRONTIER"))

	// override all config sources and prioritize one from file
	if serverConfigFileFromFlag != "" {
		options = append(options, config.WithFile(serverConfigFileFromFlag))
	}

	l := config.NewLoader(options...)
	if err := l.Load(conf); err != nil {
		return nil, err
	}

	// backward compatibility
	conf = postHook(conf)
	if conf.App.IdentityProxyHeader != "" {
		fmt.Println("WARNING: running in development mode, bypassing all authorization checks")
	}

	return conf, nil
}

func postHook(conf *Frontier) *Frontier {
	if len(conf.App.CorsOrigin) != 0 {
		conf.App.Cors.AllowedOrigins = conf.App.CorsOrigin
	}
	return conf
}
