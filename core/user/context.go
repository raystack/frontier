package user

import (
	"context"

	"github.com/raystack/shield/pkg/server/consts"
)

type contextEmailKey struct{}

func SetContextWithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, contextEmailKey{}, email)
}

func GetEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(contextEmailKey{}).(string)
	return email, ok
}

func GetUserFromContext(ctx context.Context) (*User, bool) {
	u, ok := ctx.Value(consts.AuthenticatedUserContextKey).(*User)
	return u, ok
}
