package audience

import (
	"context"
	"net/mail"
	"strings"
	"time"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, audience Audience) (Audience, error) {
	email := audience.Email
	if !isValidEmail(email) {
		return Audience{}, InvalidEmail
	}
	return s.repository.Create(ctx, Audience{
		Name:      audience.Name,
		Email:     strings.ToLower(email),
		Phone:     audience.Phone,
		Activity:  audience.Activity,
		Status:    audience.Status,
		Source:    audience.Source,
		Verified:  true, // if user is logged in on platform them we already would have already verified the email
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  audience.Metadata,
	})
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
