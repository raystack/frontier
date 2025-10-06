package session

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/pkg/server/consts"
)

type SessionMetadata struct {
	IpAddress string
	Location  struct {
		Country string
		City    string
	}
	OperatingSystem string
	Browser         string
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

// SetSessionMetadataInContext sets session metadata in context
func SetSessionMetadataInContext(ctx context.Context, metadata SessionMetadata) context.Context {
	return context.WithValue(ctx, consts.SessionContextKey, metadata)
}

// GetSessionMetadataFromContext returns session metadata from context
func GetSessionMetadataFromContext(ctx context.Context) (SessionMetadata, bool) {
	metadata, ok := ctx.Value(consts.SessionContextKey).(SessionMetadata)
	return metadata, ok
}
