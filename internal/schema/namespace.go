package schema

import (
	"context"
	"errors"

	"github.com/odpf/shield/model"
)

var NamespaceDoesntExist = errors.New("actions doesn't exist")

func (s Service) GetNamespace(ctx context.Context, id string) (model.Namespace, error) {
	return s.Store.GetNamespace(ctx, id)
}

func (s Service) CreateNamespace(ctx context.Context, ns model.Namespace) (model.Namespace, error) {
	newNamespace, err := s.Store.CreateNamespace(ctx, model.Namespace{
		Name: ns.Name,
		Id:   ns.Id,
	})

	if err != nil {
		return model.Namespace{}, err
	}

	return newNamespace, nil
}

func (s Service) ListNamespaces(ctx context.Context) ([]model.Namespace, error) {
	return s.Store.ListNamespaces(ctx)
}
