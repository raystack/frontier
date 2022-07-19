package user

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotExist    = errors.New("user doesn't exist")
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetUser(ctx context.Context, id string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	CreateUser(ctx context.Context, user User) (User, error)
	ListUsers(ctx context.Context, limit int32, page int32, keyword string) (PagedUsers, error)
	UpdateUser(ctx context.Context, toUpdate User) (User, error)
	UpdateCurrentUser(ctx context.Context, toUpdate User) (User, error)
}

type User struct {
	Id        string
	Name      string
	Email     string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PagedUsers struct {
	Count int32
	Users []User
}
