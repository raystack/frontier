package consent_test

import (
	"testing"

	"github.com/raystack/frontier/core/consent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// enabledConfig is the three-document set a deployment configures: terms,
// privacy policy and EULA. The ids are deliberately not in alphabetical order
// here, so the ordering the service applies is visible in every assertion.
func enabledConfig() consent.Config {
	return consent.Config{
		Enabled: true,
		Documents: map[string]consent.DocumentConfig{
			"terms_of_service": {
				Title:   "Terms & Conditions",
				Version: "2026-04-01",
				URL:     "https://example.org/legal/terms/2026-04-01",
			},
			"privacy_policy": {
				Title:   "Privacy Policy",
				Version: "2026-04-01",
				URL:     "https://example.org/legal/privacy/2026-04-01",
			},
			"eula": {
				Title:   "End User License Agreement",
				Version: "2026-02-14",
				URL:     "https://example.org/legal/eula/2026-02-14",
			},
		},
	}
}

func allDocuments() []consent.Document {
	return []consent.Document{
		{
			ID:      "eula",
			Title:   "End User License Agreement",
			Version: "2026-02-14",
			URL:     "https://example.org/legal/eula/2026-02-14",
		},
		{
			ID:      "privacy_policy",
			Title:   "Privacy Policy",
			Version: "2026-04-01",
			URL:     "https://example.org/legal/privacy/2026-04-01",
		},
		{
			ID:      "terms_of_service",
			Title:   "Terms & Conditions",
			Version: "2026-04-01",
			URL:     "https://example.org/legal/terms/2026-04-01",
		},
	}
}

func TestService_Documents(t *testing.T) {
	t.Run("returns every configured document ordered by id", func(t *testing.T) {
		documents := consent.NewService(enabledConfig()).Documents()

		assert.Equal(t, allDocuments(), documents)
	})

	t.Run("orders by id whatever order config was read in", func(t *testing.T) {
		// map iteration is randomised, so run it enough times that an
		// unsorted implementation cannot pass by luck
		service := consent.NewService(enabledConfig())
		for i := 0; i < 20; i++ {
			assert.Equal(t, []string{"eula", "privacy_policy", "terms_of_service"}, ids(service.Documents()))
		}
	})

	t.Run("returns empty when disabled, even with documents configured", func(t *testing.T) {
		config := enabledConfig()
		config.Enabled = false

		assert.Empty(t, consent.NewService(config).Documents())
	})

	t.Run("returns empty for the zero config", func(t *testing.T) {
		assert.Empty(t, consent.NewService(consent.Config{}).Documents())
	})
}

func TestService_Resolve(t *testing.T) {
	service := consent.NewService(enabledConfig())

	t.Run("maps known ids to their config snapshots", func(t *testing.T) {
		documents, err := service.Resolve([]string{"terms_of_service"})

		require.NoError(t, err)
		assert.Equal(t, []consent.Document{{
			ID:      "terms_of_service",
			Title:   "Terms & Conditions",
			Version: "2026-04-01",
			URL:     "https://example.org/legal/terms/2026-04-01",
		}}, documents)
	})

	t.Run("says nothing about completeness", func(t *testing.T) {
		// two of three, which ResolveAll would reject
		documents, err := service.Resolve([]string{"privacy_policy", "eula"})

		require.NoError(t, err)
		assert.Equal(t, []string{"eula", "privacy_policy"}, ids(documents))
	})

	t.Run("orders the result by id", func(t *testing.T) {
		documents, err := service.Resolve([]string{"terms_of_service", "eula", "privacy_policy"})

		require.NoError(t, err)
		assert.Equal(t, allDocuments(), documents)
	})

	t.Run("removes duplicates", func(t *testing.T) {
		documents, err := service.Resolve([]string{"eula", "eula", "eula"})

		require.NoError(t, err)
		assert.Equal(t, []string{"eula"}, ids(documents))
	})

	t.Run("rejects an unknown id and names it", func(t *testing.T) {
		_, err := service.Resolve([]string{"terms_of_service", "cookie_policy"})

		require.ErrorIs(t, err, consent.ErrUnknownDocuments)
		assert.Contains(t, err.Error(), "cookie_policy")
		assert.NotContains(t, err.Error(), "terms_of_service")
	})

	t.Run("names every unknown id", func(t *testing.T) {
		_, err := service.Resolve([]string{"cookie_policy", "acceptable_use"})

		require.ErrorIs(t, err, consent.ErrUnknownDocuments)
		assert.Contains(t, err.Error(), "cookie_policy")
		assert.Contains(t, err.Error(), "acceptable_use")
	})

	t.Run("resolves nothing for no ids", func(t *testing.T) {
		documents, err := service.Resolve(nil)

		require.NoError(t, err)
		assert.Empty(t, documents)
	})

	t.Run("ignores ids when disabled rather than rejecting them", func(t *testing.T) {
		config := enabledConfig()
		config.Enabled = false

		documents, err := consent.NewService(config).Resolve([]string{"anything_at_all"})

		require.NoError(t, err)
		assert.Empty(t, documents)
	})
}

func TestService_ResolveAll(t *testing.T) {
	service := consent.NewService(enabledConfig())

	t.Run("accepts a set covering every configured document", func(t *testing.T) {
		documents, err := service.ResolveAll([]string{"privacy_policy", "terms_of_service", "eula"})

		require.NoError(t, err)
		assert.Equal(t, allDocuments(), documents)
	})

	t.Run("removes duplicates before the check", func(t *testing.T) {
		documents, err := service.ResolveAll([]string{"eula", "privacy_policy", "eula", "terms_of_service", "eula"})

		require.NoError(t, err)
		assert.Equal(t, allDocuments(), documents)
	})

	t.Run("rejects an incomplete set and names what is missing", func(t *testing.T) {
		_, err := service.ResolveAll([]string{"terms_of_service"})

		require.ErrorIs(t, err, consent.ErrMissingDocuments)
		assert.Contains(t, err.Error(), "eula")
		assert.Contains(t, err.Error(), "privacy_policy")
		assert.NotContains(t, err.Error(), "terms_of_service")
	})

	t.Run("rejects an empty set when documents are configured", func(t *testing.T) {
		_, err := service.ResolveAll(nil)

		require.ErrorIs(t, err, consent.ErrMissingDocuments)
	})

	t.Run("rejects an id config does not know and names it", func(t *testing.T) {
		_, err := service.ResolveAll([]string{"terms_of_service", "privacy_policy", "eula", "cookie_policy"})

		require.ErrorIs(t, err, consent.ErrUnknownDocuments)
		assert.Contains(t, err.Error(), "cookie_policy")
	})

	t.Run("reports the unknown id when the set is both incomplete and unknown", func(t *testing.T) {
		// an id config does not know is a client or config mismatch, and it is
		// the more diagnostic of the two, so it is reported first
		_, err := service.ResolveAll([]string{"cookie_policy"})

		require.ErrorIs(t, err, consent.ErrUnknownDocuments)
		assert.NotErrorIs(t, err, consent.ErrMissingDocuments)
	})

	t.Run("ignores ids when disabled rather than rejecting them", func(t *testing.T) {
		config := enabledConfig()
		config.Enabled = false
		service := consent.NewService(config)

		documents, err := service.ResolveAll(nil)
		require.NoError(t, err)
		assert.Empty(t, documents)

		documents, err = service.ResolveAll([]string{"anything_at_all"})
		require.NoError(t, err)
		assert.Empty(t, documents)
	})
}

func ids(documents []consent.Document) []string {
	out := make([]string, 0, len(documents))
	for _, document := range documents {
		out = append(out, document.ID)
	}
	return out
}
