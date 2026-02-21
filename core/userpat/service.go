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

// Create generates a new personal access token and returns it with the plaintext token value.
// The plaintext token is only available at creation time.
func (s *Service) Create(ctx context.Context, req CreateRequest) (PersonalAccessToken, string, error) {
	if !s.config.Enabled {
		return PersonalAccessToken{}, "", ErrDisabled
	}

	// NOTE: CountActive + Create is not atomic (TOCTOU race). Two concurrent requests
	// could both read count=49 (assuming max limit is 50), pass the check, and create tokens exceeding the limit.
	// Acceptable for now given low concurrency on this endpoint. If this becomes an issue,
	// use an atomic INSERT ... SELECT with a count subquery in the WHERE clause.
	count, err := s.repo.CountActive(ctx, req.UserID, req.OrgID)
	if err != nil {
		return PersonalAccessToken{}, "", fmt.Errorf("counting active tokens: %w", err)
	}
	if count >= s.config.MaxTokensPerUserPerOrg {
		return PersonalAccessToken{}, "", ErrLimitExceeded
	}

	tokenValue, secretHash, err := s.generateToken()
	if err != nil {
		return PersonalAccessToken{}, "", err
	}

	pat := PersonalAccessToken{
		UserID:     req.UserID,
		OrgID:      req.OrgID,
		Title:      req.Title,
		SecretHash: secretHash,
		Metadata:   req.Metadata,
		ExpiresAt:  req.ExpiresAt,
	}

	created, err := s.repo.Create(ctx, pat)
	if err != nil {
		return PersonalAccessToken{}, "", err
	}

	// TODO: create policies for roles + project_ids

	// TODO: move audit record creation into the same transaction as token creation to avoid partial state where token exists but audit record doesn't.
	if err := s.createAuditRecord(ctx, pkgAuditRecord.PATCreatedEvent, created, created.CreatedAt, map[string]any{
		"roles":       req.Roles,
		"project_ids": req.ProjectIDs,
	}); err != nil {
		s.logger.Error("failed to create audit record for PAT", "pat_id", created.ID, "error", err)
	}

	return created, tokenValue, nil
}

// createAuditRecord logs a PAT lifecycle event with org context and token metadata.
func (s *Service) createAuditRecord(ctx context.Context, event pkgAuditRecord.Event, pat PersonalAccessToken, occurredAt time.Time, targetMetadata map[string]any) error {
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

// generateToken creates a random token string with the configured prefix and returns
// the plaintext token value along with its SHA3-256 hash for storage.
// The hash is computed over the raw secret bytes (not the formatted token string)
// to avoid coupling the stored hash to the token prefix configuration.
func (s *Service) generateToken() (tokenValue, secretHash string, err error) {
	secretBytes := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, secretBytes); err != nil {
		return "", "", fmt.Errorf("generating secret: %w", err)
	}

	tokenValue = s.config.TokenPrefix + "_" + base64.RawURLEncoding.EncodeToString(secretBytes)

	hash := sha3.Sum256(secretBytes)
	secretHash = hex.EncodeToString(hash[:])

	return tokenValue, secretHash, nil
}
