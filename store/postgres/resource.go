package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/odpf/shield/internal/resource"
	"github.com/odpf/shield/model"
	"time"
)

type Resource struct {
	Id             string       `db:"id"`
	Name           string       `db:"name"`
	ProjectId      string       `db:"project_id"`
	Project        Project      `db:"project"`
	GroupId        string       `db:"group_id"`
	Group          Group        `db:"group"`
	OrganizationId string       `db:"org_id"`
	Organization   Organization `db:"organization"`
	NamespaceId    string       `db:"namespace_id"`
	Namespace      Namespace    `db:"namespace"`
	CreatedAt      time.Time    `db:"created_at"`
	UpdatedAt      time.Time    `db:"updated_at"`
}

const (
	createResourceQuery = `
		INSERT INTO resources (
			id,
		    name,
			project_id,
			group_id,
			org_id,
			namespace_id
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
		    $6
		)
		RETURNING id, name, project_id, group_id, org_id, namespace_id, created_at, updated_at`
	listResourcesQuery = `
		SELECT
			id,
		    name,
			project_id,
			group_id,
			org_id,
			namespace_id,
			created_at,
			updated_at
		FROM resources`
	getResourcesQuery = `
		SELECT
			id,
		    name,
			project_id,
			group_id,
			org_id,
			namespace_id,
			created_at,
			updated_at
		FROM resources
		WHERE id = $1`
	updateResourceQuery = `
		UPDATE resources SET
		    name = $2,
			project_id = $3,
			group_id = $4,
			org_id = $5,
			namespace_id = $6
		WHERE id = $1
		`
)

func (s Store) CreateResource(ctx context.Context, resourceToCreate model.Resource) (model.Resource, error) {
	var newResource Resource

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newResource, createResourceQuery, resourceToCreate.Id, resourceToCreate.Name, resourceToCreate.ProjectId, resourceToCreate.GroupId, resourceToCreate.OrganizationId, resourceToCreate.NamespaceId)
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

func (s Store) ListResources(ctx context.Context) ([]model.Resource, error) {
	var fetchedResources []Resource
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
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
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedResource, getResourcesQuery, id)
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

func (s Store) UpdateResource(ctx context.Context, toUpdate model.Resource) (model.Resource, error) {
	var updatedResource Resource

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedResource, updateResourceQuery, toUpdate.Id, toUpdate.Name, toUpdate.ProjectId, toUpdate.GroupId, toUpdate.OrganizationId, toUpdate.NamespaceId)
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

func transformToResource(from Resource) (model.Resource, error) {

	return model.Resource{
		Id:             from.Id,
		Name:           from.Name,
		ProjectId:      from.ProjectId,
		NamespaceId:    from.NamespaceId,
		OrganizationId: from.OrganizationId,
		GroupId:        from.GroupId,
		CreatedAt:      from.CreatedAt,
		UpdatedAt:      from.UpdatedAt,
	}, nil
}
