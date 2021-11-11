package project

import (
	"context"
	"errors"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store Store
}

var (
	ProjectDoesntExist = errors.New("project doesn't exist")
	InvalidUUID        = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetProject(ctx context.Context, id string) (model.Project, error)
	CreateProject(ctx context.Context, org model.Project) (model.Project, error)
	ListProject(ctx context.Context) ([]model.Project, error)
	UpdateProject(ctx context.Context, toUpdate model.Project) (model.Project, error)
}

func (s Service) Get(ctx context.Context, id string) (model.Project, error) {
	return s.Store.GetProject(ctx, id)
}

func (s Service) Create(ctx context.Context, project model.Project) (model.Project, error) {
	newOrg, err := s.Store.CreateProject(ctx, model.Project{
		Name:         project.Name,
		Slug:         project.Slug,
		Metadata:     project.Metadata,
		Organization: project.Organization,
	})

	if err != nil {
		return model.Project{}, err
	}

	return newOrg, nil
}

func (s Service) List(ctx context.Context) ([]model.Project, error) {
	return s.Store.ListProject(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate model.Project) (model.Project, error) {
	return s.Store.UpdateProject(ctx, toUpdate)
}
