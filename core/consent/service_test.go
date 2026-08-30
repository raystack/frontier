package consent_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
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

func TestService_Documents(t *testing.T) {
	t.Run("returns every configured document ordered by id", func(t *testing.T) {
		documents := consent.NewService(nil, enabledConfig(), nil, nil).Documents()

		assert.Equal(t, allDocuments(), documents)
	})

	t.Run("orders by id whatever order config was read in", func(t *testing.T) {
		// map iteration is randomised, so run it enough times that an
		// unsorted implementation cannot pass by luck
		service := consent.NewService(nil, enabledConfig(), nil, nil)
		for i := 0; i < 20; i++ {
			assert.Equal(t, []string{"eula", "privacy_policy", "terms_of_service"}, ids(service.Documents()))
		}
	})

	t.Run("returns empty when disabled, even with documents configured", func(t *testing.T) {
		config := enabledConfig()
		config.Enabled = false

		assert.Empty(t, consent.NewService(nil, config, nil, nil).Documents())
	})

	t.Run("returns empty for the zero config", func(t *testing.T) {
		assert.Empty(t, consent.NewService(nil, consent.Config{}, nil, nil).Documents())
	})
}

func TestService_Resolve(t *testing.T) {
	service := consent.NewService(nil, enabledConfig(), nil, nil)

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

		documents, err := consent.NewService(nil, config, nil, nil).Resolve([]string{"anything_at_all"})

		require.NoError(t, err)
		assert.Empty(t, documents)
	})
}

