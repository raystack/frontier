package metaschema

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, metaschema MetaSchema) (MetaSchema, error)
	Get(ctx context.Context, id string) (MetaSchema, error)
	Update(ctx context.Context, id string, metaschema MetaSchema) (MetaSchema, error)
	List(ctx context.Context) ([]MetaSchema, error)
	Delete(ctx context.Context, id string) error
	CreateDefaultInDB(ctx context.Context) error
	InitMetaSchemas(ctx context.Context) (map[string]string, error)
}

// MetaSchema represents metadata schema to be validated for users/ groups/ organisations / roles
type MetaSchema struct {
	ID        string
	Name      string
	Schema    string
	CreatedAt time.Time
	UpdatedAt time.Time
}
