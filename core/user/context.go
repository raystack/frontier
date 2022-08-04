package user

import "context"

type contextEmailKey struct{}

func SetContextWithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, contextEmailKey{}, email)
}

func GetEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(contextEmailKey{}).(string)
	return email, ok
}
