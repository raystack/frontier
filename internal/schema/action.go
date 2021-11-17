package schema

import (
	"context"
	"errors"
	"time"
)

type Action struct {
	Id        string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var ActionDoesntExist = errors.New("actions doesn't exist")

func (s Service) GetAction(ctx context.Context, id string) (Action, error) {
	return s.Store.GetAction(ctx, id)
}

func (s Service) CreateAction(ctx context.Context, org Action) (Action, error) {
	newAction, err := s.Store.CreateAction(ctx, Action{
		Name: org.Name,
		Slug: org.Slug,
	})

	if err != nil {
		return Action{}, err
	}

	return newAction, nil
}

func (s Service) ListActions(ctx context.Context) ([]Action, error) {
	return s.Store.ListActions(ctx)
}
