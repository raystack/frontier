package organization

import (
	"context"
	"time"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/pkg/metadata"
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	Enabled  State = "enabled"
	Disabled State = "disabled"

	AdminPermission = schema.UpdatePermission
	AdminRole       = schema.OwnerRelationName
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Organization, error)
	GetByIDs(ctx context.Context, ids []string) ([]Organization, error)
	GetByName(ctx context.Context, name string) (Organization, error)
	Create(ctx context.Context, org Organization) (Organization, error)
	List(ctx context.Context, flt Filter) ([]Organization, error)
	UpdateByID(ctx context.Context, org Organization) (Organization, error)
	UpdateByName(ctx context.Context, org Organization) (Organization, error)
	SetState(ctx context.Context, id string, state State) error
	Delete(ctx context.Context, id string) error
}

type Organization struct {
	ID        string
	Name      string
	Title     string
	Metadata  metadata.Metadata
	State     State
	CreatedAt time.Time
	UpdatedAt time.Time
}
