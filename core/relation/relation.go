package relation

import (
	"context"
	"time"
)

type Repository interface {
	Get(ctx context.Context, id string) (RelationV2, error)
	Upsert(ctx context.Context, relation RelationV2) (RelationV2, error)
	List(ctx context.Context) ([]RelationV2, error)
	DeleteByID(ctx context.Context, id string) error
	GetByFields(ctx context.Context, rel RelationV2) ([]RelationV2, error)
}

type AuthzRepository interface {
	Check(ctx context.Context, rel RelationV2, permissionName string) (bool, error)
	Delete(ctx context.Context, rel RelationV2) error
	Add(ctx context.Context, rel RelationV2) error
	LookupSubjects(ctx context.Context, rel RelationV2) ([]string, error)
	LookupResources(ctx context.Context, rel RelationV2) ([]string, error)
	ListRelations(ctx context.Context, rel RelationV2) ([]RelationV2, error)
}

//type Relation struct {
//	ID                  string
//	SubjectNamespaceID  string `json:"subject_namespace_id"`
//	SubjectID           string `json:"subject_id"`
//	SubjectRelationName string `json:"subject_relation_name"`
//	ObjectNamespaceID   string `json:"object_namespace_id"`
//	ObjectID            string `json:"object_id"`
//	RelationName        string `json:"relation_name"`
//	CreatedAt           time.Time
//	UpdatedAt           time.Time
//}

type Object struct {
	ID        string
	Namespace string
}

type Subject struct {
	ID              string
	Namespace       string
	SubRelationName string `json:"sub_relation_name"`
}

type RelationV2 struct {
	ID           string
	Object       Object
	Subject      Subject
	RelationName string `json:"relation_name"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
