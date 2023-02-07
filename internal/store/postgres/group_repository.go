package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	newrelic "github.com/newrelic/go-agent"
	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/internal/schema"
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
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_GROUPS,
				Operation:  "GetByID",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

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
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_GROUPS,
				Operation:  "GetBySlug",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

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

func (r GroupRepository) GetByIDs(ctx context.Context, groupIDs []string) ([]group.Group, error) {
	var fetchedGroups []Group

	query, params, err := dialect.From(TABLE_GROUPS).Where(
		goqu.Ex{
			"id": goqu.Op{"in": groupIDs},
		}).ToSQL()
	if err != nil {
		return []group.Group{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_GROUPS,
				Operation:  "GetByIDs",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.SelectContext(ctx, &fetchedGroups, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []group.Group{}, group.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
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
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_GROUPS,
				Operation:  "Create",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

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
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_GROUPS,
				Operation:  "List",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

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
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_GROUPS,
				Operation:  "UpdateByID",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

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
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_GROUPS,
				Operation:  "GetBySlug",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

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
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_GROUPS,
				Operation:  "ListUserGroups",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

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

func (r GroupRepository) ListGroupRelations(ctx context.Context, objectId string, subject_type string, role string) ([]relation.RelationV2, error) {
	whereClauseExp := goqu.Ex{}
	whereClauseExp["object_id"] = objectId
	whereClauseExp["object_namespace_id"] = schema.GroupNamespace

	if subject_type != "" {
		if subject_type == "user" {
			whereClauseExp["subject_namespace_id"] = schema.UserPrincipal
		} else if subject_type == "group" {
			whereClauseExp["subject_namespace_id"] = schema.GroupPrincipal
		}
	}

	if role != "" {
		like := "%:" + role
		whereClauseExp["role_id"] = goqu.Op{"like": like}
	}

	query, params, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).Where(whereClauseExp).ToSQL()
	if err != nil {
		return []relation.RelationV2{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedRelations []Relation
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_GROUPS,
				Operation:  "ListGroupRelations",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.SelectContext(ctx, &fetchedRelations, query, params...)
	}); err != nil {
		// List should return empty list and no error instead
		if errors.Is(err, sql.ErrNoRows) {
			return []relation.RelationV2{}, nil
		}
		return []relation.RelationV2{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedRelations []relation.RelationV2
	for _, r := range fetchedRelations {
		transformedRelations = append(transformedRelations, r.transformToRelationV2())
	}

	return transformedRelations, nil
}
