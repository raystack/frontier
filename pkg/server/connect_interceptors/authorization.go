package connectinterceptors

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"github.com/raystack/frontier/internal/api/v1beta1connect"

	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/internal/metrics"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

var (
	ErrNotAvailable      = connect.NewError(connect.CodeUnavailable, fmt.Errorf("function not available at the moment"))
	ErrDeniedInvalidArgs = connect.NewError(connect.CodePermissionDenied, errors.New("invalid arguments"))
)

type AuthorizationInterceptor struct {
	h *v1beta1connect.ConnectHandler
}

func (a *AuthorizationInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)
		return conn
	})
}

func (a *AuthorizationInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		// check if authorization needs to be skipped
		if authorizationSkipEndpoints[conn.Spec().Procedure] {
			return next(ctx, conn)
		}

		if metrics.ServiceOprLatency != nil {
			promCollect := metrics.ServiceOprLatency("authenticate", "UnaryAuthorizationCheck")
			defer promCollect()
		}

		// apply authorization rules
		azFunc, azVerifier := authorizationValidationMap[conn.Spec().Procedure]
		if !azVerifier {
			// deny access if not configured by default
			// return nil, connect.NewError(codes.Unauthenticated, "unauthorized access")
			return connect.NewError(connect.CodePermissionDenied, v1beta1connect.ErrUnauthorized)
		}
		if err := azFunc(ctx, a.h, nil); err != nil {
			return err
		}

		return next(ctx, conn)
	})
}

func (a *AuthorizationInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// check if authorization needs to be skipped
		if authorizationSkipEndpoints[req.Spec().Procedure] {
			return next(ctx, req)
		}

		if metrics.ServiceOprLatency != nil {
			promCollect := metrics.ServiceOprLatency("authenticate", "UnaryAuthorizationCheck")
			defer promCollect()
		}

		// apply authorization rules
		azFunc, azVerifier := authorizationValidationMap[req.Spec().Procedure]
		if !azVerifier {
			// deny access if not configured by default
			// return nil, connect.NewError(codes.Unauthenticated, "unauthorized access")
			return nil, connect.NewError(connect.CodePermissionDenied, v1beta1connect.ErrUnauthorized)
		}
		if err := azFunc(ctx, a.h, req); err != nil {
			return nil, err
		}

		return next(ctx, req)
	})
}

func NewAuthorizationInterceptor(h *v1beta1connect.ConnectHandler) *AuthorizationInterceptor {
	return &AuthorizationInterceptor{h}
}

