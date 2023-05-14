package metaschema

import (
	"context"

	"github.com/odpf/shield/pkg/metadata"
	shielduuid "github.com/odpf/shield/pkg/uuid"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type Service struct {
	repository      Repository
	metaSchemaCache map[string]MetaSchema
}

func NewService(repository Repository) *Service {
	return &Service{
		repository:      repository,
		metaSchemaCache: make(map[string]MetaSchema),
	}
}

func (s Service) Create(ctx context.Context, toCreate MetaSchema) (MetaSchema, error) {
	mschema, err := s.repository.Create(ctx, toCreate)
	if err != nil {
		return MetaSchema{}, err
	}
	s.metaSchemaCache[mschema.Name] = mschema
	return mschema, nil
}

func (s Service) Get(ctx context.Context, idOrName string) (MetaSchema, error) {
	if s.metaSchemaCache[idOrName] != (MetaSchema{}) {
		return s.metaSchemaCache[idOrName], nil
	}

	if shielduuid.IsValid(idOrName) {
		schema, err := s.repository.Get(ctx, idOrName)
		if err != nil {
			return MetaSchema{}, err
		}

		return schema, nil
	}
	return MetaSchema{}, ErrInvalidID
}

func (s Service) List(ctx context.Context) ([]MetaSchema, error) {
	if len(s.metaSchemaCache) == 0 {
		schemas, err := s.repository.List(ctx)
		if err != nil {
			return nil, err
		}
		for _, schema := range schemas {
			s.metaSchemaCache[schema.Name] = schema
		}
		return schemas, nil
	}

	var schemas []MetaSchema
	for _, schema := range s.metaSchemaCache {
		schemas = append(schemas, schema)
	}
	return schemas, nil
}

func (s Service) Update(ctx context.Context, id string, toUpdate MetaSchema) (MetaSchema, error) {
	if shielduuid.IsValid(id) {
		schema, err := s.repository.Update(ctx, id, toUpdate)
		if err != nil {
			return MetaSchema{}, err
		}
		s.metaSchemaCache[schema.Name] = schema
		return schema, nil
	}
	return MetaSchema{}, ErrInvalidID
}

func (s Service) Delete(ctx context.Context, id string) error {
	if shielduuid.IsValid(id) {
		name, err := s.repository.Delete(ctx, id)
		if err != nil {
			return err
		}

		delete(s.metaSchemaCache, name)
		return nil
	}
	return ErrInvalidID
}

func (s Service) MigrateDefault(ctx context.Context) error {
	return s.repository.MigrateDefaults(ctx)
}

// validates the metadata against the json-schema. In case metaschema doesn't exists in the cache, it will return nil (no validation)
func (s Service) Validate(mdata metadata.Metadata, name string) error {
	if s.metaSchemaCache[name] == (MetaSchema{}) {
		return nil
	}
	metadataSchema := gojsonschema.NewStringLoader(s.metaSchemaCache[name].Schema)
	providedSchema := gojsonschema.NewGoLoader(mdata)
	results, err := gojsonschema.Validate(metadataSchema, providedSchema)
	if err != nil {
		return errors.Wrap(err, "failed to validate metadata")
	}

	if !results.Valid() {
		return errors.New("metadata doesn't match the json-schema")
	}
	return nil
}
