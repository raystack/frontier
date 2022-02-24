package org

import (
	"context"
	"errors"

	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/internal/permission"
	shieldError "github.com/odpf/shield/utils/errors"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store       Store
	Permissions permission.Permissions
}

var (
	OrgDoesntExist = errors.New("org doesn't exist")
	NoAdminsExist  = errors.New("no admins exist")
	InvalidUUID    = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetOrg(ctx context.Context, id string) (model.Organization, error)
	CreateOrg(ctx context.Context, org model.Organization) (model.Organization, error)
	ListOrg(ctx context.Context) ([]model.Organization, error)
	UpdateOrg(ctx context.Context, toUpdate model.Organization) (model.Organization, error)
	GetUsersByIds(ctx context.Context, userIds []string) ([]model.User, error)
	GetUser(ctx context.Context, userId string) (model.User, error)
	ListOrgAdmins(ctx context.Context, id string) ([]model.User, error)
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

func (s Service) AddAdmin(ctx context.Context, id string, userIds []string) ([]model.User, error) {
	currentUser, err := s.Permissions.FetchCurrentUser(ctx)
	if err != nil {
		return []model.User{}, err
	}

	org, err := s.Store.GetOrg(ctx, id)

	if err != nil {
		return []model.User{}, err
	}

	isAuthorized, err := s.Permissions.CheckPermission(ctx, currentUser, model.Resource{
		Id:        id,
		Namespace: definition.OrgNamespace,
	},
		definition.ManageOrganizationAction,
	)

	if err != nil {
		return []model.User{}, err
	}

	if !isAuthorized {
		return []model.User{}, shieldError.Unauthorzied
	}

	users, err := s.Store.GetUsersByIds(ctx, userIds)

	if err != nil {
		return []model.User{}, err
	}

	for _, user := range users {
		err = s.Permissions.AddAdminToOrg(ctx, user, org)
		if err != nil {
			return []model.User{}, err
		}
	}
	return s.ListAdmins(ctx, id)
}

func (s Service) ListAdmins(ctx context.Context, id string) ([]model.User, error) {
	return s.Store.ListOrgAdmins(ctx, id)
}

func (s Service) RemoveAdmin(ctx context.Context, id string, userId string) ([]model.User, error) {
	currentUser, err := s.Permissions.FetchCurrentUser(ctx)
	if err != nil {
		return []model.User{}, err
	}

	org, err := s.Store.GetOrg(ctx, id)

	if err != nil {
		return []model.User{}, err
	}

	isAuthorized, err := s.Permissions.CheckPermission(ctx, currentUser, model.Resource{
		Id:        id,
		Namespace: definition.OrgNamespace,
	},
		definition.ManageOrganizationAction,
	)

	if err != nil {
		return []model.User{}, err
	}

	if !isAuthorized {
		return []model.User{}, shieldError.Unauthorzied
	}

	user, err := s.Store.GetUser(ctx, userId)

	if err != nil {
		return []model.User{}, err
	}

	err = s.Permissions.RemoveAdminFromOrg(ctx, user, org)
	if err != nil {
		return []model.User{}, err
	}

	return s.ListAdmins(ctx, id)
}
