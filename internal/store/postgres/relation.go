package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/pkg/str"
)

type Relation struct {
	ID                 string         `db:"id"`
	SubjectNamespaceID string         `db:"subject_namespace_id"`
	SubjectNamespace   Namespace      `db:"subject_namespace"`
	SubjectID          string         `db:"subject_id"`
	ObjectNamespaceID  string         `db:"object_namespace_id"`
	ObjectNamespace    Namespace      `db:"object_namespace"`
	ObjectID           string         `db:"object_id"`
	RoleID             sql.NullString `db:"role_id"`
	Role               Role           `db:"role"`
	NamespaceID        sql.NullString `db:"namespace_id"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
	DeletedAt          sql.NullTime   `db:"deleted_at"`
}

type relationCols struct {
	ID                 string         `db:"id"`
	SubjectNamespaceID string         `db:"subject_namespace_id"`
	SubjectID          string         `db:"subject_id"`
	ObjectNamespaceID  string         `db:"object_namespace_id"`
	ObjectID           string         `db:"object_id"`
	RoleID             sql.NullString `db:"role_id"`
	NamespaceID        sql.NullString `db:"namespace_id"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
}

func buildCreateRelationQuery(dialect goqu.DialectWrapper) (string, error) {
	// TODO: Look for a better way to implement goqu.OnConflict with multiple columns

	createRelationQuery, _, err := dialect.Insert(TABLE_RELATION).Rows(
		goqu.Record{
			"subject_namespace_id": goqu.L("$1"),
			"subject_id":           goqu.L("$2"),
			"object_namespace_id":  goqu.L("$3"),
			"object_id":            goqu.L("$4"),
			"role_id":              goqu.L("$5"),
			"namespace_id":         goqu.L("$6"),
		}).OnConflict(goqu.DoUpdate("subject_namespace_id, subject_id, object_namespace_id,  object_id, COALESCE(role_id, ''), COALESCE(namespace_id, '')", goqu.Record{
		"subject_namespace_id": goqu.L("$1"),
	})).Returning(&relationCols{}).ToSQL()

	return createRelationQuery, err
}

func buildListRelationQuery(dialect goqu.DialectWrapper) (string, error) {
	listRelationQuery, _, err := dialect.Select(&relationCols{}).From(TABLE_RELATION).ToSQL()

	return listRelationQuery, err
}

func buildGetRelationsQuery(dialect goqu.DialectWrapper) (string, error) {
	getRelationsQuery, _, err := dialect.Select(&relationCols{}).From(TABLE_RELATION).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).ToSQL()

	return getRelationsQuery, err
}

func buildUpdateRelationQuery(dialect goqu.DialectWrapper) (string, error) {
	updateRelationQuery, _, err := goqu.Update(TABLE_RELATION).Set(
		goqu.Record{
			"subject_namespace_id": goqu.L("$2"),
			"subject_id":           goqu.L("$3"),
			"object_namespace_id":  goqu.L("$4"),
			"object_id":            goqu.L("$5"),
			"role_id":              goqu.L("$6"),
			"namespace_id":         goqu.L("$7"),
		}).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).Returning(&relationCols{}).ToSQL()

	return updateRelationQuery, err
}

func buildGetRelationByFieldsQuery(dialect goqu.DialectWrapper) (string, error) {
	getRelationByFieldsQuery, _, err := dialect.Select(&relationCols{}).From(TABLE_RELATION).Where(goqu.Ex{
		"subject_namespace_id": goqu.L("$1"),
		"subject_id":           goqu.L("$2"),
		"object_namespace_id":  goqu.L("$3"),
		"object_id":            goqu.L("$4"),
	}, goqu.And(
		goqu.Or(
			goqu.C("role_id").IsNull(),
			goqu.C("role_id").Eq(goqu.L("$5")),
		)),
		goqu.And(
			goqu.Or(
				goqu.C("namespace_id").IsNull(),
				goqu.C("namespace_id").Eq(goqu.L("$6")),
			),
		)).ToSQL()

	return getRelationByFieldsQuery, err
}

func buildDeleteRelationByIDQuery(dialect goqu.DialectWrapper) (string, error) {
	deleteRelationByIDQuery, _, err := dialect.Delete(TABLE_RELATION).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).ToSQL()

	return deleteRelationByIDQuery, err
}

func (s Store) CreateRelation(ctx context.Context, relationToCreate relation.Relation) (relation.Relation, error) {
	var newRelation Relation

	subjectNamespaceID := str.DefaultStringIfEmpty(relationToCreate.SubjectNamespace.ID, relationToCreate.SubjectNamespaceID)
	objectNamespaceID := str.DefaultStringIfEmpty(relationToCreate.ObjectNamespace.ID, relationToCreate.ObjectNamespaceID)
	roleID := str.DefaultStringIfEmpty(relationToCreate.Role.ID, relationToCreate.RoleID)
	var nsID string

	if relationToCreate.RelationType == relation.RelationTypes.Namespace {
		nsID = roleID
		roleID = ""
	}

	createRelationQuery, err := buildCreateRelationQuery(dialect)
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(
			ctx,
			&newRelation,
			createRelationQuery,
			subjectNamespaceID,
			relationToCreate.SubjectID,
			objectNamespaceID,
			relationToCreate.ObjectID,
			sql.NullString{String: roleID, Valid: roleID != ""},
			sql.NullString{String: nsID, Valid: nsID != ""},
		)
	})

	if err != nil {
		return relation.Relation{}, err
	}

	transformedRelation, err := transformToRelation(newRelation)

	if err != nil {
		return relation.Relation{}, err
	}

	return transformedRelation, nil
}

