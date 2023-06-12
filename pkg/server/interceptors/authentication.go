package interceptors

import (
	"context"
	"errors"

	"github.com/raystack/shield/internal/api/v1beta1"
	"github.com/raystack/shield/pkg/server/consts"
	"google.golang.org/grpc"
)

func UnaryAuthenticationCheck(identityHeader string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if len(identityHeader) != 0 {
			// if configured, skip
			return handler(ctx, req)
		}
		if authenticationSkipList[info.FullMethod] {
			return handler(ctx, req)
		}

		// authentication can be done via
		// - session id sent via cookies
		// - jwt token
		// - api key (not supported yet)
		serverHandler, ok := info.Server.(*v1beta1.Handler)
		if !ok {
			return nil, errors.New("miss-configured server handler")
		}

		currentUser, err := serverHandler.GetLoggedInUser(ctx)
		if err != nil {
			return nil, err
		}
		ctx = context.WithValue(ctx, consts.AuthenticatedUserContextKey, &currentUser)
		return handler(ctx, req)
	}
}

// authorizationValidationMap stores path to skip authentication, by default its enabled for all requests
var authenticationSkipList = map[string]bool{
	"/raystack.shield.v1beta1.ShieldService/ListUsers":          true,
	"/raystack.shield.v1beta1.ShieldService/ListOrganizations":  true,
	"/raystack.shield.v1beta1.ShieldService/ListPermissions":    true,
	"/raystack.shield.v1beta1.ShieldService/GetPermission":      true,
	"/raystack.shield.v1beta1.ShieldService/ListNamespaces":     true,
	"/raystack.shield.v1beta1.ShieldService/GetNamespace":       true,
	"/raystack.shield.v1beta1.ShieldService/ListAuthStrategies": true,
	"/raystack.shield.v1beta1.ShieldService/Authenticate":       true,
	"/raystack.shield.v1beta1.ShieldService/AuthCallback":       true,
	"/raystack.shield.v1beta1.ShieldService/AuthLogout":         true,
	"/raystack.shield.v1beta1.ShieldService/ListMetaSchemas":    true,
	"/raystack.shield.v1beta1.ShieldService/GetMetaSchema":      true,
}
