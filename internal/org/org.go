package org

import (
	"context"
	"errors"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store Store
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
}

func (s Service) Get(ctx context.Context, id string) (model.Organization, error) {
	return s.Store.GetOrg(ctx, id)
}

func (s Service) Create(ctx context.Context, org model.Organization) (model.Organization, error) {
	newOrg, err := s.Store.CreateOrg(ctx, model.Organization{
		Name:     org.Name,
		Slug:     org.Slug,
		Metadata: org.Metadata,
	})

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
