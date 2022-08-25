package user

import (
	"context"
	"time"

	"github.com/odpf/shield/pkg/metadata"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByIDs(ctx context.Context, userIds []string) ([]User, error)
	Create(ctx context.Context, user User) (User, error)
	List(ctx context.Context, flt Filter) ([]User, error)
	UpdateByID(ctx context.Context, toUpdate User) (User, error)
	UpdateByEmail(ctx context.Context, toUpdate User) (User, error)
}

type User struct {
	ID        string
	Name      string
	Email     string
	Metadata  metadata.Metadata
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserMetadataKey struct {
	ID          string
	Key         string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PagedUsers struct {
	Count int32
	Users []User
}
