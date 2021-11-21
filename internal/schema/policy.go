package schema

import (
	"context"
	"errors"
	"github.com/odpf/shield/model"
)

var PolicyDoesntExist = errors.New("actions doesn't exist")

func (s Service) GetPolicy(ctx context.Context, id string) (model.Policy, error) {
	return s.Store.GetPolicy(ctx, id)
}

func (s Service) ListPolicies(ctx context.Context) ([]model.Policy, error) {
	return s.Store.ListPolicies(ctx)
}
