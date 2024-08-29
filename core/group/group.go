package group

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

	AdminPermission = schema.DeletePermission

	MemberRole = schema.GroupMemberRole
	AdminRole  = schema.GroupOwnerRole
)

type Repository interface {
	Create(ctx context.Context, grp Group) (Group, error)
	GetByID(ctx context.Context, id string) (Group, error)
	GetByIDs(ctx context.Context, groupIDs []string, flt Filter) ([]Group, error)
	List(ctx context.Context, flt Filter) ([]Group, error)
	UpdateByID(ctx context.Context, toUpdate Group) (Group, error)
	SetState(ctx context.Context, id string, state State) error
	Delete(ctx context.Context, id string) error
}

type Group struct {
	ID             string
	Name           string
	Title          string
	OrganizationID string `json:"orgId"`
	Metadata       metadata.Metadata
	State          State
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Transient fields
	MemberCount int
}
