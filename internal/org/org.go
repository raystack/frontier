package org

import (
	"context"
	"errors"

	modelv1 "github.com/odpf/shield/model/v1"
)

type Service struct {
	Store Store
}

var (
	OrgDoesntExist = errors.New("org doesn't exist")
	InvalidUUID    = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetOrg(ctx context.Context, id string) (modelv1.Organization, error)
	CreateOrg(ctx context.Context, org modelv1.Organization) (modelv1.Organization, error)
	ListOrg(ctx context.Context) ([]modelv1.Organization, error)
	UpdateOrg(ctx context.Context, toUpdate modelv1.Organization) (modelv1.Organization, error)
}

func (s Service) Get(ctx context.Context, id string) (modelv1.Organization, error) {
	return s.Store.GetOrg(ctx, id)
}

func (s Service) Create(ctx context.Context, org modelv1.Organization) (modelv1.Organization, error) {
	newOrg, err := s.Store.CreateOrg(ctx, modelv1.Organization{
		Name:     org.Name,
		Slug:     org.Slug,
		Metadata: org.Metadata,
	})

	if err != nil {
		return modelv1.Organization{}, err
	}

	return newOrg, nil
}

func (s Service) List(ctx context.Context) ([]modelv1.Organization, error) {
	return s.Store.ListOrg(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate modelv1.Organization) (modelv1.Organization, error) {
	return s.Store.UpdateOrg(ctx, toUpdate)
}
