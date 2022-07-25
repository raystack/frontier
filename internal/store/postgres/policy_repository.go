package postgres

import (
	"context"
	"errors"
	"fmt"

	"database/sql"

	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/str"
)

type PolicyRepository struct {
	dbc *db.Client
}

func NewPolicyRepository(dbc *db.Client) *PolicyRepository {
	return &PolicyRepository{
		dbc: dbc,
	}
}

func (r PolicyRepository) Get(ctx context.Context, id string) (policy.Policy, error) {
	fetchedPolicy, err := r.selectPolicy(ctx, id)
	return fetchedPolicy, err
}

func (r PolicyRepository) selectPolicy(ctx context.Context, id string) (policy.Policy, error) {
	var fetchedPolicy Policy
	getPolicyQuery, err := buildGetPolicyQuery(dialect)
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedPolicy, getPolicyQuery, id)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return policy.Policy{}, policy.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return policy.Policy{}, policy.ErrInvalidUUID
		}
		return policy.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedPolicy, err := transformToPolicy(fetchedPolicy)
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedPolicy, nil
}

func (r PolicyRepository) List(ctx context.Context) ([]policy.Policy, error) {
	var fetchedPolicies []Policy
	listPolicyQuery, err := buildListPolicyQuery(dialect)
	if err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedPolicies, listPolicyQuery)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []policy.Policy{}, project.ErrNotExist
		}
		return []policy.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedPolicies []policy.Policy
	for _, p := range fetchedPolicies {
		transformedPolicy, err := transformToPolicy(p)
		if err != nil {
			return []policy.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedPolicies = append(transformedPolicies, transformedPolicy)
	}

	return transformedPolicies, nil
}

func (r PolicyRepository) Create(ctx context.Context, policyToCreate policy.Policy) ([]policy.Policy, error) {
	var newPolicy Policy

	roleID := str.DefaultStringIfEmpty(policyToCreate.Role.ID, policyToCreate.RoleID)
	actionID := str.DefaultStringIfEmpty(policyToCreate.Action.ID, policyToCreate.ActionID)
	nsID := str.DefaultStringIfEmpty(policyToCreate.Namespace.ID, policyToCreate.NamespaceID)
	createPolicyQuery, err := buildCreatePolicyQuery(dialect)
	if err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &newPolicy, createPolicyQuery, nsID, roleID, sql.NullString{String: actionID, Valid: actionID != ""})
	}); err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
	}
	return r.List(ctx)
}

func (r PolicyRepository) Update(ctx context.Context, id string, toUpdate policy.Policy) ([]policy.Policy, error) {
	var updatedPolicy Policy
	updatePolicyQuery, err := buildUpdatePolicyQuery(dialect)
	if err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &updatedPolicy, updatePolicyQuery, id, toUpdate.NamespaceID, toUpdate.RoleID, sql.NullString{String: toUpdate.ActionID, Valid: toUpdate.ActionID != ""})
	}); err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return r.List(ctx)
}