// authorizationSkipEndpoints stores path to skip authorization, by default its enabled for all requests
var authorizationSkipEndpoints = map[string]bool{
	"/raystack.frontier.v1beta1.FrontierService/GetJWKs":                 true,
	"/raystack.frontier.v1beta1.FrontierService/ListAuthStrategies":      true,
	"/raystack.frontier.v1beta1.FrontierService/Authenticate":            true,
	"/raystack.frontier.v1beta1.FrontierService/AuthCallback":            true,
	"/raystack.frontier.v1beta1.FrontierService/AuthToken":               true,
	"/raystack.frontier.v1beta1.FrontierService/AuthLogout":              true,
	"/raystack.frontier.v1beta1.FrontierService/CheckResourcePermission": true,
	"/raystack.frontier.v1beta1.FrontierService/BatchCheckPermission":    true,

	"/raystack.frontier.v1beta1.FrontierService/ListPermissions": true,
	"/raystack.frontier.v1beta1.FrontierService/GetPermission":   true,

	"/raystack.frontier.v1beta1.FrontierService/ListNamespaces": true,
	"/raystack.frontier.v1beta1.FrontierService/GetNamespace":   true,

	"/raystack.frontier.v1beta1.FrontierService/ListMetaSchemas":     true,
	"/raystack.frontier.v1beta1.FrontierService/GetMetaSchema":       true,
	"/raystack.frontier.v1beta1.FrontierService/DescribePreferences": true,

	"/raystack.frontier.v1beta1.FrontierService/ListCurrentUserGroups":          true,
	"/raystack.frontier.v1beta1.FrontierService/GetCurrentUser":                 true,
	"/raystack.frontier.v1beta1.FrontierService/UpdateCurrentUser":              true,
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationsByCurrentUser": true,
	"/raystack.frontier.v1beta1.FrontierService/ListProjectsByCurrentUser":      true,
	"/raystack.frontier.v1beta1.FrontierService/CreateCurrentUserPreferences":   true,
	"/raystack.frontier.v1beta1.FrontierService/ListCurrentUserPreferences":     true,
	"/raystack.frontier.v1beta1.FrontierService/ListCurrentUserInvitations":     true,

	"/raystack.frontier.v1beta1.FrontierService/JoinOrganization":   true,
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganization": true,

	"/raystack.frontier.v1beta1.FrontierService/GetServiceUserJWK": true,

	"/raystack.frontier.v1beta1.FrontierService/GetPlan":      true,
	"/raystack.frontier.v1beta1.FrontierService/ListPlans":    true,
	"/raystack.frontier.v1beta1.FrontierService/GetProduct":   true,
	"/raystack.frontier.v1beta1.FrontierService/ListProducts": true,
	"/raystack.frontier.v1beta1.FrontierService/ListFeatures": true,
	"/raystack.frontier.v1beta1.FrontierService/GetFeature":   true,

	"/raystack.frontier.v1beta1.FrontierService/BillingWebhookCallback": true,

	// TODO(kushsharma): for now we are allowing all requests to billing
	// entitlement checks. Ideally we should only allow requests for
	// features that are enabled for the user. One flaw with this is anyone
	// can potentially check if a feature is enabled for an org by making a
	// request to this endpoint.
	"/raystack.frontier.v1beta1.FrontierService/CheckFeatureEntitlement": true,
	"/raystack.frontier.v1beta1.FrontierService/CreateProspectPublic":    true,
}

