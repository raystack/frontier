package postgres

import (
	"context"
	"errors"
	"fmt"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/str"
)

type RelationRepository struct {
	dbc *db.Client
}

func NewRelationRepository(dbc *db.Client) *RelationRepository {
	return &RelationRepository{
		dbc: dbc,
	}
}

func (r RelationRepository) Create(ctx context.Context, relationToCreate relation.Relation) (relation.Relation, error) {
	subjectNamespaceID := str.DefaultStringIfEmpty(relationToCreate.SubjectNamespace.ID, relationToCreate.SubjectNamespaceID)
	objectNamespaceID := str.DefaultStringIfEmpty(relationToCreate.ObjectNamespace.ID, relationToCreate.ObjectNamespaceID)
	roleID := str.DefaultStringIfEmpty(relationToCreate.Role.ID, relationToCreate.RoleID)

	var nsID string
	if relationToCreate.RelationType == relation.RelationTypes.Namespace {
		nsID = roleID
		roleID = ""
	}

	if str.IsStringEmpty(subjectNamespaceID) || str.IsStringEmpty(relationToCreate.SubjectID) ||
		str.IsStringEmpty(objectNamespaceID) || str.IsStringEmpty(relationToCreate.ObjectID) {
		return relation.Relation{}, relation.ErrInvalidDetail
	}

	query, params, err := dialect.Insert(TABLE_RELATIONS).Rows(
		goqu.Record{
			"subject_namespace_id": subjectNamespaceID,
			"subject_id":           relationToCreate.SubjectID,
			"object_namespace_id":  objectNamespaceID,
			"object_id":            relationToCreate.ObjectID,
			"role_id":              sql.NullString{String: roleID, Valid: roleID != ""},
			"namespace_id":         sql.NullString{String: nsID, Valid: nsID != ""},
		}).OnConflict(
		goqu.DoUpdate("subject_namespace_id, subject_id, object_namespace_id,  object_id, COALESCE(role_id, ''), COALESCE(namespace_id, '')", goqu.Record{
			"subject_namespace_id": subjectNamespaceID,
		})).Returning(&relationCols{}).ToSQL()
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var relationModel Relation
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&relationModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errForeignKeyViolation):
			return relation.Relation{}, relation.ErrInvalidDetail
		default:
			return relation.Relation{}, err
		}
	}

	return relationModel.transformToRelation(), nil
}

func (r RelationRepository) List(ctx context.Context) ([]relation.Relation, error) {
	query, params, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).ToSQL()
	if err != nil {
		return []relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetchedRelations []Relation
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
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
		transformedRelations = append(transformedRelations, r.transformToRelation())
	}

	return transformedRelations, nil
}

func (r RelationRepository) Get(ctx context.Context, id string) (relation.Relation, error) {
	if str.IsStringEmpty(id) {
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
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &relationModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return relation.Relation{}, relation.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return relation.Relation{}, relation.ErrInvalidUUID
		default:
			return relation.Relation{}, err
		}
	}

	return relationModel.transformToRelation(), nil
}

func (r RelationRepository) DeleteByID(ctx context.Context, id string) error {
	if str.IsStringEmpty(id) {
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

func (r RelationRepository) GetByFields(ctx context.Context, rel relation.Relation) (relation.Relation, error) {
	var fetchedRelation Relation

	subjectNamespaceID := str.DefaultStringIfEmpty(rel.SubjectNamespace.ID, rel.SubjectNamespaceID)
	objectNamespaceID := str.DefaultStringIfEmpty(rel.ObjectNamespace.ID, rel.ObjectNamespaceID)
	roleID := str.DefaultStringIfEmpty(rel.Role.ID, rel.RoleID)

	var nsID string
	if rel.RelationType == relation.RelationTypes.Namespace {
		nsID = roleID
		roleID = ""
	}

	query, params, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).Where(goqu.Ex{
		"subject_namespace_id": subjectNamespaceID,
		"subject_id":           rel.SubjectID,
		"object_namespace_id":  objectNamespaceID,
		"object_id":            rel.ObjectID,
	}, goqu.And(
		goqu.Or(
			goqu.C("role_id").IsNull(),
			goqu.C("role_id").Eq(sql.NullString{String: roleID, Valid: roleID != ""}),
		)),
		goqu.And(
			goqu.Or(
				goqu.C("namespace_id").IsNull(),
				goqu.C("namespace_id").Eq(sql.NullString{String: nsID, Valid: nsID != ""}),
			),
		)).ToSQL()
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedRelation, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return relation.Relation{}, relation.ErrNotExist
		default:
			return relation.Relation{}, err
		}
	}

	return fetchedRelation.transformToRelation(), nil
}

func (r RelationRepository) Update(ctx context.Context, rel relation.Relation) (relation.Relation, error) {
	if str.IsStringEmpty(rel.ID) {
		return relation.Relation{}, relation.ErrInvalidID
	}

	subjectNamespaceID := str.DefaultStringIfEmpty(rel.SubjectNamespace.ID, rel.SubjectNamespaceID)
	objectNamespaceID := str.DefaultStringIfEmpty(rel.ObjectNamespace.ID, rel.ObjectNamespaceID)
	roleID := str.DefaultStringIfEmpty(rel.Role.ID, rel.RoleID)

	var nsID string
	if rel.RelationType == relation.RelationTypes.Namespace {
		nsID = roleID
		roleID = ""
	}

	query, params, err := goqu.Update(TABLE_RELATIONS).Set(
		goqu.Record{
			"subject_namespace_id": subjectNamespaceID,
			"subject_id":           rel.SubjectID,
			"object_namespace_id":  objectNamespaceID,
			"object_id":            rel.ObjectID,
			"role_id":              sql.NullString{String: roleID, Valid: roleID != ""},
			"namespace_id":         sql.NullString{String: nsID, Valid: nsID != ""},
		}).Where(goqu.Ex{
		"id": rel.ID,
	}).Returning(&relationCols{}).ToSQL()

	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var relationModel Relation
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &relationModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return relation.Relation{}, relation.ErrNotExist
		case errors.Is(err, errDuplicateKey):
			return relation.Relation{}, relation.ErrConflict
		case errors.Is(err, errForeignKeyViolation):
			return relation.Relation{}, relation.ErrInvalidDetail
		case errors.Is(err, errInvalidTexRepresentation):
			return relation.Relation{}, relation.ErrInvalidUUID
		default:
			return relation.Relation{}, err
		}
	}

	return relationModel.transformToRelation(), nil
}
