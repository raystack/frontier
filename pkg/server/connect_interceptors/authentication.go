package connectinterceptors

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/internal/api/v1beta1connect"
)

func UnaryAuthenticationCheck(h *v1beta1connect.ConnectHandler) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			fmt.Println("connect interceptor called")
			// if _, ok := info.Server.(*health.Handler); ok {
			// 	// pass through health handler
			// 	return handler(ctx, req)
			// }
			if authenticationSkipList[req.Spec().Procedure] {
				return next(ctx, req)
			}

			principal, err := h.GetLoggedInPrincipal(ctx)
			if err != nil {
				return nil, err
			}
			ctx = authenticate.SetContextWithPrincipal(ctx, &principal)
			ctx = audit.SetContextWithActor(ctx, audit.Actor{
				ID:   principal.ID,
				Type: principal.Type,
			})
			return next(ctx, req)
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
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
	"/raystack.frontier.v1beta1.FrontierService/AuthLogout":             true,
	"/raystack.frontier.v1beta1.FrontierService/ListMetaSchemas":        true,
	"/raystack.frontier.v1beta1.FrontierService/GetMetaSchema":          true,
	"/raystack.frontier.v1beta1.FrontierService/BillingWebhookCallback": true,
	"/raystack.frontier.v1beta1.FrontierService/CreateProspectPublic":   true,
}
