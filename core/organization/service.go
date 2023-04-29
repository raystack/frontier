package organization

import (
	"context"
	"fmt"

	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/pkg/uuid"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.RelationV2) (relation.RelationV2, error)
	LookupResources(ctx context.Context, rel relation.RelationV2) ([]string, error)
	DeleteSubjectRelations(ctx context.Context, resourceType, optionalResourceID string) error
}

type UserService interface {
	FetchCurrentUser(ctx context.Context) (user.User, error)
	GetByID(ctx context.Context, id string) (user.User, error)
}

type Service struct {
	repository      Repository
	relationService RelationService
	userService     UserService
}

func NewService(repository Repository, relationService RelationService, userService UserService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationService,
		userService:     userService,
	}
}

func (s Service) Get(ctx context.Context, idOrSlug string) (Organization, error) {
	if uuid.IsValid(idOrSlug) {
		return s.repository.GetByID(ctx, idOrSlug)
	}
	return s.repository.GetBySlug(ctx, idOrSlug)
}

func (s Service) Create(ctx context.Context, org Organization) (Organization, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return Organization{}, fmt.Errorf("%w: %s", user.ErrInvalidEmail, err.Error())
	}

	newOrg, err := s.repository.Create(ctx, Organization{
		Name:     org.Name,
		Slug:     org.Slug,
		Metadata: org.Metadata,
	})
	if err != nil {
		return Organization{}, err
	}

	// attach user as admin
	if err = s.CreateRelation(ctx, newOrg, BuildUserOrgAdminSubject(currentUser)); err != nil {
		return Organization{}, err
	}
	return newOrg, nil
}

func (s Service) List(ctx context.Context, f Filter) ([]Organization, error) {
	if f.UserID != "" {
		return s.ListByUser(ctx, f.UserID)
	}

	// state gets filtered in db
	return s.repository.List(ctx, f)
}

func (s Service) Update(ctx context.Context, org Organization) (Organization, error) {
	if org.ID != "" {
		return s.repository.UpdateByID(ctx, org)
	}
	return s.repository.UpdateBySlug(ctx, org)
}

func (s Service) CreateRelation(ctx context.Context, org Organization, subject relation.Subject) error {
	rel := relation.RelationV2{
		Object: relation.Object{
			ID:        org.ID,
			Namespace: schema.OrganizationNamespace,
		},
		Subject: subject,
	}
	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}
	return nil
}

func (s Service) ListByUser(ctx context.Context, userID string) ([]Organization, error) {
	subjectIDs, err := s.relationService.LookupResources(ctx, relation.RelationV2{
		Object: relation.Object{
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        userID,
			Namespace: schema.UserPrincipal,
			RoleID:    schema.MembershipPermission,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(subjectIDs) == 0 {
		// no organizations
		return []Organization{}, nil
	}
	return s.repository.GetByIDs(ctx, subjectIDs)
}

func (s Service) Enable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Enabled)
}

func (s Service) Disable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Disabled)
}

// DeleteModel doesn't delete the nested resource, only itself
func (s Service) DeleteModel(ctx context.Context, id string) error {
	if err := s.relationService.DeleteSubjectRelations(ctx, schema.OrganizationNamespace, id); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}
