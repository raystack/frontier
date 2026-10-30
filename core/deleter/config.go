package deleter

import (
	"fmt"
	"maps"
	"slices"

	"github.com/raystack/frontier/internal/bootstrap/schema"
)

// Parent is what is being deleted.
type Parent string

const (
	ParentProject      Parent = "project"
	ParentOrganization Parent = "organization"
)

type Rule string

const (
	RuleBlock Rule = "block"
	// RuleBlockUnlessDeletable blocks until the resource metadata has
	// deletable set to true.
	RuleBlockUnlessDeletable Rule = "block_unless_deletable"
	RuleAllow                Rule = "allow"
)

type Rules map[Parent]Rule

type Config struct {
	Enabled   bool             `yaml:"enabled" mapstructure:"enabled" default:"false"`
	Resources map[string]Rules `yaml:"resources" mapstructure:"resources"`
	// sent whether or not Enabled is set
	UserDeleteNotice UserDeleteNoticeConfig `yaml:"user_delete_notice" mapstructure:"user_delete_notice"`
}

type UserDeleteNoticeConfig struct {
	Subject string `yaml:"subject" mapstructure:"subject"`
	Body    string `yaml:"body" mapstructure:"body"`
}

func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if len(c.Resources) == 0 {
		return fmt.Errorf("app.deletion is enabled but lists no resources")
	}
	// sorted for a stable error
	for _, namespace := range slices.Sorted(maps.Keys(c.Resources)) {
		if err := validateResource(namespace, c.Resources[namespace]); err != nil {
			return err
		}
	}
	return nil
}

func validateResource(namespace string, rules Rules) error {
	if !schema.IsValidPermissionNamespace(namespace) {
		return fmt.Errorf("app.deletion resource %q must look like service/name, for example compute/machine", namespace)
	}
	if len(rules) == 0 {
		return fmt.Errorf("app.deletion resource %q lists no rules", namespace)
	}
	for _, parent := range slices.Sorted(maps.Keys(rules)) {
		if parent != ParentProject && parent != ParentOrganization {
			return fmt.Errorf("app.deletion resource %q has unknown key %q: use %s or %s",
				namespace, parent, ParentProject, ParentOrganization)
		}
		if rule := rules[parent]; rule != RuleBlock && rule != RuleBlockUnlessDeletable && rule != RuleAllow {
			return fmt.Errorf("app.deletion resource %q has unknown %s rule %q: use %s, %s or %s",
				namespace, parent, rule, RuleBlock, RuleBlockUnlessDeletable, RuleAllow)
		}
	}
	return nil
}
