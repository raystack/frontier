package consent

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
)

// Repository writes consent records and nothing else: the table is immutable.
// It takes the transaction rather than opening one, because the record has to
// land with the user row or not at all.
type Repository interface {
	Create(ctx context.Context, tx *sqlx.Tx, cnst Consent) (Consent, error)
}

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord models.AuditRecord) (models.AuditRecord, error)
}

// Service owns the document config, so it owns the checks that read it, and the
// repository that writes the records those checks guard. Config is read at boot,
// so a version change needs a restart.
type Service struct {
	logger                *slog.Logger
	config                Config
	repository            Repository
	auditRecordRepository AuditRecordRepository
}

func NewService(logger *slog.Logger, config Config, repository Repository,
	auditRecordRepository AuditRecordRepository) *Service {
	return &Service{
		logger:                logger,
		config:                config,
		repository:            repository,
		auditRecordRepository: auditRecordRepository,
	}
}

// Documents returns every configured document, ordered by id so a response built
// from it is stable. Empty when disabled, so one client build works against both
// kinds of deployment.
func (s Service) Documents() []Document {
	if !s.config.Enabled {
		return nil
	}

	ids := sortedIDs(s.config.Documents)
	documents := make([]Document, 0, len(ids))
	for _, id := range ids {
		documents = append(documents, s.document(id))
	}
	return documents
}

// Resolve maps ids to their config snapshots and rejects unknown ones, saying
// nothing about whether the set is complete — for callers where completeness is
// not yet knowable. Duplicates are removed.
//
// Disabled, it resolves nothing and rejects nothing, so the ids are ignored
// rather than refused.
func (s Service) Resolve(ids []string) ([]Document, error) {
	if !s.config.Enabled {
		return nil, nil
	}

	accepted := uniqueSorted(ids)
	var unknown []string
	for _, id := range accepted {
		if _, ok := s.config.Documents[id]; !ok {
			unknown = append(unknown, id)
		}
	}
	if len(unknown) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrUnknownDocuments, strings.Join(unknown, ", "))
	}

	documents := make([]Document, 0, len(accepted))
	for _, id := range accepted {
		documents = append(documents, s.document(id))
	}
	return documents, nil
}

// ResolveAll is Resolve plus the completeness rule: the ids must cover every
// configured document, no more and no less. Both directions are compared, so the
// error names what is wrong rather than just that something is.
func (s Service) ResolveAll(ids []string) ([]Document, error) {
	if !s.config.Enabled {
		return nil, nil
	}

	// the "no more" direction and the deduplication both come from Resolve
	documents, err := s.Resolve(ids)
	if err != nil {
		return nil, err
	}

	accepted := make(map[string]struct{}, len(documents))
	for _, document := range documents {
		accepted[document.ID] = struct{}{}
	}

	var missing []string
	for _, id := range sortedIDs(s.config.Documents) {
		if _, ok := accepted[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrMissingDocuments, strings.Join(missing, ", "))
	}
	return documents, nil
}

// Grant writes one record for the documents it is given, inside the transaction
// creating the user. No completeness rule here — ResolveAll owns that — which
// leaves room for a later re-consent covering a subset.
func (s Service) Grant(ctx context.Context, tx *sqlx.Tx, req GrantRequest) (Consent, error) {
	if req.Source == "" {
		req.Source = SourceSignup
	}
	if err := validateGrant(req); err != nil {
		return Consent{}, err
	}

	return s.repository.Create(ctx, tx, Consent{
		UserID:       req.UserID,
		UserEmail:    req.UserEmail,
		Documents:    req.Documents,
		Source:       req.Source,
		AuthStrategy: req.AuthStrategy,
		IPAddress:    req.IPAddress,
		ConsentedAt:  req.ConsentedAt,
	})
}

// RecordGranted writes the audit record for a granted consent, after the commit:
// AuditRecordRepository.Create has no transactional variant, so this cannot be
// atomic with the record it describes. The consent record is the source of truth
// and this is a breadcrumb, so a failure here is logged, not returned.
//
// Actor is set explicitly because these endpoints are skip-listed: the repository
// would enrich an empty actor from a context that has none, landing the nil UUID
// and the system actor for an act a person performed.
func (s Service) RecordGranted(ctx context.Context, granted Consent) {
	// the whole snapshot, the same four fields the record itself holds: a reader
	// working from the audit trail alone can then say what the user accepted
	// without joining back to a record they may not be able to read
	documents := make([]map[string]string, 0, len(granted.Documents))
	for _, document := range granted.Documents {
		documents = append(documents, map[string]string{
			"id":      document.ID,
			"title":   document.Title,
			"version": document.Version,
			"url":     document.URL,
		})
	}

	if _, err := s.auditRecordRepository.Create(ctx, models.AuditRecord{
		Event: pkgAuditRecord.UserConsentGrantedEvent,
		// the new user is both actor and resource: nobody else is present
		Actor: models.Actor{
			ID:   granted.UserID,
			Type: schema.UserPrincipal,
			Name: granted.UserEmail,
		},
		Resource: models.Resource{
			ID:   granted.UserID,
			Type: pkgAuditRecord.UserType,
			Name: granted.UserEmail,
		},
		Target: &models.Target{
			ID:   granted.ID,
			Type: pkgAuditRecord.ConsentType,
			Name: granted.Source,
			Metadata: map[string]any{
				"documents": documents,
			},
		},
		// the platform org, not a blank one, so readers need no special case
		OrgID: schema.PlatformOrgID.String(),
		// when the user accepted, not when this row was written.
		OccurredAt: granted.ConsentedAt,
		// IdempotencyKey is left empty: it is nullable, and a consent record is
		// written once, so there is nothing to deduplicate.
	}); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "failed to write the audit record for a granted consent",
			"consent_id", granted.ID, "user_id", granted.UserID, "error", err)
	}
}

func validateGrant(req GrantRequest) error {
	switch {
	case req.UserID == "":
		return fmt.Errorf("%w: no user id", ErrInvalidGrant)
	case req.UserEmail == "":
		return fmt.Errorf("%w: no user email", ErrInvalidGrant)
	case len(req.Documents) == 0:
		return fmt.Errorf("%w: no documents", ErrInvalidGrant)
	case req.ConsentedAt.IsZero():
		return fmt.Errorf("%w: no consented at", ErrInvalidGrant)
	}
	return nil
}

func (s Service) document(id string) Document {
	document := s.config.Documents[id]
	return Document{
		ID:      id,
		Title:   document.Title,
		Version: document.Version,
		URL:     document.URL,
	}
}

// uniqueSorted removes duplicates and orders the ids, so the same accepted set
// produces the same document list whatever order the client sent it in.
func uniqueSorted(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		seen[id] = struct{}{}
	}
	unique := make([]string, 0, len(seen))
	for id := range seen {
		unique = append(unique, id)
	}
	sort.Strings(unique)
	return unique
}
