package permission

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

func (s Service) Get(ctx context.Context, id string) (Permission, error) {
	if shielduuid.IsValid(id) {
		return s.repository.Get(ctx, id)
	}
	id = convertHashPermissionToSlug(id)
	return s.repository.GetBySlug(ctx, id)
}

func (s Service) Upsert(ctx context.Context, perm Permission) (Permission, error) {
	if perm.Slug == "" {
		perm.Slug = perm.GenerateSlug()
	}
	return s.repository.Upsert(ctx, perm)
}

func (s Service) List(ctx context.Context) ([]Permission, error) {
	return s.repository.List(ctx)
}

func (s Service) Update(ctx context.Context, perm Permission) (Permission, error) {
	if perm.Slug == "" {
		perm.Slug = perm.GenerateSlug()
	}
	return s.repository.Update(ctx, perm)
}

// Delete call over a service could be dangerous without removing all of its relations
// the method does not do it by default
func (s Service) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}
