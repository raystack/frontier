package permission

import (
	"context"
	"fmt"

	"github.com/odpf/shield/model"
	"google.golang.org/grpc/metadata"
)

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
		return "", fmt.Errorf("unable to fetch context from incoming")
	}

	var email string
	metadataValues := md.Get(headerKey)
	if len(metadataValues) > 0 {
		email = metadataValues[0]
	}
	return email, nil
}
