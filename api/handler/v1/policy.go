package v1

import (
	"context"
	"fmt"

	"github.com/odpf/shield/model"
)

type PolicyService interface {
	GetPolicy(ctx context.Context, id string) (model.Policy, error)
	ListPolicies(ctx context.Context) ([]model.Policy, error)
}

func (v Dep) GetPolicy() {
	fmt.Println("GetPolicy")
}

func (v Dep) ListPolicies() {
	fmt.Println("ListPolicies")
}
