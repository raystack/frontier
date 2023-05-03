package organization

import (
	"context"
	"time"

	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/schema"

	"github.com/odpf/shield/pkg/metadata"
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	Enabled  State = "enabled"
	Disabled State = "disabled"

	AdminPermission = schema.EditPermission
	AdminRole       = schema.OwnerRole
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Organization, error)
	GetByIDs(ctx context.Context, ids []string) ([]Organization, error)
	GetBySlug(ctx context.Context, slug string) (Organization, error)
	Create(ctx context.Context, org Organization) (Organization, error)
	List(ctx context.Context, flt Filter) ([]Organization, error)
	UpdateByID(ctx context.Context, org Organization) (Organization, error)
	UpdateBySlug(ctx context.Context, org Organization) (Organization, error)
	SetState(ctx context.Context, id string, state State) error
	Delete(ctx context.Context, id string) error
}

type Organization struct {
	ID        string
	Name      string
	Slug      string
	Metadata  metadata.Metadata
	State     State
	CreatedAt time.Time
	UpdatedAt time.Time
}

func BuildUserOrgAdminSubject(user user.User) relation.Subject {
	return relation.Subject{
		ID:        user.ID,
		Namespace: schema.UserPrincipal,
		RoleID:    schema.OwnerRole,
	}
}
