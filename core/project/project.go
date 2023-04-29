package project

import (
	"context"
	"time"

	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/pkg/metadata"
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
	GetByID(ctx context.Context, id string) (Project, error)
	GetByIDs(ctx context.Context, ids []string) ([]Project, error)
	GetBySlug(ctx context.Context, slug string) (Project, error)
	Create(ctx context.Context, org Project) (Project, error)
	List(ctx context.Context, f Filter) ([]Project, error)
	UpdateByID(ctx context.Context, toUpdate Project) (Project, error)
	UpdateBySlug(ctx context.Context, toUpdate Project) (Project, error)
	Delete(ctx context.Context, id string) error
	SetState(ctx context.Context, id string, state State) error
}

type Project struct {
	ID           string
	Name         string
	Slug         string
	Organization organization.Organization
	State        State
	Metadata     metadata.Metadata
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
