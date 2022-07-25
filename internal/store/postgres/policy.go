package postgres

import (
	"fmt"
	"time"

	"database/sql"

	"github.com/doug-martin/goqu/v9"

	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/core/role"
)

type Policy struct {
	ID          string         `db:"id"`
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
	ID          string         `db:"id"`
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

func transformToPolicy(from Policy) (policy.Policy, error) {
	var rl role.Role
	var err error

	if from.Role.ID != "" {
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
		ID:          from.ID,
		Role:        rl,
		RoleID:      from.RoleID,
		Action:      action,
		ActionID:    from.ActionID.String,
		Namespace:   namespace,
		NamespaceID: from.NamespaceID,
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
	}, nil
}
