package group

import (
	"context"
	"errors"

	"github.com/odpf/shield/internal/authz"
	"github.com/odpf/shield/internal/permission"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store       Store
	Authz       *authz.Authz
	Permissions permission.Permissions
}

type Store interface {
	CreateGroup(ctx context.Context, grp model.Group) (model.Group, error)
	GetGroup(ctx context.Context, id string) (model.Group, error)
	ListGroups(ctx context.Context) ([]model.Group, error)
	UpdateGroup(ctx context.Context, toUpdate model.Group) (model.Group, error)
}

var (
	GroupDoesntExist = errors.New("group doesn't exist")
	InvalidUUID      = errors.New("invalid syntax of uuid")
)

func (s Service) CreateGroup(ctx context.Context, grp model.Group) (model.Group, error) {
	user, err := s.Permissions.FetchCurrentUser(ctx)

	if err != nil {
		return model.Group{}, err
	}

	newGroup, err := s.Store.CreateGroup(ctx, grp)

	if err != nil {
		return model.Group{}, err
	}

	err = s.Permissions.AddTeamToOrg(ctx, newGroup, model.Organization{Id: grp.OrganizationId})

	if err != nil {
		return model.Group{}, err
	}

	err = s.Permissions.AddAdminToTeam(ctx, user, newGroup)

	if err != nil {
		return model.Group{}, err
	}

	return newGroup, nil
}

func (s Service) GetGroup(ctx context.Context, id string) (model.Group, error) {
	return s.Store.GetGroup(ctx, id)
}

func (s Service) ListGroups(ctx context.Context) ([]model.Group, error) {
	return s.Store.ListGroups(ctx)
}

func (s Service) UpdateGroup(ctx context.Context, grp model.Group) (model.Group, error) {
	return s.Store.UpdateGroup(ctx, grp)
}
