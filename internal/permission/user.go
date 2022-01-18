package permission

import (
	"context"
	"fmt"

	"github.com/odpf/shield/model"
	"google.golang.org/grpc/metadata"
)

const emailContext = "email-context"

func (s Service) FetchCurrentUser(ctx context.Context) (model.User, error) {
	email, err := fetchEmailFromMetadata(ctx, s.IdentityProxyHeader)
	if err != nil {
		return model.User{}, err
	}

	fetchedUser, err := s.Store.GetCurrentUser(ctx, email)

	if err != nil {
		return model.User{}, err
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
