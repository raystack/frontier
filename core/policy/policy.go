package policy

import (
	"context"
	"time"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
)

type Repository interface {
	Get(ctx context.Context, id string) (Policy, error)
	List(ctx context.Context) ([]Policy, error)
	Create(ctx context.Context, policy Policy) ([]Policy, error)
	Update(ctx context.Context, id string, policy Policy) ([]Policy, error)
}

type AuthzRepository interface {
	Add(ctx context.Context, policies []Policy) error
}

type Policy struct {
	ID          string
	Role        role.Role
	RoleID      string `json:"role_id"`
	Namespace   namespace.Namespace
	NamespaceID string `json:"namespace_id"`
	Action      action.Action
	ActionID    string `json:"action_id"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Filters struct {
	NamespaceID string
}
