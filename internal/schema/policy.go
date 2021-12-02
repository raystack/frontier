package schema

import (
	"context"
	"errors"

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

func (s Service) CreatePolicy(ctx context.Context, policy model.Policy) ([]model.Policy, error) {
	policies, err := s.Store.CreatePolicy(ctx, policy)
	if err != nil {
		return []model.Policy{}, err
	}
	schemas, err := s.generateSchema(ctx, policies)
	if err != nil {
		return []model.Policy{}, err
	}
	err = s.pushSchema(schemas)
	if err != nil {
		return []model.Policy{}, err
	}
	return policies, err
}

func (s Service) UpdatePolicy(ctx context.Context, id string, policy model.Policy) ([]model.Policy, error) {
	policies, err := s.Store.UpdatePolicy(ctx, id, policy)
	if err != nil {
		return []model.Policy{}, err
	}
	schemas, err := s.generateSchema(ctx, policies)
	if err != nil {
		return []model.Policy{}, err
	}
	err = s.pushSchema(schemas)
	if err != nil {
		return []model.Policy{}, err
	}
	return policies, err
}

func (s Service) generateSchema(ctx context.Context, policies []model.Policy) ([]string, error) {
	definitions, err := schema_generator.BuildPolicyDefinitions(policies)
	if err != nil {
		return []string{}, err
	}
	schemas := schema_generator.BuildSchema(definitions)
	return schemas, nil
}

func (s Service) pushSchema(schemas []string) error {
	for _, schema := range schemas {
		err := s.Authz.Policy.AddPolicy(schema)
		if err != nil {
			return err
		}
	}
	return nil
}
