package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"

	"github.com/odpf/shield/internal/resource"
	"github.com/odpf/shield/model"
	"github.com/odpf/shield/pkg/utils"
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

const (
	createResourceQuery = `
		INSERT INTO resources (
			urn,
		    name,
			project_id,
			group_id,
			org_id,
			namespace_id,
		    user_id
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
		    $6,
		    $7
		)
		ON CONFLICT ON CONSTRAINT resources_urn_unique 
		DO
		    UPDATE SET name=$2, project_id=$3, group_id=$4, org_id=$5, namespace_id=$6,user_id=$7
		RETURNING id, urn, name, project_id, group_id, org_id, namespace_id, user_id, created_at, updated_at;`
	getResourcesQueryByURN = `
		SELECT
			id,
		    urn,
		    name,
			project_id,
			group_id,
			org_id,
			namespace_id,
		    user_id,
			created_at,
			updated_at
		FROM resources
		WHERE urn = $1;`
	updateResourceQuery = `
		UPDATE resources SET
		    name = $2,
			project_id = $3,
			group_id = $4,
			org_id = $5,
			namespace_id = $6,
		    user_id = $7,
		    urn = $8,
		WHERE id = $1
		`
)

func (s Store) CreateResource(ctx context.Context, resourceToCreate model.Resource) (model.Resource, error) {
	var newResource Resource

	userId := sql.NullString{String: resourceToCreate.UserId, Valid: resourceToCreate.UserId != ""}
	groupId := sql.NullString{String: resourceToCreate.GroupId, Valid: resourceToCreate.GroupId != ""}

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		nr := newrelic.DatastoreSegment{
			Product:    newrelic.DatastorePostgres,
			Collection: fmt.Sprintf("resource"),
			Operation:  "Create Resource",
			StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
		}
		defer nr.End()

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

	selectQuery := goqu.Select("*").From("resources")

	for key, value := range filterQueryMap {
		if value != "" {
			selectQuery = selectQuery.Where(goqu.Ex{key: value})
		}
	}

	querySql, _, err := selectQuery.ToSQL()

	if err != nil {
		return []model.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		nr := newrelic.DatastoreSegment{
			Product:    newrelic.DatastorePostgres,
			Collection: fmt.Sprintf("resource"),
			Operation:  "List Resources",
			StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
		}
		defer nr.End()

		return s.DB.SelectContext(ctx, &fetchedResources, querySql)
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

	querySql, _, err := goqu.Select("*").From("resources").Where(goqu.Ex{"id": id}).ToSQL()

	if err != nil {
		return model.Resource{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		nr := newrelic.DatastoreSegment{
			Product:    newrelic.DatastorePostgres,
			Collection: fmt.Sprintf("resource"),
			Operation:  "Get Resource",
			StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
		}
		defer nr.End()

		return s.DB.GetContext(ctx, &fetchedResource, querySql)
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

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		nr := newrelic.DatastoreSegment{
			Product:    newrelic.DatastorePostgres,
			Collection: fmt.Sprintf("resource"),
			Operation:  "Update Resource",
			StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
		}
		defer nr.End()

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
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		nr := newrelic.DatastoreSegment{
			Product:    newrelic.DatastorePostgres,
			Collection: fmt.Sprintf("resource"),
			Operation:  "Get Resource from URN",
			StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
		}
		defer nr.End()

		return s.DB.GetContext(ctx, &fetchedResource, getResourcesQueryByURN, urn)
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
