package metaschema

import (
	"context"

	shielduuid "github.com/odpf/shield/pkg/uuid"
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
	return s.repository.Create(ctx, toCreate)
}

func (s Service) Get(ctx context.Context, id string) (MetaSchema, error) {
	if shielduuid.IsValid(id) {
		return s.repository.Get(ctx, id)
	}
	return MetaSchema{}, ErrInvalidID
}

func (s Service) List(ctx context.Context) ([]MetaSchema, error) {
	return s.repository.List(ctx)
}

func (s Service) Update(ctx context.Context, id string, toUpdate MetaSchema) (MetaSchema, error) {
	if shielduuid.IsValid(id) {
		return s.repository.Update(ctx, id, toUpdate)
	}
	return MetaSchema{}, ErrInvalidID
}

func (s Service) Delete(ctx context.Context, id string) (string, error) {
	mschema, err := s.repository.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return mschema.Name, s.repository.Delete(ctx, id)
}

func (s Service) InitMetaSchemas(ctx context.Context) (map[string]string, error) {
	return s.repository.InitMetaSchemas(ctx)
}

func (s Service) MigrateDefault(ctx context.Context) error {
	return s.repository.CreateDefaultInDB(ctx)
}
