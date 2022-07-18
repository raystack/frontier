package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/pkg/utils"

	"github.com/odpf/shield/internal/resource"
	"github.com/odpf/shield/model"
)

type Resource struct {
	Id             string         `db:"id"`
	Urn            string         `db:"urn"`
	Name           string         `db:"name"`
	ProjectId      string         `db:"project_id"`
	Project        Project        `db:"project"`
	GroupId        sql.NullString `db:"group_id"`
	Group          Group          `db:"group"`
	OrganizationId string         `db:"org_id"`
	Organization   Organization   `db:"organization"`
	NamespaceId    string         `db:"namespace_id"`
	Namespace      Namespace      `db:"namespace"`
	User           User           `db:"user"`
	UserId         sql.NullString `db:"user_id"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	DeletedAt      sql.NullTime   `db:"deleted_at"`
}

type ResourceCols struct {
	Id             string         `db:"id"`
	Urn            string         `db:"urn"`
	Name           string         `db:"name"`
	ProjectId      string         `db:"project_id"`
	GroupId        sql.NullString `db:"group_id"`
	OrganizationId string         `db:"org_id"`
	NamespaceId    string         `db:"namespace_id"`
	UserId         sql.NullString `db:"user_id"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

func buildListResourcesStatement(dialect goqu.DialectWrapper) *goqu.SelectDataset {
	listResourcesStatement := dialect.From(TABLE_RESOURCE)

	return listResourcesStatement
}

func buildGetResourcesByIdQuery(dialect goqu.DialectWrapper) (string, error) {
	getResourcesByIdQuery, _, err := buildListResourcesStatement(dialect).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).ToSQL()

	return getResourcesByIdQuery, err
}

func buildCreateResourceQuery(dialect goqu.DialectWrapper) (string, error) {
	createResourceQuery, _, err := dialect.Insert(TABLE_RESOURCE).Rows(
		goqu.Record{
			"urn":          goqu.L("$1"),
			"name":         goqu.L("$2"),
			"project_id":   goqu.L("$3"),
			"group_id":     goqu.L("$4"),
			"org_id":       goqu.L("$5"),
			"namespace_id": goqu.L("$6"),
			"user_id":      goqu.L("$7"),
		}).OnConflict(goqu.DoUpdate("ON CONSTRAINT resources_urn_unique", goqu.Record{
		"name":         goqu.L("$2"),
		"project_id":   goqu.L("$3"),
		"group_id":     goqu.L("$4"),
		"org_id":       goqu.L("$5"),
		"namespace_id": goqu.L("$6"),
		"user_id":      goqu.L("$7"),
	})).Returning(&ResourceCols{}).ToSQL()

	return createResourceQuery, err
}

func buildGetResourcesByURNQuery(dialect goqu.DialectWrapper) (string, error) {
	getResourcesByURNQuery, _, err := dialect.Select(&ResourceCols{}).From(TABLE_RESOURCE).Where(goqu.Ex{
		"urn": goqu.L("$1"),
	}).ToSQL()

	return getResourcesByURNQuery, err
}

func buildUpdateResourceQuery(dialect goqu.DialectWrapper) (string, error) {
	updateResourceQuery, _, err := dialect.Update(TABLE_RESOURCE).Set(
		goqu.Record{
			"name":         goqu.L("$2"),
			"project_id":   goqu.L("$3"),
			"group_id":     goqu.L("$4"),
			"org_id":       goqu.L("$5"),
			"namespace_id": goqu.L("$6"),
			"user_id":      goqu.L("$7"),
			"urn":          goqu.L("$8"),
		}).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).ToSQL()

	return updateResourceQuery, err
}

func (s Store) CreateResource(ctx context.Context, resourceToCreate model.Resource) (model.Resource, error) {
	var newResource Resource

	userId := sql.NullString{String: resourceToCreate.UserId, Valid: resourceToCreate.UserId != ""}
	groupId := sql.NullString{String: resourceToCreate.GroupId, Valid: resourceToCreate.GroupId != ""}
	createResourceQuery, err := buildCreateResourceQuery(dialect)
	if err != nil {
		return model.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newResource, createResourceQuery, resourceToCreate.Urn, resourceToCreate.Name, resourceToCreate.ProjectId, groupId, resourceToCreate.OrganizationId, resourceToCreate.NamespaceId, userId)
	})

	if err != nil {
		return model.Resource{}, err
	}

	transformedResource, err := transformToResource(newResource)

	if err != nil {
		return model.Resource{}, err
	}

	return transformedResource, nil
}

