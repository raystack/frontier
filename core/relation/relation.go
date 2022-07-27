package relation

import (
	"context"
	"time"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
)

type Repository interface {
	Get(ctx context.Context, id string) (Relation, error)
	Create(ctx context.Context, relation Relation) (Relation, error)
	List(ctx context.Context) ([]Relation, error)
	Update(ctx context.Context, toUpdate Relation) (Relation, error)
	GetByFields(ctx context.Context, relation Relation) (Relation, error)
	DeleteByID(ctx context.Context, id string) error
}

type AuthzRepository interface {
	Add(ctx context.Context, rel Relation) error
	Check(ctx context.Context, rel Relation, act action.Action) (bool, error)
	Delete(ctx context.Context, rel Relation) error
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
