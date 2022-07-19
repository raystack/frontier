package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"

	"github.com/odpf/shield/core/namespace"
)

type Namespace struct {
	Id        string       `db:"id"`
	Name      string       `db:"name"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func buildGetNamespaceQuery(dialect goqu.DialectWrapper) (string, error) {
	getNamespaceQuery, _, err := dialect.From(TABLE_NAMESPACE).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).ToSQL()

	return getNamespaceQuery, err
}
func buildCreateNamespaceQuery(dialect goqu.DialectWrapper) (string, error) {
	createNamespaceQuery, _, err := dialect.Insert(TABLE_NAMESPACE).Rows(
		goqu.Record{
			"id":   goqu.L("$1"),
			"name": goqu.L("$2"),
		}).OnConflict(goqu.DoUpdate("id", goqu.Record{
		"name": goqu.L("$2"),
	})).Returning(&Namespace{}).ToSQL()

	return createNamespaceQuery, err
}
func buildListNamespacesQuery(dialect goqu.DialectWrapper) (string, error) {
	listNamespacesQuery, _, err := dialect.From(TABLE_NAMESPACE).ToSQL()

	return listNamespacesQuery, err
}
func buildUpdateNamespaceQuery(dialect goqu.DialectWrapper) (string, error) {
	updateNamespaceQuery, _, err := dialect.Update(TABLE_NAMESPACE).Set(
		goqu.Record{
			"id":         goqu.L("$2"),
			"name":       goqu.L("$3"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).Returning(&Namespace{}).ToSQL()

	return updateNamespaceQuery, err
}

func (s Store) GetNamespace(ctx context.Context, id string) (namespace.Namespace, error) {
	fetchedNamespace, err := s.selectNamespace(ctx, id, nil)
	return fetchedNamespace, err
}

func (s Store) selectNamespace(ctx context.Context, id string, txn *sqlx.Tx) (namespace.Namespace, error) {
	var fetchedNamespace Namespace
	getNamespaceQuery, err := buildGetNamespaceQuery(dialect)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedNamespace, getNamespaceQuery, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return namespace.Namespace{}, namespace.ErrNotExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return namespace.Namespace{}, namespace.ErrInvalidUUID
	} else if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedNamespace, err := transformToNamespace(fetchedNamespace)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedNamespace, nil
}

func (s Store) CreateNamespace(ctx context.Context, namespaceToCreate namespace.Namespace) (namespace.Namespace, error) {
	var newNamespace Namespace
	createNamespaceQuery, err := buildCreateNamespaceQuery(dialect)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newNamespace, createNamespaceQuery, namespaceToCreate.Id, namespaceToCreate.Name)
	})

	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedNamespace, err := transformToNamespace(newNamespace)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedNamespace, nil
}

func (s Store) ListNamespaces(ctx context.Context) ([]namespace.Namespace, error) {
	var fetchedNamespaces []Namespace
	listNamespacesQuery, err := buildListNamespacesQuery(dialect)
	if err != nil {
		return []namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedNamespaces, listNamespacesQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []namespace.Namespace{}, namespace.ErrNotExist
	}

	if err != nil {
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

func (s Store) UpdateNamespace(ctx context.Context, id string, toUpdate namespace.Namespace) (namespace.Namespace, error) {
	var updatedNamespace Namespace
	updateNamespaceQuery, err := buildUpdateNamespaceQuery(dialect)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedNamespace, updateNamespaceQuery, id, toUpdate.Id, toUpdate.Name)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return namespace.Namespace{}, namespace.ErrNotExist
	} else if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedNamespace, err := transformToNamespace(updatedNamespace)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return transformedNamespace, nil
}

func transformToNamespace(from Namespace) (namespace.Namespace, error) {
	return namespace.Namespace{
		Id:        from.Id,
		Name:      from.Name,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
