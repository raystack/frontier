package role

import "context"

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

func (s Service) Create(ctx context.Context, toCreate Role) (Role, error) {
	return s.store.CreateRole(ctx, toCreate)
}

func (s Service) Get(ctx context.Context, id string) (Role, error) {
	return s.store.GetRole(ctx, id)
}

func (s Service) List(ctx context.Context) ([]Role, error) {
	return s.store.ListRoles(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate Role) (Role, error) {
	return s.store.UpdateRole(ctx, toUpdate)
}
