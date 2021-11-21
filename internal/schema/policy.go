package schema

import (
	"context"
	"errors"
	"fmt"
	"github.com/odpf/shield/internal/schema_generator"
	"github.com/odpf/shield/model"
)

var PolicyDoesntExist = errors.New("policies doesn't exist")

func (s Service) GetPolicy(ctx context.Context, id string) (model.Policy, error) {
	return s.Store.GetPolicy(ctx, id)
}

func (s Service) ListPolicies(ctx context.Context) ([]model.Policy, error) {
	return s.Store.ListPolicies(ctx)
}

func (s Service) CreatePolicy(ctx context.Context, policy model.Policy) (model.Policy, error) {
	policy, err := s.Store.CreatePolicy(ctx, policy)
	s.generateSchema(ctx, policy.NamespaceId)
	return policy, err
}

func (s Service) UpdatePolicy(ctx context.Context, id string, policy model.Policy) (model.Policy, error) {
	policy, err := s.Store.UpdatePolicy(ctx, id, policy)
	s.generateSchema(ctx, policy.NamespaceId)
	return policy, err
}

func (s Service) generateSchema(ctx context.Context, namespaceId string) {
	policies, err := s.Store.ListPoliciesWithFilters(ctx, PolicyFilters{NamespaceId: namespaceId})
	if err != nil {
		return
	}
	definitions, err := schema_generator.BuildPolicyDefinitions(policies)
	schema := schema_generator.BuildSchema(definitions)
	fmt.Println(schema)
}
