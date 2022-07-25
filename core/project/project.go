package project

import (
	"context"
	"errors"
	"time"

	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/user"
)

var (
	ErrNotExist      = errors.New("project doesn't exist")
	ErrNoAdminsExist = errors.New("no admins exist")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
)

type Repository interface {
	Get(ctx context.Context, id string) (Project, error)
	Create(ctx context.Context, org Project) (Project, error)
	List(ctx context.Context) ([]Project, error)
	Update(ctx context.Context, toUpdate Project) (Project, error)
	ListAdmins(ctx context.Context, id string) ([]user.User, error)
}

type Project struct {
	ID           string
	Name         string
	Slug         string
	Organization organization.Organization
	Metadata     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
