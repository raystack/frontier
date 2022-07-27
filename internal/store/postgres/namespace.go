package postgres

import (
	"time"

	"database/sql"

	"github.com/odpf/shield/core/namespace"
)

type Namespace struct {
	ID        string       `db:"id"`
	Name      string       `db:"name"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func (from Namespace) transformToNamespace() namespace.Namespace {
	return namespace.Namespace{
		ID:        from.ID,
		Name:      from.Name,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}
}
