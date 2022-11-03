package relation

import (
	"context"
	"time"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
)

type Repository interface {
	Get(ctx context.Context, id string) (RelationV2, error)
	Create(ctx context.Context, relation RelationV2) (RelationV2, error)
	List(ctx context.Context) ([]RelationV2, error)
	Update(ctx context.Context, toUpdate Relation) (Relation, error)
	DeleteByID(ctx context.Context, id string) error
}

type AuthzRepository interface {
	Add(ctx context.Context, rel Relation) error
	Check(ctx context.Context, rel Relation, act action.Action) (bool, error)
	Delete(ctx context.Context, rel Relation) error
	DeleteSubjectRelations(ctx context.Context, resourceType, optionalResourceID string) error
	AddV2(ctx context.Context, rel RelationV2) error
}

type RoleService interface {
	Get(ctx context.Context, id string) (role.Role, error)
}

type UserService interface {
	GetByEmail(ctx context.Context, email string) (user.User, error)
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

type Object struct {
	ID          string
	NamespaceID string
}

type Subject struct {
	ID        string
	Namespace string
	RoleID    string
}

type RelationV2 struct {
	ID        string
	Object    Object
	Subject   Subject
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RelationType string

var RelationTypes = struct {
	Role      RelationType
	Namespace RelationType
}{
	Role:      "role",
	Namespace: "namespace",
}
