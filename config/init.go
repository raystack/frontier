package config

import (
	"embed"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/mcuadros/go-defaults"
	"github.com/odpf/shield/internal/proxy"
	"github.com/odpf/shield/pkg/file"
	"gopkg.in/yaml.v2"
)

//go:embed resources_config/*
var resourcesConfig embed.FS

//go:embed rules/*
var rulesConfig embed.FS

func Init(resourcesURL, rulesURL, configFile string) error {
	if file.Exist(configFile) {
		return errors.New("config file already exists")
	}

	cfg := &Shield{}

	defaults.SetDefaults(cfg)

	if err := initResourcesPath(resourcesURL); err != nil {
		return err
	}
	if err := initRulesPath(rulesURL); err != nil {
		return err
	}

	cfg.App.RulesPath = rulesURL
	cfg.App.ResourcesConfigPath = resourcesURL
	// sample proxy
	cfg.Proxy = proxy.ServicesConfig{
		Services: []proxy.Config{
			{
				Name:      "base",
				Host:      "0.0.0.0",
				Port:      5556,
				RulesPath: rulesURL,
			},
		},
	}

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

func initResourcesPath(resURL string) error {
	resourceURL, err := url.Parse(resURL)
	if err != nil {
		return err
	}

	if resourceURL.Scheme != "file" {
		// skip creating
		return nil
	}

	resourcesPath := resourceURL.Path
	if !file.DirExists(resourcesPath) {
		_ = os.MkdirAll(resourcesPath, 0755)
	}

	files, err := ioutil.ReadDir(resourcesPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".yaml") || strings.HasSuffix(f.Name(), ".yml") {
			// skip creating
			return nil
		}
	}

	resourceYaml, err := resourcesConfig.ReadFile("resources_config/resources.yaml")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path.Join(resourcesPath, "resources.yaml"), resourceYaml, 0655); err != nil {
		return err
	}

	return nil
}

func initRulesPath(rURL string) error {
	rulesURL, err := url.Parse(rURL)
	if err != nil {
		return err
	}

	if rulesURL.Scheme != "file" {
		// skip creating
		return nil
	}

	rulesPath := rulesURL.Path

	if !file.DirExists(rulesPath) {
		_ = os.MkdirAll(rulesPath, 0755)
	}

	files, err := ioutil.ReadDir(rulesPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".yaml") || strings.HasSuffix(f.Name(), ".yml") {
			// skip creating
			return nil
		}
	}

	ruleRestYaml, err := rulesConfig.ReadFile("rules/sample.rest.yaml")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path.Join(rulesPath, "sample.rest.yaml"), ruleRestYaml, 0655); err != nil {
		return err
	}

	ruleRestGrpc, err := resourcesConfig.ReadFile("rules/sample.grpc.yaml")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path.Join(rulesPath, "sample.grpc.yaml"), ruleRestGrpc, 0655); err != nil {
		return err
	}

	return nil
}
