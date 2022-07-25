package postgres

import (
	"context"
	"errors"
	"fmt"

	"database/sql"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/pkg/db"
)

type NamespaceRepository struct {
	dbc *db.Client
}

func NewNamespaceRepository(dbc *db.Client) *NamespaceRepository {
	return &NamespaceRepository{
		dbc: dbc,
	}
}

func (r NamespaceRepository) Get(ctx context.Context, id string) (namespace.Namespace, error) {
	var fetchedNamespace Namespace
	getNamespaceQuery, err := buildGetNamespaceQuery(dialect)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedNamespace, getNamespaceQuery, id)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return namespace.Namespace{}, namespace.ErrNotExist
		}
		return namespace.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedNamespace, err := transformToNamespace(fetchedNamespace)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedNamespace, nil
}

// TODO this is actually an upsert
func (r NamespaceRepository) Create(ctx context.Context, namespaceToCreate namespace.Namespace) (namespace.Namespace, error) {
	if namespaceToCreate.ID == "" {
		return namespace.Namespace{}, namespace.ErrInvalidID
	}

	createNamespaceQuery, err := buildCreateNamespaceQuery(dialect)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var nsModel Namespace
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, createNamespaceQuery, namespaceToCreate.ID, namespaceToCreate.Name).StructScan(&nsModel)
	}); err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedNamespace, err := transformToNamespace(nsModel)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedNamespace, nil
}

func (r NamespaceRepository) List(ctx context.Context) ([]namespace.Namespace, error) {
	var fetchedNamespaces []Namespace
	listNamespacesQuery, err := buildListNamespacesQuery(dialect)
	if err != nil {
		return []namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedNamespaces, listNamespacesQuery)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []namespace.Namespace{}, namespace.ErrNotExist
		}
		return []namespace.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedNamespaces []namespace.Namespace

	for _, o := range fetchedNamespaces {
		transformedNamespace, err := transformToNamespace(o)
		if err != nil {
			return []namespace.Namespace{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedNamespaces = append(transformedNamespaces, transformedNamespace)
	}

	return transformedNamespaces, nil
}

func (r NamespaceRepository) Update(ctx context.Context, toUpdate namespace.Namespace) (namespace.Namespace, error) {
	if toUpdate.ID == "" {
		return namespace.Namespace{}, namespace.ErrInvalidID
	}

	updateNamespaceQuery, err := buildUpdateNamespaceQuery(dialect)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var nsModel Namespace
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, updateNamespaceQuery, toUpdate.ID, toUpdate.ID, toUpdate.Name).StructScan(&nsModel)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return namespace.Namespace{}, namespace.ErrNotExist
		}
		return namespace.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedNamespace, err := transformToNamespace(nsModel)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return transformedNamespace, nil
}
