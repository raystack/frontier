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
	ScopeType    string    `db:"scope_type"`
	ScopeID      string    `db:"scope_id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (from Preference) transformToPreference() preference.Preference {
	pref := preference.Preference{
		ID:           from.ID.String(),
		Name:         from.Name,
		Value:        from.Value,
		ResourceType: from.ResourceType,
		ResourceID:   from.ResourceID,
		CreatedAt:    from.CreatedAt,
		UpdatedAt:    from.UpdatedAt,
	}
	// Convert zero values back to empty strings for API layer
	if from.ScopeType != preference.ScopeTypeGlobal {
		pref.ScopeType = from.ScopeType
	}
	if from.ScopeID != preference.ScopeIDGlobal {
		pref.ScopeID = from.ScopeID
	}
	return pref
}
