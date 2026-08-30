package consent_test

import (
	"testing"

	"github.com/raystack/frontier/core/consent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("accepts a fully configured block", func(t *testing.T) {
		require.NoError(t, enabledConfig().Validate())
	})

	t.Run("accepts a disabled block", func(t *testing.T) {
		require.NoError(t, consent.Config{}.Validate())
	})

	t.Run("does not check a disabled block", func(t *testing.T) {
		// a half-written documents map on a deployment that has not turned
		// consent on yet is not an error; turning it on is what makes it one
		config := consent.Config{
			Documents: map[string]consent.DocumentConfig{
				"terms_of_service": {},
			},
		}

		require.NoError(t, config.Validate())

		config.Enabled = true
		require.Error(t, config.Validate())
	})

	t.Run("rejects enabled with no documents", func(t *testing.T) {
		// silently disabling itself would look identical to a working
		// deployment while asking nobody to accept anything
		err := consent.Config{Enabled: true}.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no documents")
	})

	t.Run("rejects an empty id", func(t *testing.T) {
		config := consent.Config{
			Enabled: true,
			Documents: map[string]consent.DocumentConfig{
				"": {Title: "Terms", Version: "1", URL: "https://example.org/terms"},
			},
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty id")
	})

	t.Run("rejects an empty version", func(t *testing.T) {
		config := enabledConfig()
		config.Documents["terms_of_service"] = consent.DocumentConfig{
			Title: "Terms & Conditions",
			URL:   "https://example.org/legal/terms",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "terms_of_service")
		assert.Contains(t, err.Error(), "empty version")
	})

	t.Run("rejects an empty url", func(t *testing.T) {
		config := enabledConfig()
		config.Documents["privacy_policy"] = consent.DocumentConfig{
			Title:   "Privacy Policy",
			Version: "2026-04-01",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "privacy_policy")
		assert.Contains(t, err.Error(), "empty url")
	})

	t.Run("rejects a url that does not parse", func(t *testing.T) {
		config := enabledConfig()
		config.Documents["eula"] = consent.DocumentConfig{
			Title:   "End User License Agreement",
			Version: "2026-02-14",
			URL:     "://example.org/legal/eula",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "eula")
	})

	t.Run("rejects a url the client cannot link to", func(t *testing.T) {
		// url.Parse accepts a bare path, and a document nobody can open is as
		// useless as one that does not parse at all
		config := enabledConfig()
		config.Documents["eula"] = consent.DocumentConfig{
			Title:   "End User License Agreement",
			Version: "2026-02-14",
			URL:     "example.org/legal/eula",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "absolute url")
	})

	t.Run("rejects an empty title", func(t *testing.T) {
		// the title is the label a client puts on the link, so an untitled
		// document asks the user to accept a bare url
		config := enabledConfig()
		config.Documents["eula"] = consent.DocumentConfig{
			Version: "2026-02-14",
			URL:     "https://example.org/legal/eula",
		}

		err := config.Validate()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "eula")
		assert.Contains(t, err.Error(), "empty title")
	})

	t.Run("names the same document on every run", func(t *testing.T) {
		// map iteration is randomised, so a validator that walks it unsorted
		// reports a different document each boot
		config := consent.Config{
			Enabled: true,
			Documents: map[string]consent.DocumentConfig{
				"aaa_broken": {Version: "1"},
				"zzz_broken": {Version: "1"},
			},
		}

		for i := 0; i < 20; i++ {
			err := config.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "aaa_broken")
		}
	})
}
