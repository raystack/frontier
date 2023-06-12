package policy

import (
	"context"
	"time"

	"github.com/raystack/shield/pkg/metadata"
)

type Repository interface {
	Get(ctx context.Context, id string) (Policy, error)
	List(ctx context.Context, f Filter) ([]Policy, error)
	Upsert(ctx context.Context, pol Policy) (string, error)
	Delete(ctx context.Context, id string) error
}

type Policy struct {
	ID            string
	RoleID        string
	ResourceID    string `json:"resource_id"`
	ResourceType  string `json:"resource_type"`
	PrincipalID   string `json:"principal_id"`
	PrincipalType string `json:"principal_type"`
	Metadata      metadata.Metadata

	CreatedAt time.Time
	UpdatedAt time.Time
}

type Filters struct {
	UserID  string
	GroupID string
}
