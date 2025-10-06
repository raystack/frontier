package authenticate

import (
	"context"

	"github.com/raystack/frontier/pkg/utils"
	"google.golang.org/grpc/metadata"

	"github.com/raystack/frontier/pkg/server/consts"
)

// contextEmailKey should not be used in production
// Deprecated
type contextEmailKey struct{}

// SetContextWithEmail sets email in context
// Deprecated
func SetContextWithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, contextEmailKey{}, email)
}

// GetEmailFromContext returns email from context
// Deprecated
func GetEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(contextEmailKey{}).(string)
	if !utils.IsValidEmail(email) {
		return "", false
	}
	return email, ok
}

func GetPrincipalFromContext(ctx context.Context) (*Principal, bool) {
	u, ok := ctx.Value(consts.AuthenticatedPrincipalContextKey).(*Principal)
	return u, ok
}

func SetContextWithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, consts.AuthenticatedPrincipalContextKey, p)
}

func GetTokenFromContext(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	tokenHeaders := md.Get(consts.UserTokenGatewayKey)
	if len(tokenHeaders) == 0 || len(tokenHeaders[0]) == 0 {
		return "", false
	}
	return tokenHeaders[0], true
}

func GetSecretFromContext(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}
	secretHeaders := md.Get(consts.UserSecretGatewayKey)
	if len(secretHeaders) == 0 || len(secretHeaders[0]) == 0 {
		return "", false
	}
	return secretHeaders[0], true
}

// SetSuperUserInContext sets the superuser flag in context
func SetSuperUserInContext(ctx context.Context, isSuperUser bool) context.Context {
	return context.WithValue(ctx, consts.AuthSuperUserContextKey, isSuperUser)
}

// GetSuperUserFromContext returns the superuser flag from context
func GetSuperUserFromContext(ctx context.Context) (bool, bool) {
	isSuperUser, ok := ctx.Value(consts.AuthSuperUserContextKey).(bool)
	return isSuperUser, ok
}

// GetPrincipalName returns the name/title from a principal
func GetPrincipalName(principal *Principal) string {
	if principal == nil {
		return "unknown"
	}
	if principal.User != nil {
		return principal.User.Title
	}
	if principal.ServiceUser != nil {
		return principal.ServiceUser.Title
	}
	return "unknown"
}
