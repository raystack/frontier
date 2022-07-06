package user

import (
	"context"
	"errors"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store Store
}

var (
	UserDoesntExist = errors.New("user doesn't exist")
	InvalidUUID     = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetUser(ctx context.Context, id string) (model.User, error)
	GetCurrentUser(ctx context.Context, email string) (model.User, error)
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	ListUsers(ctx context.Context, limit int32, page int32, keyword string) ([]model.User, error)
	//ListUsers(ctx context.Context) ([]model.User, error)
	UpdateUser(ctx context.Context, toUpdate model.User) (model.User, error)
	UpdateCurrentUser(ctx context.Context, toUpdate model.User) (model.User, error)
	ListUserGroups(ctx context.Context, userId string, roleId string) ([]model.Group, error)
}

func (s Service) GetUser(ctx context.Context, id string) (model.User, error) {
	return s.Store.GetUser(ctx, id)
}

func (s Service) GetCurrentUser(ctx context.Context, email string) (model.User, error) {
	return s.Store.GetCurrentUser(ctx, email)
}

func (s Service) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	newUser, err := s.Store.CreateUser(ctx, model.User{
		Name:     user.Name,
		Email:    user.Email,
		Metadata: user.Metadata,
	})

	if err != nil {
		return model.User{}, err
	}

	return newUser, nil
}

func (s Service) ListUsers(ctx context.Context, limit int32, page int32, keyword string) ([]model.User, error) {
	return s.Store.ListUsers(ctx, limit, page, keyword)
}

func (s Service) UpdateUser(ctx context.Context, toUpdate model.User) (model.User, error) {
	return s.Store.UpdateUser(ctx, toUpdate)
}

func (s Service) UpdateCurrentUser(ctx context.Context, toUpdate model.User) (model.User, error) {
	return s.Store.UpdateCurrentUser(ctx, toUpdate)
}

func (s Service) ListUserGroups(ctx context.Context, userId string, roleId string) ([]model.Group, error) {
	return s.Store.ListUserGroups(ctx, userId, roleId)
}
