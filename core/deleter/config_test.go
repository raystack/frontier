package deleter_test

import (
	"fmt"
	"testing"

	"github.com/raystack/frontier/core/deleter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("accepts every rule", func(t *testing.T) {
		require.NoError(t, enabledDeletionConfig().Validate())
	})

	t.Run("accepts a disabled block", func(t *testing.T) {
		require.NoError(t, deleter.Config{}.Validate())
	})

	t.Run("does not check a disabled block", func(t *testing.T) {
		config := deleter.Config{
			Resources: map[string]deleter.Rules{
				"compute/machine": {deleter.ParentProject: "blok"},
			},
		}

		require.NoError(t, config.Validate())

		config.Enabled = true
		require.Error(t, config.Validate())
	})

	t.Run("rejects enabled with no resources", func(t *testing.T) {
		err := deleter.Config{Enabled: true}.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "lists no resources")
	})

	t.Run("rejects a resource with no rules", func(t *testing.T) {
		for _, rules := range []deleter.Rules{nil, {}} {
			config := enabledDeletionConfig()
			config.Resources["compute/disk"] = rules

			err := config.Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), `"compute/disk" lists no rules`)
		}
	})

	for _, parent := range []deleter.Parent{"projct", "org", "workspace"} {
		t.Run(fmt.Sprintf("rejects key %q", parent), func(t *testing.T) {
			config := enabledDeletionConfig()
			config.Resources["compute/disk"] = deleter.Rules{parent: deleter.RuleBlock}

			err := config.Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), `"compute/disk" has unknown key "`+string(parent)+`"`)
		})
	}

	for _, rule := range []deleter.Rule{"blok", "", "BLOCK", "block_unless_finished"} {
		t.Run(fmt.Sprintf("rejects rule %q", rule), func(t *testing.T) {
			config := enabledDeletionConfig()
			config.Resources["compute/order"][deleter.ParentOrganization] = rule

			err := config.Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), `"compute/order" has unknown organization rule "`+string(rule)+`"`)
		})
	}

	for _, namespace := range []string{"machine", "compute/", "compute/machine/disk", "compute_machine"} {
		t.Run(fmt.Sprintf("rejects namespace %q", namespace), func(t *testing.T) {
			config := enabledDeletionConfig()
			config.Resources[namespace] = deleter.Rules{deleter.ParentProject: deleter.RuleBlock}

			err := config.Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), `"`+namespace+`" must look like service/name`)
		})
	}
}

func enabledDeletionConfig() deleter.Config {
	return deleter.Config{
		Enabled: true,
		Resources: map[string]deleter.Rules{
			"compute/machine": {
				deleter.ParentProject:      deleter.RuleBlock,
				deleter.ParentOrganization: deleter.RuleAllow,
			},
			"compute/order": {
				deleter.ParentProject:      deleter.RuleBlockUnlessDeletable,
				deleter.ParentOrganization: deleter.RuleBlockUnlessDeletable,
			},
			"storage/file": {
				deleter.ParentOrganization: deleter.RuleAllow,
			},
		},
	}
}
