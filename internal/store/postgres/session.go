package postgres

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/raystack/frontier/core/authenticate/session"

	"github.com/google/uuid"
)

type Session struct {
	ID              uuid.UUID  `db:"id"`
	UserID          uuid.UUID  `db:"user_id"`
	AuthenticatedAt time.Time  `db:"authenticated_at"`
	ExpiresAt       time.Time  `db:"expires_at"`
	Metadata        []byte     `db:"metadata"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
	DeletedAt       *time.Time `db:"deleted_at"`
}

func (s *Session) transformToSession() (*session.Session, error) {
	var unmarshalledMetadata map[string]any
	if err := json.Unmarshal(s.Metadata, &unmarshalledMetadata); err != nil {
		return nil, fmt.Errorf("error marshaling session: %w", err)
	}

	return &session.Session{
		ID:              s.ID,
		UserID:          s.UserID.String(),
		AuthenticatedAt: s.AuthenticatedAt,
		ExpiresAt:       s.ExpiresAt,
		Metadata:        unmarshalledMetadata,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
		DeletedAt:       s.DeletedAt,
	}, nil
}
