package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/str"
)

type Resource struct {
	ID             string         `db:"id"`
	URN            string         `db:"urn"`
	Name           string         `db:"name"`
	ProjectID      string         `db:"project_id"`
	Project        Project        `db:"project"`
	GroupID        sql.NullString `db:"group_id"`
	Group          Group          `db:"group"`
	OrganizationID string         `db:"org_id"`
	Organization   Organization   `db:"organization"`
	NamespaceID    string         `db:"namespace_id"`
	Namespace      Namespace      `db:"namespace"`
	User           User           `db:"user"`
	UserID         sql.NullString `db:"user_id"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	DeletedAt      sql.NullTime   `db:"deleted_at"`
}

type ResourceCols struct {
	ID             string         `db:"id"`
	URN            string         `db:"urn"`
	Name           string         `db:"name"`
	ProjectID      string         `db:"project_id"`
	GroupID        sql.NullString `db:"group_id"`
	OrganizationID string         `db:"org_id"`
	NamespaceID    string         `db:"namespace_id"`
	UserID         sql.NullString `db:"user_id"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

func buildListResourcesStatement(dialect goqu.DialectWrapper) *goqu.SelectDataset {
	listResourcesStatement := dialect.From(TABLE_RESOURCE)

	return listResourcesStatement
}

func buildGetResourcesByIDQuery(dialect goqu.DialectWrapper) (string, error) {
	getResourcesByIDQuery, _, err := buildListResourcesStatement(dialect).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).ToSQL()

	return getResourcesByIDQuery, err
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

func buildGetResourcesByNamespaceQuery(dialect goqu.DialectWrapper, withResourceType bool) (string, error) {
	namespaceQueryExpression := goqu.Ex{
		"backend": goqu.L("$2"),
	}

	if withResourceType {
		namespaceQueryExpression["resouce_type"] = goqu.L("$3")
	}

	getNamespaceQuery := dialect.Select(&Namespace{}).From(TABLE_NAMESPACE).Where(namespaceQueryExpression)
	getResourcesByURNQuery, _, err := dialect.Select(&ResourceCols{}).From(TABLE_RESOURCE).Where(goqu.Ex{
		"urn":          goqu.L("$1"),
		"namespace_id": goqu.Op{"in": getNamespaceQuery},
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

func (s Store) CreateResource(ctx context.Context, resourceToCreate resource.Resource) (resource.Resource, error) {
	var newResource Resource

	userID := sql.NullString{String: resourceToCreate.UserID, Valid: resourceToCreate.UserID != ""}
	groupID := sql.NullString{String: resourceToCreate.GroupID, Valid: resourceToCreate.GroupID != ""}
	createResourceQuery, err := buildCreateResourceQuery(dialect)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newResource, createResourceQuery, resourceToCreate.URN, resourceToCreate.Name, resourceToCreate.ProjectID, groupID, resourceToCreate.OrganizationID, resourceToCreate.NamespaceID, userID)
	})

	if err != nil {
		return resource.Resource{}, err
	}

	transformedResource, err := transformToResource(newResource)

	if err != nil {
		return resource.Resource{}, err
	}

	return transformedResource, nil
}

func (s Store) ListResources(ctx context.Context, filters resource.Filters) ([]resource.Resource, error) {
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

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedResources, listResourcesQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []resource.Resource{}, resource.ErrNotExist
	}

	if err != nil {
		return []resource.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedResources []resource.Resource

	for _, r := range fetchedResources {
		transformedResource, err := transformToResource(r)
		if err != nil {
			return []resource.Resource{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedResources = append(transformedResources, transformedResource)
	}

	return transformedResources, nil
}

func (s Store) GetResource(ctx context.Context, id string) (resource.Resource, error) {
	var fetchedResource Resource

	getResourcesByIDQuery, err := buildGetResourcesByIDQuery(dialect)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedResource, getResourcesByIDQuery, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return resource.Resource{}, resource.ErrNotExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return resource.Resource{}, resource.ErrInvalidUUID
	} else if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	if err != nil {
		return resource.Resource{}, err
	}

	transformedResource, err := transformToResource(fetchedResource)
	if err != nil {
		return resource.Resource{}, err
	}

	return transformedResource, nil
}

func (s Store) UpdateResource(ctx context.Context, id string, toUpdate resource.Resource) (resource.Resource, error) {
	var updatedResource Resource

	userID := sql.NullString{String: toUpdate.UserID, Valid: toUpdate.UserID != ""}
	groupID := sql.NullString{String: toUpdate.GroupID, Valid: toUpdate.GroupID != ""}
	updateResourceQuery, err := buildUpdateResourceQuery(dialect)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedResource, updateResourceQuery, id, toUpdate.Name, toUpdate.ProjectID, groupID, toUpdate.OrganizationID, toUpdate.NamespaceID, userID, toUpdate.URN)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return resource.Resource{}, resource.ErrNotExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return resource.Resource{}, fmt.Errorf("%w: %s", resource.ErrInvalidUUID, err)
	} else if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	toUpdate, err = transformToResource(updatedResource)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}

func (s Store) GetResourceByNamespace(ctx context.Context, name string, ns namespace.Namespace) (resource.Resource, error) {
	var fetchedResource Resource

	//build query
	getResourceByNamespace, err := buildGetResourcesByNamespaceQuery(dialect, ns.ResourceType != "")
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedResource, getResourceByNamespace, name, ns.Backend, ns.ResourceType)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return resource.Resource{}, resource.ErrNotExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return resource.Resource{}, resource.ErrInvalidUUID
	} else if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	if err != nil {
		return resource.Resource{}, err
	}

	transformedResource, err := transformToResource(fetchedResource)
	if err != nil {
		return resource.Resource{}, err
	}

	return transformedResource, nil
}

func (s Store) GetResourceByURN(ctx context.Context, urn string) (resource.Resource, error) {
	var fetchedResource Resource
	getResourcesByURNQuery, err := buildGetResourcesByURNQuery(dialect)
	if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedResource, getResourcesByURNQuery, urn)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return resource.Resource{}, resource.ErrNotExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return resource.Resource{}, resource.ErrInvalidUUID
	} else if err != nil {
		return resource.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	if err != nil {
		return resource.Resource{}, err
	}

	transformedResource, err := transformToResource(fetchedResource)
	if err != nil {
		return resource.Resource{}, err
	}

	return transformedResource, nil
}

func transformToResource(from Resource) (resource.Resource, error) {
	// TODO: remove *ID
	return resource.Resource{
		Idxa:           from.ID,
		URN:            from.URN,
		Name:           from.Name,
		Project:        project.Project{ID: from.ProjectID},
		ProjectID:      from.ProjectID,
		Namespace:      namespace.Namespace{ID: from.NamespaceID},
		NamespaceID:    from.NamespaceID,
		Organization:   organization.Organization{ID: from.OrganizationID},
		OrganizationID: from.OrganizationID,
		GroupID:        from.GroupID.String,
		Group:          group.Group{ID: from.GroupID.String},
		User:           user.User{ID: from.UserID.String},
		CreatedAt:      from.CreatedAt,
		UpdatedAt:      from.UpdatedAt,
	}, nil
}
