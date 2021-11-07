package project

import (
	"context"
	"errors"
	"time"
)

type Project struct {
	Id        string
	Name      string
	Slug      string
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Service struct {
	Store Store
}

var (
	ProjectDoesntExist = errors.New("project doesn't exist")
	InvalidUUID        = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetProject(ctx context.Context, id string) (Project, error)
	CreateProject(ctx context.Context, org Project) (Project, error)
	ListProject(ctx context.Context) ([]Project, error)
	UpdateProject(ctx context.Context, toUpdate Project) (Project, error)
}

func (s Service) GetProject(ctx context.Context, id string) (Project, error) {
	return s.Store.GetProject(ctx, id)
}

func (s Service) CreateProject(ctx context.Context, project Project) (Project, error) {
	newOrg, err := s.Store.CreateProject(ctx, Project{
		Name:     project.Name,
		Slug:     project.Slug,
		Metadata: project.Metadata,
	})

	if err != nil {
		return Project{}, err
	}

	return newOrg, nil
}

func (s Service) ListProject(ctx context.Context) ([]Project, error) {
	return s.Store.ListProject(ctx)
}

func (s Service) UpdateProject(ctx context.Context, toUpdate Project) (Project, error) {
	return s.Store.UpdateProject(ctx, toUpdate)
}
