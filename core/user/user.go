package user

import (
	"context"
	"time"

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
	GetByID(ctx context.Context, id string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByIDs(ctx context.Context, userIds []string) ([]User, error)
	GetBySlug(ctx context.Context, slug string) (User, error)
	Create(ctx context.Context, user User) (User, error)
	List(ctx context.Context, flt Filter) ([]User, error)
	UpdateByID(ctx context.Context, toUpdate User) (User, error)
	UpdateBySlug(ctx context.Context, toUpdate User) (User, error)
	UpdateByEmail(ctx context.Context, toUpdate User) (User, error)
	Delete(ctx context.Context, id string) error
	SetState(ctx context.Context, id string, state State) error
}

type User struct {
	ID        string
	Name      string
	Email     string
	State     State
	Slug      string
	Metadata  metadata.Metadata
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PagedUsers struct {
	Count int32
	Users []User
}