func (s Store) ListResources(ctx context.Context, filters model.ResourceFilters) ([]model.Resource, error) {
	var fetchedResources []Resource
	filterQueryMap, err := utils.StructToStringMap(filters)
	if err != nil {
		return []model.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	listResourcesStatement := buildListResourcesStatement(dialect)

	for key, value := range filterQueryMap {
		if value != "" {
			listResourcesStatement = listResourcesStatement.Where(goqu.Ex{key: value})
		}
	}

	listResourcesQuery, _, err := listResourcesStatement.ToSQL()
	if err != nil {
		return []model.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedResources, listResourcesQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.Resource{}, resource.ResourceDoesntExist
	}

	if err != nil {
		return []model.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedResources []model.Resource

	for _, r := range fetchedResources {
		transformedResource, err := transformToResource(r)
		if err != nil {
			return []model.Resource{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedResources = append(transformedResources, transformedResource)
	}

	return transformedResources, nil
}

func (s Store) GetResource(ctx context.Context, id string) (model.Resource, error) {
	var fetchedResource Resource

	getResourcesByIdQuery, err := buildGetResourcesByIdQuery(dialect)
	if err != nil {
		return model.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedResource, getResourcesByIdQuery, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Resource{}, resource.ResourceDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.Resource{}, resource.InvalidUUID
	} else if err != nil {
		return model.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	if err != nil {
		return model.Resource{}, err
	}

	transformedResource, err := transformToResource(fetchedResource)
	if err != nil {
		return model.Resource{}, err
	}

	return transformedResource, nil
}

func (s Store) UpdateResource(ctx context.Context, id string, toUpdate model.Resource) (model.Resource, error) {
	var updatedResource Resource

	userId := sql.NullString{String: toUpdate.UserId, Valid: toUpdate.UserId != ""}
	groupId := sql.NullString{String: toUpdate.GroupId, Valid: toUpdate.GroupId != ""}
	updateResourceQuery, err := buildUpdateResourceQuery(dialect)
	if err != nil {
		return model.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedResource, updateResourceQuery, id, toUpdate.Name, toUpdate.ProjectId, groupId, toUpdate.OrganizationId, toUpdate.NamespaceId, userId, toUpdate.Urn)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Resource{}, resource.ResourceDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.Resource{}, fmt.Errorf("%w: %s", resource.InvalidUUID, err)
	} else if err != nil {
		return model.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	toUpdate, err = transformToResource(updatedResource)
	if err != nil {
		return model.Resource{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}

func (s Store) GetResourceByURN(ctx context.Context, urn string) (model.Resource, error) {
	var fetchedResource Resource
	getResourcesByURNQuery, err := buildGetResourcesByURNQuery(dialect)
	if err != nil {
		return model.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedResource, getResourcesByURNQuery, urn)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Resource{}, resource.ResourceDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.Resource{}, resource.InvalidUUID
	} else if err != nil {
		return model.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	if err != nil {
		return model.Resource{}, err
	}

	transformedResource, err := transformToResource(fetchedResource)
	if err != nil {
		return model.Resource{}, err
	}

	return transformedResource, nil
}

func transformToResource(from Resource) (model.Resource, error) {
	// TODO: remove *Id
	return model.Resource{
		Idxa:           from.Id,
		Urn:            from.Urn,
		Name:           from.Name,
		Project:        model.Project{Id: from.ProjectId},
		ProjectId:      from.ProjectId,
		Namespace:      model.Namespace{Id: from.NamespaceId},
		NamespaceId:    from.NamespaceId,
		Organization:   model.Organization{Id: from.OrganizationId},
		OrganizationId: from.OrganizationId,
		GroupId:        from.GroupId.String,
		Group:          model.Group{Id: from.GroupId.String},
		User:           model.User{Id: from.UserId.String},
		CreatedAt:      from.CreatedAt,
		UpdatedAt:      from.UpdatedAt,
	}, nil
}
