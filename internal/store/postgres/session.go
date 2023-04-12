package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/odpf/shield/core/authenticate"
)

type Session struct {
	ID              uuid.UUID `db:"id"`
	UserID          string    `db:"user_id"`
	AuthenticatedAt time.Time `db:"authenticated_at"`
	ExpiresAt       time.Time `db:"expires_at"`
	CreatedAt       time.Time `db:"created_at"`
}

func (s *Session) transformToSession() *authenticate.Session {
	return &authenticate.Session{
		ID:              s.ID,
		UserID:          s.UserID,
		AuthenticatedAt: s.AuthenticatedAt,
		ExpiresAt:       s.ExpiresAt,
		CreatedAt:       s.CreatedAt,
	}
}
