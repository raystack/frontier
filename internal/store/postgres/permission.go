package postgres

import (
	"encoding/json"
	"time"

	"github.com/odpf/shield/core/permission"
)

type Permission struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Slug        string    `db:"slug"`
	Namespace   Namespace `db:"namespace"`
	NamespaceID string    `db:"namespace_name"`
	Metadata    []byte    `db:"metadata"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type returnedColumns struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Slug        string    `db:"slug"`
	NamespaceID string    `db:"namespace_name"`
	Metadata    []byte    `db:"metadata"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (from Permission) transformToPermission() (permission.Permission, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return permission.Permission{}, err
		}
	}

	return permission.Permission{
		ID:          from.ID,
		Name:        from.Name,
		Slug:        from.Slug,
		NamespaceID: from.NamespaceID,
		Metadata:    unmarshalledMetadata,
	}, nil
}
