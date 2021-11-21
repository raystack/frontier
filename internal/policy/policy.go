package policy

import (
	"context"
	"fmt"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store Store
}

type Store interface {
	GetPolicy(ctx context.Context, id string) (model.Policy, error)
	ListPolicies(ctx context.Context) ([]model.Policy, error)
}

func (s Service) GetPolicy(ctx context.Context, id string) (model.Policy, error) {
	fmt.Println("GetPolicy")
	return s.Store.GetPolicy(ctx, id)
}

func (s Service) ListPolicies(ctx context.Context) ([]model.Policy, error) {
	return s.Store.ListPolicies(ctx)
}