// authorizationValidationMap stores path to validation function
var authorizationValidationMap = map[string]func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error{
	// user
	"/raystack.frontier.v1beta1.FrontierService/ListUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		prefs, err := handler.ListPlatformPreferences(ctx)
		if err != nil {
			return connect.NewError(connect.CodeUnavailable, err)
		}
		if prefs[preference.PlatformDisableUsersListing] == "true" {
			return ErrNotAvailable
		}
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return ErrNotAvailable
	},
	"/raystack.frontier.v1beta1.FrontierService/GetUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return ErrNotAvailable
	},
	"/raystack.frontier.v1beta1.FrontierService/ListUserGroups": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return ErrNotAvailable
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return ErrNotAvailable
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return ErrNotAvailable
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return ErrNotAvailable
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return ErrNotAvailable
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationsByUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return ErrNotAvailable
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectsByUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return ErrNotAvailable
	},
	"/raystack.frontier.v1beta1.FrontierService/ListUserInvitations": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return ErrNotAvailable
	},

	// serviceuser
	"/raystack.frontier.v1beta1.FrontierService/ListServiceUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListServiceUsersRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateServiceUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateServiceUserRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.ServiceUserManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetServiceUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetServiceUserRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.Msg.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteServiceUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteServiceUserRequest])
		svuser, err := handler.GetServiceUser(ctx, connect.NewRequest(&frontierv1beta1.GetServiceUserRequest{
			Id: pbreq.Msg.GetId(),
		}))
		if err != nil {
			return err
		}
		if pbreq.Msg.GetOrgId() != svuser.Msg.GetServiceuser().GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.ServiceUserManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListServiceUserProjects": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListServiceUserProjectsRequest])
		svuser, err := handler.GetServiceUser(ctx, connect.NewRequest(&frontierv1beta1.GetServiceUserRequest{
			Id: pbreq.Msg.GetId(),
		}))
		if err != nil {
			return err
		}
		if pbreq.Msg.GetOrgId() != svuser.Msg.GetServiceuser().GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.ServiceUserManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListServiceUserJWKs": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListServiceUserJWKsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.Msg.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateServiceUserJWK": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateServiceUserJWKRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.Msg.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteServiceUserJWK": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteServiceUserJWKRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.Msg.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateServiceUserCredential": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateServiceUserCredentialRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.Msg.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListServiceUserCredentials": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListServiceUserCredentialsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.Msg.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteServiceUserCredential": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteServiceUserCredentialRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.Msg.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateServiceUserToken": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateServiceUserTokenRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.Msg.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListServiceUserTokens": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListServiceUserTokensRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.Msg.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteServiceUserToken": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteServiceUserTokenRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.Msg.GetId()}, schema.ManagePermission)
	},

	// org
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizations": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		// check if true or not
		prefs, err := handler.ListPlatformPreferences(ctx)
		if err != nil {
			return connect.NewError(connect.CodeUnavailable, err)
		}
		if prefs[preference.PlatformDisableOrgsListing] == "true" {
			return ErrNotAvailable
		}
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganization": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetOrganizationRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateOrganization": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.UpdateOrganizationRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListOrganizationUsersRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationKyc": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetOrganizationKycRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationAdmins": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListOrganizationAdminsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationServiceUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListOrganizationServiceUsersRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationProjects": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListOrganizationProjectsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetId()}, schema.ProjectListPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/AddOrganizationUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.AddOrganizationUsersRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/RemoveOrganizationUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.RemoveOrganizationUserRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationInvitations": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListOrganizationInvitationsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.InvitationListPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationInvitation": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateOrganizationInvitationRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.InvitationCreatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationInvitation": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetOrganizationInvitationRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.InvitationNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/AcceptOrganizationInvitation": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.AcceptOrganizationInvitationRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.InvitationNamespace, ID: pbreq.Msg.GetId()}, schema.AcceptPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteOrganizationInvitation": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteOrganizationInvitationRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.InvitationNamespace, ID: pbreq.Msg.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationDomain": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateOrganizationDomainRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteOrganizationDomain": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteOrganizationDomainRequest])
		domain, err := handler.GetOrganizationDomain(ctx, connect.NewRequest(&frontierv1beta1.GetOrganizationDomainRequest{
			OrgId: pbreq.Msg.GetOrgId(),
			Id:    pbreq.Msg.GetId(),
		}))
		if err != nil {
			return err
		}
		if domain.Msg.GetDomain().GetOrgId() != pbreq.Msg.GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationDomains": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListOrganizationDomainsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationsByDomain": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationDomain": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetOrganizationDomainRequest])
		domain, err := handler.GetOrganizationDomain(ctx, connect.NewRequest(&frontierv1beta1.GetOrganizationDomainRequest{
			OrgId: pbreq.Msg.GetOrgId(),
			Id:    pbreq.Msg.GetId(),
		}))
		if err != nil {
			return err
		}
		if domain.Msg.GetDomain().GetOrgId() != pbreq.Msg.GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/VerifyOrganizationDomain": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.VerifyOrganizationDomainRequest])
		domain, err := handler.GetOrganizationDomain(ctx, connect.NewRequest(&frontierv1beta1.GetOrganizationDomainRequest{
			OrgId: pbreq.Msg.GetOrgId(),
			Id:    pbreq.Msg.GetId(),
		}))
		if err != nil {
			return err
		}
		if domain.Msg.GetDomain().GetOrgId() != pbreq.Msg.GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableOrganization": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableOrganization": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteOrganization": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteOrganizationRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetId()}, schema.DeletePermission)
	},

	// group
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationGroups": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListOrganizationGroupsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GroupListPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateGroup": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateGroupRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GroupCreatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetGroup": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetGroupRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateGroup": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.UpdateGroupRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.Msg.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListGroupUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListGroupUsersRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/AddGroupUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.AddGroupUsersRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.Msg.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/RemoveGroupUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.RemoveGroupUserRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.Msg.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableGroup": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.EnableGroupRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.Msg.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableGroup": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DisableGroupRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.Msg.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteGroup": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteGroupRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.Msg.GetId()}, schema.DeletePermission)
	},

	// project
	"/raystack.frontier.v1beta1.FrontierService/CreateProject": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateProjectRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetBody().GetOrgId()}, schema.ProjectCreatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetProject": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetProjectRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateProject": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.UpdateProjectRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectAdmins": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListProjectAdminsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListProjectUsersRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectServiceUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListProjectServiceUsersRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectGroups": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListProjectGroupsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableProject": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.EnableProjectRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableProject": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DisableProjectRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteProject": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteProjectRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetId()}, schema.DeletePermission)
	},

	// roles
	"/raystack.frontier.v1beta1.FrontierService/ListRoles": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationRoles": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListOrganizationRolesRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationRole": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateOrganizationRoleRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.RoleManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationRole": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetOrganizationRoleRequest])
		if err := ensureRoleBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateOrganizationRole": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.UpdateOrganizationRoleRequest])
		if err := ensureRoleBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.RoleManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteOrganizationRole": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteOrganizationRoleRequest])
		if err := ensureRoleBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.RoleManagePermission)
	},

	// policies
	"/raystack.frontier.v1beta1.FrontierService/CreatePolicy": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreatePolicyRequest])
		ns, id, err := schema.SplitNamespaceAndResourceID(pbreq.Msg.GetBody().GetResource())
		if err != nil {
			return err
		}

		switch ns {
		case schema.OrganizationNamespace, schema.ProjectNamespace:
			return handler.IsAuthorized(ctx, relation.Object{Namespace: ns, ID: id}, schema.PolicyManagePermission)
		case schema.GroupNamespace:
			return handler.IsAuthorized(ctx, relation.Object{Namespace: ns, ID: id}, group.AdminPermission)
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: ns, ID: id}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreatePolicyForProject": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreatePolicyForProjectRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetProjectId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListPolicies": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListPoliciesRequest])
		if pbreq.Msg.GetOrgId() != "" {
			return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.PolicyManagePermission)
		}
		if pbreq.Msg.GetProjectId() != "" {
			return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetProjectId()}, schema.PolicyManagePermission)
		}
		if pbreq.Msg.GetGroupId() != "" {
			return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.Msg.GetGroupId()}, group.AdminPermission)
		}
		if pbreq.Msg.GetUserId() != "" {
			principal, err := handler.GetLoggedInPrincipal(ctx)
			if err != nil {
				return err
			}
			if pbreq.Msg.GetUserId() == principal.ID {
				// can self introspect
				return nil
			}
		}

		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetPolicy": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdatePolicy": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return ErrNotAvailable
	},
	"/raystack.frontier.v1beta1.FrontierService/DeletePolicy": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeletePolicyRequest])
		policyResp, err := handler.GetPolicy(ctx, connect.NewRequest(&frontierv1beta1.GetPolicyRequest{Id: pbreq.Msg.GetId()}))
		if err != nil {
			return err
		}
		ns, id, err := schema.SplitNamespaceAndResourceID(policyResp.Msg.GetPolicy().GetResource())
		if err != nil {
			return err
		}

		switch ns {
		case schema.OrganizationNamespace, schema.ProjectNamespace:
			return handler.IsAuthorized(ctx, relation.Object{Namespace: ns, ID: id}, schema.PolicyManagePermission)
		case schema.GroupNamespace:
			return handler.IsAuthorized(ctx, relation.Object{Namespace: ns, ID: id}, group.AdminPermission)
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: ns, ID: id}, schema.DeletePermission)
	},

	// relations
	"/raystack.frontier.v1beta1.FrontierService/CreateRelation": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetRelation": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteRelation": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},

	// resources
	"/raystack.frontier.v1beta1.FrontierService/ListProjectResources": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListProjectResourcesRequest])
		isAuthed := handler.IsAuthorized(ctx, relation.Object{
			Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetProjectId(),
		}, schema.ResourceListPermission)
		if isAuthed != nil {
			return isAuthed
		}

		return handler.CheckPlanEntitlement(ctx, relation.Object{
			Namespace: schema.ProjectNamespace,
			ID:        pbreq.Msg.GetProjectId(),
		})
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateProjectResource": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateProjectResourceRequest])
		isAuthed := handler.IsAuthorized(ctx, relation.Object{
			Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetProjectId(),
		}, schema.GetPermission)
		if isAuthed != nil {
			return isAuthed
		}

		return handler.CheckPlanEntitlement(ctx, relation.Object{
			Namespace: schema.ProjectNamespace,
			ID:        pbreq.Msg.GetProjectId(),
		})
	},
	"/raystack.frontier.v1beta1.FrontierService/GetProjectResource": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetProjectResourceRequest])
		resp, err := handler.GetProjectResource(ctx, connect.NewRequest(&frontierv1beta1.GetProjectResourceRequest{Id: pbreq.Msg.GetId()}))
		if err != nil {
			return err
		}
		isAuthed := handler.IsAuthorized(ctx, relation.Object{
			Namespace: resp.Msg.GetResource().GetNamespace(), ID: resp.Msg.GetResource().GetId()},
			schema.GetPermission)
		if isAuthed != nil {
			return isAuthed
		}

		return handler.CheckPlanEntitlement(ctx, relation.Object{
			Namespace: schema.ProjectNamespace,
			ID:        pbreq.Msg.GetProjectId(),
		})
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateProjectResource": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.UpdateProjectResourceRequest])
		isAuthed := handler.IsAuthorized(ctx, relation.Object{
			Namespace: pbreq.Msg.GetBody().GetNamespace(), ID: pbreq.Msg.GetId()}, schema.UpdatePermission,
		)
		if isAuthed != nil {
			return isAuthed
		}

		return handler.CheckPlanEntitlement(ctx, relation.Object{
			Namespace: schema.ProjectNamespace,
			ID:        pbreq.Msg.GetProjectId(),
		})
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteProjectResource": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteProjectResourceRequest])
		resp, err := handler.GetProjectResource(ctx, connect.NewRequest(&frontierv1beta1.GetProjectResourceRequest{Id: pbreq.Msg.GetId()}))
		if err != nil {
			return err
		}
		isAuthed := handler.IsAuthorized(ctx, relation.Object{
			Namespace: resp.Msg.GetResource().GetNamespace(), ID: resp.Msg.GetResource().GetId(),
		}, schema.DeletePermission)
		if isAuthed != nil {
			return isAuthed
		}

		return handler.CheckPlanEntitlement(ctx, relation.Object{
			Namespace: schema.ProjectNamespace,
			ID:        pbreq.Msg.GetProjectId(),
		})
	},

	// audit logs
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationAuditLogs": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListOrganizationAuditLogsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationAuditLogs": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateOrganizationAuditLogsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationAuditLog": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetOrganizationAuditLogRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.UpdatePermission)
	},

	// preferences
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationPreferences": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateOrganizationPreferencesRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationPreferences": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListOrganizationPreferencesRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateProjectPreferences": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateProjectPreferencesRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectPreferences": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListProjectPreferencesRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateGroupPreferences": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateGroupPreferencesRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupPrincipal, ID: pbreq.Msg.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListGroupPreferences": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListGroupPreferencesRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupPrincipal, ID: pbreq.Msg.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateUserPreferences": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListUserPreferences": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},

	// metaschemas
	"/raystack.frontier.v1beta1.FrontierService/UpdateMetaSchema": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateMetaSchema": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},

	// billing customer
	"/raystack.frontier.v1beta1.FrontierService/CreateBillingAccount": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateBillingAccountRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListBillingAccounts": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListBillingAccountsRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetBillingAccount": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetBillingAccountRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetBillingBalance": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetBillingBalanceRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateBillingAccount": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.UpdateBillingAccountRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteBillingAccount": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DeleteBillingAccountRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableBillingAccount": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.EnableBillingAccountRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableBillingAccount": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.DisableBillingAccountRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/HasTrialed": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.HasTrialedRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/RegisterBillingAccount": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.RegisterBillingAccountRequest])
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},

	// subscriptions
	"/raystack.frontier.v1beta1.FrontierService/GetSubscription": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetSubscriptionRequest])
		if err := ensureSubscriptionBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetBillingId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListSubscriptions": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListSubscriptionsRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateSubscription": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.UpdateSubscriptionRequest])
		if err := ensureSubscriptionBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetBillingId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CancelSubscription": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CancelSubscriptionRequest])
		if err := ensureSubscriptionBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetBillingId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ChangeSubscription": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ChangeSubscriptionRequest])
		if err := ensureSubscriptionBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetBillingId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateCheckout": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.CreateCheckoutRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListCheckouts": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.ListCheckoutsRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetCheckout": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbreq := req.(*connect.Request[frontierv1beta1.GetCheckoutRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.Msg.GetOrgId(), pbreq.Msg.GetBillingId()); err != nil {
			return err
		}
		if err := ensureCheckoutBelongToOrg(ctx, handler, pbreq.Msg.GetBillingId(), pbreq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.Msg.GetOrgId()}, schema.DeletePermission)
	},

	// plans
	"/raystack.frontier.v1beta1.FrontierService/CreatePlan": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdatePlan": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},

	// products
	"/raystack.frontier.v1beta1.FrontierService/CreateProduct": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateProduct": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},

	// features
	"/raystack.frontier.v1beta1.FrontierService/CreateFeature": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateFeature": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},

	// usage
	"/raystack.frontier.v1beta1.FrontierService/CreateBillingUsage": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbReq := req.(*connect.Request[frontierv1beta1.CreateBillingUsageRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbReq.Msg.GetOrgId(), pbReq.Msg.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.PlatformNamespace, ID: schema.PlatformID}, schema.PlatformCheckPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListBillingTransactions": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbReq := req.(*connect.Request[frontierv1beta1.ListBillingTransactionsRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbReq.Msg.GetOrgId(), pbReq.Msg.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbReq.Msg.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/TotalDebitedTransactions": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbReq := req.(*connect.Request[frontierv1beta1.TotalDebitedTransactionsRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbReq.Msg.GetOrgId(), pbReq.Msg.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbReq.Msg.GetOrgId()}, schema.GetPermission)
	},

	// invoice
	"/raystack.frontier.v1beta1.FrontierService/ListInvoices": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbReq := req.(*connect.Request[frontierv1beta1.ListInvoicesRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbReq.Msg.GetOrgId(), pbReq.Msg.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbReq.Msg.GetOrgId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetUpcomingInvoice": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbReq := req.(*connect.Request[frontierv1beta1.GetUpcomingInvoiceRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbReq.Msg.GetOrgId(), pbReq.Msg.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbReq.Msg.GetOrgId()}, schema.UpdatePermission)
	},

	// admin APIs
	"/raystack.frontier.v1beta1.AdminService/ListAllUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListGroups": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListAllOrganizations": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SearchOrganizations": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/AdminCreateOrganization": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SearchOrganizationUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SearchOrganizationProjects": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SearchProjectUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SearchOrganizationInvoices": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SearchOrganizationTokens": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SearchOrganizationServiceUserCredentials": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SearchUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SearchUserOrganizations": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SearchUserProjects": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SearchInvoices": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ExportOrganizations": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ExportOrganizationUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ExportOrganizationProjects": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ExportOrganizationTokens": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ExportUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/SetOrganizationKyc": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListOrganizationsKyc": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListProjects": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListRelations": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListResources": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/CreateRole": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/UpdateRole": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/DeleteRole": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/CreatePermission": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/UpdatePermission": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/DeletePermission": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return ErrNotAvailable
	},
	"/raystack.frontier.v1beta1.AdminService/CreatePreferences": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListPreferences": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/AddPlatformUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/RemovePlatformUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListPlatformUsers": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/CheckFederatedResourcePermission": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.PlatformNamespace, ID: schema.PlatformID}, schema.PlatformCheckPermission)
	},
	"/raystack.frontier.v1beta1.AdminService/DelegatedCheckout": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListAllInvoices": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListAllBillingAccounts": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/GenerateInvoices": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/UpdateBillingAccountLimits": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/GetBillingAccountDetails": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbReq := req.(*connect.Request[frontierv1beta1.GetBillingAccountDetailsRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbReq.Msg.GetOrgId(), pbReq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/UpdateBillingAccountDetails": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		pbReq := req.(*connect.Request[frontierv1beta1.UpdateBillingAccountDetailsRequest])
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbReq.Msg.GetOrgId(), pbReq.Msg.GetId()); err != nil {
			return err
		}
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/RevertBillingUsage": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.PlatformNamespace, ID: schema.PlatformID}, schema.PlatformCheckPermission)
	},
	"/raystack.frontier.v1beta1.AdminService/CreateWebhook": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListWebhooks": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/UpdateWebhook": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/DeleteWebhook": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/CreateProspect": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/GetProspect": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListProspects": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/UpdateProspect": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/DeleteProspect": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/GetCurrentAdminUser": func(ctx context.Context, handler *v1beta1connect.ConnectHandler, req connect.AnyRequest) error {
		return handler.IsSuperUser(ctx)
	},
}

