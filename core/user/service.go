package user

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
)

var emailContext = struct{}{}

type Service struct {
	repository          Repository
	identityProxyHeader string
}

func NewService(identityProxyHeader string, repository Repository) *Service {
	return &Service{
		identityProxyHeader: identityProxyHeader,
		repository:          repository,
	}
}

func (s Service) GetByID(ctx context.Context, id string) (User, error) {
	return s.repository.GetByID(ctx, id)
}

func (s Service) GetByIDs(ctx context.Context, userIDs []string) ([]User, error) {
	return s.repository.GetByIDs(ctx, userIDs)
}

func (s Service) GetByEmail(ctx context.Context, email string) (User, error) {
	return s.repository.GetByEmail(ctx, email)
}

func (s Service) Create(ctx context.Context, user User) (User, error) {
	newUser, err := s.repository.Create(ctx, User{
		Name:     user.Name,
		Email:    user.Email,
		Metadata: user.Metadata,
	})
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

func (s Service) List(ctx context.Context, flt Filter) (PagedUsers, error) {
	users, err := s.repository.List(ctx, flt)
	if err != nil {
		return PagedUsers{}, err
	}
	//TODO might better to do this in handler level
	return PagedUsers{
		Count: int32(len(users)),
		Users: users,
	}, nil
}

func (s Service) UpdateByID(ctx context.Context, toUpdate User) (User, error) {
	return s.repository.UpdateByID(ctx, toUpdate)
}

func (s Service) UpdateByEmail(ctx context.Context, toUpdate User) (User, error) {
	return s.repository.UpdateByEmail(ctx, toUpdate)
}

func (s Service) FetchCurrentUser(ctx context.Context) (User, error) {
	email, err := fetchEmailFromMetadata(ctx, s.identityProxyHeader)
	if err != nil {
		return User{}, err
	}

	fetchedUser, err := s.repository.GetByEmail(ctx, email)
	if err != nil {
		return User{}, err
	}

	return fetchedUser, nil
}

// TODO need to simplify this, service package should not depend on grpc metadata
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
