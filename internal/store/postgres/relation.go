package postgres

import (
	"time"

	"database/sql"

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

func (from Relation) transformToRelation() relation.Relation {
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
	}
}
