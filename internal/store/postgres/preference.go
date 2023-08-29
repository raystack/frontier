package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/preference"
)

type Preference struct {
	ID           uuid.UUID `db:"id"`
	Name         string    `db:"name"`
	Value        string    `db:"value"`
	ResourceType string    `db:"resource_type"`
	ResourceID   string    `db:"resource_id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (from Preference) transformToPreference() preference.Preference {
	return preference.Preference{
		ID:           from.ID.String(),
		Name:         from.Name,
		Value:        from.Value,
		ResourceType: from.ResourceType,
		ResourceID:   from.ResourceID,
		CreatedAt:    from.CreatedAt,
		UpdatedAt:    from.UpdatedAt,
	}
}
