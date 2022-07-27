package postgres

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"github.com/odpf/shield/core/role"
)

type Role struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	Types       pq.StringArray `db:"types"`
	Namespace   Namespace      `db:"namespace"`
	NamespaceID string         `db:"namespace_id"`
	Metadata    []byte         `db:"metadata"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

func (from Role) transformToRole() (role.Role, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return role.Role{}, err
		}
	}

	return role.Role{
		ID:          from.ID,
		Name:        from.Name,
		Types:       from.Types,
		Namespace:   from.Namespace.transformToNamespace(),
		NamespaceID: from.NamespaceID,
		Metadata:    unmarshalledMetadata,
		CreatedAt:   from.CreatedAt,
		UpdatedAt:   from.UpdatedAt,
	}, nil
}
