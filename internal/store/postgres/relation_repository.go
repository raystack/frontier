package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/pkg/db"
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
			"object_namespace_id":  relationToCreate.Object.NamespaceID,
			"object_id":            relationToCreate.Object.ID,
			"role_id":              schema.GetRoleID(relationToCreate.Object.NamespaceID, relationToCreate.Subject.RoleID),
		}).OnConflict(
		goqu.DoUpdate("subject_namespace_id, subject_id, object_namespace_id,  object_id, role_id", goqu.Record{
			"subject_namespace_id": relationToCreate.Subject.Namespace,
		})).Returning(&relationCols{}).ToSQL()
	if err != nil {
		return relation.RelationV2{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var relationModel Relation
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
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
