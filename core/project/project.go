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

type Store interface {
	GetProject(ctx context.Context, id string) (Project, error)
	CreateProject(ctx context.Context, org Project) (Project, error)
	ListProject(ctx context.Context) ([]Project, error)
	UpdateProject(ctx context.Context, toUpdate Project) (Project, error)
	GetUsersByIds(ctx context.Context, userIds []string) ([]user.User, error)
	GetUser(ctx context.Context, userId string) (user.User, error)
	ListProjectAdmins(ctx context.Context, id string) ([]user.User, error)
}

type Project struct {
	Id           string
	Name         string
	Slug         string
	Organization organization.Organization
	Metadata     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