func (s Store) ListRelations(ctx context.Context) ([]relation.Relation, error) {
	var fetchedRelations []Relation
	listRelationQuery, err := buildListRelationQuery(dialect)
	if err != nil {
		return []relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedRelations, listRelationQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []relation.Relation{}, relation.ErrNotExist
	}

	if err != nil {
		return []relation.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedRelations []relation.Relation

	for _, r := range fetchedRelations {
		transformedRelation, err := transformToRelation(r)
		if err != nil {
			return []relation.Relation{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedRelations = append(transformedRelations, transformedRelation)
	}

	return transformedRelations, nil
}

func (s Store) GetRelation(ctx context.Context, id string) (relation.Relation, error) {
	var fetchedRelation Relation
	getRelationsQuery, err := buildGetRelationsQuery(dialect)
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedRelation, getRelationsQuery, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return relation.Relation{}, relation.ErrNotExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return relation.Relation{}, relation.ErrInvalidUUID
	} else if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	if err != nil {
		return relation.Relation{}, err
	}

	transformedRelation, err := transformToRelation(fetchedRelation)
	if err != nil {
		return relation.Relation{}, err
	}

	return transformedRelation, nil
}

func (s Store) DeleteRelationByID(ctx context.Context, id string) error {
	deleteRelationByIDQuery, err := buildDeleteRelationByIDQuery(dialect)
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		result, err := s.DB.ExecContext(ctx, deleteRelationByIDQuery, id)
		if err == nil {
			count, err := result.RowsAffected()
			if err == nil {
				if count == 1 {
					return nil
				}
			}
		}
		return err
	})
	return err
}

func (s Store) GetRelationByFields(ctx context.Context, rel relation.Relation) (relation.Relation, error) {
	var fetchedRelation Relation

	subjectNamespaceID := str.DefaultStringIfEmpty(rel.SubjectNamespace.ID, rel.SubjectNamespaceID)
	objectNamespaceID := str.DefaultStringIfEmpty(rel.ObjectNamespace.ID, rel.ObjectNamespaceID)
	roleID := str.DefaultStringIfEmpty(rel.Role.ID, rel.RoleID)
	var nsID string

	if rel.RelationType == relation.RelationTypes.Namespace {
		nsID = roleID
		roleID = ""
	}

	getRelationByFieldsQuery, err := buildGetRelationByFieldsQuery(dialect)
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx,
			&fetchedRelation,
			getRelationByFieldsQuery,
			subjectNamespaceID,
			rel.SubjectID,
			objectNamespaceID,
			rel.ObjectID,
			sql.NullString{String: roleID, Valid: roleID != ""},
			sql.NullString{String: nsID, Valid: nsID != ""},
		)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return relation.Relation{}, relation.ErrNotExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return relation.Relation{}, relation.ErrInvalidUUID
	} else if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	if err != nil {
		return relation.Relation{}, err
	}

	transformedRelation, err := transformToRelation(fetchedRelation)
	if err != nil {
		return relation.Relation{}, err
	}

	return transformedRelation, nil
}

func (s Store) UpdateRelation(ctx context.Context, id string, toUpdate relation.Relation) (relation.Relation, error) {
	var updatedRelation Relation

	subjectNamespaceID := str.DefaultStringIfEmpty(toUpdate.SubjectNamespace.ID, toUpdate.SubjectNamespaceID)
	objectNamespaceID := str.DefaultStringIfEmpty(toUpdate.ObjectNamespace.ID, toUpdate.ObjectNamespaceID)
	roleID := str.DefaultStringIfEmpty(toUpdate.Role.ID, toUpdate.RoleID)
	var nsID string

	if toUpdate.RelationType == relation.RelationTypes.Namespace {
		nsID = roleID
		roleID = ""
	}

	updateRelationQuery, err := buildUpdateRelationQuery(dialect)
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(
			ctx,
			&updatedRelation,
			updateRelationQuery,
			id,
			subjectNamespaceID,
			toUpdate.SubjectID,
			objectNamespaceID,
			toUpdate.ObjectID,
			sql.NullString{String: roleID, Valid: roleID != ""},
			sql.NullString{String: nsID, Valid: nsID != ""},
		)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return relation.Relation{}, relation.ErrNotExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return relation.Relation{}, fmt.Errorf("%w: %s", relation.ErrInvalidUUID, err)
	} else if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	toUpdate, err = transformToRelation(updatedRelation)
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}

func transformToRelation(from Relation) (relation.Relation, error) {
	relationType := relation.RelationTypes.Role
	roleID := from.RoleID.String

	if from.NamespaceID.Valid {
		roleID = from.NamespaceID.String
		relationType = relation.RelationTypes.Namespace
	}

	return relation.Relation{
		ID:                 from.ID,
		SubjectNamespaceID: from.SubjectNamespaceID,
		SubjectID:          from.SubjectID,
		ObjectNamespaceID:  from.ObjectNamespaceID,
		ObjectID:           from.ObjectID,
		RoleID:             roleID,
		RelationType:       relationType,
		CreatedAt:          from.CreatedAt,
		UpdatedAt:          from.UpdatedAt,
	}, nil
}
