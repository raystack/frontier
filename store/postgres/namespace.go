package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/shield/internal/schema"
)

type Namespace struct {
	Id        string    `db:"id"`
	Name      string    `db:"name"`
	Slug      string    `db:"slug"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

const (
	getNamespaceQuery             = `SELECT id, name, slug, created_at, updated_at from namespaces where id=$1;`
	createNamespaceQuery          = `INSERT INTO namespaces(name, slug) values($1, $2) RETURNING id, name, slug, created_at, updated_at;`
	listNamespacesQuery           = `SELECT id, name, slug, created_at, updated_at from namespaces;`
	selectNamespaceForUpdateQuery = `SELECT id, name, slug, version, updated_at from namespaces where id=$1;`
	updateNamespaceQuery          = `UPDATE namespaces set name = $2, slug = $3 updated_at = now() where id = $1 RETURNING id, name, slug, created_at, updated_at;`
)

func (s Store) GetNamespace(ctx context.Context, id string) (schema.Namespace, error) {
	fetchedNamespace, err := s.selectNamespace(ctx, id, false, nil)
	return fetchedNamespace, err
}

func (s Store) selectNamespace(ctx context.Context, id string, forUpdate bool, txn *sqlx.Tx) (schema.Namespace, error) {
	var fetchedNamespace Namespace

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		if forUpdate {
			return txn.GetContext(ctx, &fetchedNamespace, selectNamespaceForUpdateQuery, id)
		} else {
			return s.DB.GetContext(ctx, &fetchedNamespace, getNamespaceQuery, id)
		}
	})

	if errors.Is(err, sql.ErrNoRows) {
		return schema.Namespace{}, schema.NamespaceDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return schema.Namespace{}, schema.InvalidUUID
	} else if err != nil {
		return schema.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedNamespace, err := transformToNamespace(fetchedNamespace)
	if err != nil {
		return schema.Namespace{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedNamespace, nil
}

func (s Store) CreateNamespace(ctx context.Context, namespaceToCreate schema.Namespace) (schema.Namespace, error) {

	var newNamespace Namespace
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newNamespace, createNamespaceQuery, namespaceToCreate.Name, namespaceToCreate.Slug)
	})

	if err != nil {
		return schema.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedNamespace, err := transformToNamespace(newNamespace)
	if err != nil {
		return schema.Namespace{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedNamespace, nil
}

func (s Store) ListNamespaces(ctx context.Context) ([]schema.Namespace, error) {
	var fetchedNamespaces []Namespace
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedNamespaces, listNamespacesQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []schema.Namespace{}, schema.NamespaceDoesntExist
	}

	if err != nil {
		return []schema.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedNamespaces []schema.Namespace

	for _, o := range fetchedNamespaces {
		transformedNamespace, err := transformToNamespace(o)
		if err != nil {
			return []schema.Namespace{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedNamespaces = append(transformedNamespaces, transformedNamespace)
	}

	return transformedNamespaces, nil
}

func transformToNamespace(from Namespace) (schema.Namespace, error) {

	return schema.Namespace{
		Id:        from.Id,
		Name:      from.Name,
		Slug:      from.Slug,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
