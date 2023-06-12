package postgres

import (
	"time"

	"github.com/raystack/shield/core/authenticate/session"

	"github.com/google/uuid"
)

type Session struct {
	ID              uuid.UUID `db:"id"`
	UserID          uuid.UUID `db:"user_id"`
	AuthenticatedAt time.Time `db:"authenticated_at"`
	ExpiresAt       time.Time `db:"expires_at"`
	CreatedAt       time.Time `db:"created_at"`
}

func (s *Session) transformToSession() *session.Session {
	return &session.Session{
		ID:              s.ID,
		UserID:          s.UserID.String(),
		AuthenticatedAt: s.AuthenticatedAt,
		ExpiresAt:       s.ExpiresAt,
		CreatedAt:       s.CreatedAt,
	}
}
