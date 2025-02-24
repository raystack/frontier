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
		Phone:    audience.Phone,
		Activity: audience.Activity,
		Status:   audience.Status,
		Source:   audience.Source,
		Verified: true, // if user is logged in on platform them we already would have already verified the email
		Metadata: audience.Metadata,
	})
}
