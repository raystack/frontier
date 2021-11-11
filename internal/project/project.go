package project

import (
	"context"
	"errors"

	modelv1 "github.com/odpf/shield/model/v1"
)

type Service struct {
	Store Store
}

var (
	ProjectDoesntExist = errors.New("project doesn't exist")
	InvalidUUID        = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetProject(ctx context.Context, id string) (modelv1.Project, error)
	CreateProject(ctx context.Context, org modelv1.Project) (modelv1.Project, error)
	ListProject(ctx context.Context) ([]modelv1.Project, error)
	UpdateProject(ctx context.Context, toUpdate modelv1.Project) (modelv1.Project, error)
}

func (s Service) Get(ctx context.Context, id string) (modelv1.Project, error) {
	return s.Store.GetProject(ctx, id)
}

func (s Service) Create(ctx context.Context, project modelv1.Project) (modelv1.Project, error) {
	newOrg, err := s.Store.CreateProject(ctx, modelv1.Project{
		Name:         project.Name,
		Slug:         project.Slug,
		Metadata:     project.Metadata,
		Organization: project.Organization,
	})

	if err != nil {
		return modelv1.Project{}, err
	}

	return newOrg, nil
}

func (s Service) List(ctx context.Context) ([]modelv1.Project, error) {
	return s.Store.ListProject(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate modelv1.Project) (modelv1.Project, error) {
	return s.Store.UpdateProject(ctx, toUpdate)
}
