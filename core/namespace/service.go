package namespace

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

func (s Service) GetNamespace(ctx context.Context, id string) (Namespace, error) {
	return s.store.GetNamespace(ctx, id)
}

func (s Service) CreateNamespace(ctx context.Context, ns Namespace) (Namespace, error) {
	return s.store.CreateNamespace(ctx, ns)
}

func (s Service) ListNamespaces(ctx context.Context) ([]Namespace, error) {
	return s.store.ListNamespaces(ctx)
}

func (s Service) UpdateNamespace(ctx context.Context, id string, ns Namespace) (Namespace, error) {
	updatedNamespace, err := s.store.UpdateNamespace(ctx, id, Namespace{
		Name: ns.Name,
		ID:   ns.ID,
	})
	if err != nil {
		return Namespace{}, err
	}
	return updatedNamespace, nil
}
