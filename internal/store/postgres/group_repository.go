package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/db"
)

type GroupRepository struct {
	dbc *db.Client
}

func NewGroupRepository(dbc *db.Client) *GroupRepository {
	return &GroupRepository{
		dbc: dbc,
	}
}

func (r GroupRepository) Get(ctx context.Context, id string) (group.Group, error) {
	var fetchedGroup Group
	var getGroupsQuery string
	var err error
	id = strings.TrimSpace(id)
	isUuid := isUUID(id)

	if isUuid {
		getGroupsQuery, err = buildGetGroupsByIDQuery(dialect)
	} else {
		getGroupsQuery, err = buildGetGroupsBySlugQuery(dialect)
	}
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if isUuid {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &fetchedGroup, getGroupsQuery, id, id)
		})
	} else {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &fetchedGroup, getGroupsQuery, id)
		})
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return group.Group{}, group.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return group.Group{}, group.ErrInvalidUUID
		}
		return group.Group{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedGroup, err := transformToGroup(fetchedGroup)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedGroup, nil
}

func (r GroupRepository) Create(ctx context.Context, grp group.Group) (group.Group, error) {
	marshaledMetadata, err := json.Marshal(grp.Metadata)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	createGroupsQuery, err := buildCreateGroupQuery(dialect)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var newGroup Group
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &newGroup, createGroupsQuery, grp.Name, grp.Slug, grp.OrganizationID, marshaledMetadata)
	}); err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedGroup, err := transformToGroup(newGroup)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedGroup, nil
}

func (r GroupRepository) List(ctx context.Context, org organization.Organization) ([]group.Group, error) {
	var fetchedGroups []Group

	query, err := buildListGroupsQuery(dialect)
	if err != nil {
		return []group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if org.ID != "" {
		query = query + fmt.Sprintf(" WHERE org_id='%s'", org.ID)
	}

	query = query + ";"
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedGroups, query)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []group.Group{}, group.ErrNotExist
		}
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

func (r GroupRepository) Update(ctx context.Context, toUpdate group.Group) (group.Group, error) {
	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var updateGroupQuery string
	toUpdate.ID = strings.TrimSpace(toUpdate.ID)
	isUuid := isUUID(toUpdate.ID)

	if isUuid {
		updateGroupQuery, err = buildUpdateGroupByIDQuery(dialect)
	} else {
		updateGroupQuery, err = buildUpdateGroupBySlugQuery(dialect)
	}
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var updatedGroup Group

	if isUuid {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &updatedGroup, updateGroupQuery, toUpdate.ID, toUpdate.ID, toUpdate.Name, toUpdate.Slug, toUpdate.Organization.ID, marshaledMetadata)
		})
	} else {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &updatedGroup, updateGroupQuery, toUpdate.ID, toUpdate.Name, toUpdate.Slug, toUpdate.Organization.ID, marshaledMetadata)
		})
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return group.Group{}, group.ErrNotExist
		}
		return group.Group{}, fmt.Errorf("%s: %w", dbErr, err)
	}

	updated, err := transformToGroup(updatedGroup)
	if err != nil {
		return group.Group{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return updated, nil
}

func (r GroupRepository) ListUsers(ctx context.Context, groupID string, roleID string) ([]user.User, error) {
	var role = role.DefinitionTeamMember.ID
	if roleID != "" {
		role = roleID
	}

	groupID = strings.TrimSpace(groupID) //groupID can be uuid or slug
	fetchedGroup, err := r.Get(ctx, groupID)
	if err != nil {
		return []user.User{}, err
	}
	groupID = fetchedGroup.ID

	listGroupUsersQuery, err := buildListGroupUsersQuery(dialect)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedUsers []User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedUsers, listGroupUsersQuery, groupID, role)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []user.User{}, user.ErrNotExist
		}
		return []user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []user.User
	for _, u := range fetchedUsers {
		transformedUser, err := u.transformToUser()
		if err != nil {
			return []user.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}

func (r GroupRepository) ListUserGroupRelations(ctx context.Context, userID string, groupID string) ([]relation.Relation, error) {
	var fetchedRelations []Relation

	listUserGroupRelationsQuery, err := buildListUserGroupRelationsQuery(dialect)
	if err != nil {
		return []relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	groupID = strings.TrimSpace(groupID)
	fetchedGroup, err := r.Get(ctx, groupID)
	if err != nil {
		return []relation.Relation{}, err
	}
	groupID = fetchedGroup.ID

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedRelations, listUserGroupRelationsQuery, userID, groupID)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []relation.Relation{}, sql.ErrNoRows
		}
		return []relation.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedRelations []relation.Relation
	for _, v := range fetchedRelations {
		transformedRelations = append(transformedRelations, v.transformToRelation())
	}

	return transformedRelations, nil
}

func (r GroupRepository) ListUserGroups(ctx context.Context, userID string, roleID string) ([]group.Group, error) {
	rlID := role.DefinitionTeamMember.ID

	if roleID == role.DefinitionTeamAdmin.ID {
		rlID = role.DefinitionTeamAdmin.ID
	}

	var fetchedGroups []Group

	listUserGroupsQuery, err := buildListUserGroupsQuery(dialect)
	if err != nil {
		return []group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedGroups, listUserGroupsQuery, userID, rlID)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []group.Group{}, group.ErrNotExist
		}
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
