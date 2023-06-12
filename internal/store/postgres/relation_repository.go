package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	newrelic "github.com/newrelic/go-agent"
	"github.com/raystack/shield/core/relation"
	"github.com/raystack/shield/internal/schema"
	"github.com/raystack/shield/pkg/db"
)

type RelationRepository struct {
	dbc *db.Client
}

func NewRelationRepository(dbc *db.Client) *RelationRepository {
	return &RelationRepository{
		dbc: dbc,
	}
}

func (r RelationRepository) Create(ctx context.Context, relationToCreate relation.RelationV2) (relation.RelationV2, error) {
	query, params, err := dialect.Insert(TABLE_RELATIONS).Rows(
		goqu.Record{
			"subject_namespace_id": relationToCreate.Subject.Namespace,
			"subject_id":           relationToCreate.Subject.ID,
			"object_namespace_id":  relationToCreate.Object.Namespace,
			"object_id":            relationToCreate.Object.ID,
			"role_id":              schema.GetRoleID(relationToCreate.Object.Namespace, relationToCreate.Subject.RoleID),
		}).OnConflict(
		goqu.DoUpdate("subject_namespace_id, subject_id, object_namespace_id,  object_id, role_id", goqu.Record{
			"subject_namespace_id": relationToCreate.Subject.Namespace,
		})).Returning(&relationCols{}).ToSQL()
	if err != nil {
		return relation.RelationV2{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var relationModel Relation
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_RELATIONS,
				Operation:  "Create",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&relationModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errForeignKeyViolation):
			return relation.RelationV2{}, fmt.Errorf("%w: %s", relation.ErrInvalidDetail, err)
		default:
			return relation.RelationV2{}, err
		}
	}

	return relationModel.transformToRelationV2(), nil
}

func (r RelationRepository) List(ctx context.Context) ([]relation.RelationV2, error) {
	query, params, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).ToSQL()
	if err != nil {
		return []relation.RelationV2{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedRelations []Relation
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_RELATIONS,
				Operation:  "List",
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

func (r RelationRepository) Get(ctx context.Context, id string) (relation.RelationV2, error) {
	if strings.TrimSpace(id) == "" {
		return relation.RelationV2{}, relation.ErrInvalidID
	}

	query, params, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).
		Where(goqu.Ex{
			"id": id,
		}).ToSQL()
	if err != nil {
		return relation.RelationV2{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var relationModel Relation
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_RELATIONS,
				Operation:  "Get",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.GetContext(ctx, &relationModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return relation.RelationV2{}, relation.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return relation.RelationV2{}, relation.ErrInvalidUUID
		default:
			return relation.RelationV2{}, err
		}
	}

	return relationModel.transformToRelationV2(), nil
}

func (r RelationRepository) DeleteByID(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return relation.ErrInvalidID
	}
	query, params, err := dialect.Delete(TABLE_RELATIONS).Where(goqu.Ex{
		"id": id,
	}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_RELATIONS,
				Operation:  "DeleteByID",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		result, err := r.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, errInvalidTexRepresentation):
				return relation.ErrInvalidUUID
			default:
				return err
			}
		}

		count, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if count > 0 {
			return nil
		}

		// TODO make this idempotent
		return relation.ErrNotExist
	})
}

// Update TO_DEPRECIATE
func (r RelationRepository) Update(ctx context.Context, rel relation.Relation) (relation.Relation, error) {
	return relation.Relation{}, nil
}

func (r RelationRepository) GetByFields(ctx context.Context, rel relation.RelationV2) (relation.RelationV2, error) {
	var fetchedRelation Relation
	like := "%:" + rel.Subject.RoleID

	query, _, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).Where(goqu.Ex{
		"subject_id": rel.Subject.ID,
		"object_id":  rel.Object.ID,
		"role_id":    goqu.Op{"like": like},
	}).ToSQL()
	if err != nil {
		return relation.RelationV2{}, fmt.Errorf("%w: %s", queryErr, err)
	}
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_RELATIONS,
				Operation:  "GetByFields",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.GetContext(ctx, &fetchedRelation, query)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return relation.RelationV2{}, relation.ErrNotExist
		default:
			return relation.RelationV2{}, err
		}
	}

	return fetchedRelation.transformToRelationV2(), nil
}

func (r RelationRepository) ListByFields(ctx context.Context, rel relation.RelationV2) ([]relation.RelationV2, error) {
	var fetchedRelation []Relation
	like := "%:" + rel.Subject.RoleID

	var exprs []goqu.Expression
	if len(rel.Subject.ID) != 0 {
		exprs = append(exprs, goqu.Ex{"subject_id": rel.Subject.ID})
	}
	if len(rel.Subject.RoleID) != 0 {
		exprs = append(exprs, goqu.Ex{"role_id": goqu.Op{"like": like}})
	}
	if len(rel.Object.ID) != 0 {
		exprs = append(exprs, goqu.Ex{"object_id": rel.Object.ID})
	}
	query, _, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).Where(exprs...).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		nrCtx := newrelic.FromContext(ctx)
		if nrCtx != nil {
			nr := newrelic.DatastoreSegment{
				Product:    newrelic.DatastorePostgres,
				Collection: TABLE_RELATIONS,
				Operation:  "GetByFields",
				StartTime:  nrCtx.StartSegmentNow(),
			}
			defer nr.End()
		}

		return r.dbc.SelectContext(ctx, &fetchedRelation, query)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, relation.ErrNotExist
		default:
			return nil, err
		}
	}

	var relations []relation.RelationV2
	for _, fr := range fetchedRelation {
		relations = append(relations, fr.transformToRelationV2())
	}
	return relations, nil
}
