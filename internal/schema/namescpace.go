package schema

import (
	"context"
	"errors"
	"time"
)

type Namespace struct {
	Id        string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var NamespaceDoesntExist = errors.New("actions doesn't exist")

func (s Service) GetNamespace(ctx context.Context, id string) (Namespace, error) {
	return s.Store.GetNamespace(ctx, id)
}

func (s Service) CreateNamespace(ctx context.Context, org Namespace) (Namespace, error) {
	newNamespace, err := s.Store.CreateNamespace(ctx, Namespace{
		Name: org.Name,
		Slug: org.Slug,
	})

	if err != nil {
		return Namespace{}, err
	}

	return newNamespace, nil
}

func (s Service) ListNamespaces(ctx context.Context) ([]Namespace, error) {
	return s.Store.ListNamespaces(ctx)
}
