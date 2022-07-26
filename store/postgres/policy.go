package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/pkg/str"
)

type Policy struct {
	Id          string         `db:"id"`
	Role        Role           `db:"role"`
	RoleID      string         `db:"role_id"`
	Namespace   Namespace      `db:"namespace"`
	NamespaceID string         `db:"namespace_id"`
	Action      Action         `db:"action"`
	ActionID    sql.NullString `db:"action_id"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

type PolicyCols struct {
	Id          string         `db:"id"`
	RoleID      string         `db:"role_id"`
	NamespaceID string         `db:"namespace_id"`
	ActionID    sql.NullString `db:"action_id"`
}

func buildPolicySelectStatement(dialect goqu.DialectWrapper) *goqu.SelectDataset {
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
	).From(goqu.T(TABLE_POLICY).As("p"))

	return selectStatement
}

func buildPolicyJoinStatement(selectStatement *goqu.SelectDataset) *goqu.SelectDataset {
	joinStatement := selectStatement.Join(goqu.T(TABLE_ROLES), goqu.On(
		goqu.I("roles.id").Eq(goqu.I("p.role_id")),
	)).Join(goqu.T(TABLE_ACTION), goqu.On(
		goqu.I("actions.id").Eq(goqu.I("p.action_id")),
	)).Join(goqu.T(TABLE_NAMESPACE), goqu.On(
		goqu.I("namespaces.id").Eq(goqu.I("p.namespace_id")),
	))

	return joinStatement
}

func buildCreatePolicyQuery(dialect goqu.DialectWrapper) (string, error) {
	createPolicyQuery, _, err := dialect.Insert(TABLE_POLICY).Rows(
		goqu.Record{
			"namespace_id": goqu.L("$1"),
			"role_id":      goqu.L("$2"),
			"action_id":    goqu.L("$3"),
		}).OnConflict(goqu.DoUpdate("role_id, namespace_id, action_id", goqu.Record{
		"namespace_id": goqu.L("$1"),
	})).Returning(&PolicyCols{}).ToSQL()

	return createPolicyQuery, err
}

func buildGetPolicyQuery(dialect goqu.DialectWrapper) (string, error) {
	selectStatement := buildPolicySelectStatement(dialect)
	joinStatement := buildPolicyJoinStatement(selectStatement)
	getPolicyQuery, _, err := joinStatement.Where(goqu.Ex{
		"p.id": goqu.L("$1"),
	}).ToSQL()

	return getPolicyQuery, err
}
func buildListPolicyQuery(dialect goqu.DialectWrapper) (string, error) {
	selectStatement := buildPolicySelectStatement(dialect)
	joinStatement := buildPolicyJoinStatement(selectStatement)
	listPolicyQuery, _, err := joinStatement.ToSQL()

	return listPolicyQuery, err
}
func buildUpdatePolicyQuery(dialect goqu.DialectWrapper) (string, error) {
	updatePolicyQuery, _, err := dialect.Update(TABLE_POLICY).Set(
		goqu.Record{
			"namespace_id": goqu.L("$2"),
			"role_id":      goqu.L("$3"),
			"action_id":    goqu.L("$4"),
			"updated_at":   goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": goqu.L("$1"),
	}).Returning(&PolicyCols{}).ToSQL()

	return updatePolicyQuery, err
}

func (s Store) GetPolicy(ctx context.Context, id string) (policy.Policy, error) {
	fetchedPolicy, err := s.selectPolicy(ctx, id)
	return fetchedPolicy, err
}

func (s Store) selectPolicy(ctx context.Context, id string) (policy.Policy, error) {
	var fetchedPolicy Policy
	getPolicyQuery, err := buildGetPolicyQuery(dialect)
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedPolicy, getPolicyQuery, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return policy.Policy{}, policy.ErrNotExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return policy.Policy{}, policy.ErrInvalidUUID
	} else if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedPolicy, err := transformToPolicy(fetchedPolicy)
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedPolicy, nil
}

func (s Store) ListPolicies(ctx context.Context) ([]policy.Policy, error) {
	var fetchedPolicies []Policy
	listPolicyQuery, err := buildListPolicyQuery(dialect)
	if err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedPolicies, listPolicyQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []policy.Policy{}, project.ErrNotExist
	} else if err != nil {
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

func (s Store) CreatePolicy(ctx context.Context, policyToCreate policy.Policy) ([]policy.Policy, error) {
	var newPolicy Policy

	roleId := str.DefaultStringIfEmpty(policyToCreate.Role.Id, policyToCreate.RoleId)
	actionId := str.DefaultStringIfEmpty(policyToCreate.Action.Id, policyToCreate.ActionId)
	nsId := str.DefaultStringIfEmpty(policyToCreate.Namespace.Id, policyToCreate.NamespaceId)
	createPolicyQuery, err := buildCreatePolicyQuery(dialect)
	if err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newPolicy, createPolicyQuery, nsId, roleId, sql.NullString{String: actionId, Valid: actionId != ""})
	})
	if err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
	}
	return s.ListPolicies(ctx)
}

func (s Store) UpdatePolicy(ctx context.Context, id string, toUpdate policy.Policy) ([]policy.Policy, error) {
	var updatedPolicy Policy
	updatePolicyQuery, err := buildUpdatePolicyQuery(dialect)
	if err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedPolicy, updatePolicyQuery, id, toUpdate.NamespaceId, toUpdate.RoleId, sql.NullString{String: toUpdate.ActionId, Valid: toUpdate.ActionId != ""})
	})

	if err != nil {
		return []policy.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return s.ListPolicies(ctx)
}

func transformToPolicy(from Policy) (policy.Policy, error) {
	var rl role.Role
	var err error

	if from.Role.Id != "" {
		rl, err = transformToRole(from.Role)
		if err != nil {
			return policy.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
		}
	}

	action, err := transformToAction(from.Action)
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	namespace, err := transformToNamespace(from.Namespace)
	if err != nil {
		return policy.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return policy.Policy{
		Id:          from.Id,
		Role:        rl,
		RoleId:      from.RoleID,
		Action:      action,
		ActionId:    from.ActionID.String,
		Namespace:   namespace,
		NamespaceId: from.NamespaceID,
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
	}, nil
}