func ensureRoleBelongToOrg(ctx context.Context, handler *v1beta1connect.ConnectHandler, orgID, roleID string) error {
	role, err := handler.GetOrganizationRole(ctx, connect.NewRequest(&frontierv1beta1.GetOrganizationRoleRequest{
		OrgId: orgID,
		Id:    roleID,
	}))
	if err != nil {
		return err
	}
	if role.Msg.GetRole().GetOrgId() != orgID {
		return ErrDeniedInvalidArgs
	}
	return nil
}

func ensureBillingAccountBelongToOrg(ctx context.Context, handler *v1beta1connect.ConnectHandler, orgID, billingID string) error {
	acc, err := handler.GetBillingAccount(ctx, connect.NewRequest(&frontierv1beta1.GetBillingAccountRequest{
		OrgId: orgID,
		Id:    billingID,
	}))
	if err != nil {
		return err
	}
	if acc.Msg.GetBillingAccount().GetOrgId() != orgID {
		return ErrDeniedInvalidArgs
	}
	return nil
}

func ensureSubscriptionBelongToOrg(ctx context.Context, handler *v1beta1connect.ConnectHandler, orgID, billingID, subID string) error {
	sub, err := handler.GetSubscription(ctx, connect.NewRequest(&frontierv1beta1.GetSubscriptionRequest{
		OrgId:     orgID,
		BillingId: billingID,
		Id:        subID,
	}))
	if err != nil {
		return err
	}
	if sub.Msg.GetSubscription().GetCustomerId() != billingID {
		return ErrDeniedInvalidArgs
	}
	acc, err := handler.GetBillingAccount(ctx, connect.NewRequest(&frontierv1beta1.GetBillingAccountRequest{
		OrgId: orgID,
		Id:    billingID,
	}))
	if err != nil {
		return err
	}
	if acc.Msg.GetBillingAccount().GetOrgId() != orgID {
		return ErrDeniedInvalidArgs
	}
	return nil
}

func ensureCheckoutBelongToOrg(ctx context.Context, handler *v1beta1connect.ConnectHandler, billingID, checkoutID string) error {
	checkout, err := handler.GetRawCheckout(ctx, checkoutID)
	if err != nil {
		return err
	}
	if checkout.CustomerID != billingID {
		return ErrDeniedInvalidArgs
	}
	return nil
}
