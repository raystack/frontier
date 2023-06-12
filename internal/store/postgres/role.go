package postgres

import (
	"encoding/json"
	"time"

	"github.com/raystack/shield/core/role"
)

type Role struct {
	ID          string    `db:"id"`
	OrgID       string    `db:"org_id"`
	Name        string    `db:"name"`
	Permissions []byte    `db:"permissions"`
	State       string    `db:"state"`
	Metadata    []byte    `db:"metadata"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (from Role) transformToRole() (role.Role, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return role.Role{}, err
		}
	}
	var unmarshalledPermissions []string
	if len(from.Permissions) > 0 {
		if err := json.Unmarshal(from.Permissions, &unmarshalledPermissions); err != nil {
			return role.Role{}, err
		}
	}

	return role.Role{
		ID:          from.ID,
		Name:        from.Name,
		OrgID:       from.OrgID,
		Permissions: unmarshalledPermissions,
		Metadata:    unmarshalledMetadata,
		State:       role.State(from.State),
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
	}, nil
}
