package v1

import (
	"context"
	"fmt"

	"github.com/odpf/shield/model"
)

type PolicyService interface {
	GetPolicy(ctx context.Context, id string) (model.Policy, error)
	ListPolicies(ctx context.Context) ([]model.Policy, error)
	CreatePolicy(ctx context.Context, policy model.Policy) (model.Policy, error)
	UpdatePolicy(ctx context.Context, id string, policy model.Policy) (model.Policy, error)
}

func (v Dep) GetPolicy(ctx context.Context) {
	fmt.Println("GetPolicy")
}

func (v Dep) ListPolicies(ctx context.Context) {
	fmt.Println("ListPolicies")
}

func (v Dep) CreatePolicy(ctx context.Context, policy model.Policy) {
	fmt.Println("Create Policy")
}

func (v Dep) UpdatePolicy(ctx context.Context, id string, policy model.Policy) {
	fmt.Println("Create Policy")
}
