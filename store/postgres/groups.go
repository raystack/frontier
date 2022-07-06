package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"time"

	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/internal/group"
	"github.com/odpf/shield/internal/user"
	"github.com/odpf/shield/model"
)

type Group struct {
	Id        string       `db:"id"`
	Name      string       `db:"name"`
	Slug      string       `db:"slug"`
	OrgID     string       `db:"org_id"`
	Metadata  []byte       `db:"metadata"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func buildCreateGroupQuery(dialect goqu.DialectWrapper) (string, error) {
	createGroupsQuery, _, err := dialect.Insert("groups").Rows(
		goqu.Record{
			"name":     goqu.L("$1"),
			"slug":     goqu.L("$2"),
			"org_id":   goqu.L("$3"),
			"metadata": goqu.L("$4"),
		}).Returning(&Group{}).ToSQL()
	return createGroupsQuery, err
}

func buildGetGroupsQuery(dialect goqu.DialectWrapper) (string, error) {
	getGroupsQuery, _, err := dialect.Select(&Group{}).From("groups").Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).ToSQL()

	return getGroupsQuery, err
}

func buildListGroupsQuery(dialect goqu.DialectWrapper) (string, error) {
	listGroupsQuery, _, err := dialect.From("groups").ToSQL()

	return listGroupsQuery, err
}

func buildUpdateGroupQuery(dialect goqu.DialectWrapper) (string, error) {
	updateGroupQuery, _, err := dialect.Update("groups").
		Set(goqu.Record{
			"name":       goqu.L("$2"),
			"slug":       goqu.L("$3"),
			"org_id":     goqu.L("$4"),
			"metadata":   goqu.L("$5"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{"id": goqu.L("$1")}).
		Returning(&Group{}).ToSQL()

	return updateGroupQuery, err
}

func buildListGroupUsersQuery(dialect goqu.DialectWrapper) (string, error) {
	listGroupUsersQuery, _, err := dialect.Select(
		goqu.I("u.id").As("u"),
		goqu.I("u.name").As("name"),
		goqu.I("u.email").As("email"),
		goqu.I("u.metadata").As("metadata"),
		goqu.I("u.created_at").As("created_at"),
		goqu.I("u.updated_at").As("updated_at"),
	).From(goqu.L("relations r")).
		Join(goqu.L("users u"), goqu.On(
			goqu.I("u.id").Cast("VARCHAR").
				Eq(goqu.I("r.subject_id")),
		)).Where(goqu.Ex{
		"r.object_id":            goqu.L("$1"),
		"r.role_id":              goqu.L("$2"),
		"r.subject_namespace_id": definition.UserNamespace.Id,
		"r.object_namespace_id":  definition.TeamNamespace.Id,
	}).ToSQL()

	return listGroupUsersQuery, err
}

func buildListUserGroupRelationsQuery(dialect goqu.DialectWrapper) (string, error) {
	listUserGroupRelationsQuery, _, err := dialect.From("relations").
		Where(goqu.Ex{
			"subject_namespace_id": goqu.L(definition.UserNamespace.Id),
			"object_namespace_id":  goqu.L(definition.TeamNamespace.Id),
			"subject_id":           goqu.L("$1"),
			"object_id":            goqu.L("$2"),
		}).ToSQL()

	return listUserGroupRelationsQuery, err
}

func (s Store) GetGroup(ctx context.Context, id string) (model.Group, error) {
	var fetchedGroup Group
	getGroupsQuery, err := buildGetGroupsQuery(dialect)
	if err != nil {
		return model.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
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

	createGroupsQuery, err := buildCreateGroupQuery(dialect)
	if err != nil {
		return model.Group{}, fmt.Errorf("%w: %s", queryErr, err)
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

	query, err := buildListGroupsQuery(dialect)
	if err != nil {
		return []model.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if org.Id != "" {
		query = query + fmt.Sprintf(" WHERE org_id='%s'", org.Id)
	}

	query = query + ";"
	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedGroups, query)
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

	updateGroupQuery, err := buildUpdateGroupQuery(dialect)
	if err != nil {
		return model.Group{}, fmt.Errorf("%w: %s", queryErr, err)
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

func (s Store) ListGroupUsers(ctx context.Context, groupId string, roleId string) ([]model.User, error) {
	var role = definition.TeamMemberRole.Id
	if roleId != "" {
		role = roleId
	}

	listGroupUsersQuery, err := buildListGroupUsersQuery(dialect)
	if err != nil {
		return []model.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedUsers []User
	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedUsers, listGroupUsersQuery, groupId, role)
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

func (s Store) ListUserGroupRelations(ctx context.Context, userId string, groupId string) ([]model.Relation, error) {
	var fetchedRelations []Relation

	listUserGroupRelationsQuery, err := buildListUserGroupRelationsQuery(dialect)
	if err != nil {
		return []model.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedRelations, listUserGroupRelationsQuery, userId, groupId)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.Relation{}, sql.ErrNoRows
	}

	if err != nil {
		return []model.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedRelations []model.Relation

	for _, v := range fetchedRelations {
		transformedGroup, err := transformToRelation(v)
		if err != nil {
			return []model.Relation{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedRelations = append(transformedRelations, transformedGroup)
	}

	return transformedRelations, nil
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
