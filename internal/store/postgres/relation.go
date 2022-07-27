package postgres

import (
	"time"

	"database/sql"

	"github.com/doug-martin/goqu/v9"

	"github.com/odpf/shield/core/relation"
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

	createRelationQuery, _, err := dialect.Insert(TABLE_RELATIONS).Rows(
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
	listRelationQuery, _, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).ToSQL()

	return listRelationQuery, err
}

func buildGetRelationsQuery(dialect goqu.DialectWrapper) (string, error) {
	getRelationsQuery, _, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).ToSQL()

	return getRelationsQuery, err
}

func buildUpdateRelationQuery(dialect goqu.DialectWrapper) (string, error) {
	updateRelationQuery, _, err := goqu.Update(TABLE_RELATIONS).Set(
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
	getRelationByFieldsQuery, _, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).Where(goqu.Ex{
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
	deleteRelationByIDQuery, _, err := dialect.Delete(TABLE_RELATIONS).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).ToSQL()

	return deleteRelationByIDQuery, err
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
