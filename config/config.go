package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/odpf/salt/config"
	"github.com/odpf/shield/internal/proxy"
	"github.com/odpf/shield/internal/server"
	"github.com/odpf/shield/internal/store/spicedb"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/logger"
)

type Shield struct {
	// configuration version
	Version  int                  `yaml:"version"`
	Proxy    proxy.ServicesConfig `yaml:"proxy"`
	Log      logger.Config        `yaml:"log"`
	NewRelic NewRelic             `yaml:"new_relic"`
	App      server.Config        `yaml:"app"`
	DB       db.Config            `yaml:"db"`
	SpiceDB  spicedb.Config       `yaml:"spice_db"`
}

type NewRelic struct {
	AppName string `yaml:"app_name" mapstructure:"app_name"`
	License string `yaml:"license" mapstructure:"license"`
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
}

func Load(configfileFromFlag string) (*Shield, error) {
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

	// override all config sources and prioritize one from file
	if configfileFromFlag != "" {
		options = []config.LoaderOption{config.WithFile(configfileFromFlag)}
	}

	l := config.NewLoader(options...)
	if err := l.Load(conf); err != nil {
		if errors.As(err, &config.ConfigFileNotFoundError{}) {
			return nil, errors.New("config file not found")
		} else {
			return nil, err
		}
	}
	return conf, nil
}
