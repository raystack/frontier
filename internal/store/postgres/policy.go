package postgres

import (
	"fmt"
	"time"

	"database/sql"

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

func (from Policy) transformToPolicy() (policy.Policy, error) {
	var rl role.Role
	var err error

	if from.Role.ID != "" {
		rl, err = from.Role.transformToRole()
		if err != nil {
			return policy.Policy{}, fmt.Errorf("%w: %s", parseErr, err)
		}
	}

	act := from.Action.transformToAction()
	ns := from.Namespace.transformToNamespace()

	return policy.Policy{
		ID:          from.ID,
		RoleID:      rl.ID,
		ActionID:    act.ID,
		NamespaceID: ns.ID,
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,

		// @TODO(krtkvrm): issues/171
		DepreciatedAction:    from.Action.transformToAction(),
		DepreciatedRole:      rl,
		DepreciatedNamespace: from.Namespace.transformToNamespace(),
	}, nil
}
