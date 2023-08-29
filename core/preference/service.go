package preference

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Set(ctx context.Context, preference Preference) (Preference, error)
	Get(ctx context.Context, id uuid.UUID) (Preference, error)
	List(ctx context.Context, filter Filter) ([]Preference, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, preference Preference) (Preference, error) {
	// only allow creating preferences for which a trait exists
	allowCreate := false
	for _, trait := range s.Describe(ctx) {
		if trait.Name == preference.Name && trait.ResourceType == preference.ResourceType {
			allowCreate = true
			break
		}
	}
	if !allowCreate {
		return Preference{}, ErrTraitNotFound
	}
	return s.repo.Set(ctx, preference)
}

func (s *Service) Get(ctx context.Context, id string) (Preference, error) {
	prefID, err := uuid.Parse(id)
	if err != nil {
		return Preference{}, ErrInvalidID
	}
	return s.repo.Get(ctx, prefID)
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Preference, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Describe(ctx context.Context) []Trait {
	return DefaultTraits
}
