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
}

type Service struct {
	store           Store
	relationService RelationService
	userService     UserService
}

func NewService(store Store, relationService RelationService, userService UserService) *Service {
	return &Service{
		store:           store,
		relationService: relationService,
		userService:     userService,
	}
}

func (s Service) Get(ctx context.Context, id string) (Project, error) {
	return s.store.GetProject(ctx, id)
}

func (s Service) Create(ctx context.Context, prj Project) (Project, error) {
	user, err := s.userService.FetchCurrentUser(ctx)

	if err != nil {
		return Project{}, err
	}

	newProject, err := s.store.CreateProject(ctx, Project{
		Name:         prj.Name,
		Slug:         prj.Slug,
		Metadata:     prj.Metadata,
		Organization: prj.Organization,
	})
	if err != nil {
		return Project{}, err
	}

	err = s.AddAdminToProject(ctx, user, newProject)

	if err != nil {
		return Project{}, err
	}

	err = s.AddProjectToOrg(ctx, newProject, prj.Organization)

	if err != nil {
		return Project{}, err
	}

	return newProject, nil
}

func (s Service) List(ctx context.Context) ([]Project, error) {
	return s.store.ListProject(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate Project) (Project, error) {
	return s.store.UpdateProject(ctx, toUpdate)
}

func (s Service) AddAdmin(ctx context.Context, id string, userIds []string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	id = strings.TrimSpace(id)
	project, err := s.store.GetProject(ctx, id)

	if err != nil {
		return []user.User{}, err
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionProject, project.Id, action.DefinitionManageProject)
	if err != nil {
		return []user.User{}, err
	}

	if !isAuthorized {
		return []user.User{}, errors.Unauthorized
	}

	users, err := s.store.GetUsersByIds(ctx, userIds)

	if err != nil {
		return []user.User{}, err
	}

	for _, usr := range users {
		err = s.AddAdminToProject(ctx, usr, project)
		if err != nil {
			return []user.User{}, err
		}
	}
	return s.ListAdmins(ctx, id)
}

func (s Service) ListAdmins(ctx context.Context, id string) ([]user.User, error) {
	return s.store.ListProjectAdmins(ctx, id)
}

func (s Service) RemoveAdmin(ctx context.Context, id string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	id = strings.TrimSpace(id)
	project, err := s.store.GetProject(ctx, id)
	if err != nil {
		return []user.User{}, err
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionProject, project.Id, action.DefinitionManageProject)
	if err != nil {
		return []user.User{}, err
	}

	if !isAuthorized {
		return []user.User{}, errors.Unauthorized
	}

	usr, err := s.store.GetUser(ctx, userId)
	if err != nil {
		return []user.User{}, err
	}

	err = s.RemoveAdminFromProject(ctx, usr, project)
	if err != nil {
		return []user.User{}, err
	}

	return s.ListAdmins(ctx, id)
}

func (s Service) AddAdminToProject(ctx context.Context, usr user.User, prj Project) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionProject,
		ObjectId:         prj.Id,
		SubjectId:        usr.Id,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			Id:        role.DefinitionProjectAdmin.Id,
			Namespace: namespace.DefinitionProject,
		},
	}
	_, err := s.relationService.Create(ctx, rel)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) RemoveAdminFromProject(ctx context.Context, usr user.User, prj Project) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionProject,
		ObjectId:         prj.Id,
		SubjectId:        usr.Id,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			Id:        role.DefinitionProjectAdmin.Id,
			Namespace: namespace.DefinitionProject,
		},
	}
	return s.relationService.Delete(ctx, rel)
}

func (s Service) AddProjectToOrg(ctx context.Context, prj Project, org organization.Organization) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionProject,
		ObjectId:         prj.Id,
		SubjectId:        org.Id,
		SubjectNamespace: namespace.DefinitionOrg,
		Role: role.Role{
			Id:        namespace.DefinitionOrg.Id,
			Namespace: namespace.DefinitionProject,
		},
		RelationType: relation.RelationTypes.Namespace,
	}
	_, err := s.relationService.Create(ctx, rel)
	if err != nil {
		return err
	}

	return nil
}
