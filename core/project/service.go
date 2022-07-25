package project

import (
	"context"
	"strings"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/errors"
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

func (s Service) Get(ctx context.Context, id string) (Project, error) {
	return s.repository.Get(ctx, id)
}

func (s Service) Create(ctx context.Context, prj Project) (Project, error) {
	user, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return Project{}, err
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

	if err = s.AddAdminToProject(ctx, user, newProject); err != nil {
		return Project{}, err
	}

	if err = s.AddProjectToOrg(ctx, newProject, prj.Organization); err != nil {
		return Project{}, err
	}

	return newProject, nil
}

func (s Service) List(ctx context.Context) ([]Project, error) {
	return s.repository.List(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate Project) (Project, error) {
	return s.repository.Update(ctx, toUpdate)
}

func (s Service) AddAdmin(ctx context.Context, id string, userIds []string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	id = strings.TrimSpace(id)
	project, err := s.repository.Get(ctx, id)

	if err != nil {
		return []user.User{}, err
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionProject, project.ID, action.DefinitionManageProject)
	if err != nil {
		return []user.User{}, err
	}

	if !isAuthorized {
		return []user.User{}, errors.Unauthorized
	}

	users, err := s.userService.GetByIDs(ctx, userIds)
	if err != nil {
		return []user.User{}, err
	}

	for _, usr := range users {
		if err = s.AddAdminToProject(ctx, usr, project); err != nil {
			return []user.User{}, err
		}
	}
	return s.ListAdmins(ctx, id)
}

func (s Service) ListAdmins(ctx context.Context, id string) ([]user.User, error) {
	return s.repository.ListAdmins(ctx, id)
}

func (s Service) RemoveAdmin(ctx context.Context, id string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	id = strings.TrimSpace(id)
	project, err := s.repository.Get(ctx, id)
	if err != nil {
		return []user.User{}, err
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionProject, project.ID, action.DefinitionManageProject)
	if err != nil {
		return []user.User{}, err
	}

	if !isAuthorized {
		return []user.User{}, errors.Unauthorized
	}

	usr, err := s.userService.GetByID(ctx, userId)
	if err != nil {
		return []user.User{}, err
	}

	if err = s.RemoveAdminFromProject(ctx, usr, project); err != nil {
		return []user.User{}, err
	}

	return s.ListAdmins(ctx, id)
}

func (s Service) AddAdminToProject(ctx context.Context, usr user.User, prj Project) error {
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

func (s Service) RemoveAdminFromProject(ctx context.Context, usr user.User, prj Project) error {
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

func (s Service) AddProjectToOrg(ctx context.Context, prj Project, org organization.Organization) error {
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
