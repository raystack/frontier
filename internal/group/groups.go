package group

import (
	"context"
	"errors"
	"time"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store Store
}

type Group struct {
	Id           string
	Name         string
	Slug         string
	Organization model.Organization
	Metadata     map[string]string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Store interface {
	CreateGroup(ctx context.Context, grp Group) (Group, error)
	GetGroup(ctx context.Context, id string) (Group, error)
	ListGroups(ctx context.Context) ([]Group, error)
	UpdateGroup(ctx context.Context, toUpdate Group) (Group, error)
}

var (
	GroupDoesntExist = errors.New("org doesn't exist")
	InvalidUUID      = errors.New("invalid syntax of uuid")
)

func (s Service) CreateGroup(ctx context.Context, grp Group) (Group, error) {
	return s.Store.CreateGroup(ctx, grp)
}

func (s Service) GetGroup(ctx context.Context, id string) (Group, error) {
	return s.Store.GetGroup(ctx, id)
}

func (s Service) ListGroups(ctx context.Context) ([]Group, error) {
	return s.Store.ListGroups(ctx)
}

func (s Service) UpdateGroup(ctx context.Context, grp Group) (Group, error) {
	return s.Store.UpdateGroup(ctx, grp)
}
