package org

import (
	"context"
	"errors"

	"github.com/odpf/shield/internal/permission"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store       Store
	Permissions permission.Permissions
}

var (
	OrgDoesntExist = errors.New("org doesn't exist")
	InvalidUUID    = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetOrg(ctx context.Context, id string) (model.Organization, error)
	CreateOrg(ctx context.Context, org model.Organization) (model.Organization, error)
	ListOrg(ctx context.Context) ([]model.Organization, error)
	UpdateOrg(ctx context.Context, toUpdate model.Organization) (model.Organization, error)
	AddOrgAdmin(ctx context.Context, id string, toAdd []model.User) ([]model.User, error)
}

func (s Service) Get(ctx context.Context, id string) (model.Organization, error) {
	return s.Store.GetOrg(ctx, id)
}

func (s Service) Create(ctx context.Context, org model.Organization) (model.Organization, error) {
	user, err := s.Permissions.FetchCurrentUser(ctx)
	if err != nil {
		return model.Organization{}, err
	}

	newOrg, err := s.Store.CreateOrg(ctx, model.Organization{
		Name:     org.Name,
		Slug:     org.Slug,
		Metadata: org.Metadata,
	})

	if err != nil {
		return model.Organization{}, err
	}

	err = s.Permissions.AddAdminToOrg(ctx, user, newOrg)

	if err != nil {
		return model.Organization{}, err
	}

	return newOrg, nil
}

func (s Service) List(ctx context.Context) ([]model.Organization, error) {
	return s.Store.ListOrg(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate model.Organization) (model.Organization, error) {
	return s.Store.UpdateOrg(ctx, toUpdate)
}

func (s Service) AddAdmin(ctx context.Context, id string, toAdd []model.User) ([]model.User, error) {
	return s.Store.AddOrgAdmin(ctx, id, toAdd)
}
