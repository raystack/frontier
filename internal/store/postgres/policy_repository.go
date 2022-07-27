package postgres

import (
	"context"
	"errors"
	"fmt"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/policy"
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

func (r PolicyRepository) buildListQuery() *goqu.SelectDataset {
	selectStatement := dialect.Select(
		"p.id",
		"p.namespace_id",
		goqu.I("roles.id").As(goqu.C("role.id")),
		goqu.I("roles.name").As(goqu.C("role.name")),
		goqu.I("roles.types").As(goqu.C("role.types")),
		goqu.I("roles.namespace_id").As(goqu.C("role.namespace_id")),
		goqu.I("roles.namespace_id").As(goqu.C("role.namespace.id")),
		goqu.I("roles.metadata").As(goqu.C("role.metadata")),
		goqu.I("namespaces.id").As(goqu.C("namespace.id")),
		goqu.I("namespaces.name").As(goqu.C("namespace.name")),
		goqu.I("actions.id").As(goqu.C("action.id")),
		goqu.I("actions.name").As(goqu.C("action.name")),
		goqu.I("actions.namespace_id").As(goqu.C("action.namespace_id")),
		goqu.I("actions.namespace_id").As(goqu.C("action.namespace.id")),
	).From(goqu.T(TABLE_POLICIES).As("p"))

	return selectStatement.Join(goqu.T(TABLE_ROLES), goqu.On(
		goqu.I("roles.id").Eq(goqu.I("p.role_id")),
	)).Join(goqu.T(TABLE_ACTIONS), goqu.On(
		goqu.I("actions.id").Eq(goqu.I("p.action_id")),
	)).Join(goqu.T(TABLE_NAMESPACES), goqu.On(
		goqu.I("namespaces.id").Eq(goqu.I("p.namespace_id")),
	))

}

func (r PolicyRepository) Get(ctx context.Context, id string) (policy.Policy, error) {
	if id == "" {
		return policy.Policy{}, policy.ErrInvalidID
	}

	query, params, err := r.buildListQuery().
		Where(
			goqu.Ex{
				"p.id": id,
			},
		).ToSQL()
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var policyModel Policy
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &policyModel, query, params...)
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

	transformedPolicy, err := transformToPolicy(policyModel)
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedPolicy, nil
}

func (r PolicyRepository) List(ctx context.Context) ([]policy.Policy, error) {
	var fetchedPolicies []Policy
	query, params, err := r.buildListQuery().ToSQL()
	if err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedPolicies, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, errInvalidTexRepresentation) {
			return []policy.Policy{}, policy.ErrInvalidUUID
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

//TODO this is actually upsert
func (r PolicyRepository) Create(ctx context.Context, pol policy.Policy) (string, error) {
	// TODO need to check actionID != ""

	//TODO need to find a way to deprecate this
	roleID := str.DefaultStringIfEmpty(pol.Role.ID, pol.RoleID)
	actionID := str.DefaultStringIfEmpty(pol.Action.ID, pol.ActionID)
	nsID := str.DefaultStringIfEmpty(pol.Namespace.ID, pol.NamespaceID)

	query, params, err := dialect.Insert(TABLE_POLICIES).Rows(
		goqu.Record{
			"namespace_id": nsID,
			"role_id":      roleID,
			"action_id":    sql.NullString{String: actionID, Valid: actionID != ""},
		}).OnConflict(goqu.DoUpdate("role_id, namespace_id, action_id", goqu.Record{
		"namespace_id": nsID,
	})).Returning("id").ToSQL()
	if err != nil {
		return "", fmt.Errorf("%w: %s", queryErr, err)
	}

	var policyID string
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).Scan(&policyID)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, errForeignKeyViolation) {
			return "", policy.ErrNotExist
		}
		return "", fmt.Errorf("%w: %s", dbErr, err)
	}

	return policyID, nil
}

func (r PolicyRepository) Update(ctx context.Context, toUpdate policy.Policy) (string, error) {
	// TODO need to check actionID != ""

	query, params, err := dialect.Update(TABLE_POLICIES).Set(
		goqu.Record{
			"namespace_id": toUpdate.NamespaceID,
			"role_id":      toUpdate.RoleID,
			"action_id":    sql.NullString{String: toUpdate.ActionID, Valid: toUpdate.ActionID != ""},
			"updated_at":   goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning("id").ToSQL()
	if err != nil {
		return "", fmt.Errorf("%w: %s", queryErr, err)
	}

	var policyID string
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).Scan(&policyID)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return "", policy.ErrNotExist
		}
		if errors.Is(err, errForeignKeyViolation) {
			return "", policy.ErrNotExist
		}
		return "", fmt.Errorf("%w: %s", dbErr, err)
	}

	return policyID, nil
}
