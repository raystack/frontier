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

func (s Service) CreateNamespace(ctx context.Context, org model.Namespace) (model.Namespace, error) {
	newNamespace, err := s.Store.CreateNamespace(ctx, model.Namespace{
		Name: org.Name,
		Slug: org.Slug,
	})

	if err != nil {
		return model.Namespace{}, err
	}

	return newNamespace, nil
}

func (s Service) ListNamespaces(ctx context.Context) ([]model.Namespace, error) {
	return s.Store.ListNamespaces(ctx)
}
