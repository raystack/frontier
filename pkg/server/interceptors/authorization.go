package interceptors

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/relation"

	"github.com/raystack/frontier/core/preference"

	"github.com/raystack/frontier/core/group"

	"github.com/raystack/frontier/pkg/server/health"

	"github.com/raystack/frontier/internal/api/v1beta1"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotAvailable      = fmt.Errorf("function not available at the moment")
	ErrDeniedInvalidArgs = status.Error(codes.PermissionDenied, "invalid arguments")
)

// TODO(kushsharma): some of the authz checks do double fetch of resources which
// can be optimized

// UnaryAuthorizationCheck returns a unary server interceptor that checks for authorization
func UnaryAuthorizationCheck() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if _, ok := info.Server.(*health.Handler); ok {
			// pass through health handler
			return handler(ctx, req)
		}

		// check if authorization needs to be skipped
		if authorizationSkipEndpoints[info.FullMethod] {
			return handler(ctx, req)
		}

		serverHandler, ok := info.Server.(*v1beta1.Handler)
		if !ok {
			return nil, errors.New("miss-configured server handler")
		}

		// apply authorization rules
		azFunc, azVerifier := authorizationValidationMap[info.FullMethod]
		if !azVerifier {
			// deny access if not configured by default
			return nil, status.Error(codes.Unauthenticated, "unauthorized access")
		}
		if err = azFunc(ctx, serverHandler, req); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
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

	"/raystack.frontier.v1beta1.FrontierService/GetServiceUserKey": true,

	"/raystack.frontier.v1beta1.FrontierService/GetPlan":      true,
	"/raystack.frontier.v1beta1.FrontierService/ListPlans":    true,
	"/raystack.frontier.v1beta1.FrontierService/GetProduct":   true,
	"/raystack.frontier.v1beta1.FrontierService/ListProducts": true,
	"/raystack.frontier.v1beta1.FrontierService/ListFeatures": true,
	"/raystack.frontier.v1beta1.FrontierService/GetFeature":   true,

	// TODO(kushsharma): for now we are allowing all requests to billing
	// entitlement checks. Ideally we should only allow requests for
	// features that are enabled for the user. One flaw with this is anyone
	// can potentially check if a feature is enabled for an org by making a
	// request to this endpoint.
	"/raystack.frontier.v1beta1.FrontierService/CheckFeatureEntitlement": true,
}

