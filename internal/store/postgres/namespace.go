package postgres

import (
	"time"

	"database/sql"

	"github.com/raystack/shield/core/namespace"
)

type Namespace struct {
	ID           string       `db:"id"`
	Name         string       `db:"name"`
	Backend      string       `db:"backend"`
	ResourceType string       `db:"resource_type"`
	CreatedAt    time.Time    `db:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at"`
	DeletedAt    sql.NullTime `db:"deleted_at"`
}

func (from Namespace) transformToNamespace() namespace.Namespace {
	return namespace.Namespace{
		ID:           from.ID,
		Name:         from.Name,
		Backend:      from.Backend,
		ResourceType: from.ResourceType,
		CreatedAt:    from.CreatedAt,
		UpdatedAt:    from.UpdatedAt,
	}
}
