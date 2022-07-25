package policy

import (
	"context"
)

type Service struct {
	store      Store
	authzStore AuthzStore
}

func NewService(store Store, authzStore AuthzStore) *Service {
	return &Service{
		store:      store,
		authzStore: authzStore,
	}
}

func (s Service) GetPolicy(ctx context.Context, id string) (Policy, error) {
	return s.store.GetPolicy(ctx, id)
}

func (s Service) ListPolicies(ctx context.Context) ([]Policy, error) {
	return s.store.ListPolicies(ctx)
}

func (s Service) CreatePolicy(ctx context.Context, policy Policy) ([]Policy, error) {
	policies, err := s.store.CreatePolicy(ctx, policy)
	if err != nil {
		return []Policy{}, err
	}
	err = s.authzStore.AddPolicy(ctx, policies)
	if err != nil {
		return []Policy{}, err
	}
	return policies, err
}

func (s Service) UpdatePolicy(ctx context.Context, id string, policy Policy) ([]Policy, error) {
	policies, err := s.store.UpdatePolicy(ctx, id, policy)
	if err != nil {
		return []Policy{}, err
	}

	err = s.authzStore.AddPolicy(ctx, policies)
	if err != nil {
		return []Policy{}, err
	}
	return policies, err
}
