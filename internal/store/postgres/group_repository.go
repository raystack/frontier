package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/pkg/db"
)

type GroupRepository struct {
	dbc *db.Client
}

func NewGroupRepository(dbc *db.Client) *GroupRepository {
	return &GroupRepository{
		dbc: dbc,
	}
}

var notDisabledGroupExp = goqu.Or(
	goqu.Ex{
		"state": nil,
	},
	goqu.Ex{
		"state": goqu.Op{"neq": group.Disabled},
	},
)

func (r GroupRepository) GetByID(ctx context.Context, id string) (group.Group, error) {
	if strings.TrimSpace(id) == "" {
		return group.Group{}, group.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_GROUPS).Where(
		goqu.Ex{
			"id": id,
		}).Where(notDisabledGroupExp).ToSQL()
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var groupModel Group
	if err = r.dbc.WithTimeout(ctx, TABLE_GROUPS, "GetByID", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &groupModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return group.Group{}, group.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
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

func (r GroupRepository) GetByIDs(ctx context.Context, groupIDs []string, flt group.Filter) ([]group.Group, error) {
	var fetchedGroups []Group

	sqlStatement := dialect.From(TABLE_GROUPS)
	if flt.OrganizationID != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"org_id": flt.OrganizationID})
	}

	for _, id := range groupIDs {
		if strings.TrimSpace(id) == "" {
			return []group.Group{}, group.ErrInvalidID
		}
	}

	query, params, err := sqlStatement.Where(
		goqu.Ex{
			"id": goqu.Op{"in": groupIDs},
		}).Where(notDisabledGroupExp).ToSQL()
	if err != nil {
		return []group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	// this query will return empty array of groups if no UUID is not matched
	// TODO: check and fox what to do in this scenerio
	if err = r.dbc.WithTimeout(ctx, TABLE_GROUPS, "GetByIDs", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedGroups, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []group.Group{}, group.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return []group.Group{}, group.ErrInvalidUUID
		default:
			return []group.Group{}, err
		}
	}

	var transformedGroups []group.Group
	for _, g := range fetchedGroups {
		transformedGroup, err := g.transformToGroup()
		if err != nil {
			return []group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedGroups = append(transformedGroups, transformedGroup)
	}

	return transformedGroups, nil
}

func (r GroupRepository) Create(ctx context.Context, grp group.Group) (group.Group, error) {
	if strings.TrimSpace(grp.Name) == "" {
		return group.Group{}, group.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(grp.Metadata)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	insertRow := goqu.Record{
		"name":     grp.Name,
		"title":    grp.Title,
		"org_id":   grp.OrganizationID,
		"metadata": marshaledMetadata,
	}
	if grp.State != "" {
		insertRow["state"] = grp.State
	}
	query, params, err := dialect.Insert(TABLE_GROUPS).Rows(insertRow).Returning(&Group{}).ToSQL()
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var groupModel Group
	if err = r.dbc.WithTimeout(ctx, TABLE_GROUPS, "Upsert", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&groupModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrForeignKeyViolation):
			return group.Group{}, organization.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return group.Group{}, organization.ErrInvalidUUID
		case errors.Is(err, ErrDuplicateKey):
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
	if flt.State != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"state": flt.State})
	} else {
		sqlStatement = sqlStatement.Where(notDisabledGroupExp)
	}

	query, params, err := sqlStatement.ToSQL()
	if err != nil {
		return []group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedGroups []Group
	if err = r.dbc.WithTimeout(ctx, TABLE_GROUPS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedGroups, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []group.Group{}, nil
		case errors.Is(err, ErrInvalidTextRepresentation):
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

	if strings.TrimSpace(grp.Name) == "" {
		return group.Group{}, group.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(grp.Metadata)
	if err != nil {
		return group.Group{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_GROUPS).Set(
		goqu.Record{
			"title":      grp.Title,
			"name":       grp.Name,
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
	if err = r.dbc.WithTimeout(ctx, TABLE_GROUPS, "Update", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&groupModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return group.Group{}, group.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return group.Group{}, group.ErrInvalidUUID
		case errors.Is(err, ErrDuplicateKey):
			return group.Group{}, group.ErrConflict
		case errors.Is(err, ErrForeignKeyViolation):
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

// TODO(kushsharma): no longer in use, delete if needed
func (r GroupRepository) ListGroupRelations(ctx context.Context, objectId string, subject_type string, role string) ([]relation.Relation, error) {
	whereClauseExp := goqu.Ex{}
	whereClauseExp["object_id"] = objectId
	whereClauseExp["object_namespace_name"] = schema.GroupNamespace

	if subject_type != "" {
		if subject_type == "user" {
			whereClauseExp["subject_namespace_name"] = schema.UserPrincipal
		} else if subject_type == "group" {
			whereClauseExp["subject_namespace_name"] = schema.GroupPrincipal
		}
	}

	if role != "" {
		like := "%:" + role
		whereClauseExp["role_id"] = goqu.Op{"like": like}
	}

	query, params, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).Where(whereClauseExp).ToSQL()
	if err != nil {
		return []relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedRelations []Relation
	if err = r.dbc.WithTimeout(ctx, TABLE_GROUPS, "ListGroupRelations", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedRelations, query, params...)
	}); err != nil {
		// List should return empty list and no error instead
		if errors.Is(err, sql.ErrNoRows) {
			return []relation.Relation{}, nil
		}
		return []relation.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedRelations []relation.Relation
	for _, r := range fetchedRelations {
		transformedRelations = append(transformedRelations, r.transformToRelationV2())
	}

	return transformedRelations, nil
}

func (r GroupRepository) SetState(ctx context.Context, id string, state group.State) error {
	query, params, err := dialect.Update(TABLE_GROUPS).Set(
		goqu.Record{
			"state": state,
		}).Where(
		goqu.Ex{
			"id": id,
		},
	).Returning(&Group{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	var groupModel Group
	if err = r.dbc.WithTimeout(ctx, TABLE_GROUPS, "SetState", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&groupModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return group.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return group.ErrInvalidUUID
		default:
			return err
		}
	}
	return nil
}

func (r GroupRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_GROUPS).Where(
		goqu.Ex{
			"id": id,
		},
	).Returning(&Group{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	var groupModel Group
	if err = r.dbc.WithTimeout(ctx, TABLE_GROUPS, "Delete", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&groupModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return group.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return group.ErrInvalidUUID
		default:
			return err
		}
	}
	return nil
}
