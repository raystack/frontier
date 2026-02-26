package userpat

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"time"

	"github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/organization"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/salt/log"
	"golang.org/x/crypto/sha3"
)

type OrganizationService interface {
	GetRaw(ctx context.Context, id string) (organization.Organization, error)
}

type AuditRecordRepository interface {
	Create(ctx context.Context, auditRecord models.AuditRecord) (models.AuditRecord, error)
}

type Service struct {
	repo                  Repository
	config                Config
	logger                log.Logger
	orgService            OrganizationService
	auditRecordRepository AuditRecordRepository
}

func NewService(logger log.Logger, repo Repository, config Config, orgService OrganizationService, auditRecordRepository AuditRecordRepository) *Service {
	return &Service{
		repo:                  repo,
		config:                config,
		logger:                logger,
		orgService:            orgService,
		auditRecordRepository: auditRecordRepository,
	}
}

type CreateRequest struct {
	UserID     string
	OrgID      string
	Title      string
	Roles      []string
	ProjectIDs []string
	ExpiresAt  time.Time
	Metadata   map[string]any
}

// ValidateExpiry checks that the given expiry time is in the future and within
// the configured maximum PAT lifetime.
func (s *Service) ValidateExpiry(expiresAt time.Time) error {
	if !expiresAt.After(time.Now()) {
		return ErrExpiryInPast
	}
	if expiresAt.After(time.Now().Add(s.config.MaxExpiry())) {
		return ErrExpiryExceeded
	}
	return nil
}

// Create generates a new PAT and returns it with the plaintext value.
// The plaintext value is only available at creation time.
func (s *Service) Create(ctx context.Context, req CreateRequest) (PAT, string, error) {
	if !s.config.Enabled {
		return PAT{}, "", ErrDisabled
	}

	// NOTE: CountActive + Create is not atomic (TOCTOU race). Two concurrent requests
	// could both read count=49 (assuming max limit is 50), pass the check, and create PATs exceeding the limit.
	// Acceptable for now given low concurrency on this endpoint. If this becomes an issue,
	// use an atomic INSERT ... SELECT with a count subquery in the WHERE clause.
	count, err := s.repo.CountActive(ctx, req.UserID, req.OrgID)
	if err != nil {
		return PAT{}, "", fmt.Errorf("counting active PATs: %w", err)
	}
	if count >= s.config.MaxPerUserPerOrg {
		return PAT{}, "", ErrLimitExceeded
	}

	patValue, secretHash, err := s.generatePAT()
	if err != nil {
		return PAT{}, "", err
	}

	pat := PAT{
		UserID:     req.UserID,
		OrgID:      req.OrgID,
		Title:      req.Title,
		SecretHash: secretHash,
		Metadata:   req.Metadata,
		ExpiresAt:  req.ExpiresAt,
	}

	created, err := s.repo.Create(ctx, pat)
	if err != nil {
		return PAT{}, "", err
	}

	// TODO: create policies for roles + project_ids

	// TODO: move audit record creation into the same transaction as PAT creation to avoid partial state where PAT exists but audit record doesn't.
	if err := s.createAuditRecord(ctx, pkgAuditRecord.PATCreatedEvent, created, created.CreatedAt, map[string]any{
		"roles":       req.Roles,
		"project_ids": req.ProjectIDs,
	}); err != nil {
		s.logger.Error("failed to create audit record for PAT", "pat_id", created.ID, "error", err)
	}

	return created, patValue, nil
}

// createAuditRecord logs a PAT lifecycle event with org context and PAT metadata.
func (s *Service) createAuditRecord(ctx context.Context, event pkgAuditRecord.Event, pat PAT, occurredAt time.Time, targetMetadata map[string]any) error {
	orgName := ""
	if org, err := s.orgService.GetRaw(ctx, pat.OrgID); err == nil {
		orgName = org.Title
	}

	metadata := make(map[string]any, len(targetMetadata)+1)
	maps.Copy(metadata, targetMetadata)
	metadata["user_id"] = pat.UserID

	if _, err := s.auditRecordRepository.Create(ctx, models.AuditRecord{
		Event: event,
		Resource: models.Resource{
			ID:   pat.OrgID,
			Type: pkgAuditRecord.OrganizationType,
			Name: orgName,
		},
		Target: &models.Target{
			ID:       pat.ID,
			Type:     pkgAuditRecord.PATType,
			Name:     pat.Title,
			Metadata: metadata,
		},
		OrgID:      pat.OrgID,
		OccurredAt: occurredAt,
	}); err != nil {
		return fmt.Errorf("creating audit record: %w", err)
	}
	return nil
}

// generatePAT creates a random PAT string with the configured prefix and returns
// the plaintext value along with its SHA3-256 hash for storage.
// The hash is computed over the raw secret bytes (not the formatted PAT string)
// to avoid coupling the stored hash to the prefix configuration.
func (s *Service) generatePAT() (patValue, secretHash string, err error) {
	secretBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, secretBytes); err != nil {
		return "", "", fmt.Errorf("generating secret: %w", err)
	}

	patValue = s.config.Prefix + "_" + base64.RawURLEncoding.EncodeToString(secretBytes)

	hash := sha3.Sum256(secretBytes)
	secretHash = hex.EncodeToString(hash[:])

	return patValue, secretHash, nil
}
