package org

import (
	"context"
	"errors"
	"time"
)

type Organization struct {
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
	OrgDoesntExist = errors.New("org doesn't exist")
	InvalidUUID    = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetOrg(ctx context.Context, id string) (Organization, error)
	CreateOrg(ctx context.Context, org Organization) (Organization, error)
	ListOrg(ctx context.Context) ([]Organization, error)
	UpdateOrg(ctx context.Context, toUpdate Organization) (Organization, error)
}

func (s Service) GetOrganization(ctx context.Context, id string) (Organization, error) {
	return s.Store.GetOrg(ctx, id)
}

func (s Service) CreateOrganization(ctx context.Context, org Organization) (Organization, error) {
	newOrg, err := s.Store.CreateOrg(ctx, Organization{
		Name:     org.Name,
		Slug:     org.Slug,
		Metadata: org.Metadata,
	})

	if err != nil {
		return Organization{}, err
	}

	return newOrg, nil
}

func (s Service) ListOrganizations(ctx context.Context) ([]Organization, error) {
	return s.Store.ListOrg(ctx)
}

func (s Service) UpdateOrganization(ctx context.Context, toUpdate Organization) (Organization, error) {
	return s.Store.UpdateOrg(ctx, toUpdate)
}
