package role

import (
	"context"
	"fmt"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s Service) Create(ctx context.Context, toCreate Role) (Role, error) {
	roleID, err := s.repository.Create(ctx, toCreate)
	if err != nil {
		fmt.Printf("err creating role: %v\n", err)
		return Role{}, err
	}
	return s.repository.Get(ctx, roleID)
}

func (s Service) Get(ctx context.Context, id string) (Role, error) {
	return s.repository.Get(ctx, id)
}

func (s Service) List(ctx context.Context) ([]Role, error) {
	return s.repository.List(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate Role) (Role, error) {
	roleID, err := s.repository.Update(ctx, toUpdate)
	if err != nil {
		return Role{}, err
	}
	return s.repository.Get(ctx, roleID)
}
