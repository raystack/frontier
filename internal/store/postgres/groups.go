package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
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

// *Get Groups Query
func buildGetGroupsBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	getGroupsBySlugQuery, _, err := dialect.From(TABLE_GROUPS).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).ToSQL()

	return getGroupsBySlugQuery, err
}

func buildGetGroupsByIdQuery(dialect goqu.DialectWrapper) (string, error) {
	getGroupsByIdQuery, _, err := dialect.From(TABLE_GROUPS).Where(goqu.ExOr{
		"id":   goqu.L("$1"),
		"slug": goqu.L("$2"),
	}).ToSQL()

	return getGroupsByIdQuery, err
}

// *Create Group Query
func buildCreateGroupQuery(dialect goqu.DialectWrapper) (string, error) {
	createGroupsQuery, _, err := dialect.Insert(TABLE_GROUPS).Rows(
		goqu.Record{
			"name":     goqu.L("$1"),
			"slug":     goqu.L("$2"),
			"org_id":   goqu.L("$3"),
			"metadata": goqu.L("$4"),
		}).Returning(&Group{}).ToSQL()
	return createGroupsQuery, err
}

// *List Groups Query
func buildListGroupsQuery(dialect goqu.DialectWrapper) (string, error) {
	listGroupsQuery, _, err := dialect.From(TABLE_GROUPS).ToSQL()

	return listGroupsQuery, err
}

func buildListGroupUsersQuery(dialect goqu.DialectWrapper) (string, error) {
	listGroupUsersQuery, _, err := dialect.Select(
		goqu.I("u.id").As("id"),
		goqu.I("u.name").As("name"),
		goqu.I("u.email").As("email"),
		goqu.I("u.metadata").As("metadata"),
		goqu.I("u.created_at").As("created_at"),
		goqu.I("u.updated_at").As("updated_at"),
	).From(goqu.T(TABLE_RELATION).As("r")).
		Join(goqu.T(TABLE_USER).As("u"), goqu.On(
			goqu.I("u.id").Cast("VARCHAR").
				Eq(goqu.I("r.subject_id")),
		)).Where(goqu.Ex{
		"r.object_id":            goqu.L("$1"),
		"r.role_id":              goqu.L("$2"),
		"r.subject_namespace_id": namespace.DefinitionUser.Id,
		"r.object_namespace_id":  namespace.DefinitionTeam.Id,
	}).ToSQL()

	return listGroupUsersQuery, err
}

func buildListUserGroupRelationsQuery(dialect goqu.DialectWrapper) (string, error) {
	listUserGroupRelationsQuery, _, err := dialect.From(TABLE_RELATION).Where(goqu.Ex{
		"subject_namespace_id": goqu.L(namespace.DefinitionUser.Id),
		"object_namespace_id":  goqu.L(namespace.DefinitionTeam.Id),
		"subject_id":           goqu.L("$1"),
		"object_id":            goqu.L("$2"),
	}).ToSQL()

	return listUserGroupRelationsQuery, err
}

// *Update Group Query
func buildUpdateGroupBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	updateGroupQuery, _, err := dialect.Update(TABLE_GROUPS).Set(
		goqu.Record{
			"name":       goqu.L("$2"),
			"slug":       goqu.L("$3"),
			"org_id":     goqu.L("$4"),
			"metadata":   goqu.L("$5"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).Returning(&Group{}).ToSQL()

	return updateGroupQuery, err
}

func buildUpdateGroupByIdQuery(dialect goqu.DialectWrapper) (string, error) {
	updateGroupQuery, _, err := dialect.Update(TABLE_GROUPS).Set(
		goqu.Record{
			"name":       goqu.L("$3"),
			"slug":       goqu.L("$4"),
			"org_id":     goqu.L("$5"),
			"metadata":   goqu.L("$6"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.ExOr{
		"id":   goqu.L("$1"),
		"slug": goqu.L("$2"),
	}).Returning(&Group{}).ToSQL()

	return updateGroupQuery, err
}

func (s Store) GetGroup(ctx context.Context, id string) (group.Group, error) {
	var fetchedGroup Group
	var getGroupsQuery string
	var err error
	id = strings.TrimSpace(id)
	isUuid := isUUID(id)

	if isUuid {
		getGroupsQuery, err = buildGetGroupsByIdQuery(dialect)
	} else {
		getGroupsQuery, err = buildGetGroupsBySlugQuery(dialect)
	}
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if isUuid {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &fetchedGroup, getGroupsQuery, id, id)
		})
	} else {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &fetchedGroup, getGroupsQuery, id)
		})
	}

	if errors.Is(err, sql.ErrNoRows) {
		return group.Group{}, group.ErrNotExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return group.Group{}, group.ErrInvalidUUID
	} else if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedGroup, err := transformToGroup(fetchedGroup)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedGroup, nil
}

