package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/raystack/frontier/pkg/utils"

	"database/sql"

	goqu "github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/pkg/db"
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
	if strings.TrimSpace(res.ID) == "" {
		res.ID = uuid.New().String()
	} else {
		// check if the uuid is already in use
		existingResource, err := r.GetByID(ctx, res.ID)
		if err != nil {
			if !errors.Is(err, resource.ErrNotExist) {
				return resource.Resource{}, err
			}
		} else {
			if existingResource.ID == res.ID {
				return resource.Resource{}, resource.ErrConflict
			}
		}
	}

	principalID := sql.NullString{String: res.PrincipalID, Valid: res.PrincipalID != ""}
	principalType := sql.NullString{String: res.PrincipalType, Valid: res.PrincipalType != ""}

	marshaledMetadata, err := json.Marshal(res.Metadata)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("resource metadata: %w: %s", parseErr, err)
	}

	query, params, err := dialect.Insert(TABLE_RESOURCES).Rows(
		goqu.Record{
			"id":             res.ID,
			"urn":            res.URN,
			"name":           res.Name,
			"title":          res.Title,
			"project_id":     res.ProjectID,
			"namespace_name": res.NamespaceID,
			"principal_id":   principalID,
			"principal_type": principalType,
			"metadata":       marshaledMetadata,
		}).OnConflict(
		goqu.DoUpdate("urn", goqu.Record{
			"name":           res.Name,
			"project_id":     res.ProjectID,
			"namespace_name": res.NamespaceID,
			"principal_id":   principalID,
			"principal_type": principalType,
			"metadata":       marshaledMetadata,
		})).Returning(&ResourceCols{}).ToSQL()
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var resourceModel Resource
	if err = r.dbc.WithTimeout(ctx, TABLE_RESOURCES, "Upsert", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&resourceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrForeignKeyViolation):
			return resource.Resource{}, fmt.Errorf("%s: %w", err.Error(), resource.ErrInvalidDetail)
		case errors.Is(err, ErrInvalidTextRepresentation):
			return resource.Resource{}, fmt.Errorf("%s: %w", err.Error(), resource.ErrInvalidUUID)
		default:
			return resource.Resource{}, err
		}
	}

	return resourceModel.transformToResource()
}

func (r ResourceRepository) List(ctx context.Context, flt resource.Filter) ([]resource.Resource, error) {
	var fetchedResources []Resource

	sqlStatement := dialect.From(TABLE_RESOURCES)
	if flt.ProjectID != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"project_id": flt.ProjectID})
	}
	if flt.UserID != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"principal_id": flt.UserID})
	}
	if flt.ServiceUserID != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"principal_id": flt.ServiceUserID})
	}
	if flt.NamespaceID != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"namespace_name": flt.NamespaceID})
	}
	query, params, err := sqlStatement.ToSQL()
	if err != nil {
		return nil, err
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_RESOURCES, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedResources, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return []resource.Resource{}, nil
		}
		if errors.Is(err, ErrInvalidTextRepresentation) {
			return []resource.Resource{}, nil
		}
		return []resource.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedResources []resource.Resource
	for _, r := range fetchedResources {
		res, err := r.transformToResource()
		if err != nil {
			return nil, fmt.Errorf("failed to transform resource from db: %w", err)
		}
		transformedResources = append(transformedResources, res)
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
	if err = r.dbc.WithTimeout(ctx, TABLE_RESOURCES, "GetByID", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &resourceModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return resource.Resource{}, resource.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return resource.Resource{}, resource.ErrInvalidUUID
		default:
			return resource.Resource{}, err
		}
	}

	return resourceModel.transformToResource()
}

func (r ResourceRepository) Update(ctx context.Context, res resource.Resource) (resource.Resource, error) {
	if strings.TrimSpace(res.ID) == "" || !utils.IsValidUUID(res.ID) {
		return resource.Resource{}, resource.ErrInvalidID
	}

	marshaledMetadata, err := json.Marshal(res.Metadata)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("resource metadata: %w: %s", parseErr, err)
	}
	query, params, err := dialect.Update(TABLE_RESOURCES).Set(
		goqu.Record{
			"title":    res.Title,
			"metadata": marshaledMetadata,
		},
	).Where(goqu.Ex{"id": res.ID}).Returning(&ResourceCols{}).ToSQL()
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var resourceModel Resource
	if err = r.dbc.WithTimeout(ctx, TABLE_RESOURCES, "Update", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&resourceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return resource.Resource{}, resource.ErrNotExist
		case errors.Is(err, ErrDuplicateKey):
			return resource.Resource{}, resource.ErrConflict
		case errors.Is(err, ErrForeignKeyViolation):
			return resource.Resource{}, resource.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return resource.Resource{}, resource.ErrInvalidDetail
		default:
			return resource.Resource{}, err
		}
	}

	return resourceModel.transformToResource()
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
	if err = r.dbc.WithTimeout(ctx, TABLE_RESOURCES, "GetByURN", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &resourceModel, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return resource.Resource{}, resource.ErrNotExist
		}
		return resource.Resource{}, err
	}

	return resourceModel.transformToResource()
}

func (r ResourceRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_RESOURCES).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_RESOURCES, "Delete", func(ctx context.Context) error {
		if _, err = r.dbc.DB.ExecContext(ctx, query, params...); err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return resource.ErrNotExist
		default:
			return err
		}
	}
	return nil
}
