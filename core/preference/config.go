package preference

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TraitsConfig represents the YAML config file structure for custom traits
type TraitsConfig struct {
	Traits []Trait `yaml:"traits"`
}

// LoadTraitsFromFile loads custom traits from a YAML configuration file
// and merges them with DefaultTraits
func LoadTraitsFromFile(path string) ([]Trait, error) {
	if path == "" {
		return DefaultTraits, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read traits file: %w", err)
	}

	var config TraitsConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse traits file: %w", err)
	}

	// Merge custom traits with default traits
	// Custom traits with same resource_type and name override defaults
	traits := make([]Trait, 0, len(DefaultTraits)+len(config.Traits))
	traits = append(traits, DefaultTraits...)

	for _, customTrait := range config.Traits {
		// Check if custom trait overrides a default trait
		found := false
		for i, defaultTrait := range traits {
			if defaultTrait.ResourceType == customTrait.ResourceType && defaultTrait.Name == customTrait.Name {
				traits[i] = customTrait
				found = true
				break
			}
		}
		if !found {
			traits = append(traits, customTrait)
		}
	}

	return traits, nil
}
