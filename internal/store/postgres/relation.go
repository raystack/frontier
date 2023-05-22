package postgres

import (
	"time"

	"database/sql"

	"github.com/odpf/shield/core/relation"
)

type Relation struct {
	ID                   string         `db:"id"`
	SubjectNamespaceID   string         `db:"subject_namespace_name"`
	SubjectNamespace     Namespace      `db:"subject_namespace"`
	SubjectID            string         `db:"subject_id"`
	SubjectSubRelationID sql.NullString `db:"subject_subrelation_name"`
	ObjectNamespaceID    string         `db:"object_namespace_name"`
	ObjectNamespace      Namespace      `db:"object_namespace"`
	ObjectID             string         `db:"object_id"`
	RelationName         string         `db:"relation_name"`
	CreatedAt            time.Time      `db:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at"`
	DeletedAt            sql.NullTime   `db:"deleted_at"`
}

type relationCols struct {
	ID                   string         `db:"id"`
	SubjectNamespaceID   string         `db:"subject_namespace_name"`
	SubjectID            string         `db:"subject_id"`
	SubjectSubRelationID string         `db:"subject_subrelation_name"`
	ObjectNamespaceID    string         `db:"object_namespace_name"`
	ObjectID             string         `db:"object_id"`
	RelationName         sql.NullString `db:"relation_name"`
	CreatedAt            time.Time      `db:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at"`
}

func (from Relation) transformToRelationV2() relation.Relation {
	return relation.Relation{
		ID: from.ID,
		Subject: relation.Subject{
			ID:              from.SubjectID,
			Namespace:       from.SubjectNamespaceID,
			SubRelationName: from.SubjectSubRelationID.String,
		},
		Object: relation.Object{
			ID:        from.ObjectID,
			Namespace: from.ObjectNamespaceID,
		},
		RelationName: from.RelationName,
		CreatedAt:    from.CreatedAt,
		UpdatedAt:    from.UpdatedAt,
	}
}
