package project

import (
	"context"

	"github.com/raystack/shield/core/action"
	"github.com/raystack/shield/core/namespace"
	"github.com/raystack/shield/core/relation"
	"github.com/raystack/shield/core/user"
	"github.com/raystack/shield/internal/schema"
	"github.com/raystack/shield/pkg/uuid"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.RelationV2) (relation.RelationV2, error)
	CheckPermission(ctx context.Context, usr user.User, resourceNS namespace.Namespace, resourceIdxa string, action action.Action) (bool, error)
	LookupSubjects(ctx context.Context, rel relation.RelationV2) ([]string, error)
	ListRelations(ctx context.Context, rel relation.RelationV2) ([]relation.RelationV2, error)
}

type UserService interface {
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

func (s Service) Get(ctx context.Context, idOrSlug string) (Project, error) {
	if uuid.IsValid(idOrSlug) {
		return s.repository.GetByID(ctx, idOrSlug)
	}
	return s.repository.GetBySlug(ctx, idOrSlug)
}

func (s Service) GetByIDs(ctx context.Context, ids []string) ([]Project, error) {
	return s.repository.GetByIDs(ctx, ids)
}

func (s Service) Create(ctx context.Context, prj Project) (Project, error) {
	newProject, err := s.repository.Create(ctx, Project{
		Name:         prj.Name,
		Slug:         prj.Slug,
		Metadata:     prj.Metadata,
		Organization: prj.Organization,
	})
	if err != nil {
		return Project{}, err
	}

	if err = s.addProjectToOrg(ctx, newProject, prj.Organization.ID); err != nil {
		return Project{}, err
	}

	return newProject, nil
}

func (s Service) List(ctx context.Context) ([]Project, error) {
	return s.repository.List(ctx)
}

func (s Service) Update(ctx context.Context, prj Project) (Project, error) {
	if prj.ID != "" {
		return s.repository.UpdateByID(ctx, prj)
	}
	return s.repository.UpdateBySlug(ctx, prj)
}

func (s Service) AddAdmins(ctx context.Context, idOrSlug string, userIds []string) ([]user.User, error) {
	// TODO(discussion): can be done with create relations
	return []user.User{}, nil
}

func (s Service) ListUsers(ctx context.Context, id string, permissionFilter string) ([]user.User, error) {
	userIDs, err := s.relationService.LookupSubjects(ctx, relation.RelationV2{
		Object: relation.Object{
			ID:        id,
			Namespace: schema.ProjectNamespace,
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

func (s Service) RemoveAdmin(ctx context.Context, idOrSlug string, userId string) ([]user.User, error) {
	// TODO(discussion): can be done with delete relations
	return []user.User{}, nil
}

func (s Service) addProjectToOrg(ctx context.Context, prj Project, orgID string) error {
	rel := relation.RelationV2{
		Object: relation.Object{
			ID:        prj.ID,
			Namespace: schema.ProjectNamespace,
		},
		Subject: relation.Subject{
			ID:        orgID,
			Namespace: schema.OrganizationNamespace,
			RoleID:    schema.OrganizationRelationName,
		},
	}

	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}

	return nil
}

func (s Service) ListByOrganization(ctx context.Context, id string) ([]Project, error) {
	relations, err := s.relationService.ListRelations(ctx, relation.RelationV2{
		Object: relation.Object{
			Namespace: schema.ProjectNamespace,
		},
		Subject: relation.Subject{
			ID:        id,
			Namespace: schema.OrganizationNamespace,
			RoleID:    schema.OrganizationRelationName,
		},
	})
	if err != nil {
		return nil, err
	}

	var projectIDs []string
	for _, rel := range relations {
		projectIDs = append(projectIDs, rel.Object.ID)
	}
	if len(projectIDs) == 0 {
		// no projects
		return []Project{}, nil
	}
	return s.repository.GetByIDs(ctx, projectIDs)
}
