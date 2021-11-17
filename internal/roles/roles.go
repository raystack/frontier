package roles

import (
	"context"
	"errors"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store Store
}

var (
	RoleDoesntExist = errors.New("role doesn't exist")
	InvalidUUID     = errors.New("invalid syntax of uuid")
)

type Store interface {
	CreateRole(ctx context.Context, role model.Role) (model.Role, error)
	GetRole(ctx context.Context, id string) (model.Role, error)
	ListRoles(ctx context.Context) ([]model.Role, error)
	UpdateRole(ctx context.Context, toUpdate model.Role) (model.Role, error)
}

func (s Service) Create(ctx context.Context, toCreate model.Role) (model.Role, error) {
	return s.Store.CreateRole(ctx, toCreate)
}

func (s Service) Get(ctx context.Context, id string) (model.Role, error) {
	return s.Store.GetRole(ctx, id)
}

func (s Service) List(ctx context.Context) ([]model.Role, error) {
	return s.Store.ListRoles(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate model.Role) (model.Role, error) {
	return s.Store.UpdateRole(ctx, toUpdate)
}
