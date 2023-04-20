package organization

import (
	"context"
	"fmt"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/pkg/uuid"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.RelationV2) (relation.RelationV2, error)
	CheckPermission(ctx context.Context, usr user.User, resourceNS namespace.Namespace, resourceIdxa string, action action.Action) (bool, error)
	LookupSubjects(ctx context.Context, rel relation.RelationV2) ([]string, error)
	LookupResources(ctx context.Context, rel relation.RelationV2) ([]string, error)
}

type UserService interface {
	FetchCurrentUser(ctx context.Context) (user.User, error)
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error)
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
	if err = s.CreateRelation(ctx, newOrg, relation.BuildUserResourceAdminSubject(currentUser)); err != nil {
		return Organization{}, err
	}
	return newOrg, nil
}

func (s Service) List(ctx context.Context) ([]Organization, error) {
	return s.repository.List(ctx)
}

func (s Service) Update(ctx context.Context, org Organization) (Organization, error) {
	if org.ID != "" {
		return s.repository.UpdateByID(ctx, org)
	}
	return s.repository.UpdateBySlug(ctx, org)
}

func (s Service) ListAdmins(ctx context.Context, idOrSlug string) ([]user.User, error) {
	var org Organization
	var err error
	if uuid.IsValid(idOrSlug) {
		return s.repository.ListAdminsByOrgID(ctx, idOrSlug)
	}
	org, err = s.repository.GetBySlug(ctx, idOrSlug)
	if err != nil {
		return []user.User{}, err
	}
	return s.repository.ListAdminsByOrgID(ctx, org.ID)
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

func (s Service) ListUsers(ctx context.Context, id string, permissionFilter string) ([]user.User, error) {
	userIDs, err := s.relationService.LookupSubjects(ctx, relation.RelationV2{
		Object: relation.Object{
			ID:        id,
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
		return []user.User{}, nil
	}
	return s.userService.GetByIDs(ctx, userIDs)
}

func (s Service) ListByUser(ctx context.Context, userID string) ([]Organization, error) {
	subjectIDs, err := s.relationService.LookupResources(ctx, relation.RelationV2{
		Object: relation.Object{
			Namespace: schema.OrganizationNamespace,
		},
		Subject: relation.Subject{
			ID:        userID,
			Namespace: schema.UserPrincipal,
			RoleID:    schema.ViewPermission,
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
