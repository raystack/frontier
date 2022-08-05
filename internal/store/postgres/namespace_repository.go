package postgres

import (
	"context"
	"errors"
	"fmt"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/str"
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
	if str.IsStringEmpty(id) {
		return namespace.Namespace{}, namespace.ErrInvalidID
	}

	query, params, err := dialect.Select(&Namespace{}).From(TABLE_NAMESPACES).Where(goqu.Ex{
		"id": id,
	}).ToSQL()
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedNamespace Namespace
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedNamespace, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return namespace.Namespace{}, namespace.ErrNotExist
		}
		return namespace.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return fetchedNamespace.transformToNamespace(), nil
}

// TODO this is actually an upsert
func (r NamespaceRepository) Create(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error) {
	if str.IsStringEmpty(ns.ID) {
		return namespace.Namespace{}, namespace.ErrInvalidID
	}

	if str.IsStringEmpty(ns.Name) {
		return namespace.Namespace{}, namespace.ErrInvalidDetail
	}

	query, params, err := dialect.Insert(TABLE_NAMESPACES).Rows(
		goqu.Record{
			"id":   ns.ID,
			"name": ns.Name,
		}).OnConflict(
		goqu.DoUpdate("id", goqu.Record{
			"name":       ns.Name,
			"updated_at": goqu.L("now()"),
		})).Returning(&Namespace{}).ToSQL()
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var nsModel Namespace
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&nsModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errDuplicateKey):
			return namespace.Namespace{}, namespace.ErrConflict
		default:
			return namespace.Namespace{}, err
		}
	}

	return nsModel.transformToNamespace(), nil
}

func (r NamespaceRepository) List(ctx context.Context) ([]namespace.Namespace, error) {
	query, params, err := dialect.Select(&Namespace{}).From(TABLE_NAMESPACES).ToSQL()
	if err != nil {
		return []namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedNamespaces []Namespace
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedNamespaces, query, params...)
	}); err != nil {
		// should not throw error but return empty instead
		if errors.Is(err, sql.ErrNoRows) {
			return []namespace.Namespace{}, nil
		}
		return []namespace.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedNamespaces []namespace.Namespace
	for _, o := range fetchedNamespaces {
		transformedNamespaces = append(transformedNamespaces, o.transformToNamespace())
	}

	return transformedNamespaces, nil
}

func (r NamespaceRepository) Update(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error) {
	if str.IsStringEmpty(ns.ID) {
		return namespace.Namespace{}, namespace.ErrInvalidID
	}

	if str.IsStringEmpty(ns.Name) {
		return namespace.Namespace{}, namespace.ErrInvalidDetail
	}

	query, params, err := dialect.Update(TABLE_NAMESPACES).Set(
		goqu.Record{
			"id":         ns.ID,
			"name":       ns.Name,
			"updated_at": goqu.L("now()"),
		}).Where(
		goqu.Ex{
			"id": ns.ID,
		},
	).Returning(&Namespace{}).ToSQL()
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var nsModel Namespace
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&nsModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return namespace.Namespace{}, namespace.ErrNotExist
		case errors.Is(err, errDuplicateKey):
			return namespace.Namespace{}, namespace.ErrConflict
		default:
			return namespace.Namespace{}, err
		}
	}

	return nsModel.transformToNamespace(), nil
}
