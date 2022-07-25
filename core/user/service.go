package user

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
)

var emailContext = struct{}{}

type Service struct {
	store               Store
	identityProxyHeader string
}

func NewService(identityProxyHeader string, store Store) *Service {
	return &Service{
		identityProxyHeader: identityProxyHeader,
		store:               store,
	}
}

func (s Service) GetUser(ctx context.Context, id string) (User, error) {
	return s.store.GetUser(ctx, id)
}

func (s Service) GetCurrentUser(ctx context.Context, email string) (User, error) {
	return s.store.GetCurrentUser(ctx, email)
}

func (s Service) CreateUser(ctx context.Context, user User) (User, error) {
	newUser, err := s.store.CreateUser(ctx, User{
		Name:     user.Name,
		Email:    user.Email,
		Metadata: user.Metadata,
	})

	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

func (s Service) ListUsers(ctx context.Context, limit int32, page int32, keyword string) (PagedUsers, error) {
	return s.store.ListUsers(ctx, limit, page, keyword)
}

func (s Service) UpdateUser(ctx context.Context, toUpdate User) (User, error) {
	return s.store.UpdateUser(ctx, toUpdate)
}

func (s Service) UpdateCurrentUser(ctx context.Context, toUpdate User) (User, error) {
	return s.store.UpdateCurrentUser(ctx, toUpdate)
}

func (s Service) FetchCurrentUser(ctx context.Context) (User, error) {
	email, err := fetchEmailFromMetadata(ctx, s.identityProxyHeader)
	if err != nil {
		return User{}, err
	}

	fetchedUser, err := s.store.GetCurrentUser(ctx, email)
	if err != nil {
		return User{}, err
	}

	return fetchedUser, nil
}

func fetchEmailFromMetadata(ctx context.Context, headerKey string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		val, ok := GetEmailFromContext(ctx)
		if !ok {
			return "", fmt.Errorf("unable to fetch context from incoming")
		}

		return val, nil
	}

	var email string
	metadataValues := md.Get(headerKey)
	if len(metadataValues) > 0 {
		email = metadataValues[0]
	}
	return email, nil
}

func SetEmailToContext(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, emailContext, email)
}

func GetEmailFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(emailContext).(string)

	return val, ok
}