func (s Store) CreateGroup(ctx context.Context, grp group.Group) (group.Group, error) {
	marshaledMetadata, err := json.Marshal(grp.Metadata)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	createGroupsQuery, err := buildCreateGroupQuery(dialect)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var newGroup Group
	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newGroup, createGroupsQuery, grp.Name, grp.Slug, grp.OrganizationId, marshaledMetadata)
	})

	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedGroup, err := transformToGroup(newGroup)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedGroup, nil
}

func (s Store) ListGroups(ctx context.Context, org organization.Organization) ([]group.Group, error) {
	var fetchedGroups []Group

	query, err := buildListGroupsQuery(dialect)
	if err != nil {
		return []group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if org.Id != "" {
		query = query + fmt.Sprintf(" WHERE org_id='%s'", org.Id)
	}

	query = query + ";"
	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedGroups, query)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []group.Group{}, group.ErrNotExist
	}

	if err != nil {
		return []group.Group{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedGroups []group.Group

	for _, v := range fetchedGroups {
		transformedGroup, err := transformToGroup(v)
		if err != nil {
			return []group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedGroups = append(transformedGroups, transformedGroup)
	}

	return transformedGroups, nil
}

func (s Store) UpdateGroup(ctx context.Context, toUpdate group.Group) (group.Group, error) {
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var updateGroupQuery string
	toUpdate.Id = strings.TrimSpace(toUpdate.Id)
	isUuid := isUUID(toUpdate.Id)

	if isUuid {
		updateGroupQuery, err = buildUpdateGroupByIdQuery(dialect)
	} else {
		updateGroupQuery, err = buildUpdateGroupBySlugQuery(dialect)
	}
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var updatedGroup Group

	if isUuid {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &updatedGroup, updateGroupQuery, toUpdate.Id, toUpdate.Id, toUpdate.Name, toUpdate.Slug, toUpdate.Organization.Id, marshaledMetadata)
		})
	} else {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &updatedGroup, updateGroupQuery, toUpdate.Id, toUpdate.Name, toUpdate.Slug, toUpdate.Organization.Id, marshaledMetadata)
		})
	}

	if errors.Is(err, sql.ErrNoRows) {
		return group.Group{}, group.ErrNotExist
	} else if err != nil {
		return group.Group{}, fmt.Errorf("%s: %w", dbErr, err)
	}

	updated, err := transformToGroup(updatedGroup)
	if err != nil {
		return group.Group{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return updated, nil
}

func (s Store) ListGroupUsers(ctx context.Context, groupId string, roleId string) ([]user.User, error) {
	var role = role.DefinitionTeamMember.Id
	if roleId != "" {
		role = roleId
	}

	groupId = strings.TrimSpace(groupId) //groupId can be uuid or slug
	fetchedGroup, err := s.GetGroup(ctx, groupId)
	if err != nil {
		return []user.User{}, err
	}
	groupId = fetchedGroup.Id

	listGroupUsersQuery, err := buildListGroupUsersQuery(dialect)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedUsers []User
	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedUsers, listGroupUsersQuery, groupId, role)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []user.User{}, user.ErrNotExist
	}

	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []user.User

	for _, u := range fetchedUsers {
		transformedUser, err := transformToUser(u)
		if err != nil {
			return []user.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}

func (s Store) ListUserGroupRelations(ctx context.Context, userId string, groupId string) ([]relation.Relation, error) {
	var fetchedRelations []Relation

	listUserGroupRelationsQuery, err := buildListUserGroupRelationsQuery(dialect)
	if err != nil {
		return []relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	groupId = strings.TrimSpace(groupId)
	fetchedGroup, err := s.GetGroup(ctx, groupId)
	if err != nil {
		return []relation.Relation{}, err
	}
	groupId = fetchedGroup.Id

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedRelations, listUserGroupRelationsQuery, userId, groupId)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []relation.Relation{}, sql.ErrNoRows
	}

	if err != nil {
		return []relation.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedRelations []relation.Relation

	for _, v := range fetchedRelations {
		transformedGroup, err := transformToRelation(v)
		if err != nil {
			return []relation.Relation{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedRelations = append(transformedRelations, transformedGroup)
	}

	return transformedRelations, nil
}

func transformToGroup(from Group) (group.Group, error) {
	var unmarshalledMetadata map[string]any
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return group.Group{}, err
	}

	return group.Group{
		Id:             from.Id,
		Name:           from.Name,
		Slug:           from.Slug,
		Organization:   organization.Organization{Id: from.OrgID},
		OrganizationId: from.OrgID,
		Metadata:       unmarshalledMetadata,
		CreatedAt:      from.CreatedAt,
		UpdatedAt:      from.UpdatedAt,
	}, nil
}
