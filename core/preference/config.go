package preference

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TraitsConfig represents the YAML config file structure for additional traits
type TraitsConfig struct {
	Traits []Trait `yaml:"traits"`
}

// LoadTraitsFromFile loads additional traits from a YAML configuration file
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

	// Merge additional traits with default traits
	// Additional traits with same resource_type and name override defaults
	traits := make([]Trait, 0, len(DefaultTraits)+len(config.Traits))
	traits = append(traits, DefaultTraits...)

	for _, additionalTrait := range config.Traits {
		// Check if additional trait overrides a default trait
		found := false
		for i, defaultTrait := range traits {
			if defaultTrait.ResourceType == additionalTrait.ResourceType && defaultTrait.Name == additionalTrait.Name {
				traits[i] = additionalTrait
				found = true
				break
			}
		}
		if !found {
			traits = append(traits, additionalTrait)
		}
	}

	return traits, nil
}
