package project

import (
	"context"
	"fmt"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/errors"
	"github.com/odpf/shield/pkg/uuid"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
	CheckPermission(ctx context.Context, usr user.User, resourceNS namespace.Namespace, resourceIdxa string, action action.Action) (bool, error)
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

func (s Service) Get(ctx context.Context, idOrSlug string) (Project, error) {
	if uuid.IsValid(idOrSlug) {
		return s.repository.GetByID(ctx, idOrSlug)
	}
	return s.repository.GetBySlug(ctx, idOrSlug)
}

func (s Service) Create(ctx context.Context, prj Project) (Project, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return Project{}, fmt.Errorf("%w: %s", user.ErrInvalidEmail, err.Error())
	}

	newProject, err := s.repository.Create(ctx, Project{
		Name:         prj.Name,
		Slug:         prj.Slug,
		Metadata:     prj.Metadata,
		Organization: prj.Organization,
	})
	if err != nil {
		return Project{}, err
	}

	if err = s.addAdminToProject(ctx, currentUser, newProject); err != nil {
		return Project{}, err
	}

	if err = s.addProjectToOrg(ctx, newProject, prj.Organization); err != nil {
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
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", user.ErrInvalidEmail, err.Error())
	} else if len(userIds) < 1 {
		return nil, user.ErrInvalidID
	}

	var prj Project
	if uuid.IsValid(idOrSlug) {
		prj, err = s.repository.GetByID(ctx, idOrSlug)
	} else {
		prj, err = s.repository.GetBySlug(ctx, idOrSlug)
	}
	if err != nil {
		return []user.User{}, err
	}

	isAllowed, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionProject, prj.ID, action.DefinitionManageProject)
	if err != nil {
		return []user.User{}, err
	} else if !isAllowed {
		return []user.User{}, errors.ErrForbidden
	}

	users, err := s.userService.GetByIDs(ctx, userIds)
	if err != nil {
		return []user.User{}, err
	}

	for _, usr := range users {
		if err = s.addAdminToProject(ctx, usr, prj); err != nil {
			return []user.User{}, err
		}
	}
	return s.ListAdmins(ctx, prj.ID)
}

func (s Service) ListAdmins(ctx context.Context, id string) ([]user.User, error) {
	return s.repository.ListAdmins(ctx, id)
}

func (s Service) RemoveAdmin(ctx context.Context, idOrSlug string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", user.ErrInvalidEmail, err.Error())
	}

	var prj Project
	if uuid.IsValid(idOrSlug) {
		prj, err = s.repository.GetByID(ctx, idOrSlug)
	} else {
		prj, err = s.repository.GetBySlug(ctx, idOrSlug)
	}
	if err != nil {
		return []user.User{}, err
	}

	isAllowed, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionProject, prj.ID, action.DefinitionManageProject)
	if err != nil {
		return []user.User{}, err
	} else if !isAllowed {
		return []user.User{}, errors.ErrForbidden
	}

	removedUser, err := s.userService.GetByID(ctx, userId)
	if err != nil {
		return []user.User{}, err
	}

	if err = s.removeAdminFromProject(ctx, removedUser, prj); err != nil {
		return []user.User{}, err
	}

	return s.ListAdmins(ctx, prj.ID)
}

func (s Service) addAdminToProject(ctx context.Context, usr user.User, prj Project) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionProject,
		ObjectID:         prj.ID,
		SubjectID:        usr.ID,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			ID:        role.DefinitionProjectAdmin.ID,
			Namespace: namespace.DefinitionProject,
		},
	}
	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}

	return nil
}

func (s Service) removeAdminFromProject(ctx context.Context, usr user.User, prj Project) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionProject,
		ObjectID:         prj.ID,
		SubjectID:        usr.ID,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			ID:        role.DefinitionProjectAdmin.ID,
			Namespace: namespace.DefinitionProject,
		},
	}
	return s.relationService.Delete(ctx, rel)
}

func (s Service) addProjectToOrg(ctx context.Context, prj Project, org organization.Organization) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionProject,
		ObjectID:         prj.ID,
		SubjectID:        org.ID,
		SubjectNamespace: namespace.DefinitionOrg,
		Role: role.Role{
			ID:        namespace.DefinitionOrg.ID,
			Namespace: namespace.DefinitionProject,
		},
		RelationType: relation.RelationTypes.Namespace,
	}

	if _, err := s.relationService.Create(ctx, rel); err != nil {
		return err
	}

	return nil
}
