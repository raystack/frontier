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

type Repository interface {
	Get(ctx context.Context, id string) (Organization, error)
	Create(ctx context.Context, org Organization) (Organization, error)
	List(ctx context.Context) ([]Organization, error)
	Update(ctx context.Context, toUpdate Organization) (Organization, error)
	ListAdmins(ctx context.Context, id string) ([]user.User, error)
}

type Organization struct {
	ID        string
	Name      string
	Slug      string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}
