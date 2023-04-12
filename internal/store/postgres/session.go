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

type Flow struct {
	ID        uuid.UUID `db:"id"`
	Method    string    `db:"method"`
	StartURL  string    `db:"start_url"`
	FinishURL string    `db:"finish_url"`
	Nonce     string    `db:"nonce"`
	CreatedAt time.Time `db:"created_at"`
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

func (f *Flow) transformToFlow() *authenticate.Flow {
	return &authenticate.Flow{
		ID:        f.ID,
		Method:    f.Method,
		StartURL:  f.StartURL,
		FinishURL: f.FinishURL,
		Nonce:     f.Nonce,
		CreatedAt: f.CreatedAt,
	}
}
