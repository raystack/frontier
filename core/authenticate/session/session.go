package session

import (
	"time"

	"github.com/raystack/shield/pkg/metadata"

	"github.com/google/uuid"
)

// Session is created on successful authentication of users
type Session struct {
	ID uuid.UUID

	// UserID is a unique identifier for logged in users
	UserID string

	// AuthenticatedAt is set when a user is successfully authn
	AuthenticatedAt time.Time

	// ExpiresAt is ideally now() + lifespan of session, e.g. 7 days
	ExpiresAt time.Time
	CreatedAt time.Time

	Metadata metadata.Metadata
}

func (s Session) IsValid(now time.Time) bool {
	if s.ExpiresAt.After(now) && !s.AuthenticatedAt.IsZero() {
		return true
	}
	return false
}