func TestService_ResolveAll(t *testing.T) {
	service := consent.NewService(nil, enabledConfig(), nil, nil)

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
		service := consent.NewService(nil, config, nil, nil)

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

// fakeRepository stands in for the postgres repository. core/consent has no
// generated mocks, and the write is one method, so a fake here reads better
// than a mock and keeps the transaction out of the picture: what Grant hands
// the repository is the whole question.
type fakeRepository struct {
	created []consent.Consent
	id      string
	err     error
}

func (f *fakeRepository) Create(_ context.Context, _ *sqlx.Tx, cnst consent.Consent) (consent.Consent, error) {
	if f.err != nil {
		return consent.Consent{}, f.err
	}
	f.created = append(f.created, cnst)
	cnst.ID = f.id
	cnst.CreatedAt = time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	return cnst, nil
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

func TestService_Grant(t *testing.T) {
	t.Run("writes one record carrying everything that describes the act", func(t *testing.T) {
		repo := &fakeRepository{id: "consent-id"}
		service := consent.NewService(nil, enabledConfig(), repo, nil)

		req := grantRequest()
		granted, err := service.Grant(context.Background(), nil, req)
		require.NoError(t, err)

		require.Len(t, repo.created, 1)
		written := repo.created[0]
		assert.Equal(t, req.UserID, written.UserID)
		assert.Equal(t, req.UserEmail, written.UserEmail)
		assert.Equal(t, allDocuments(), written.Documents)
		assert.Equal(t, consent.SourceSignup, written.Source)
		assert.Equal(t, "mailotp", written.AuthStrategy)
		assert.Equal(t, "203.0.113.9", written.IPAddress)
		assert.Equal(t, req.ConsentedAt, written.ConsentedAt)

		assert.Equal(t, "consent-id", granted.ID)
	})

	t.Run("has no completeness rule of its own", func(t *testing.T) {
		// one document out of the three configured. ResolveAll is what decides
		// a signup covers everything; keeping that out of Grant is what leaves
		// room for a re-consent covering a subset without a second write path.
		repo := &fakeRepository{id: "consent-id"}
		service := consent.NewService(nil, enabledConfig(), repo, nil)

		req := grantRequest()
		req.Documents = allDocuments()[:1]
		_, err := service.Grant(context.Background(), nil, req)
		require.NoError(t, err)

		require.Len(t, repo.created, 1)
		assert.Equal(t, []string{"eula"}, ids(repo.created[0].Documents))
	})

	t.Run("defaults the source to signup", func(t *testing.T) {
		repo := &fakeRepository{id: "consent-id"}
		service := consent.NewService(nil, enabledConfig(), repo, nil)

		req := grantRequest()
		req.Source = ""
		_, err := service.Grant(context.Background(), nil, req)
		require.NoError(t, err)

		require.Len(t, repo.created, 1)
		assert.Equal(t, consent.SourceSignup, repo.created[0].Source)
	})

	t.Run("rejects a request the record cannot be written from", func(t *testing.T) {
		cases := map[string]func(*consent.GrantRequest){
			"no user id":      func(r *consent.GrantRequest) { r.UserID = "" },
			"no user email":   func(r *consent.GrantRequest) { r.UserEmail = "" },
			"no documents":    func(r *consent.GrantRequest) { r.Documents = nil },
			"no consented at": func(r *consent.GrantRequest) { r.ConsentedAt = time.Time{} },
		}
		for name, break_ := range cases {
			t.Run(name, func(t *testing.T) {
				repo := &fakeRepository{id: "consent-id"}
				service := consent.NewService(nil, enabledConfig(), repo, nil)

				req := grantRequest()
				break_(&req)
				_, err := service.Grant(context.Background(), nil, req)
				assert.ErrorIs(t, err, consent.ErrInvalidGrant)
				// nothing reached the repository, so the surrounding
				// transaction has nothing to roll back
				assert.Empty(t, repo.created)
			})
		}
	})

	t.Run("surfaces a repository failure so the transaction rolls back", func(t *testing.T) {
		repo := &fakeRepository{err: consent.ErrConsentExists}
		service := consent.NewService(nil, enabledConfig(), repo, nil)

		_, err := service.Grant(context.Background(), nil, grantRequest())
		assert.ErrorIs(t, err, consent.ErrConsentExists)
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
		service := consent.NewService(nil, enabledConfig(), nil, auditRepo)

		service.RecordGranted(context.Background(), granted)

		require.Len(t, auditRepo.created, 1)
		record := auditRepo.created[0]
		assert.Equal(t, pkgAuditRecord.UserConsentGrantedEvent, record.Event)

		// the actor cannot be left to enrichment: these endpoints are on the
		// authentication skip list, so the context holds no actor and the
		// record would land as the system actor for an act a person performed
		assert.Equal(t, granted.UserID, record.Actor.ID)
		assert.Equal(t, schema.UserPrincipal, record.Actor.Type)
		assert.Equal(t, granted.UserEmail, record.Actor.Name)

		assert.Equal(t, granted.UserID, record.Resource.ID)
		assert.Equal(t, pkgAuditRecord.UserType, record.Resource.Type)

		require.NotNil(t, record.Target)
		assert.Equal(t, granted.ID, record.Target.ID)
		assert.Equal(t, pkgAuditRecord.ConsentType, record.Target.Type)
		assert.Equal(t, consent.SourceSignup, record.Target.Name)
		// the whole snapshot, not just the id and the version: the audit trail
		// has to say what the user accepted without reading the record back
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

		// when the user accepted, not when the row was written
		assert.Equal(t, granted.ConsentedAt, record.OccurredAt)
		assert.Equal(t, schema.PlatformOrgID.String(), record.OrgID)
		// nullable, and a consent record is written once
		assert.Empty(t, record.IdempotencyKey)
	})

	t.Run("carries on when the audit write fails", func(t *testing.T) {
		// the audit record cannot be atomic with the consent record, which is
		// why that record is the source of truth and this one is a breadcrumb
		auditRepo := &fakeAuditRecordRepository{err: errors.New("audit is down")}
		service := consent.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)),
			enabledConfig(), nil, auditRepo)

		assert.NotPanics(t, func() {
			service.RecordGranted(context.Background(), granted)
		})
		assert.Len(t, auditRepo.created, 1)
	})
}
