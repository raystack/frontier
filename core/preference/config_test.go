package preference

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTraitsFromFile(t *testing.T) {
	t.Run("empty path returns default traits", func(t *testing.T) {
		traits, err := LoadTraitsFromFile("")
		require.NoError(t, err)
		assert.Equal(t, DefaultTraits, traits)
	})

	t.Run("non-existent file returns error", func(t *testing.T) {
		_, err := LoadTraitsFromFile("/non/existent/path.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read traits file")
	})

	t.Run("valid custom traits file merges with defaults", func(t *testing.T) {
		content := `
traits:
  - resource_type: "app/user"
    name: "custom_trait"
    title: "Custom Trait"
    description: "A custom trait"
    input: "text"
    default: "custom_default"
`
		tmpFile := filepath.Join(t.TempDir(), "traits.yaml")
		err := os.WriteFile(tmpFile, []byte(content), 0644)
		require.NoError(t, err)

		traits, err := LoadTraitsFromFile(tmpFile)
		require.NoError(t, err)

		// Should have all default traits plus the custom one
		assert.Len(t, traits, len(DefaultTraits)+1)

		// Find the custom trait
		var customTrait *Trait
		for i := range traits {
			if traits[i].Name == "custom_trait" {
				customTrait = &traits[i]
				break
			}
		}
		require.NotNil(t, customTrait)
		assert.Equal(t, "Custom Trait", customTrait.Title)
		assert.Equal(t, "custom_default", customTrait.Default)
	})

	t.Run("custom trait overrides default trait", func(t *testing.T) {
		content := `
traits:
  - resource_type: "app/platform"
    name: "disable_orgs_on_create"
    title: "Custom Title"
    description: "Custom description"
    input: "checkbox"
    input_hints: "true,false"
    default: "true"
`
		tmpFile := filepath.Join(t.TempDir(), "traits.yaml")
		err := os.WriteFile(tmpFile, []byte(content), 0644)
		require.NoError(t, err)

		traits, err := LoadTraitsFromFile(tmpFile)
		require.NoError(t, err)

		// Should have same number as defaults (override, not add)
		assert.Len(t, traits, len(DefaultTraits))

		// Find the overridden trait
		var overriddenTrait *Trait
		for i := range traits {
			if traits[i].Name == "disable_orgs_on_create" {
				overriddenTrait = &traits[i]
				break
			}
		}
		require.NotNil(t, overriddenTrait)
		assert.Equal(t, "Custom Title", overriddenTrait.Title)
		assert.Equal(t, "true", overriddenTrait.Default)
	})

	t.Run("invalid yaml returns error", func(t *testing.T) {
		content := `invalid: yaml: content: [`
		tmpFile := filepath.Join(t.TempDir(), "traits.yaml")
		err := os.WriteFile(tmpFile, []byte(content), 0644)
		require.NoError(t, err)

		_, err = LoadTraitsFromFile(tmpFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse traits file")
	})
}
