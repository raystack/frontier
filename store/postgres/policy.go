package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/odpf/shield/internal/project"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/model"
)

type Policy struct {
	Id          string    `db:"id"`
	Role        Role      `db:"role"`
	RoleID      string    `db:"role_id"`
	Namespace   Namespace `db:"namespace"`
	NamespaceID string    `db:"namespace_id"`
	Action      Action    `db:"action"`
	ActionID    string    `db:"action_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

const selectStatement = `p.id, roles.id "role.id",roles.name "role.name", roles.namespace_id "role.namespace_id", roles.metadata "role.metadata", namespaces.id "namespace.id", namespaces.name "namespace.name", actions.id "action.id", actions.name "action.name", actions.namespace_id "action.namespace_id"`
const joinStatement = `JOIN roles ON roles.id = p.role_id JOIN actions ON actions.id = p.action_id JOIN namespaces on namespaces.id = p.namespace_id`

var (
	getPolicyQuery  = fmt.Sprintf(`SELECT %s FROM policies p %s WHERE p.id = $1`, selectStatement, joinStatement)
	listPolicyQuery = fmt.Sprintf(`SELECT %s FROM policies p %s`, selectStatement, joinStatement)
)

func (s Store) GetPolicy(ctx context.Context, id string) (model.Policy, error) {
	fetchedPolicy, err := s.selectPolicy(ctx, id, nil)
	return fetchedPolicy, err
}

func (s Store) selectPolicy(ctx context.Context, id string, txn *sqlx.Tx) (model.Policy, error) {
	var fetchedPolicy Policy

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedPolicy, getPolicyQuery, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Policy{}, schema.PolicyDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.Policy{}, schema.InvalidUUID
	} else if err != nil {
		return model.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedPolicy, err := transformToPolicy(fetchedPolicy)
	if err != nil {
		return model.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedPolicy, nil
}

func (s Store) ListPolicies(ctx context.Context) ([]model.Policy, error) {
	var fetchedPolicies []Policy
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedPolicies, listPolicyQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.Policy{}, project.ProjectDoesntExist
	} else if err != nil {
		return []model.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedPolicies []model.Policy
	for _, p := range fetchedPolicies {
		transformedPolicy, err := transformToPolicy(p)
		if err != nil {
			return []model.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedPolicies = append(transformedPolicies, transformedPolicy)
	}

	return transformedPolicies, nil
}

func transformToPolicy(from Policy) (model.Policy, error) {
	role, err := transformToRole(from.Role)
	if err != nil {
		return model.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	action, err := transformToAction(from.Action)
	if err != nil {
		return model.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	namespace, err := transformToNamespace(from.Namespace)
	if err != nil {
		return model.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return model.Policy{
		Id:        from.Id,
		Role:      role,
		Action:    action,
		Namespace: namespace,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
