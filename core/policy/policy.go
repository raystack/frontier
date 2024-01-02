package policy

import (
	"context"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

type Repository interface {
	Get(ctx context.Context, id string) (Policy, error)
	List(ctx context.Context, f Filter) ([]Policy, error)
	Count(ctx context.Context, f Filter) (int64, error)
	Upsert(ctx context.Context, pol Policy) (Policy, error)
	Delete(ctx context.Context, id string) error
	GroupMemberCount(ctx context.Context, IDs []string) ([]MemberCount, error)
	ProjectMemberCount(ctx context.Context, IDs []string) ([]MemberCount, error)
	OrgMemberCount(ctx context.Context, ID string) (MemberCount, error)
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

type MemberCount struct {
	ID    string `db:"id"`
	Count int    `db:"count"`
}
