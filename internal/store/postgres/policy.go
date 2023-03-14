package postgres

import (
	"time"

	"database/sql"

	"github.com/odpf/shield/core/policy"
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
	return policy.Policy{
		ID:          from.ID,
		RoleID:      from.RoleID,
		ActionID:    from.ActionID.String,
		NamespaceID: from.NamespaceID,
	}, nil
}
