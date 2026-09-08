package consent_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/consent"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
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

func TestService_Enabled(t *testing.T) {
	t.Run("reports on when documents are configured", func(t *testing.T) {
		assert.True(t, consent.NewService(nil, enabledConfig(), nil).Enabled())
	})

	t.Run("reports off when the flag is off, even with documents configured", func(t *testing.T) {
		config := enabledConfig()
		config.Enabled = false

		assert.False(t, consent.NewService(nil, config, nil).Enabled())
	})

	t.Run("reports off when the flag is on but nothing is configured", func(t *testing.T) {
		// boot validation rejects this, so it is only reachable by a caller that
		// skipped Validate: a deployment that asks for nothing
		assert.False(t, consent.NewService(nil, consent.Config{Enabled: true}, nil).Enabled())
	})

	t.Run("reports off for the zero config", func(t *testing.T) {
		assert.False(t, consent.NewService(nil, consent.Config{}, nil).Enabled())
	})
}

func TestService_Documents(t *testing.T) {
	t.Run("returns every configured document ordered by id", func(t *testing.T) {
		documents := consent.NewService(nil, enabledConfig(), nil).Documents()

		assert.Equal(t, allDocuments(), documents)
	})

	t.Run("orders by id whatever order config was read in", func(t *testing.T) {
		// map iteration is randomised, so run it enough times that an
		// unsorted implementation cannot pass by luck
		service := consent.NewService(nil, enabledConfig(), nil)
		for i := 0; i < 20; i++ {
			assert.Equal(t, []string{"eula", "privacy_policy", "terms_of_service"}, ids(service.Documents()))
		}
	})

	t.Run("returns empty when disabled, even with documents configured", func(t *testing.T) {
		config := enabledConfig()
		config.Enabled = false

		assert.Empty(t, consent.NewService(nil, config, nil).Documents())
	})

	t.Run("returns empty for the zero config", func(t *testing.T) {
		assert.Empty(t, consent.NewService(nil, consent.Config{}, nil).Documents())
	})
}

func TestService_Resolve(t *testing.T) {
	service := consent.NewService(nil, enabledConfig(), nil)

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

		documents, err := consent.NewService(nil, config, nil).Resolve([]string{"anything_at_all"})

		require.NoError(t, err)
		assert.Empty(t, documents)
	})
}

func TestService_ResolveAll(t *testing.T) {
	service := consent.NewService(nil, enabledConfig(), nil)

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
		service := consent.NewService(nil, config, nil)

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

type fakeAuditRecordRepository struct {
	created []models.AuditRecord
	err     error
}

func (f *fakeAuditRecordRepository) Create(_ context.Context, record models.AuditRecord) (models.AuditRecord, error) {
	f.created = append(f.created, record)
	return record, f.err
}

func grantRequest() consent.GrantRequest {
	return consent.GrantRequest{
		UserID:       "8814cdf1-0000-0000-0000-000000000001",
		UserEmail:    "new@example.com",
		Documents:    allDocuments(),
		Source:       consent.SourceSignup,
		AuthStrategy: "mailotp",
		IPAddress:    "203.0.113.9",
		ConsentedAt:  time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC),
	}
}

