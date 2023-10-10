package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/raystack/frontier/billing"

	"github.com/raystack/frontier/pkg/server"

	"github.com/raystack/frontier/internal/store/spicedb"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/logger"
	"github.com/raystack/salt/config"
)

type Frontier struct {
	// configuration version
	Version  int            `yaml:"version"`
	Log      logger.Config  `yaml:"log"`
	NewRelic NewRelic       `yaml:"new_relic"`
	App      server.Config  `yaml:"app"`
	DB       db.Config      `yaml:"db"`
	SpiceDB  spicedb.Config `yaml:"spicedb"`
	Billing  billing.Config `yaml:"billing"`
}

type NewRelic struct {
	AppName string `yaml:"app_name" mapstructure:"app_name"`
	License string `yaml:"license" mapstructure:"license"`
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
}

func Load(serverConfigFileFromFlag string) (*Frontier, error) {
	conf := &Frontier{}

	var options []config.LoaderOption
	options = append(options, config.WithName("config"))
	options = append(options, config.WithEnvKeyReplacer(".", "_"))
	options = append(options, config.WithEnvPrefix("FRONTIER"))
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

	// override all config sources and prioritize one from file
	if serverConfigFileFromFlag != "" {
		options = append(options, config.WithFile(serverConfigFileFromFlag))
	}

	l := config.NewLoader(options...)
	if err := l.Load(conf); err != nil {
		if !errors.As(err, &config.ConfigFileNotFoundError{}) {
			return nil, err
		}
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
