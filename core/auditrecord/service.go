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
	paterrors "github.com/raystack/frontier/core/userpat/errors"
	patModels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/server/consts"
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

type UserPATService interface {
	GetByID(ctx context.Context, id string) (patModels.PAT, error)
}

type Service struct {
	repository     Repository
	userService    UserService
	serviceUser    ServiceUserService
	sessionService SessionService
	userPATService UserPATService
}

func NewService(repository Repository, userService UserService, serviceUserService ServiceUserService, sessionService SessionService, userPATService UserPATService) *Service {
	return &Service{
		repository:     repository,
		userService:    userService,
		serviceUser:    serviceUserService,
		sessionService: sessionService,
		userPATService: userPATService,
	}
}

func (s *Service) Create(ctx context.Context, auditRecord AuditRecord) (AuditRecord, bool, error) {
	if auditRecord.IdempotencyKey != "" {
		existingRecord, isReplay, err := s.checkIdempotency(ctx, auditRecord)
		if err != nil {
			return AuditRecord{}, false, err
		}
		if isReplay {
			return existingRecord, true, nil
		}
	}

	enrichedActor, err := s.enrichActor(ctx, auditRecord.Actor)
	if err != nil {
		return AuditRecord{}, false, err
	}
	auditRecord.Actor = enrichedActor

	createdRecord, err := s.repository.Create(ctx, auditRecord)
	return createdRecord, false, err
}

// checkIdempotency returns (existingRecord, true, nil) if this is a replay of a previous request,
// or (empty, false, nil) if no duplicate exists and creation should proceed.
func (s *Service) checkIdempotency(ctx context.Context, auditRecord AuditRecord) (AuditRecord, bool, error) {
	existingRecord, err := s.repository.GetByIdempotencyKey(ctx, auditRecord.IdempotencyKey)
	if errors.Is(err, ErrNotFound) {
		return AuditRecord{}, false, nil
	}
	if err != nil {
		return AuditRecord{}, false, err
	}
	if computeHash(existingRecord) == computeHash(auditRecord) {
		return existingRecord, true, nil
	}
	return AuditRecord{}, false, ErrIdempotencyKeyConflict
}

func (s *Service) enrichActor(ctx context.Context, actor Actor) (Actor, error) {
	switch {
	case actor.ID == uuid.Nil.String():
		return s.enrichSystemActor(actor), nil
	case actor.Type == schema.UserPrincipal:
		return s.enrichUserActor(ctx, actor)
	case actor.Type == schema.ServiceUserPrincipal:
		return s.enrichServiceUserActor(ctx, actor)
	case actor.Type == schema.PATPrincipal:
		return s.enrichPATActor(ctx, actor)
	default:
		return actor, nil
	}
}

func (s *Service) enrichSystemActor(actor Actor) Actor {
	actor.Type = auditrecord.SystemActor
	actor.Name = auditrecord.SystemActor
	return actor
}

func (s *Service) enrichUserActor(ctx context.Context, actor Actor) (Actor, error) {
	sessionUUID, err := uuid.Parse(actor.ID)
	if err != nil {
		return Actor{}, ErrActorNotFound
	}
	session, err := s.sessionService.Get(ctx, sessionUUID)
	if err != nil {
		if errors.Is(err, frontiersession.ErrNoSession) {
			return Actor{}, ErrActorNotFound
		}
		return Actor{}, err
	}
	user, err := s.userService.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, userpkg.ErrNoUsersFound) {
			return Actor{}, ErrActorNotFound
		}
		return Actor{}, err
	}
	actor.Name = user.Name
	actor.Title = user.Title

	if actor.Metadata == nil {
		actor.Metadata = make(map[string]any)
	}
	actor.Metadata[consts.AuditSessionMetadataKey] = session.Metadata

	isSudo, err := s.userService.IsSudo(ctx, user.ID, schema.PlatformSudoPermission)
	if err != nil {
		return Actor{}, err
	}
	if isSudo {
		actor.Metadata[consts.AuditActorSuperUserKey] = true
	}

	return actor, nil
}

func (s *Service) enrichServiceUserActor(ctx context.Context, actor Actor) (Actor, error) {
	serviceUser, err := s.serviceUser.Get(ctx, actor.ID)
	if err != nil {
		return Actor{}, err
	}
	actor.Title = serviceUser.Title
	return actor, nil
}

func (s *Service) enrichPATActor(ctx context.Context, actor Actor) (Actor, error) {
	pat, err := s.userPATService.GetByID(ctx, actor.ID)
	if err != nil {
		if errors.Is(err, paterrors.ErrNotFound) {
			return Actor{}, ErrActorNotFound
		}
		return Actor{}, err
	}
	actor.Name = pat.Title
	actor.Title = pat.Title

	user, err := s.userService.GetByID(ctx, pat.UserID)
	if err != nil {
		if errors.Is(err, userpkg.ErrNotExist) {
			return Actor{}, ErrActorNotFound
		}
		return Actor{}, err
	}
	if actor.Metadata == nil {
		actor.Metadata = make(map[string]any)
	}
	actor.Metadata["user"] = map[string]any{
		"id":    user.ID,
		"name":  user.Name,
		"title": user.Title,
		"email": user.Email,
	}
	return actor, nil
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
	var metadataMap map[string]interface{}
	if actor.Metadata != nil {
		metadataMap = make(map[string]interface{}, len(actor.Metadata))
		for k, v := range actor.Metadata {
			metadataMap[k] = v
		}
	}
	actorMap := map[string]interface{}{
		"id":       actor.ID,
		"type":     actor.Type,
		"name":     actor.Name,
		"title":    actor.Title,
		"metadata": metadataMap,
	}
	return context.WithValue(ctx, consts.AuditRecordActorContextKey, actorMap)
}
