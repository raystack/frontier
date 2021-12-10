package schema

import (
	"context"
	"errors"

	"github.com/odpf/shield/internal/authz"
	"github.com/odpf/shield/model"
)

type PolicyFilters struct {
	NamespaceId string
}

type Service struct {
	Store Store
	Authz *authz.Authz
}

var InvalidUUID = errors.New("invalid syntax of uuid")

type Store interface {
	GetAction(ctx context.Context, id string) (model.Action, error)
	CreateAction(ctx context.Context, action model.Action) (model.Action, error)
	ListActions(ctx context.Context) ([]model.Action, error)
	UpdateAction(ctx context.Context, action model.Action) (model.Action, error)
	GetNamespace(ctx context.Context, id string) (model.Namespace, error)
	CreateNamespace(ctx context.Context, namespace model.Namespace) (model.Namespace, error)
	ListNamespaces(ctx context.Context) ([]model.Namespace, error)
	UpdateNamespace(ctx context.Context, namespace model.Namespace) (model.Namespace, error)
	GetPolicy(ctx context.Context, id string) (model.Policy, error)
	ListPolicies(ctx context.Context) ([]model.Policy, error)
	CreatePolicy(ctx context.Context, policy model.Policy) ([]model.Policy, error)
	UpdatePolicy(ctx context.Context, id string, policy model.Policy) ([]model.Policy, error)
}
