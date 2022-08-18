package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/relation"
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

func (r GroupRepository) GetByID(ctx context.Context, id string) (group.Group, error) {
	if strings.TrimSpace(id) == "" {
		return group.Group{}, group.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_GROUPS).Where(
		goqu.Ex{
			"id": id,
		}).ToSQL()
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var groupModel Group
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &groupModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return group.Group{}, group.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return group.Group{}, group.ErrInvalidUUID
		default:
			return group.Group{}, err
		}
	}

	transformedGroup, err := groupModel.transformToGroup()
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedGroup, nil
}

func (r GroupRepository) GetBySlug(ctx context.Context, slug string) (group.Group, error) {
	if strings.TrimSpace(slug) == "" {
		return group.Group{}, group.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_GROUPS).Where(goqu.Ex{
		"slug": slug,
	}).ToSQL()

	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var groupModel Group
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &groupModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return group.Group{}, group.ErrNotExist
		default:
			return group.Group{}, err
		}
	}

	transformedGroup, err := groupModel.transformToGroup()
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedGroup, nil
}

func (r GroupRepository) Create(ctx context.Context, grp group.Group) (group.Group, error) {
	if strings.TrimSpace(grp.Name) == "" || strings.TrimSpace(grp.Slug) == "" {
		return group.Group{}, group.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(grp.Metadata)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Insert(TABLE_GROUPS).Rows(
		goqu.Record{
			"name":     grp.Name,
			"slug":     grp.Slug,
			"org_id":   grp.OrganizationID,
			"metadata": marshaledMetadata,
		}).Returning(&Group{}).ToSQL()
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var groupModel Group
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&groupModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errForeignKeyViolation):
			return group.Group{}, organization.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return group.Group{}, organization.ErrInvalidUUID
		case errors.Is(err, errDuplicateKey):
			return group.Group{}, group.ErrConflict
		default:
			return group.Group{}, err
		}
	}

	transformedGroup, err := groupModel.transformToGroup()
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedGroup, nil
}

