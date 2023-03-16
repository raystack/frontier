package postgres

import (
	"time"

	"database/sql"

	"github.com/goto/shield/core/relation"
)

type Relation struct {
	ID                 string       `db:"id"`
	SubjectNamespaceID string       `db:"subject_namespace_id"`
	SubjectNamespace   Namespace    `db:"subject_namespace"`
	SubjectID          string       `db:"subject_id"`
	ObjectNamespaceID  string       `db:"object_namespace_id"`
	ObjectNamespace    Namespace    `db:"object_namespace"`
	ObjectID           string       `db:"object_id"`
	RoleID             string       `db:"role_id"`
	Role               Role         `db:"role"`
	CreatedAt          time.Time    `db:"created_at"`
	UpdatedAt          time.Time    `db:"updated_at"`
	DeletedAt          sql.NullTime `db:"deleted_at"`
}

type relationCols struct {
	ID                 string         `db:"id"`
	SubjectNamespaceID string         `db:"subject_namespace_id"`
	SubjectID          string         `db:"subject_id"`
	ObjectNamespaceID  string         `db:"object_namespace_id"`
	ObjectID           string         `db:"object_id"`
	RoleID             sql.NullString `db:"role_id"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
}

func (from Relation) transformToRelationV2() relation.RelationV2 {
	return relation.RelationV2{
		ID: from.ID,
		Subject: relation.Subject{
			ID:        from.SubjectID,
			Namespace: from.SubjectNamespaceID,
			RoleID:    from.RoleID,
		},
		Object: relation.Object{
			ID:          from.ObjectID,
			NamespaceID: from.ObjectNamespaceID,
		},
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}
}