// authorizationValidationMap stores path to validation function
var authorizationValidationMap = map[string]func(ctx context.Context, handler *v1beta1.Handler, req any) error{
	// user
	"/raystack.frontier.v1beta1.FrontierService/ListUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		prefs, err := handler.ListPlatformPreferences(ctx)
		if err != nil {
			return status.Error(codes.Unavailable, err.Error())
		}
		if prefs[preference.PlatformDisableUsersListing] == "true" {
			return status.Error(codes.Unavailable, ErrNotAvailable.Error())
		}
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/GetUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/ListUserGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationsByUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectsByUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/ListUserInvitations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},

	// serviceuser
	"/raystack.frontier.v1beta1.FrontierService/ListServiceUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListServiceUsersRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateServiceUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateServiceUserRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.ServiceUserManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetServiceUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetServiceUserRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteServiceUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteServiceUserRequest)
		svuser, err := handler.GetServiceUser(ctx, &frontierv1beta1.GetServiceUserRequest{
			Id: pbreq.GetId(),
		})
		if err != nil {
			return err
		}
		if pbreq.GetOrgId() != svuser.GetServiceuser().GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.ServiceUserManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListServiceUserKeys": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListServiceUserKeysRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateServiceUserKey": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateServiceUserKeyRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteServiceUserKey": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteServiceUserKeyRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateServiceUserSecret": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateServiceUserSecretRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListServiceUserSecrets": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListServiceUserSecretsRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.GetId()}, schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteServiceUserSecret": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteServiceUserSecretRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ServiceUserPrincipal, ID: pbreq.GetId()}, schema.ManagePermission)
	},

	// org
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		// check if true or not
		prefs, err := handler.ListPlatformPreferences(ctx)
		if err != nil {
			return status.Error(codes.Unavailable, err.Error())
		}
		if prefs[preference.PlatformDisableOrgsListing] == "true" {
			return status.Error(codes.Unavailable, ErrNotAvailable.Error())
		}
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetOrganizationRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateOrganizationRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationUsersRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationServiceUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationServiceUsersRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationProjects": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationProjectsRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.ProjectListPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/AddOrganizationUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.AddOrganizationUsersRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/RemoveOrganizationUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.RemoveOrganizationUserRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationInvitations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationInvitationsRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.InvitationListPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.InvitationCreatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.InvitationNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/AcceptOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.AcceptOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.InvitationNamespace, ID: pbreq.GetId()}, schema.AcceptPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.InvitationNamespace, ID: pbreq.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationDomain": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateOrganizationDomainRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteOrganizationDomain": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteOrganizationDomainRequest)
		domain, err := handler.GetOrganizationDomain(ctx, &frontierv1beta1.GetOrganizationDomainRequest{
			OrgId: pbreq.GetOrgId(),
			Id:    pbreq.GetId(),
		})
		if err != nil {
			return err
		}
		if domain.GetDomain().GetOrgId() != pbreq.GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationDomains": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationDomainsRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationsByDomain": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationDomain": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetOrganizationDomainRequest)
		domain, err := handler.GetOrganizationDomain(ctx, &frontierv1beta1.GetOrganizationDomainRequest{
			OrgId: pbreq.GetOrgId(),
			Id:    pbreq.GetId(),
		})
		if err != nil {
			return err
		}
		if domain.GetDomain().GetOrgId() != pbreq.GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/VerifyOrganizationDomain": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.VerifyOrganizationDomainRequest)
		domain, err := handler.GetOrganizationDomain(ctx, &frontierv1beta1.GetOrganizationDomainRequest{
			OrgId: pbreq.GetOrgId(),
			Id:    pbreq.GetId(),
		})
		if err != nil {
			return err
		}
		if domain.GetDomain().GetOrgId() != pbreq.GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		prefs, err := handler.ListPlatformPreferences(ctx)
		if err != nil {
			return status.Error(codes.Unavailable, err.Error())
		}
		if prefs[preference.PlatformDisableOrgsOnCreate] == "true" {
			return handler.IsSuperUser(ctx)
		}
		pbreq := req.(*frontierv1beta1.EnableOrganizationRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DisableOrganizationRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteOrganizationRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.DeletePermission)
	},

	// group
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationGroupsRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GroupListPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateGroupRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GroupCreatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetGroupRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateGroupRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListGroupUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListGroupUsersRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/AddGroupUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.AddGroupUsersRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/RemoveGroupUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.RemoveGroupUserRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.EnableGroupRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DisableGroupRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteGroupRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.GetId()}, schema.DeletePermission)
	},

	// project
	"/raystack.frontier.v1beta1.FrontierService/CreateProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateProjectRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetBody().GetOrgId()}, schema.ProjectCreatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetProjectRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateProjectRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectAdmins": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectAdminsRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectUsersRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectServiceUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectServiceUsersRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectGroupsRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.EnableProjectRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DisableProjectRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteProjectRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetId()}, schema.DeletePermission)
	},

	// roles
	"/raystack.frontier.v1beta1.FrontierService/ListRoles": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationRoles": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationRolesRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.RoleManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetOrganizationRoleRequest)
		role, err := handler.GetOrganizationRole(ctx, &frontierv1beta1.GetOrganizationRoleRequest{
			OrgId: pbreq.GetOrgId(),
			Id:    pbreq.GetId(),
		})
		if err != nil {
			return err
		}
		if role.GetRole().GetOrgId() != pbreq.GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateOrganizationRoleRequest)
		role, err := handler.GetOrganizationRole(ctx, &frontierv1beta1.GetOrganizationRoleRequest{
			OrgId: pbreq.GetOrgId(),
			Id:    pbreq.GetId(),
		})
		if err != nil {
			return err
		}
		if role.GetRole().GetOrgId() != pbreq.GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.RoleManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteOrganizationRoleRequest)
		role, err := handler.GetOrganizationRole(ctx, &frontierv1beta1.GetOrganizationRoleRequest{
			OrgId: pbreq.GetOrgId(),
			Id:    pbreq.GetId(),
		})
		if err != nil {
			return err
		}
		if role.GetRole().GetOrgId() != pbreq.GetOrgId() {
			return ErrDeniedInvalidArgs
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.RoleManagePermission)
	},

	// policies
	"/raystack.frontier.v1beta1.FrontierService/CreatePolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreatePolicyRequest)
		ns, id, err := schema.SplitNamespaceAndResourceID(pbreq.GetBody().GetResource())
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
	"/raystack.frontier.v1beta1.FrontierService/ListPolicies": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListPoliciesRequest)
		if pbreq.GetOrgId() != "" {
			return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.PolicyManagePermission)
		}
		if pbreq.GetProjectId() != "" {
			return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetProjectId()}, schema.PolicyManagePermission)
		}
		if pbreq.GetGroupId() != "" {
			return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupNamespace, ID: pbreq.GetGroupId()}, group.AdminPermission)
		}
		if pbreq.GetUserId() != "" {
			principal, err := handler.GetLoggedInPrincipal(ctx)
			if err != nil {
				return err
			}
			if pbreq.GetUserId() == principal.ID {
				// can self introspect
				return nil
			}
		}

		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetPolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdatePolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/DeletePolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeletePolicyRequest)
		policyResp, err := handler.GetPolicy(ctx, &frontierv1beta1.GetPolicyRequest{Id: pbreq.GetId()})
		if err != nil {
			return err
		}
		ns, id, err := schema.SplitNamespaceAndResourceID(policyResp.GetPolicy().GetResource())
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
	"/raystack.frontier.v1beta1.FrontierService/CreateRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},

	// resources
	"/raystack.frontier.v1beta1.FrontierService/ListProjectResources": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectResourcesRequest)
		isAuthed := handler.IsAuthorized(ctx, relation.Object{
			Namespace: schema.ProjectNamespace, ID: pbreq.GetProjectId(),
		}, schema.ResourceListPermission)
		if isAuthed != nil {
			return isAuthed
		}

		return handler.CheckPlanEntitlement(ctx, relation.Object{
			Namespace: schema.ProjectNamespace,
			ID:        pbreq.GetProjectId(),
		})
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateProjectResourceRequest)
		isAuthed := handler.IsAuthorized(ctx, relation.Object{
			Namespace: schema.ProjectNamespace, ID: pbreq.GetProjectId(),
		}, schema.GetPermission)
		if isAuthed != nil {
			return isAuthed
		}

		return handler.CheckPlanEntitlement(ctx, relation.Object{
			Namespace: schema.ProjectNamespace,
			ID:        pbreq.GetProjectId(),
		})
	},
	"/raystack.frontier.v1beta1.FrontierService/GetProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetProjectResourceRequest)
		resp, err := handler.GetProjectResource(ctx, &frontierv1beta1.GetProjectResourceRequest{Id: pbreq.GetId()})
		if err != nil {
			return err
		}
		isAuthed := handler.IsAuthorized(ctx, relation.Object{
			Namespace: resp.GetResource().GetNamespace(), ID: resp.GetResource().GetId()},
			schema.GetPermission)
		if isAuthed != nil {
			return isAuthed
		}

		return handler.CheckPlanEntitlement(ctx, relation.Object{
			Namespace: schema.ProjectNamespace,
			ID:        pbreq.GetProjectId(),
		})
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateProjectResourceRequest)
		isAuthed := handler.IsAuthorized(ctx, relation.Object{
			Namespace: pbreq.GetBody().GetNamespace(), ID: pbreq.GetId()}, schema.UpdatePermission,
		)
		if isAuthed != nil {
			return isAuthed
		}

		return handler.CheckPlanEntitlement(ctx, relation.Object{
			Namespace: schema.ProjectNamespace,
			ID:        pbreq.GetProjectId(),
		})
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteProjectResourceRequest)
		resp, err := handler.GetProjectResource(ctx, &frontierv1beta1.GetProjectResourceRequest{Id: pbreq.GetId()})
		if err != nil {
			return err
		}
		isAuthed := handler.IsAuthorized(ctx, relation.Object{
			Namespace: resp.GetResource().GetNamespace(), ID: resp.GetResource().GetId(),
		}, schema.DeletePermission)
		if isAuthed != nil {
			return isAuthed
		}

		return handler.CheckPlanEntitlement(ctx, relation.Object{
			Namespace: schema.ProjectNamespace,
			ID:        pbreq.GetProjectId(),
		})
	},

	// audit logs
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationAuditLogs": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationAuditLogsRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationAuditLogs": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateOrganizationAuditLogsRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationAuditLog": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetOrganizationAuditLogRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.UpdatePermission)
	},

	// preferences
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateOrganizationPreferencesRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationPreferencesRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateProjectPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateProjectPreferencesRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectPreferencesRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.ProjectNamespace, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateGroupPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateGroupPreferencesRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupPrincipal, ID: pbreq.GetId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListGroupPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListGroupPreferencesRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.GroupPrincipal, ID: pbreq.GetId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateUserPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListUserPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},

	// metaschemas
	"/raystack.frontier.v1beta1.FrontierService/UpdateMetaSchema": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateMetaSchema": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},

	// billing customer
	"/raystack.frontier.v1beta1.FrontierService/CreateBillingAccount": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateBillingAccountRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListBillingAccounts": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListBillingAccountsRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetBillingAccount": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetBillingAccountRequest)
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.GetOrgId(), pbreq.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetBillingBalance": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetBillingBalanceRequest)
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.GetOrgId(), pbreq.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateBillingAccount": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateBillingAccountRequest)
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.GetOrgId(), pbreq.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteBillingAccount": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteBillingAccountRequest)
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.GetOrgId(), pbreq.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/HasTrialed": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.HasTrialedRequest)
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.GetOrgId(), pbreq.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},

	// subscriptions
	"/raystack.frontier.v1beta1.FrontierService/GetSubscription": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetSubscriptionRequest)
		if err := ensureSubscriptionBelongToOrg(ctx, handler, pbreq.GetOrgId(), pbreq.GetBillingId(), pbreq.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListSubscriptions": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListSubscriptionsRequest)
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbreq.GetOrgId(), pbreq.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateSubscription": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateSubscriptionRequest)
		if err := ensureSubscriptionBelongToOrg(ctx, handler, pbreq.GetOrgId(), pbreq.GetBillingId(), pbreq.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CancelSubscription": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CancelSubscriptionRequest)
		if err := ensureSubscriptionBelongToOrg(ctx, handler, pbreq.GetOrgId(), pbreq.GetBillingId(), pbreq.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ChangeSubscription": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ChangeSubscriptionRequest)
		if err := ensureSubscriptionBelongToOrg(ctx, handler, pbreq.GetOrgId(), pbreq.GetBillingId(), pbreq.GetId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateCheckout": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateCheckoutRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListCheckouts": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListCheckoutsRequest)
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbreq.GetOrgId()}, schema.DeletePermission)
	},

	// plans
	"/raystack.frontier.v1beta1.FrontierService/CreatePlan": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdatePlan": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},

	// products
	"/raystack.frontier.v1beta1.FrontierService/CreateProduct": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateProduct": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},

	// features
	"/raystack.frontier.v1beta1.FrontierService/CreateFeature": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateFeature": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},

	// usage
	"/raystack.frontier.v1beta1.FrontierService/CreateBillingUsage": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbReq := req.(*frontierv1beta1.CreateBillingUsageRequest)
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbReq.GetOrgId(), pbReq.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbReq.GetOrgId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListBillingTransactions": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbReq := req.(*frontierv1beta1.ListBillingTransactionsRequest)
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbReq.GetOrgId(), pbReq.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbReq.GetOrgId()}, schema.UpdatePermission)
	},

	// invoice
	"/raystack.frontier.v1beta1.FrontierService/ListInvoices": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbReq := req.(*frontierv1beta1.ListInvoicesRequest)
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbReq.GetOrgId(), pbReq.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbReq.GetOrgId()}, schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetUpcomingInvoice": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbReq := req.(*frontierv1beta1.GetUpcomingInvoiceRequest)
		if err := ensureBillingAccountBelongToOrg(ctx, handler, pbReq.GetOrgId(), pbReq.GetBillingId()); err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.OrganizationNamespace, ID: pbReq.GetOrgId()}, schema.UpdatePermission)
	},

	// admin APIs
	"/raystack.frontier.v1beta1.AdminService/ListAllUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListAllOrganizations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListProjects": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListRelations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListResources": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/CreateRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/UpdateRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/DeleteRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/CreatePermission": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/UpdatePermission": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/DeletePermission": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.AdminService/CreatePreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/AddPlatformUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListPlatformUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/CheckFederatedResourcePermission": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsAuthorized(ctx, relation.Object{Namespace: schema.PlatformNamespace, ID: schema.PlatformID}, schema.PlatformCheckPermission)
	},
	"/raystack.frontier.v1beta1.AdminService/DelegatedCheckout": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListAllInvoices": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/ListAllBillingAccounts": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.AdminService/RevertBillingUsage": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
}

