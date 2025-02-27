package audience

import (
	"context"
	"strings"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, audience Audience) (Audience, error) {
	return s.repository.Create(ctx, Audience{
		Name:     audience.Name,
		Email:    strings.ToLower(audience.Email),
		Verified: audience.Verified,
		Phone:    audience.Phone,
		Activity: audience.Activity,
		Status:   audience.Status,
		Source:   audience.Source,
		Metadata: audience.Metadata,
	})
}

func (s *Service) List(ctx context.Context, filters Filter) ([]Audience, error) {
	return s.repository.List(ctx, filters)
}

func (s *Service) Update(ctx context.Context, audience Audience) (Audience, error) {
	return s.repository.Update(ctx, Audience{
		ID:       audience.ID,
		Name:     audience.Name,
		Email:    strings.ToLower(audience.Email),
		Phone:    audience.Phone,
		Activity: audience.Activity,
		Status:   audience.Status,
		Source:   audience.Source,
		Verified: audience.Verified,
		Metadata: audience.Metadata,
	})
}
