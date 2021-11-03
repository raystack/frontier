package user

import (
	"context"
	"errors"
	"time"
)

type User struct {
	Id        string
	Name      string
	Email     string
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Service struct {
	Store Store
}

var (
	UserDoesntExist = errors.New("user doesn't exist")
	InvalidUUID     = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetUser(ctx context.Context, id string) (User, error)
	CreateUser(ctx context.Context, user User) (User, error)
	ListUsers(ctx context.Context) ([]User, error)
	UpdateUser(ctx context.Context, toUpdate User) (User, error)
}

func (s Service) GetUser(ctx context.Context, id string) (User, error) {
	return s.Store.GetUser(ctx, id)
}

func (s Service) CreateUser(ctx context.Context, user User) (User, error) {
	newUser, err := s.Store.CreateUser(ctx, User{
		Name:     user.Name,
		Email:    user.Email,
		Metadata: user.Metadata,
	})

	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

func (s Service) ListUsers(ctx context.Context) ([]User, error) {
	return s.Store.ListUsers(ctx)
}

func (s Service) UpdateUser(ctx context.Context, toUpdate User) (User, error) {
	return s.Store.UpdateUser(ctx, toUpdate)
}
