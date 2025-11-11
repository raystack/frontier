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
		Country   string
		City      string
		Latitude  string
		Longitude string
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
// It accepts a SessionMetadata struct but stores it as a map with the same structure to avoid layer violations in repositories
func SetSessionMetadataInContext(ctx context.Context, metadata SessionMetadata) context.Context {
	metadataMap := map[string]interface{}{
		"IpAddress": metadata.IpAddress,
		"Location": map[string]interface{}{
			"Country":   metadata.Location.Country,
			"City":      metadata.Location.City,
			"Latitude":  metadata.Location.Latitude,
			"Longitude": metadata.Location.Longitude,
		},
		"OperatingSystem": metadata.OperatingSystem,
		"Browser":         metadata.Browser,
	}
	return context.WithValue(ctx, consts.SessionContextKey, metadataMap)
}