func (r GroupRepository) List(ctx context.Context, flt group.Filter) ([]group.Group, error) {
	sqlStatement := dialect.From(TABLE_GROUPS)
	if flt.OrganizationID != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"org_id": flt.OrganizationID})
	}
	query, params, err := sqlStatement.ToSQL()
	if err != nil {
		return []group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedGroups []Group
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedGroups, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []group.Group{}, nil
		case errors.Is(err, errInvalidTexRepresentation):
			return []group.Group{}, nil
		default:
			return []group.Group{}, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	var transformedGroups []group.Group
	for _, v := range fetchedGroups {
		transformedGroup, err := v.transformToGroup()
		if err != nil {
			return []group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedGroups = append(transformedGroups, transformedGroup)
	}

	return transformedGroups, nil
}

func (r GroupRepository) UpdateByID(ctx context.Context, grp group.Group) (group.Group, error) {
	if strings.TrimSpace(grp.ID) == "" {
		return group.Group{}, group.ErrInvalidID
	}

	if strings.TrimSpace(grp.Name) == "" || strings.TrimSpace(grp.Slug) == "" {
		return group.Group{}, group.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(grp.Metadata)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_GROUPS).Set(
		goqu.Record{
			"name":       grp.Name,
			"slug":       grp.Slug,
			"org_id":     grp.OrganizationID,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(goqu.ExOr{
		"id": grp.ID,
	}).Returning(&Group{}).ToSQL()
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var groupModel Group
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&groupModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return group.Group{}, group.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return group.Group{}, group.ErrInvalidUUID
		case errors.Is(err, errDuplicateKey):
			return group.Group{}, group.ErrConflict
		case errors.Is(err, errForeignKeyViolation):
			return group.Group{}, organization.ErrNotExist
		default:
			return group.Group{}, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	updated, err := groupModel.transformToGroup()
	if err != nil {
		return group.Group{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return updated, nil
}

func (r GroupRepository) UpdateBySlug(ctx context.Context, grp group.Group) (group.Group, error) {
	if strings.TrimSpace(grp.Slug) == "" {
		return group.Group{}, group.ErrInvalidID
	}

	if strings.TrimSpace(grp.Name) == "" {
		return group.Group{}, group.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(grp.Metadata)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_GROUPS).Set(
		goqu.Record{
			"name":       grp.Name,
			"org_id":     grp.OrganizationID,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"slug": grp.Slug,
	}).Returning(&Group{}).ToSQL()
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var groupModel Group
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&groupModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return group.Group{}, group.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return group.Group{}, organization.ErrInvalidUUID
		case errors.Is(err, errDuplicateKey):
			return group.Group{}, group.ErrConflict
		case errors.Is(err, errForeignKeyViolation):
			return group.Group{}, organization.ErrNotExist
		default:
			return group.Group{}, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	updated, err := groupModel.transformToGroup()
	if err != nil {
		return group.Group{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return updated, nil
}

func (r GroupRepository) buildListUsersByGroupIDQuery(groupID, roleID string) (string, []interface{}, error) {
	sqlStatement := dialect.Select(
		goqu.I("u.id").As("id"),
		goqu.I("u.name").As("name"),
		goqu.I("u.email").As("email"),
		goqu.I("u.metadata").As("metadata"),
		goqu.I("u.created_at").As("created_at"),
		goqu.I("u.updated_at").As("updated_at"),
	).
		From(goqu.T(TABLE_RELATIONS).As("r")).
		Join(goqu.T(TABLE_USERS).As("u"), goqu.On(
			goqu.I("u.id").Cast("VARCHAR").
				Eq(goqu.I("r.subject_id")),
		)).
		Where(goqu.Ex{
			"r.object_id":            groupID,
			"r.subject_namespace_id": namespace.DefinitionUser.ID,
			"r.object_namespace_id":  namespace.DefinitionTeam.ID,
		})

	if strings.TrimSpace(roleID) != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{
			"r.role_id": roleID,
		})
	}

	return sqlStatement.ToSQL()
}

func (r GroupRepository) ListUsersByGroupID(ctx context.Context, groupID string, roleID string) ([]user.User, error) {
	if strings.TrimSpace(groupID) == "" {
		return nil, group.ErrInvalidID
	}

	query, params, err := r.buildListUsersByGroupIDQuery(groupID, roleID)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedUsers []User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedUsers, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []user.User{}, nil
		default:
			return []user.User{}, err
		}
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

func (r GroupRepository) ListUsersByGroupSlug(ctx context.Context, groupSlug string, roleID string) ([]user.User, error) {
	if strings.TrimSpace(groupSlug) == "" {
		return nil, group.ErrInvalidID
	}

	fetchedGroup, err := r.GetBySlug(ctx, groupSlug)
	if err != nil {
		return []user.User{}, err
	}

	query, params, err := r.buildListUsersByGroupIDQuery(fetchedGroup.ID, roleID)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedUsers []User
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedUsers, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []user.User{}, nil
		default:
			return []user.User{}, err
		}
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

func (r GroupRepository) ListUserGroupIDRelations(ctx context.Context, userID string, groupID string) ([]relation.Relation, error) {
	if strings.TrimSpace(groupID) == "" || strings.TrimSpace(userID) == "" {
		return nil, group.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_RELATIONS).Where(goqu.Ex{
		"subject_namespace_id": namespace.DefinitionUser.ID,
		"object_namespace_id":  namespace.DefinitionTeam.ID,
		"subject_id":           userID,
		"object_id":            groupID,
	}).ToSQL()
	if err != nil {
		return []relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedRelations []Relation
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedRelations, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []relation.Relation{}, nil
		default:
			return []relation.Relation{}, err
		}
	}

	var transformedRelations []relation.Relation
	for _, v := range fetchedRelations {
		transformedRelations = append(transformedRelations, v.transformToRelation())
	}

	return transformedRelations, nil
}

func (r GroupRepository) ListUserGroupSlugRelations(ctx context.Context, userID string, groupSlug string) ([]relation.Relation, error) {
	if strings.TrimSpace(groupSlug) == "" || strings.TrimSpace(userID) == "" {
		return nil, group.ErrInvalidID
	}
	var fetchedRelations []Relation

	fetchedGroup, err := r.GetBySlug(ctx, groupSlug)
	if err != nil {
		return []relation.Relation{}, err
	}

	query, params, err := dialect.From(TABLE_RELATIONS).Where(goqu.Ex{
		"subject_namespace_id": namespace.DefinitionUser.ID,
		"object_namespace_id":  namespace.DefinitionTeam.ID,
		"subject_id":           userID,
		"object_id":            fetchedGroup.ID,
	}).ToSQL()
	if err != nil {
		return []relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedRelations, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []relation.Relation{}, sql.ErrNoRows
		case errors.Is(err, errInvalidTexRepresentation):
			return []relation.Relation{}, group.ErrInvalidUUID
		default:
			return []relation.Relation{}, err
		}
	}

	var transformedRelations []relation.Relation
	for _, v := range fetchedRelations {
		transformedRelations = append(transformedRelations, v.transformToRelation())
	}

	return transformedRelations, nil
}

func (r GroupRepository) ListUserGroups(ctx context.Context, userID string, roleID string) ([]group.Group, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, group.ErrInvalidID
	}

	sqlStatement := dialect.Select(
		goqu.I("g.id").As("id"),
		goqu.I("g.metadata").As("metadata"),
		goqu.I("g.name").As("name"),
		goqu.I("g.slug").As("slug"),
		goqu.I("g.updated_at").As("updated_at"),
		goqu.I("g.created_at").As("created_at"),
		goqu.I("g.org_id").As("org_id"),
	).
		From(goqu.L("relations r")).
		Join(goqu.L("groups g"), goqu.On(
			goqu.I("g.id").Cast("VARCHAR").
				Eq(goqu.I("r.object_id")),
		)).
		Where(goqu.Ex{
			"r.object_namespace_id": namespace.DefinitionTeam.ID,
			"subject_namespace_id":  namespace.DefinitionUser.ID,
			"subject_id":            userID,
		})

	if strings.TrimSpace(roleID) != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{
			"role_id": roleID,
		})
	}

	query, params, err := sqlStatement.ToSQL()
	if err != nil {
		return []group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedGroups []Group
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedGroups, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []group.Group{}, nil
		}
		return []group.Group{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedGroups []group.Group
	for _, v := range fetchedGroups {
		transformedGroup, err := v.transformToGroup()
		if err != nil {
			return []group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedGroups = append(transformedGroups, transformedGroup)
	}

	return transformedGroups, nil
}
