package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/odpf/shield/internal/group"
	"time"
)

type Group struct {
	Id           string    `db:"id"`
	Name         string    `db:"name"`
	Slug         string    `db:"slug"`
	Metadata     []byte    `db:"metadata"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	Organization `db:"organizations"`
}

const (
	createGroupsQuery = `
			WITH insert_into_groups AS (
				INSERT INTO groups(name, slug, org_id, metadata) values($1, $2, $3, $4) RETURNING id, name, slug, org_id, metadata, created_at, updated_at
			)
			SELECT 
			    insert_into_groups.id, insert_into_groups.name, insert_into_groups.slug, insert_into_groups.metadata, insert_into_groups.created_at, insert_into_groups.updated_at,
				organizations.id as "organizations.id",
			    organizations.name as "organizations.name",
			    organizations.slug as "organizations.slug",
			    organizations.metadata as "organizations.metadata",
			    organizations.created_at as "organizations.created_at",
			    organizations.updated_at as "organizations.updated_at"
			FROM 
			    insert_into_groups, organizations
			WHERE 
			    insert_into_groups.org_id = organizations.id;`
)

func (s Store) CreateGroup(ctx context.Context, grp group.Group) (group.Group, error) {
	marshaledMetadata, err := json.Marshal(grp.Metadata)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var newGroup Group
	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newGroup, createGroupsQuery, grp.Name, grp.Slug, grp.Organization.Id, marshaledMetadata)
	})

	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedOrg, err := transformToGroup(newGroup)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedOrg, nil
}

func transformToGroup(from Group) (group.Group, error) {
	var unmarshalledMetadata map[string]string
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return group.Group{}, err
	}

	transformedOrg, err := transformToOrg(from.Organization)
	if err != nil {
		return group.Group{}, err
	}

	return group.Group{
		Id:           from.Id,
		Name:         from.Name,
		Slug:         from.Slug,
		Organization: transformedOrg,
		Metadata:     unmarshalledMetadata,
		CreatedAt:    from.CreatedAt,
		UpdatedAt:    from.UpdatedAt,
	}, nil
}
