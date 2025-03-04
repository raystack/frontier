package interceptors

import (
	"context"
	"errors"

	"github.com/raystack/frontier/core/audit"

	"github.com/raystack/frontier/pkg/server/health"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/internal/api/v1beta1"
	"google.golang.org/grpc"
)

func UnaryAuthenticationCheck() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if _, ok := info.Server.(*health.Handler); ok {
			// pass through health handler
			return handler(ctx, req)
		}
		if authenticationSkipList[info.FullMethod] {
			return handler(ctx, req)
		}

		// authentication can be done via various authenticate.ClientAssertion
		serverHandler, ok := info.Server.(*v1beta1.Handler)
		if !ok {
			return nil, errors.New("miss-configured server handler")
		}

		principal, err := serverHandler.GetLoggedInPrincipal(ctx)
		if err != nil {
			return nil, err
		}
		ctx = authenticate.SetContextWithPrincipal(ctx, &principal)
		ctx = audit.SetContextWithActor(ctx, audit.Actor{
			ID:   principal.ID,
			Type: principal.Type,
		})
		return handler(ctx, req)
	}
}

// authenticationSkipList stores path to skip authentication, by default its enabled for all requests
var authenticationSkipList = map[string]bool{
	"/raystack.frontier.v1beta1.FrontierService/GetJWKs":                true,
	"/raystack.frontier.v1beta1.FrontierService/GetServiceUserKey":      true,
	"/raystack.frontier.v1beta1.FrontierService/ListUsers":              true,
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizations":      true,
	"/raystack.frontier.v1beta1.FrontierService/ListPermissions":        true,
	"/raystack.frontier.v1beta1.FrontierService/GetPermission":          true,
	"/raystack.frontier.v1beta1.FrontierService/ListNamespaces":         true,
	"/raystack.frontier.v1beta1.FrontierService/GetNamespace":           true,
	"/raystack.frontier.v1beta1.FrontierService/ListAuthStrategies":     true,
	"/raystack.frontier.v1beta1.FrontierService/Authenticate":           true,
	"/raystack.frontier.v1beta1.FrontierService/AuthCallback":           true,
	"/raystack.frontier.v1beta1.FrontierService/AuthToken":              true,
	"/raystack.frontier.v1beta1.FrontierService/AuthLogout":             true,
	"/raystack.frontier.v1beta1.FrontierService/ListMetaSchemas":        true,
	"/raystack.frontier.v1beta1.FrontierService/GetMetaSchema":          true,
	"/raystack.frontier.v1beta1.FrontierService/BillingWebhookCallback": true,
	"/raystack.frontier.v1beta1.FrontierService/CreateProspect":         true,
}
