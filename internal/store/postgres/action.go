package postgres

import (
	"time"

	"github.com/odpf/shield/core/action"
)

type Action struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Namespace   Namespace `db:"namespace"`
	NamespaceID string    `db:"namespace_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type returnedActionColumns struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	NamespaceID string    `db:"namespace_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (from Action) transformToAction() action.Action {
	from.Namespace.ID = from.NamespaceID

	return action.Action{
		ID:          from.ID,
		Name:        from.Name,
		NamespaceID: from.NamespaceID,
		Namespace:   from.Namespace.transformToNamespace(),
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
	}
}
