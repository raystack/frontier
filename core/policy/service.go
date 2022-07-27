package policy

import (
	"context"
)

type Service struct {
	repository      Repository
	authzRepository AuthzRepository
}

func NewService(repository Repository, authzRepository AuthzRepository) *Service {
	return &Service{
		repository:      repository,
		authzRepository: authzRepository,
	}
}

func (s Service) Get(ctx context.Context, id string) (Policy, error) {
	return s.repository.Get(ctx, id)
}

func (s Service) List(ctx context.Context) ([]Policy, error) {
	return s.repository.List(ctx)
}

func (s Service) Create(ctx context.Context, policy Policy) ([]Policy, error) {
	if _, err := s.repository.Create(ctx, policy); err != nil {
		return []Policy{}, err
	}
	policies, err := s.repository.List(ctx)
	if err != nil {
		return []Policy{}, err
	}
	if err = s.authzRepository.Add(ctx, policies); err != nil {
		return []Policy{}, err
	}
	return policies, err
}

func (s Service) Update(ctx context.Context, id string, policy Policy) ([]Policy, error) {
	if _, err := s.repository.Update(ctx, id, policy); err != nil {
		return []Policy{}, err
	}
	policies, err := s.repository.List(ctx)
	if err != nil {
		return []Policy{}, err
	}
	if err = s.authzRepository.Add(ctx, policies); err != nil {
		return []Policy{}, err
	}
	return policies, err
}
