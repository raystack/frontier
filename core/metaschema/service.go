package metaschema

import (
	"context"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s Service) Create(ctx context.Context, toCreate MetaSchema) (MetaSchema, error) {
	name, err := s.repository.Create(ctx, toCreate)
	if err != nil {
		return MetaSchema{}, err
	}
	return s.repository.Get(ctx, name)
}

func (s Service) Get(ctx context.Context, name string) (MetaSchema, error) {
	return s.repository.Get(ctx, name)
}

func (s Service) List(ctx context.Context) ([]MetaSchema, error) {
	return s.repository.List(ctx)
}

func (s Service) Update(ctx context.Context, name string, toUpdate MetaSchema) (MetaSchema, error) {
	name, err := s.repository.Update(ctx, name, toUpdate)
	if err != nil {
		return MetaSchema{}, err
	}
	return s.repository.Get(ctx, name)
}

func (s Service) Delete(ctx context.Context, name string) error {
	return s.repository.Delete(ctx, name)
}

func (s Service) InitMetaSchemas(ctx context.Context) (map[string]string, error) {
	return s.repository.InitMetaSchemas(ctx)
}

func (s Service) MigrateDefault(ctx context.Context) error {
	return s.repository.CreateDefaultInDB(ctx)
}
