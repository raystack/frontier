package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/odpf/shield/internal/user"

	"github.com/odpf/shield/internal/group"
	"github.com/odpf/shield/model"
)

type Group struct {
	Id        string    `db:"id"`
	Name      string    `db:"name"`
	Slug      string    `db:"slug"`
	OrgID     string    `db:"org_id"`
	Metadata  []byte    `db:"metadata"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

const (
	createGroupsQuery   = `INSERT INTO groups(name, slug, org_id, metadata) values($1, $2, $3, $4) RETURNING id, name, slug, org_id, metadata, created_at, updated_at;`
	getGroupsQuery      = `SELECT id, name, slug, org_id, metadata, created_at, updated_at from groups where id=$1;`
	listGroupsQuery     = `SELECT id, name, slug, org_id, metadata, created_at, updated_at from groups`
	updateGroupQuery    = `UPDATE groups set name = $2, slug = $3, org_id = $4, metadata = $5, updated_at = now() where id = $1 RETURNING id, name, slug, org_id, metadata, created_at, updated_at;`
	listGroupUsersQuery = `SELECT u.id as id, u."name" as name, u.email as email, u.metadata as metadata, u.created_at as created_at, u.updated_at as updated_at from relations r JOIN users u ON CAST(u.id as VARCHAR) = r.subject_id WHERE r.object_id=$1`
)

func (s Store) GetGroup(ctx context.Context, id string) (model.Group, error) {
	var fetchedGroup Group
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedGroup, getGroupsQuery, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Group{}, group.GroupDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.Group{}, group.InvalidUUID
	} else if err != nil {
		return model.Group{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedGroup, err := transformToGroup(fetchedGroup)
	if err != nil {
		return model.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedGroup, nil
}

func (s Store) CreateGroup(ctx context.Context, grp model.Group) (model.Group, error) {
	marshaledMetadata, err := json.Marshal(grp.Metadata)
	if err != nil {
		return model.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var newGroup Group
	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newGroup, createGroupsQuery, grp.Name, grp.Slug, grp.OrganizationId, marshaledMetadata)
	})

	if err != nil {
		return model.Group{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedGroup, err := transformToGroup(newGroup)
	if err != nil {
		return model.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedGroup, nil
}

func (s Store) ListGroups(ctx context.Context, org model.Organization) ([]model.Group, error) {
	var fetchedGroups []Group

	query := listGroupsQuery
	if org.Id != "" {
		query = query + `WHERE org_id=$1`
	}

	query = query + ";"
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedGroups, query, org.Id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.Group{}, group.GroupDoesntExist
	}

	if err != nil {
		return []model.Group{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedGroups []model.Group

	for _, v := range fetchedGroups {
		transformedGroup, err := transformToGroup(v)
		if err != nil {
			return []model.Group{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedGroups = append(transformedGroups, transformedGroup)
	}

	return transformedGroups, nil
}

func (s Store) UpdateGroup(ctx context.Context, toUpdate model.Group) (model.Group, error) {
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return model.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var updatedGroup Group
	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedGroup, updateGroupQuery, toUpdate.Id, toUpdate.Name, toUpdate.Slug, toUpdate.Organization.Id, marshaledMetadata)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Group{}, group.GroupDoesntExist
	} else if err != nil {
		return model.Group{}, fmt.Errorf("%s: %w", dbErr, err)
	}

	updated, err := transformToGroup(updatedGroup)
	if err != nil {
		return model.Group{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return updated, nil
}

func (s Store) ListGroupUsers(ctx context.Context, groupId string) ([]model.User, error) {
	var fetchedUsers []User
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedUsers, listGroupUsersQuery, groupId)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.User{}, user.UserDoesntExist
	}

	if err != nil {
		return []model.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []model.User

	for _, u := range fetchedUsers {
		transformedUser, err := transformToUser(u)
		if err != nil {
			return []model.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}

func transformToGroup(from Group) (model.Group, error) {
	var unmarshalledMetadata map[string]string
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return model.Group{}, err
	}

	return model.Group{
		Id:             from.Id,
		Name:           from.Name,
		Slug:           from.Slug,
		Organization:   model.Organization{Id: from.OrgID},
		OrganizationId: from.OrgID,
		Metadata:       unmarshalledMetadata,
		CreatedAt:      from.CreatedAt,
		UpdatedAt:      from.UpdatedAt,
	}, nil
}