func ensureSubscriptionBelongToOrg(ctx context.Context, handler *v1beta1.Handler, orgID, billingID, subID string) error {
	sub, err := handler.GetSubscription(ctx, &frontierv1beta1.GetSubscriptionRequest{
		OrgId:     orgID,
		BillingId: billingID,
		Id:        subID,
	})
	if err != nil {
		return err
	}
	if sub.GetSubscription().GetCustomerId() != billingID {
		return ErrDeniedInvalidArgs
	}
	acc, err := handler.GetBillingAccount(ctx, &frontierv1beta1.GetBillingAccountRequest{
		OrgId: orgID,
		Id:    billingID,
	})
	if err != nil {
		return err
	}
	if acc.GetBillingAccount().GetOrgId() != orgID {
		return ErrDeniedInvalidArgs
	}
	return nil
}

func ensureBillingAccountBelongToOrg(ctx context.Context, handler *v1beta1.Handler, orgID, billingID string) error {
	acc, err := handler.GetBillingAccount(ctx, &frontierv1beta1.GetBillingAccountRequest{
		OrgId: orgID,
		Id:    billingID,
	})
	if err != nil {
		return err
	}
	if acc.GetBillingAccount().GetOrgId() != orgID {
		return ErrDeniedInvalidArgs
	}
	return nil
}
