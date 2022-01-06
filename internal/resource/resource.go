package resource

import (
	"context"
	"errors"
	"fmt"
	"github.com/odpf/shield/internal/permission"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store       Store
	Permissions permission.Permissions
}

var (
	ResourceDoesntExist = errors.New("resource doesn't exist")
	InvalidUUID         = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetResource(ctx context.Context, id string) (model.Resource, error)
	CreateResource(ctx context.Context, resource model.Resource) (model.Resource, error)
	ListResources(ctx context.Context) ([]model.Resource, error)
	UpdateResource(ctx context.Context, id string, resource model.Resource) (model.Resource, error)
}

func (s Service) Get(ctx context.Context, id string) (model.Resource, error) {
	return s.Store.GetResource(ctx, id)
}

func (s Service) Create(ctx context.Context, resource model.Resource) (model.Resource, error) {
	id := createResourceUrl(resource)
	newResource, err := s.Store.CreateResource(ctx, model.Resource{
		Id:             id,
		Name:           resource.Name,
		OrganizationId: resource.OrganizationId,
		ProjectId:      resource.ProjectId,
		GroupId:        resource.GroupId,
		NamespaceId:    resource.NamespaceId,
	})

	if err != nil {
		return model.Resource{}, err
	}

	err = s.Permissions.AddTeamToResource(ctx, model.Group{Id: resource.GroupId}, newResource)

	if err != nil {
		return model.Resource{}, err
	}

	err = s.Permissions.AddProjectToResource(ctx, model.Project{Id: resource.ProjectId}, newResource)

	if err != nil {
		return model.Resource{}, err
	}

	err = s.Permissions.AddOrgToResource(ctx, model.Organization{Id: resource.OrganizationId}, newResource)

	if err != nil {
		return model.Resource{}, err
	}

	return newResource, nil
}

func (s Service) List(ctx context.Context) ([]model.Resource, error) {
	return s.Store.ListResources(ctx)
}

func (s Service) Update(ctx context.Context, id string, resource model.Resource) (model.Resource, error) {
	return s.Store.UpdateResource(ctx, id, resource)
}

func createResourceUrl(resource model.Resource) string {
	//return fmt.Sprintf("organizations/%s/projects/%s/teams/%s/%s/%s", resource.OrganizationId, resource.ProjectId, resource.GroupId, resource.NamespaceId, resource.Name)
	return fmt.Sprintf("%s/%s", resource.NamespaceId, resource.Name)
}
