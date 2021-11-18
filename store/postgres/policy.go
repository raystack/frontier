package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/model"
)

type Policy struct {
	Id        string    `db:"id"`
	RoleID    string    `db:"role_id"`
	Role      Role      `db:"roles"`
	Namespace Namespace `db:"namespaces"`
	Action    Action    `db:"actions"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

const (
	getPolicyQuery = `SELECT p.id, p.namespace_id, p.role_id, p.action_id, p.created_at, p.updated_at, roles, actions, namespaces FROM policies p JOIN roles ON roles.id = p.role_id JOIN actions ON actions.id = p.action_id JOIN namespaces on namespaces.id = p.namespace_id WHERE p.id = $1`
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
