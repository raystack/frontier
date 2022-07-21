package project

import (
	"context"
	"errors"
	"strings"

	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/internal/permission"
	"github.com/odpf/shield/model"
	shieldError "github.com/odpf/shield/utils/errors"
)

type Service struct {
	Store       Store
	Permissions permission.Permissions
}

var (
	ProjectDoesntExist = errors.New("project doesn't exist")
	NoAdminsExist      = errors.New("no admins exist")
	InvalidUUID        = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetProject(ctx context.Context, id string) (model.Project, error)
	CreateProject(ctx context.Context, org model.Project) (model.Project, error)
	ListProject(ctx context.Context) ([]model.Project, error)
	UpdateProject(ctx context.Context, toUpdate model.Project) (model.Project, error)
	GetUsersByIds(ctx context.Context, userIds []string) ([]model.User, error)
	GetUser(ctx context.Context, userId string) (model.User, error)
	ListProjectAdmins(ctx context.Context, id string) ([]model.User, error)
}

func (s Service) Get(ctx context.Context, id string) (model.Project, error) {
	return s.Store.GetProject(ctx, id)
}

func (s Service) Create(ctx context.Context, project model.Project) (model.Project, error) {
	user, err := s.Permissions.FetchCurrentUser(ctx)

	if err != nil {
		return model.Project{}, err
	}

	newProject, err := s.Store.CreateProject(ctx, model.Project{
		Name:         project.Name,
		Slug:         project.Slug,
		Metadata:     project.Metadata,
		Organization: project.Organization,
	})

	if err != nil {
		return model.Project{}, err
	}

	err = s.Permissions.AddAdminToProject(ctx, user, newProject)

	if err != nil {
		return model.Project{}, err
	}

	err = s.Permissions.AddProjectToOrg(ctx, newProject, project.Organization)

	if err != nil {
		return model.Project{}, err
	}

	return newProject, nil
}

func (s Service) List(ctx context.Context) ([]model.Project, error) {
	return s.Store.ListProject(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate model.Project) (model.Project, error) {
	return s.Store.UpdateProject(ctx, toUpdate)
}

func (s Service) AddAdmin(ctx context.Context, id string, userIds []string) ([]model.User, error) {
	currentUser, err := s.Permissions.FetchCurrentUser(ctx)
	if err != nil {
		return []model.User{}, err
	}

	id = strings.TrimSpace(id)
	project, err := s.Store.GetProject(ctx, id)

	if err != nil {
		return []model.User{}, err
	}

	isAuthorized, err := s.Permissions.CheckPermission(ctx, currentUser, model.Resource{
		Idxa:      project.Id,
		Namespace: definition.ProjectNamespace,
	},
		definition.ManageProjectAction,
	)

	if err != nil {
		return []model.User{}, err
	}

	if !isAuthorized {
		return []model.User{}, shieldError.Unauthorzied
	}

	users, err := s.Store.GetUsersByIds(ctx, userIds)

	if err != nil {
		return []model.User{}, err
	}

	for _, user := range users {
		err = s.Permissions.AddAdminToProject(ctx, user, project)
		if err != nil {
			return []model.User{}, err
		}
	}
	return s.ListAdmins(ctx, id)
}

func (s Service) ListAdmins(ctx context.Context, id string) ([]model.User, error) {
	return s.Store.ListProjectAdmins(ctx, id)
}

func (s Service) RemoveAdmin(ctx context.Context, id string, userId string) ([]model.User, error) {
	currentUser, err := s.Permissions.FetchCurrentUser(ctx)
	if err != nil {
		return []model.User{}, err
	}

	id = strings.TrimSpace(id)
	project, err := s.Store.GetProject(ctx, id)

	if err != nil {
		return []model.User{}, err
	}

	isAuthorized, err := s.Permissions.CheckPermission(ctx, currentUser, model.Resource{
		Idxa:      project.Id,
		Namespace: definition.ProjectNamespace,
	},
		definition.ManageProjectAction,
	)

	if err != nil {
		return []model.User{}, err
	}

	if !isAuthorized {
		return []model.User{}, shieldError.Unauthorzied
	}

	user, err := s.Store.GetUser(ctx, userId)

	if err != nil {
		return []model.User{}, err
	}

	err = s.Permissions.RemoveAdminFromProject(ctx, user, project)
	if err != nil {
		return []model.User{}, err
	}

	return s.ListAdmins(ctx, id)
}