func TestService_PrepareGrant(t *testing.T) {
	t.Run("carries everything that describes the act onto the record", func(t *testing.T) {
		service := consent.NewService(nil, enabledConfig(), nil)

		req := grantRequest()
		prepared, err := service.PrepareGrant(req)
		require.NoError(t, err)

		assert.Equal(t, req.UserID, prepared.UserID)
		assert.Equal(t, req.UserEmail, prepared.UserEmail)
		assert.Equal(t, allDocuments(), prepared.Documents)
		assert.Equal(t, consent.SourceSignup, prepared.Source)
		assert.Equal(t, "mailotp", prepared.AuthStrategy)
		assert.Equal(t, "203.0.113.9", prepared.IPAddress)
		assert.Equal(t, req.ConsentedAt, prepared.ConsentedAt)
	})

	t.Run("has no completeness rule of its own", func(t *testing.T) {
		// one document out of the three configured: ResolveAll owns completeness
		service := consent.NewService(nil, enabledConfig(), nil)

		req := grantRequest()
		req.Documents = allDocuments()[:1]
		prepared, err := service.PrepareGrant(req)
		require.NoError(t, err)

		assert.Equal(t, []string{"eula"}, ids(prepared.Documents))
	})

	t.Run("defaults the source to signup", func(t *testing.T) {
		service := consent.NewService(nil, enabledConfig(), nil)

		req := grantRequest()
		req.Source = ""
		prepared, err := service.PrepareGrant(req)
		require.NoError(t, err)

		assert.Equal(t, consent.SourceSignup, prepared.Source)
	})

	t.Run("rejects a request the record cannot be written from", func(t *testing.T) {
		cases := map[string]func(*consent.GrantRequest){
			"no documents":    func(r *consent.GrantRequest) { r.Documents = nil },
			"no consented at": func(r *consent.GrantRequest) { r.ConsentedAt = time.Time{} },
		}
		for name, break_ := range cases {
			t.Run(name, func(t *testing.T) {
				service := consent.NewService(nil, enabledConfig(), nil)

				req := grantRequest()
				break_(&req)
				_, err := service.PrepareGrant(req)
				assert.ErrorIs(t, err, consent.ErrInvalidGrant)
			})
		}
	})

	t.Run("does not require the identity, which the writer stamps", func(t *testing.T) {
		service := consent.NewService(nil, enabledConfig(), nil)

		req := grantRequest()
		req.UserID, req.UserEmail = "", ""
		prepared, err := service.PrepareGrant(req)
		require.NoError(t, err)
		assert.Empty(t, prepared.UserID)
	})
}

func TestService_RecordGranted(t *testing.T) {
	granted := consent.Consent{
		ID:           "consent-id",
		UserID:       "8814cdf1-0000-0000-0000-000000000001",
		UserEmail:    "new@example.com",
		Documents:    allDocuments(),
		Source:       consent.SourceSignup,
		AuthStrategy: "mailotp",
		ConsentedAt:  time.Date(2026, 8, 30, 10, 0, 0, 0, time.UTC),
	}

	t.Run("sets every field explicitly, the actor included", func(t *testing.T) {
		auditRepo := &fakeAuditRecordRepository{}
		service := consent.NewService(nil, enabledConfig(), auditRepo)

		service.RecordGranted(context.Background(), granted)

		require.Len(t, auditRepo.created, 1)
		record := auditRepo.created[0]
		assert.Equal(t, pkgAuditRecord.UserConsentGrantedEvent, record.Event)

		assert.Equal(t, granted.UserID, record.Actor.ID)
		assert.Equal(t, schema.UserPrincipal, record.Actor.Type)
		assert.Equal(t, granted.UserEmail, record.Actor.Name)

		assert.Equal(t, granted.UserID, record.Resource.ID)
		assert.Equal(t, pkgAuditRecord.UserType, record.Resource.Type)

		require.NotNil(t, record.Target)
		assert.Equal(t, granted.ID, record.Target.ID)
		assert.Equal(t, pkgAuditRecord.ConsentType, record.Target.Type)
		assert.Equal(t, consent.SourceSignup, record.Target.Name)
		// the whole snapshot: the audit trail has to say what the user accepted
		assert.Equal(t, []map[string]string{
			{
				"id":      "eula",
				"title":   "End User License Agreement",
				"version": "2026-02-14",
				"url":     "https://example.org/legal/eula/2026-02-14",
			},
			{
				"id":      "privacy_policy",
				"title":   "Privacy Policy",
				"version": "2026-04-01",
				"url":     "https://example.org/legal/privacy/2026-04-01",
			},
			{
				"id":      "terms_of_service",
				"title":   "Terms & Conditions",
				"version": "2026-04-01",
				"url":     "https://example.org/legal/terms/2026-04-01",
			},
		}, record.Target.Metadata["documents"])

		assert.Equal(t, granted.ConsentedAt, record.OccurredAt)
		assert.Equal(t, schema.PlatformOrgID.String(), record.OrgID)
		assert.Empty(t, record.IdempotencyKey)
	})

	t.Run("carries on when the audit write fails", func(t *testing.T) {
		auditRepo := &fakeAuditRecordRepository{err: errors.New("audit is down")}
		service := consent.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)),
			enabledConfig(), auditRepo)

		assert.NotPanics(t, func() {
			service.RecordGranted(context.Background(), granted)
		})
		assert.Len(t, auditRepo.created, 1)
	})
}
