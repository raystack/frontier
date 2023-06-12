package postgres

import (
	"encoding/json"
	"time"

	"database/sql"

	"github.com/raystack/shield/core/namespace"
)

type Namespace struct {
	ID        string       `db:"id"`
	Name      string       `db:"name"`
	Metadata  []byte       `db:"metadata"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

func (from Namespace) transformToNamespace() (namespace.Namespace, error) {
	var unmarshalledMetadata map[string]any
	if len(from.Metadata) > 0 {
		if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
			return namespace.Namespace{}, err
		}
	}

	return namespace.Namespace{
		ID:        from.ID,
		Name:      from.Name,
		Metadata:  unmarshalledMetadata,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}, nil
}
