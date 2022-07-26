package relation

import (
	"context"
	"errors"
	"time"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
)

var (
	ErrNotExist    = errors.New("relation doesn't exist")
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetRelation(ctx context.Context, id string) (Relation, error)
	CreateRelation(ctx context.Context, relation Relation) (Relation, error)
	ListRelations(ctx context.Context) ([]Relation, error)
	UpdateRelation(ctx context.Context, id string, toUpdate Relation) (Relation, error)
	GetRelationByFields(ctx context.Context, relation Relation) (Relation, error)
	DeleteRelationByID(ctx context.Context, id string) error
}

type AuthzStore interface {
	AddRelation(ctx context.Context, rel Relation) error
	CheckRelation(ctx context.Context, rel Relation, act action.Action) (bool, error)
	DeleteRelation(ctx context.Context, rel Relation) error
	DeleteSubjectRelations(ctx context.Context, resourceType, optionalResourceID string) error
}

type Relation struct {
	ID                 string
	SubjectNamespace   namespace.Namespace
	SubjectNamespaceID string `json:"subject_namespace_id"`
	SubjectID          string `json:"subject_id"`
	SubjectRoleID      string `json:"subject_role_id"`
	ObjectNamespace    namespace.Namespace
	ObjectNamespaceID  string `json:"object_namespace_id"`
	ObjectID           string `json:"object_id"`
	Role               role.Role
	RoleID             string       `json:"role_id"`
	RelationType       RelationType `json:"role_type"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type RelationType string

var RelationTypes = struct {
	Role      RelationType
	Namespace RelationType
}{
	Role:      "role",
	Namespace: "namespace",
}
