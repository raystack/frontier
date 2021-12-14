package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store Store
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
	return s.Store.CreateResource(ctx, model.Resource{
		Id:             id,
		Name:           resource.Name,
		OrganizationId: resource.OrganizationId,
		ProjectId:      resource.ProjectId,
		GroupId:        resource.GroupId,
		NamespaceId:    resource.NamespaceId,
	})
}

func (s Service) List(ctx context.Context) ([]model.Resource, error) {
	return s.Store.ListResources(ctx)
}

func (s Service) Update(ctx context.Context, id string, resource model.Resource) (model.Resource, error) {
	return s.Store.UpdateResource(ctx, id, resource)
}

func createResourceUrl(resource model.Resource) string {
	return fmt.Sprintf("organizations/%s/projects/%s/teams/%s/%s/%s", resource.OrganizationId, resource.ProjectId, resource.GroupId, resource.NamespaceId, resource.Name)
}
