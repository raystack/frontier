package auditrecord

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/serviceuser"
	userpkg "github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/salt/rql"
)

var SuperUserActorMetadataKey = "is_super_user"

type Repository interface {
	Create(ctx context.Context, auditRecord AuditRecord) (AuditRecord, error)
	GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (AuditRecord, error)
	List(ctx context.Context, query *rql.Query) (AuditRecordsList, error)
}

type UserService interface {
	GetByID(ctx context.Context, id string) (userpkg.User, error)
	IsSudo(ctx context.Context, id string, permissionName string) (bool, error)
}

type ServiceUserService interface {
	Get(ctx context.Context, id string) (serviceuser.ServiceUser, error)
}

type Service struct {
	repository  Repository
	userService UserService
	serviceUser ServiceUserService
}

func NewService(repository Repository, userService UserService, serviceUserService ServiceUserService) *Service {
	return &Service{
		repository:  repository,
		userService: userService,
		serviceUser: serviceUserService,
	}
}

func (s *Service) Create(ctx context.Context, auditRecord AuditRecord) (AuditRecord, bool, error) {
	// Check for idempotency key conflicts
	if auditRecord.IdempotencyKey != "" {
		existingRecord, err := s.repository.GetByIdempotencyKey(ctx, auditRecord.IdempotencyKey)
		if err == nil {
			existingHash := computeHash(existingRecord)
			newHash := computeHash(auditRecord)

			if existingHash == newHash {
				// Same request - return existing (idempotent success)
				// Return true to indicate this was an idempotency replay
				return existingRecord, true, nil
			} else {
				// Different request with same key - conflict
				return AuditRecord{}, false, ErrIdempotencyKeyConflict
			}
		} else if !errors.Is(err, ErrNotFound) {
			return AuditRecord{}, false, err
		}
		// If err is ErrNotFound, proceed to create the record
	}

	// todo(later): check what enrichment is done at service's end and modify this accordingly.
	// enrich actor info
	switch {
	case auditRecord.Actor.ID == uuid.Nil.String():
		auditRecord.Actor.Type = systemActor
		auditRecord.Actor.Name = systemActor

	case auditRecord.Actor.Type == schema.UserPrincipal:
		user, err := s.userService.GetByID(ctx, auditRecord.Actor.ID)
		if err != nil {
			return AuditRecord{}, false, err
		}
		auditRecord.Actor.Name = user.Title

		// check if the user is a superuser
		if isSudo, err := s.userService.IsSudo(ctx, user.ID, schema.PlatformSudoPermission); err != nil {
			return AuditRecord{}, false, err
		} else if isSudo {
			if auditRecord.Actor.Metadata == nil {
				auditRecord.Actor.Metadata = make(map[string]any)
			}
			auditRecord.Actor.Metadata[SuperUserActorMetadataKey] = true
		}

	case auditRecord.Actor.Type == schema.ServiceUserPrincipal:
		serviceUser, err := s.serviceUser.Get(ctx, auditRecord.Actor.ID)
		if err != nil {
			return AuditRecord{}, false, err
		}
		auditRecord.Actor.Name = serviceUser.Title
	}

	createdRecord, err := s.repository.Create(ctx, auditRecord)
	return createdRecord, false, err
}

func (s *Service) List(ctx context.Context, query *rql.Query) (AuditRecordsList, error) {
	return s.repository.List(ctx, query)
}

func computeHash(auditRecord AuditRecord) string {
	// Normalize event and IDs - trim spaces and lowercase for consistency
	normalisedEvent := strings.ToLower(strings.TrimSpace(auditRecord.Event))
	normalisedActorID := strings.ToLower(strings.TrimSpace(auditRecord.Actor.ID))
	normalisedResourceID := strings.ToLower(strings.TrimSpace(auditRecord.Resource.ID))
	normalisedOrgID := strings.ToLower(strings.TrimSpace(auditRecord.OrgID))

	var normalisedTargetID string
	if auditRecord.Target != nil {
		normalisedTargetID = strings.ToLower(strings.TrimSpace(auditRecord.Target.ID))
	}

	inputString := fmt.Sprintf("%s|%s|%s|%s|%s",
		normalisedEvent,
		normalisedActorID,
		normalisedResourceID,
		normalisedOrgID,
		normalisedTargetID)

	hasher := sha256.New()
	hasher.Write([]byte(inputString))
	hashBytes := hasher.Sum(nil)

	// Convert to hex string
	return hex.EncodeToString(hashBytes)
}
