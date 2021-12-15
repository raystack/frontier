package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/odpf/shield/internal/project"
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

const selectStatement = `p.id, p.namespace_id, roles.id "role.id", roles.name "role.name", roles.types "role.types", roles.namespace_id "role.namespace_id", roles.namespace_id "role.namespace.id", roles.metadata "role.metadata", namespaces.id "namespace.id", namespaces.name "namespace.name", actions.id "action.id", actions.name "action.name", actions.namespace_id "action.namespace_id", actions.namespace_id "action.namespace.id"`
const joinStatement = `JOIN roles ON roles.id = p.role_id JOIN actions ON actions.id = p.action_id JOIN namespaces on namespaces.id = p.namespace_id`

var (
	createPolicyQuery = fmt.Sprintf(`INSERT into policies(namespace_id, role_id, action_id) values($1, $2, $3) RETURNING id, namespace_id, role_id, action_id`)
	getPolicyQuery    = fmt.Sprintf(`SELECT %s FROM policies p %s WHERE p.id = $1`, selectStatement, joinStatement)
	listPolicyQuery   = fmt.Sprintf(`SELECT %s FROM policies p %s`, selectStatement, joinStatement)
	updatePolicyQuery = fmt.Sprintf(`UPDATE policies SET namespace_id = $2, role_id = $3, action_id = $4, updated_at = now() where id = $1 RETURNING id, namespace_id, role_id, action_id;`)
)

func (s Store) GetPolicy(ctx context.Context, id string) (model.Policy, error) {
	fetchedPolicy, err := s.selectPolicy(ctx, id)
	return fetchedPolicy, err
}

func (s Store) selectPolicy(ctx context.Context, id string) (model.Policy, error) {
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

func (s Store) fetchNamespacePolicies(ctx context.Context, namespaceId string) ([]model.Policy, error) {
	var fetchedPolicies []Policy

	query := fmt.Sprintf("%s %s", listPolicyQuery, "WHERE p.namespace_id = $1")

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedPolicies, query, namespaceId)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.Policy{}, schema.PolicyDoesntExist
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

func (s Store) CreatePolicy(ctx context.Context, policyToCreate model.Policy) ([]model.Policy, error) {
	var newPolicy Policy

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newPolicy, createPolicyQuery, policyToCreate.NamespaceId, policyToCreate.RoleId, policyToCreate.ActionId)
	})
	if err != nil {
		return []model.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
	}
	return s.fetchNamespacePolicies(ctx, newPolicy.NamespaceID)
}

func (s Store) UpdatePolicy(ctx context.Context, id string, toUpdate model.Policy) ([]model.Policy, error) {
	var updatedPolicy Policy

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedPolicy, updatePolicyQuery, id, toUpdate.NamespaceId, toUpdate.RoleId, toUpdate.ActionId)
	})

	if err != nil {
		return []model.Policy{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	return s.fetchNamespacePolicies(ctx, updatedPolicy.NamespaceID)
}

func transformToPolicy(from Policy) (model.Policy, error) {
	var role model.Role
	var err error

	if from.Role.Id != "" {
		role, err = transformToRole(from.Role)
		if err != nil {
			return model.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
		}
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
		Id:          from.Id,
		Role:        role,
		RoleId:      from.RoleID,
		Action:      action,
		ActionId:    from.ActionID,
		Namespace:   namespace,
		NamespaceId: from.NamespaceID,
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
	}, nil
}
