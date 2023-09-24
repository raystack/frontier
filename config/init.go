package config

import (
	_ "embed"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mcuadros/go-defaults"
	"github.com/raystack/frontier/pkg/file"
	"gopkg.in/yaml.v2"
)

func Init(configFile string) error {
	if file.Exist(configFile) {
		return errors.New("config file already exists")
	}

	cfg := &Frontier{}
	defaults.SetDefaults(cfg)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if !file.DirExists(configFile) {
			_ = os.MkdirAll(filepath.Dir(configFile), 0755)
		}
	}

	if err := ioutil.WriteFile(configFile, data, 0655); err != nil {
		return err
	}

	return nil
}
