package project

import (
	"context"
	"time"

	"github.com/raystack/frontier/core/organization"

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
)

var AdminPermission = schema.DeletePermission
var MemberPermission = schema.GetPermission
var OwnerRole = schema.RoleProjectOwner

type Repository interface {
	GetByID(ctx context.Context, id string) (Project, error)
	GetByName(ctx context.Context, slug string) (Project, error)
	Create(ctx context.Context, org Project) (Project, error)
	List(ctx context.Context, f Filter) ([]Project, error)
	UpdateByID(ctx context.Context, toUpdate Project) (Project, error)
	UpdateByName(ctx context.Context, toUpdate Project) (Project, error)
	Delete(ctx context.Context, id string) error
	SetState(ctx context.Context, id string, state State) error
}

type Project struct {
	ID           string
	Name         string
	Title        string
	Organization organization.Organization
	State        State
	Metadata     metadata.Metadata

	CreatedAt time.Time
	UpdatedAt time.Time

	// Transient fields
	MemberCount int
}

type Principal struct {
	Type string
	ID   string
}
