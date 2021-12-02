package v1

import (
	"context"

	"github.com/odpf/shield/model"
)

type PolicyService interface {
	GetPolicy(ctx context.Context, id string) (model.Policy, error)
	ListPolicies(ctx context.Context) ([]model.Policy, error)
	CreatePolicy(ctx context.Context, policy model.Policy) ([]model.Policy, error)
	UpdatePolicy(ctx context.Context, id string, policy model.Policy) ([]model.Policy, error)
}
