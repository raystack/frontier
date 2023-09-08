package relation

import (
	"context"
	"time"
)

type Repository interface {
	Get(ctx context.Context, id string) (Relation, error)
	Upsert(ctx context.Context, relation Relation) (Relation, error)
	List(ctx context.Context) ([]Relation, error)
	DeleteByID(ctx context.Context, id string) error
	GetByFields(ctx context.Context, rel Relation) ([]Relation, error)
}

type AuthzRepository interface {
	Check(ctx context.Context, rel Relation) (bool, error)
	BatchCheck(ctx context.Context, relations []Relation) ([]CheckPair, error)
	Delete(ctx context.Context, rel Relation) error
	Add(ctx context.Context, rel Relation) error
	LookupSubjects(ctx context.Context, rel Relation) ([]string, error)
	LookupResources(ctx context.Context, rel Relation) ([]string, error)
	ListRelations(ctx context.Context, rel Relation) ([]Relation, error)
}

type CheckPair struct {
	Relation Relation
	Status   bool
}

type Object struct {
	ID        string
	Namespace string
}

type Subject struct {
	ID              string
	Namespace       string
	SubRelationName string `json:"subject_sub_relation"`
}

type Relation struct {
	ID           string
	Object       Object
	Subject      Subject
	RelationName string `json:"relation_name"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
