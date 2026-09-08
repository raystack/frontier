package consent

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
)

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord models.AuditRecord) (models.AuditRecord, error)
}

type Service struct {
	logger                *slog.Logger
	config                Config
	auditRecordRepository AuditRecordRepository
}

func NewService(logger *slog.Logger, config Config,
	auditRecordRepository AuditRecordRepository) *Service {
	return &Service{
		logger:                logger,
		config:                config,
		auditRecordRepository: auditRecordRepository,
	}
}

// Enabled reports whether this deployment asks for consent at all
func (s Service) Enabled() bool {
	return s.config.Enabled && len(s.config.Documents) > 0
}

// Documents returns every configured document, ordered by id
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

// Resolve maps ids to their config snapshots and rejects unknown ones
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
// configured document, no more and no less
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

func (s Service) PrepareGrant(req GrantRequest) (Consent, error) {
	if req.Source == "" {
		req.Source = SourceSignup
	}
	if err := validateGrant(req); err != nil {
		return Consent{}, err
	}

	return Consent{
		UserID:       req.UserID,
		UserEmail:    req.UserEmail,
		Documents:    req.Documents,
		Source:       req.Source,
		AuthStrategy: req.AuthStrategy,
		IPAddress:    req.IPAddress,
		ConsentedAt:  req.ConsentedAt,
	}, nil
}

// RecordGranted writes the audit record after the commit
func (s Service) RecordGranted(ctx context.Context, granted Consent) {
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
		// set explicitly: these endpoints are skip-listed, so the context holds no
		// actor and enrichment would land the system actor
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
		// the platform org, so readers need no special case
		OrgID:      schema.PlatformOrgID.String(),
		OccurredAt: granted.ConsentedAt,
		// IdempotencyKey stays empty: a consent record is written once.
	}); err != nil && s.logger != nil {
		s.logger.ErrorContext(ctx, "failed to write the audit record for a granted consent",
			"consent_id", granted.ID, "user_id", granted.UserID, "error", err)
	}
}

func validateGrant(req GrantRequest) error {
	switch {
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
