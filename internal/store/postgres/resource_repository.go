package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/uuid"
)

type ResourceRepository struct {
	dbc *db.Client
}

func NewResourceRepository(dbc *db.Client) *ResourceRepository {
	return &ResourceRepository{
		dbc: dbc,
	}
}

func (r ResourceRepository) Create(ctx context.Context, res resource.Resource) (resource.Resource, error) {
	if strings.TrimSpace(res.URN) == "" {
		return resource.Resource{}, resource.ErrInvalidURN
	}

	userID := sql.NullString{String: res.UserID, Valid: res.UserID != ""}
	groupID := sql.NullString{String: res.GroupID, Valid: res.GroupID != ""}

	query, params, err := dialect.Insert(TABLE_RESOURCES).Rows(
		goqu.Record{
			"urn":          res.URN,
			"name":         res.Name,
			"project_id":   res.ProjectID,
			"group_id":     groupID,
			"org_id":       res.OrganizationID,
			"namespace_id": res.NamespaceID,
			"user_id":      userID,
		}).OnConflict(
		goqu.DoUpdate("ON CONSTRAINT resources_urn_unique", goqu.Record{
			"name":         res.Name,
			"project_id":   res.ProjectID,
			"group_id":     groupID,
			"org_id":       res.OrganizationID,
			"namespace_id": res.NamespaceID,
			"user_id":      userID,
		})).Returning(&ResourceCols{}).ToSQL()
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var resourceModel Resource
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&resourceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errForeignKeyViolation):
			return resource.Resource{}, resource.ErrInvalidDetail
		case errors.Is(err, errInvalidTexRepresentation):
			return resource.Resource{}, resource.ErrInvalidUUID
		default:
			return resource.Resource{}, err
		}
	}

	return resourceModel.transformToResource(), nil
}

func (r ResourceRepository) List(ctx context.Context, flt resource.Filter) ([]resource.Resource, error) {
	var fetchedResources []Resource

	sqlStatement := dialect.From(TABLE_RESOURCES)
	if flt.ProjectID != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"project_id": flt.ProjectID})
	}
	if flt.GroupID != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"group_id": flt.GroupID})
	}
	if flt.OrganizationID != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"org_id": flt.OrganizationID})
	}
	if flt.NamespaceID != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"namespace_id": flt.NamespaceID})
	}
	query, params, err := sqlStatement.ToSQL()
	if err != nil {
		return nil, err
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedResources, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return []resource.Resource{}, nil
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return []resource.Resource{}, nil
		}
		return []resource.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedResources []resource.Resource
	for _, r := range fetchedResources {
		transformedResources = append(transformedResources, r.transformToResource())
	}

	return transformedResources, nil
}

func (r ResourceRepository) GetByID(ctx context.Context, id string) (resource.Resource, error) {
	if strings.TrimSpace(id) == "" {
		return resource.Resource{}, resource.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_RESOURCES).Where(goqu.Ex{
		"id": id,
	}).ToSQL()
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var resourceModel Resource
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &resourceModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return resource.Resource{}, resource.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return resource.Resource{}, resource.ErrInvalidUUID
		default:
			return resource.Resource{}, err
		}
	}

	return resourceModel.transformToResource(), nil
}

func (r ResourceRepository) Update(ctx context.Context, id string, res resource.Resource) (resource.Resource, error) {
	if strings.TrimSpace(id) == "" {
		return resource.Resource{}, resource.ErrInvalidID
	}

	if !uuid.IsValid(id) {
		return resource.Resource{}, resource.ErrInvalidUUID
	}

	if strings.TrimSpace(res.URN) == "" {
		return resource.Resource{}, resource.ErrInvalidURN
	}

	userID := sql.NullString{String: res.UserID, Valid: res.UserID != ""}
	groupID := sql.NullString{String: res.GroupID, Valid: res.GroupID != ""}

	query, params, err := dialect.Update(TABLE_RESOURCES).Set(
		goqu.Record{
			"name":         res.Name,
			"project_id":   res.ProjectID,
			"group_id":     groupID,
			"org_id":       res.OrganizationID,
			"namespace_id": res.NamespaceID,
			"user_id":      userID,
			"urn":          res.URN,
		},
	).Where(goqu.Ex{
		"id": id,
	}).Returning(&ResourceCols{}).ToSQL()
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var resourceModel Resource
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&resourceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return resource.Resource{}, resource.ErrNotExist
		case errors.Is(err, errDuplicateKey):
			return resource.Resource{}, resource.ErrConflict
		case errors.Is(err, errForeignKeyViolation):
			return resource.Resource{}, resource.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return resource.Resource{}, resource.ErrInvalidDetail
		default:
			return resource.Resource{}, err
		}
	}

	return resourceModel.transformToResource(), nil
}

func (r ResourceRepository) GetByURN(ctx context.Context, urn string) (resource.Resource, error) {
	if strings.TrimSpace(urn) == "" {
		return resource.Resource{}, resource.ErrInvalidURN
	}

	query, params, err := dialect.Select(&ResourceCols{}).From(TABLE_RESOURCES).Where(
		goqu.Ex{
			"urn": urn,
		}).ToSQL()
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var resourceModel Resource
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &resourceModel, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return resource.Resource{}, resource.ErrNotExist
		}
		return resource.Resource{}, err
	}

	return resourceModel.transformToResource(), nil
}

func buildGetResourcesByNamespaceQuery(dialect goqu.DialectWrapper, name string, ns namespace.Namespace) (string, interface{}, error) {
	namespaceQueryExpression := goqu.Ex{
		"backend": goqu.L(ns.Backend),
	}

	if ns.ResourceType != "" {
		namespaceQueryExpression["resource_type"] = goqu.L(ns.ResourceType)
	}

	getNamespaceQuery := dialect.Select("id").From(TABLE_NAMESPACES).Where(namespaceQueryExpression)
	getResourcesByURNQuery, params, err := dialect.Select(&ResourceCols{}).From(TABLE_RESOURCES).Where(goqu.Ex{
		"name":         goqu.L(name),
		"namespace_id": goqu.Op{"in": getNamespaceQuery},
	}).ToSQL()

	return getResourcesByURNQuery, params, err
}

func (r ResourceRepository) GetByNamespace(ctx context.Context, name string, ns namespace.Namespace) (resource.Resource, error) {
	var fetchedResource Resource

	query, params, err := buildGetResourcesByNamespaceQuery(dialect, name, ns)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedResource, query, params)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return resource.Resource{}, resource.ErrNotExist
	} else if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	if err != nil {
		return resource.Resource{}, err
	}

	return fetchedResource.transformToResource(), nil
}
