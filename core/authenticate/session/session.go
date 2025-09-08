package session

import (
	"time"

	"github.com/google/uuid"
)

type SessionMetadata struct {
	IP      string
	Location struct {
		Country string
		City    string
	}
	OS      string
	Browser string
}

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
	UpdatedAt time.Time
	DeletedAt *time.Time // Soft delete timestamp (nil = not deleted)

	Metadata SessionMetadata
}

func (s Session) IsValid(now time.Time) bool {
	if s.ExpiresAt.After(now) && !s.AuthenticatedAt.IsZero() && s.DeletedAt == nil {
		return true
	}
	return false
}
