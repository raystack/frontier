package group

import (
	"context"
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
}

func (s Service) CreateGroup(ctx context.Context, grp Group) (Group, error) {
	newGroup, err := s.Store.CreateGroup(ctx, grp)
	if err != nil {
		return Group{}, err
	}

	return newGroup, nil
}
