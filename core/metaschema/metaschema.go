package metaschema

import "context"

type Repository interface {
	Create(ctx context.Context, metaschema MetaSchema) (string, error)
	Get(ctx context.Context, name string) (MetaSchema, error)
	Update(ctx context.Context, name string, metaschema MetaSchema) (string, error)
	List(ctx context.Context) ([]MetaSchema, error)
	Delete(ctx context.Context, name string) error
	CreateDefaultInDB(ctx context.Context) error
	InitMetaSchemas(ctx context.Context) error
}

// MetaSchema represents metadata schema to be validated for users/ groups/ organisations / roles
type MetaSchema struct {
	Name   string
	Schema string
}
