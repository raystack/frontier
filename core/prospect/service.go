package prospect

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

func (s *Service) Get(ctx context.Context, prospectId string) (Prospect, error) {
	return s.repository.Get(ctx, prospectId)
}

func (s *Service) Delete(ctx context.Context, prospectId string) error {
	return s.repository.Delete(ctx, prospectId)
}

func (s *Service) Create(ctx context.Context, prospect Prospect) (Prospect, error) {
	return s.repository.Create(ctx, Prospect{
		Name:     prospect.Name,
		Email:    strings.ToLower(prospect.Email),
		Verified: prospect.Verified,
		Phone:    prospect.Phone,
		Activity: prospect.Activity,
		Status:   prospect.Status,
		Source:   prospect.Source,
		Metadata: prospect.Metadata,
	})
}

func (s *Service) List(ctx context.Context, filters Filter) ([]Prospect, error) {
	return s.repository.List(ctx, filters)
}

func (s *Service) Update(ctx context.Context, prospect Prospect) (Prospect, error) {
	return s.repository.Update(ctx, Prospect{
		ID:       prospect.ID,
		Name:     prospect.Name,
		Email:    strings.ToLower(prospect.Email),
		Phone:    prospect.Phone,
		Activity: prospect.Activity,
		Status:   prospect.Status,
		Source:   prospect.Source,
		Verified: prospect.Verified,
		Metadata: prospect.Metadata,
	})
}
