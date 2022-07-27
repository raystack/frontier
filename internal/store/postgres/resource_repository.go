package postgres

import (
	"context"
	"errors"
	"fmt"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/str"
)

type ResourceRepository struct {
	dbc *db.Client
}

func NewResourceRepository(dbc *db.Client) *ResourceRepository {
	return &ResourceRepository{
		dbc: dbc,
	}
}

func (r ResourceRepository) Create(ctx context.Context, resourceToCreate resource.Resource) (resource.Resource, error) {
	var newResource Resource

	userID := sql.NullString{String: resourceToCreate.UserID, Valid: resourceToCreate.UserID != ""}
	groupID := sql.NullString{String: resourceToCreate.GroupID, Valid: resourceToCreate.GroupID != ""}
	createResourceQuery, err := buildCreateResourceQuery(dialect)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &newResource, createResourceQuery, resourceToCreate.URN, resourceToCreate.Name, resourceToCreate.ProjectID, groupID, resourceToCreate.OrganizationID, resourceToCreate.NamespaceID, userID)
	}); err != nil {
		return resource.Resource{}, err
	}

	return newResource.transformToResource(), nil
}

func (r ResourceRepository) List(ctx context.Context, filters resource.Filters) ([]resource.Resource, error) {
	var fetchedResources []Resource
	filterQueryMap, err := str.StructToStringMap(filters)
	if err != nil {
		return []resource.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	listResourcesStatement := buildListResourcesStatement(dialect)

	for key, value := range filterQueryMap {
		if value != "" {
			listResourcesStatement = listResourcesStatement.Where(goqu.Ex{key: value})
		}
	}

	listResourcesQuery, _, err := listResourcesStatement.ToSQL()
	if err != nil {
		return []resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedResources, listResourcesQuery)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []resource.Resource{}, resource.ErrNotExist
		}
		return []resource.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedResources []resource.Resource
	for _, r := range fetchedResources {
		transformedResources = append(transformedResources, r.transformToResource())
	}

	return transformedResources, nil
}

func (r ResourceRepository) Get(ctx context.Context, id string) (resource.Resource, error) {
	var fetchedResource Resource

	getResourcesByIDQuery, err := buildGetResourcesByIDQuery(dialect)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedResource, getResourcesByIDQuery, id)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return resource.Resource{}, resource.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return resource.Resource{}, resource.ErrInvalidUUID
		}
		return resource.Resource{}, err
	}

	return fetchedResource.transformToResource(), nil
}

func (r ResourceRepository) Update(ctx context.Context, id string, toUpdate resource.Resource) (resource.Resource, error) {
	var updatedResource Resource

	userID := sql.NullString{String: toUpdate.UserID, Valid: toUpdate.UserID != ""}
	groupID := sql.NullString{String: toUpdate.GroupID, Valid: toUpdate.GroupID != ""}
	updateResourceQuery, err := buildUpdateResourceQuery(dialect)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &updatedResource, updateResourceQuery, id, toUpdate.Name, toUpdate.ProjectID, groupID, toUpdate.OrganizationID, toUpdate.NamespaceID, userID, toUpdate.URN)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return resource.Resource{}, resource.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return resource.Resource{}, fmt.Errorf("%w: %s", resource.ErrInvalidUUID, err)
		}
		return resource.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return updatedResource.transformToResource(), nil
}

func (r ResourceRepository) GetByURN(ctx context.Context, urn string) (resource.Resource, error) {
	var fetchedResource Resource
	getResourcesByURNQuery, err := buildGetResourcesByURNQuery(dialect)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedResource, getResourcesByURNQuery, urn)
	}); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return resource.Resource{}, resource.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return resource.Resource{}, resource.ErrInvalidUUID
		}
		return resource.Resource{}, err
	}

	return fetchedResource.transformToResource(), nil
}
