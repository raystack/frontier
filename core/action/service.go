package action

import (
	"context"
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

func (s Service) GetAction(ctx context.Context, id string) (Action, error) {
	return s.store.GetAction(ctx, id)
}

func (s Service) CreateAction(ctx context.Context, action Action) (Action, error) {
	newAction, err := s.store.CreateAction(ctx, action)

	if err != nil {
		return Action{}, err
	}

	return newAction, nil
}

func (s Service) ListActions(ctx context.Context) ([]Action, error) {
	return s.store.ListActions(ctx)
}

func (s Service) UpdateAction(ctx context.Context, id string, action Action) (Action, error) {
	updatedAction, err := s.store.UpdateAction(ctx, Action{
		Name:        action.Name,
		Id:          id,
		NamespaceId: action.NamespaceId,
	})

	if err != nil {
		return Action{}, err
	}

	return updatedAction, nil
}
