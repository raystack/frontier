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
	newAction, err := s.Store.CreateAction(ctx, action)

	if err != nil {
		return model.Action{}, err
	}

	return newAction, nil
}

func (s Service) ListActions(ctx context.Context) ([]model.Action, error) {
	return s.Store.ListActions(ctx)
}

func (s Service) UpdateAction(ctx context.Context, id string, action model.Action) (model.Action, error) {
	updatedAction, err := s.Store.UpdateAction(ctx, model.Action{
		Name: action.Name,
		Id:   id,
	})

	if err != nil {
		return model.Action{}, err
	}

	return updatedAction, nil
}
