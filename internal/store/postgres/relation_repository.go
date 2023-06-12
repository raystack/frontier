package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/shield/core/relation"
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

func (r RelationRepository) Upsert(ctx context.Context, relationToCreate relation.Relation) (relation.Relation, error) {
	query, params, err := dialect.Insert(TABLE_RELATIONS).Rows(
		goqu.Record{
			"subject_namespace_name":   relationToCreate.Subject.Namespace,
			"subject_id":               relationToCreate.Subject.ID,
			"subject_subrelation_name": relationToCreate.Subject.SubRelationName,
			"object_namespace_name":    relationToCreate.Object.Namespace,
			"object_id":                relationToCreate.Object.ID,
			"relation_name":            relationToCreate.RelationName,
		}).OnConflict(
		goqu.DoUpdate("subject_namespace_name, subject_id, object_namespace_name, object_id, relation_name", goqu.Record{
			"subject_namespace_name": relationToCreate.Subject.Namespace,
		})).Returning(&relationCols{}).ToSQL()
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var relationModel Relation
	if err = r.dbc.WithTimeout(ctx, TABLE_RELATIONS, "Upsert", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&relationModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrForeignKeyViolation):
			return relation.Relation{}, fmt.Errorf("%w: %s", relation.ErrInvalidDetail, err)
		default:
			return relation.Relation{}, err
		}
	}

	return relationModel.transformToRelationV2(), nil
}

func (r RelationRepository) List(ctx context.Context) ([]relation.Relation, error) {
	query, params, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).ToSQL()
	if err != nil {
		return []relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedRelations []Relation
	if err = r.dbc.WithTimeout(ctx, TABLE_RELATIONS, "List", func(ctx context.Context) error {
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

func (r RelationRepository) Get(ctx context.Context, id string) (relation.Relation, error) {
	if strings.TrimSpace(id) == "" {
		return relation.Relation{}, relation.ErrInvalidID
	}

	query, params, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).
		Where(goqu.Ex{
			"id": id,
		}).ToSQL()
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var relationModel Relation
	if err = r.dbc.WithTimeout(ctx, TABLE_RELATIONS, "Get", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &relationModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return relation.Relation{}, relation.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return relation.Relation{}, relation.ErrInvalidUUID
		default:
			return relation.Relation{}, err
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

	return r.dbc.WithTimeout(ctx, TABLE_RELATIONS, "DeleteByID", func(ctx context.Context) error {
		result, err := r.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, ErrInvalidTextRepresentation):
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

func (r RelationRepository) GetByFields(ctx context.Context, rel relation.Relation) ([]relation.Relation, error) {
	var fetchedRelations []Relation
	stmt := dialect.Select(&relationCols{}).From(TABLE_RELATIONS)
	if rel.Object.ID != "" {
		stmt = stmt.Where(goqu.Ex{
			"object_id": rel.Object.ID,
		})
	}
	if rel.Object.Namespace != "" {
		stmt = stmt.Where(goqu.Ex{
			"object_namespace_name": rel.Object.Namespace,
		})
	}
	if rel.Subject.ID != "" {
		stmt = stmt.Where(goqu.Ex{
			"subject_id": rel.Subject.ID,
		})
	}
	if rel.Subject.Namespace != "" {
		stmt = stmt.Where(goqu.Ex{
			"subject_namespace_name": rel.Subject.Namespace,
		})
	}
	if rel.RelationName != "" {
		stmt = stmt.Where(goqu.Ex{
			"relation_name": rel.RelationName,
		})
	}

	query, _, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}
	if err = r.dbc.WithTimeout(ctx, TABLE_RELATIONS, "GetByFields", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedRelations, query)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, relation.ErrNotExist
		default:
			return nil, err
		}
	}

	var rels []relation.Relation
	for _, dbRel := range fetchedRelations {
		rels = append(rels, dbRel.transformToRelationV2())
	}
	return rels, nil
}

func (r RelationRepository) ListByFields(ctx context.Context, rel relation.Relation) ([]relation.Relation, error) {
	var fetchedRelation []Relation
	like := "%:" + rel.Subject.SubRelationName

	var exprs []goqu.Expression
	if len(rel.Subject.ID) != 0 {
		exprs = append(exprs, goqu.Ex{"subject_id": rel.Subject.ID})
	}
	if len(rel.RelationName) != 0 {
		exprs = append(exprs, goqu.Ex{"relation_name": goqu.Op{"like": like}})
	}
	if len(rel.Object.ID) != 0 {
		exprs = append(exprs, goqu.Ex{"object_id": rel.Object.ID})
	}
	query, _, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).Where(exprs...).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}
	if err = r.dbc.WithTimeout(ctx, TABLE_RELATIONS, "GetByFields", func(ctx context.Context) error {
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

	var relations []relation.Relation
	for _, fr := range fetchedRelation {
		relations = append(relations, fr.transformToRelationV2())
	}
	return relations, nil
}
