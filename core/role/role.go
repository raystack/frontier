package role

import (
	"context"
	"time"

	"github.com/raystack/shield/pkg/metadata"
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	Enabled  State = "enabled"
	Disabled State = "disabled"
)

type Repository interface {
	Get(ctx context.Context, id string) (Role, error)
	GetByName(ctx context.Context, orgID, name string) (Role, error)
	List(ctx context.Context, f Filter) ([]Role, error)
	Upsert(ctx context.Context, role Role) (string, error)
	Update(ctx context.Context, toUpdate Role) (string, error)
	Delete(ctx context.Context, roleID string) error
}

type Role struct {
	ID          string
	OrgID       string
	Name        string
	Permissions []string
	State       State
	Metadata    metadata.Metadata
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
