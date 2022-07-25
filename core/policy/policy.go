package policy

import (
	"context"
	"errors"
	"time"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
)

var (
	ErrNotExist    = errors.New("policies doesn't exist")
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetPolicy(ctx context.Context, id string) (Policy, error)
	ListPolicies(ctx context.Context) ([]Policy, error)
	CreatePolicy(ctx context.Context, policy Policy) ([]Policy, error)
	UpdatePolicy(ctx context.Context, id string, policy Policy) ([]Policy, error)
}

type AuthzStore interface {
	AddPolicy(ctx context.Context, policies []Policy) error
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
