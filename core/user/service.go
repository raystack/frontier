package user

import (
	"context"
	"net/mail"
	"strings"

	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/internal/schema"

	shieldsession "github.com/odpf/shield/core/authenticate/session"
	"github.com/odpf/shield/pkg/errors"
	shielduuid "github.com/odpf/shield/pkg/uuid"
)

type SessionService interface {
	ExtractFromContext(ctx context.Context) (*shieldsession.Session, error)
}

type RelationRepository interface {
	DeleteSubjectRelations(ctx context.Context, resourceType, optionalResourceID string) error
	LookupSubjects(ctx context.Context, rel relation.RelationV2) ([]string, error)
	LookupResources(ctx context.Context, rel relation.RelationV2) ([]string, error)
}

type Service struct {
	repository      Repository
	relationService RelationRepository
	sessionService  SessionService
}

func NewService(repository Repository, sessionService SessionService, relationRepo RelationRepository) *Service {
	return &Service{
		repository:      repository,
		sessionService:  sessionService,
		relationService: relationRepo,
	}
}

// Get by user uuid, email or slug
func (s Service) GetByID(ctx context.Context, id string) (User, error) {
	if isValidEmail(id) {
		return s.repository.GetByEmail(ctx, id)
	}
	if shielduuid.IsValid(id) {
		return s.repository.GetByID(ctx, id)
	}
	return s.repository.GetBySlug(ctx, id)
}

func (s Service) GetByIDs(ctx context.Context, userIDs []string) ([]User, error) {
	return s.repository.GetByIDs(ctx, userIDs)
}

func (s Service) GetByEmail(ctx context.Context, email string) (User, error) {
	return s.repository.GetByEmail(ctx, email)
}

func (s Service) Create(ctx context.Context, user User) (User, error) {
	newUser, err := s.repository.Create(ctx, User{
		Name:     user.Name,
		Email:    user.Email,
		Slug:     user.Slug,
		Metadata: user.Metadata,
	})
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

func (s Service) CreateMetadataKey(ctx context.Context, key UserMetadataKey) (UserMetadataKey, error) {
	newUserMetadataKey, err := s.repository.CreateMetadataKey(ctx, UserMetadataKey{
		Key:         key.Key,
		Description: key.Description,
	})
	if err != nil {
		return UserMetadataKey{}, err
	}

	return newUserMetadataKey, nil
}

func (s Service) List(ctx context.Context, flt Filter) ([]User, error) {
	if flt.OrgID != "" {
		return s.ListByOrg(ctx, flt.OrgID, schema.MembershipPermission)
	}
	if flt.GroupID != "" {
		return s.ListByGroup(ctx, flt.GroupID, schema.MembershipPermission)
	}

	// state gets filtered in db
	return s.repository.List(ctx, flt)
}

// Update by user uuid, email or slug
func (s Service) UpdateByID(ctx context.Context, toUpdate User) (User, error) {
	id := toUpdate.ID
	if isValidEmail(id) {
		return s.repository.UpdateByEmail(ctx, toUpdate)
	}
	if shielduuid.IsValid(id) {
		return s.repository.UpdateByID(ctx, toUpdate)
	}
	return s.repository.UpdateBySlug(ctx, toUpdate)
}

func (s Service) UpdateByEmail(ctx context.Context, toUpdate User) (User, error) {
	return s.repository.UpdateByEmail(ctx, toUpdate)
}

func (s Service) FetchCurrentUser(ctx context.Context) (User, error) {
	var currentUser User

	// extract user from session if present
	session, err := s.sessionService.ExtractFromContext(ctx)
	if err == nil && session.IsValid() && shielduuid.IsValid(session.UserID) {
		// userID is a valid uuid
		currentUser, err = s.GetByID(ctx, session.UserID)
		if err != nil {
			return User{}, err
		}
		return currentUser, nil
	}

	// check if header with user email is set
	if val, ok := GetEmailFromContext(ctx); ok && len(val) > 0 {
		currentUser, err = s.GetByEmail(ctx, strings.TrimSpace(val))
		if err != nil {
			return User{}, err
		}
		return currentUser, nil
	}
	return User{}, errors.ErrUnauthenticated
}

func (s Service) Enable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Enabled)
}

func (s Service) Disable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Disabled)
}

func (s Service) Delete(ctx context.Context, id string) error {
	if err := s.relationService.DeleteSubjectRelations(ctx, schema.ProjectNamespace, id); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}

func (s Service) ListByOrg(ctx context.Context, orgID string, permissionFilter string) ([]User, error) {
	userIDs, err := s.relationService.LookupSubjects(ctx, relation.RelationV2{
		Object: relation.Object{
			ID:        orgID,
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			Namespace: schema.UserPrincipal,
			RoleID:    permissionFilter,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(userIDs) == 0 {
		// no users
		return []User{}, nil
	}
	return s.repository.GetByIDs(ctx, userIDs)
}

func (s Service) ListByGroup(ctx context.Context, groupID string, permissionFilter string) ([]User, error) {
	userIDs, err := s.relationService.LookupSubjects(ctx, relation.RelationV2{
		Object: relation.Object{
			ID:        groupID,
			Namespace: schema.GroupPrincipal,
		},
		Subject: relation.Subject{
			Namespace: schema.UserPrincipal,
			RoleID:    permissionFilter,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(userIDs) == 0 {
		// no users
		return []User{}, nil
	}
	return s.repository.GetByIDs(ctx, userIDs)
}

func isValidEmail(str string) bool {
	_, err := mail.ParseAddress(str)
	return err == nil
}
