package organization

import (
	"context"
	"fmt"

	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/bootstrap/schema"
	"github.com/odpf/shield/pkg/uuid"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.RelationV2) (relation.RelationV2, error)
	LookupResources(ctx context.Context, rel relation.RelationV2) ([]string, error)
	Delete(ctx context.Context, rel relation.RelationV2) error
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

func NewService(repository Repository, relationService RelationService,
	userService UserService) *Service {
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

	// attach user as owner
	if err = s.AddOwner(ctx, newOrg, currentUser); err != nil {
		return newOrg, err
	}

	// attach org to central platform
	if err = s.AttachToPlatform(ctx, newOrg.ID); err != nil {
		return newOrg, err
	}

	return newOrg, nil
}

func (s Service) AddOwner(ctx context.Context, newOrg Organization, currentUser user.User) error {
	if _, err := s.relationService.Create(ctx, relation.RelationV2{
		Object: relation.Object{
			ID:        newOrg.ID,
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        currentUser.ID,
			Namespace: schema.UserPrincipal,
		},
		RelationName: schema.OwnerRelation,
	}); err != nil {
		return err
	}
	return nil
}

func (s Service) AttachToPlatform(ctx context.Context, orgID string) error {
	if _, err := s.relationService.Create(ctx, relation.RelationV2{
		Object: relation.Object{
			ID:        orgID,
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
		RelationName: schema.PlatformRelationName,
	}); err != nil {
		return err
	}
	return nil
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

func (s Service) ListByUser(ctx context.Context, userID string) ([]Organization, error) {
	subjectIDs, err := s.relationService.LookupResources(ctx, relation.RelationV2{
		Object: relation.Object{
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        userID,
			Namespace: schema.UserPrincipal,
		},
		RelationName: schema.MembershipPermission,
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
	if err := s.relationService.Delete(ctx, relation.RelationV2{Object: relation.Object{
		ID:        id,
		Namespace: schema.OrganizationNamespace,
	}}); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}
