package organization

import (
	"context"
	"time"

	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/metadata"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Organization, error)
	GetBySlug(ctx context.Context, slug string) (Organization, error)
	Create(ctx context.Context, org Organization) (Organization, error)
	List(ctx context.Context) ([]Organization, error)
	UpdateByID(ctx context.Context, org Organization) (Organization, error)
	UpdateBySlug(ctx context.Context, org Organization) (Organization, error)
	ListAdminsByOrgID(ctx context.Context, id string) ([]user.User, error)
}

type Organization struct {
	ID        string
	Name      string
	Slug      string
	Metadata  metadata.Metadata
	CreatedAt time.Time
	UpdatedAt time.Time
}
