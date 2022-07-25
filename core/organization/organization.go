package organization

import (
	"context"
	"errors"
	"time"

	"github.com/odpf/shield/core/user"
)

var (
	ErrNotExist      = errors.New("org doesn't exist")
	ErrNoAdminsExist = errors.New("no admins exist")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetOrg(ctx context.Context, id string) (Organization, error)
	CreateOrg(ctx context.Context, org Organization) (Organization, error)
	ListOrg(ctx context.Context) ([]Organization, error)
	UpdateOrg(ctx context.Context, toUpdate Organization) (Organization, error)
	GetUsersByIds(ctx context.Context, userIds []string) ([]user.User, error)
	GetUser(ctx context.Context, userId string) (user.User, error)
	ListOrgAdmins(ctx context.Context, id string) ([]user.User, error)
}

type Organization struct {
	Id        string
	Name      string
	Slug      string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}
