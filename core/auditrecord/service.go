package auditrecord

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/serviceuser"
	userpkg "github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/server/consts"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/rql"
)

type Repository interface {
	Create(ctx context.Context, auditRecord AuditRecord) (AuditRecord, error)
	GetByIdempotencyKey(ctx context.Context, idempotencyKey string) (AuditRecord, error)
	List(ctx context.Context, query *rql.Query) (AuditRecordsList, error)
	Export(ctx context.Context, query *rql.Query) (io.Reader, string, error)
}

type UserService interface {
	GetByID(ctx context.Context, id string) (userpkg.User, error)
	IsSudo(ctx context.Context, id string, permissionName string) (bool, error)
}

type SessionService interface {
	Get(ctx context.Context, id uuid.UUID) (*frontiersession.Session, error)
}

type ServiceUserService interface {
	Get(ctx context.Context, id string) (serviceuser.ServiceUser, error)
}

type Service struct {
	repository     Repository
	userService    UserService
	serviceUser    ServiceUserService
	sessionService SessionService
}

func NewService(repository Repository, userService UserService, serviceUserService ServiceUserService, sessionService SessionService) *Service {
	return &Service{
		repository:     repository,
		userService:    userService,
		serviceUser:    serviceUserService,
		sessionService: sessionService,
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

	// enrich actor info
	switch {
	case auditRecord.Actor.ID == uuid.Nil.String():
		auditRecord.Actor.Type = auditrecord.SystemActor
		auditRecord.Actor.Name = auditrecord.SystemActor

	case auditRecord.Actor.Type == schema.UserPrincipal:
		actorUUID, err := uuid.Parse(auditRecord.Actor.ID)
		if err != nil {
			return AuditRecord{}, false, ErrActorNotFound
		}
		session, err := s.sessionService.Get(ctx, actorUUID)
		if err != nil {
			if errors.Is(err, frontiersession.ErrNoSession) {
				return AuditRecord{}, false, ErrActorNotFound
			}
			return AuditRecord{}, false, err
		}
		user, err := s.userService.GetByID(ctx, session.UserID)
		if err != nil {
			if errors.Is(err, userpkg.ErrNoUsersFound) {
				return AuditRecord{}, false, ErrActorNotFound
			}
			return AuditRecord{}, false, err
		}
		auditRecord.Actor.Name = utils.Coalesce(user.Name, user.Title)

		// check if the user is a superuser
		if isSudo, err := s.userService.IsSudo(ctx, user.ID, schema.PlatformSudoPermission); err != nil {
			return AuditRecord{}, false, err
		} else if isSudo {
			if auditRecord.Actor.Metadata == nil {
				auditRecord.Actor.Metadata = make(map[string]any)
			}
			auditRecord.Actor.Metadata[consts.AuditActorSuperUserKey] = true
			auditRecord.Actor.Metadata[consts.AuditSessionMetadataKey] = session.Metadata
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

func (s *Service) Export(ctx context.Context, query *rql.Query) (io.Reader, string, error) {
	return s.repository.Export(ctx, query)
}

func computeHash(auditRecord AuditRecord) string {
	// Normalize event and IDs - trim spaces and lowercase for consistency
	normalisedEvent := strings.ToLower(strings.TrimSpace(auditRecord.Event.String()))
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

// SetAuditRecordActorContext sets the audit record actor in context
// It accepts an Actor struct but stores it as a map to avoid layer violations in repositories
func SetAuditRecordActorContext(ctx context.Context, actor Actor) context.Context {
	actorMap := map[string]interface{}{
		"id":       actor.ID,
		"type":     actor.Type,
		"name":     actor.Name,
		"metadata": actor.Metadata,
	}
	return context.WithValue(ctx, consts.AuditRecordActorContextKey, actorMap)
}
