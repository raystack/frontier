package schema

import (
	"context"
	"errors"

	"github.com/odpf/shield/model"
)

var ActionDoesntExist = errors.New("actions doesn't exist")

func (s Service) GetAction(ctx context.Context, id string) (model.Action, error) {
	return s.Store.GetAction(ctx, id)
}

func (s Service) CreateAction(ctx context.Context, action model.Action) (model.Action, error) {
	newAction, err := s.Store.CreateAction(ctx, model.Action{
		Name: action.Name,
		Id:   action.Id,
	})

	if err != nil {
		return model.Action{}, err
	}

	return newAction, nil
}

func (s Service) ListActions(ctx context.Context) ([]model.Action, error) {
	return s.Store.ListActions(ctx)
}
