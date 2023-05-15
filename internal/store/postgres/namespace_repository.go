package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	shielduuid "github.com/odpf/shield/pkg/uuid"

	"github.com/doug-martin/goqu/v9"
	newrelic "github.com/newrelic/go-agent"
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
	if strings.TrimSpace(id) == "" {
		return namespace.Namespace{}, namespace.ErrInvalidID
	}

	stmt := dialect.Select(&Namespace{}).From(TABLE_NAMESPACES)
	if shielduuid.IsValid(id) {
		stmt = stmt.Where(goqu.Ex{
			"id": id,
		})
	} else {
		stmt = stmt.Where(goqu.Ex{
			"name": id,
		})
	}
	query, params, err := stmt.ToSQL()
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

	return fetchedNamespace.transformToNamespace()
}

// Upsert inserts a new namespace if it doesn't exist. If it does, update the details and return no error
func (r NamespaceRepository) Upsert(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error) {
	if strings.TrimSpace(ns.Name) == "" {
		return namespace.Namespace{}, namespace.ErrInvalidDetail
	}
	if ns.ID == "" {
		ns.ID = uuid.New().String()
	}
	marshaledMetadata, err := json.Marshal(ns.Metadata)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("namespace metadata: %w: %s", parseErr, err)
	}
	query, params, err := dialect.Insert(TABLE_NAMESPACES).Rows(
		goqu.Record{
			"id":       ns.ID,
			"name":     ns.Name,
			"metadata": marshaledMetadata,
		}).OnConflict(
		goqu.DoUpdate("name", goqu.Record{
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		})).Returning(&Namespace{}).ToSQL()
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var nsModel Namespace
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_NAMESPACES,
				Operation:  "Upsert",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&nsModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrDuplicateKey):
			return namespace.Namespace{}, namespace.ErrConflict
		default:
			return namespace.Namespace{}, err
		}
	}

	return nsModel.transformToNamespace()
}

func (r NamespaceRepository) List(ctx context.Context) ([]namespace.Namespace, error) {
	query, params, err := dialect.Select(&Namespace{}).From(TABLE_NAMESPACES).ToSQL()
	if err != nil {
		return []namespace.Namespace{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedNamespaces []Namespace
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_NAMESPACES,
				Operation:  "List",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.SelectContext(ctx, &fetchedNamespaces, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []namespace.Namespace{}, nil
		}
		return []namespace.Namespace{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedNamespaces []namespace.Namespace
	for _, o := range fetchedNamespaces {
		res, err := o.transformToNamespace()
		if err != nil {
			return nil, fmt.Errorf("failed to transform namespace: %w", err)
		}
		transformedNamespaces = append(transformedNamespaces, res)
	}

	return transformedNamespaces, nil
}

func (r NamespaceRepository) Update(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error) {
	if strings.TrimSpace(ns.ID) == "" {
		return namespace.Namespace{}, namespace.ErrInvalidID
	}

	marshaledMetadata, err := json.Marshal(ns.Metadata)
	if err != nil {
		return namespace.Namespace{}, fmt.Errorf("namespace metadata: %w: %s", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_NAMESPACES).Set(
		goqu.Record{
			"metadata":   marshaledMetadata,
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
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_NAMESPACES,
				Operation:  "Update",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&nsModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return namespace.Namespace{}, namespace.ErrNotExist
		case errors.Is(err, ErrDuplicateKey):
			return namespace.Namespace{}, namespace.ErrConflict
		default:
			return namespace.Namespace{}, err
		}
	}

	return nsModel.transformToNamespace()
}
