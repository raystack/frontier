package postgres

import (
	"time"

	"github.com/raystack/shield/core/metaschema"
)

type MetaSchema struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Schema    string    `db:"schema"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (from MetaSchema) tranformtoMetadataSchema() metaschema.MetaSchema {
	return metaschema.MetaSchema{
		ID:        from.ID,
		Name:      from.Name,
		Schema:    from.Schema,
		CreatedAt: from.CreatedAt,
		UpdatedAt: from.UpdatedAt,
	}
}
