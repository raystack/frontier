package interceptors

import (
	"context"
	"errors"
	"fmt"

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
	ErrNotAvailable = fmt.Errorf("function not available at the moment")
)

func UnaryAuthorizationCheck(identityHeader string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if _, ok := info.Server.(*health.Handler); ok {
			// pass through health handler
			return handler(ctx, req)
		}
		if len(identityHeader) != 0 {
			// if configured, skip
			return handler(ctx, req)
		}

		// TODO(kushsharma): refactor to check request call with proto options instead of static map
		// use a proto field option to map user permission with the required permission to access this path
		// something like
		// option (frontier.v1beta1.auth_option) = {
		//      permission: "app_project_update";
		// };

		// check if authorization needs to be skipped
		if authorizationSkipList[info.FullMethod] {
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

// authorizationSkipList stores path to skip authorization, by default its enabled for all requests
var authorizationSkipList = map[string]bool{
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

	"/raystack.frontier.v1beta1.FrontierService/GetServiceUserKey": true,

	"/raystack.frontier.v1beta1.FrontierService/CreateOrganization": true,
}

// authorizationValidationMap stores path to validation function
var authorizationValidationMap = map[string]func(ctx context.Context, handler *v1beta1.Handler, req any) error{
	// user
	"/raystack.frontier.v1beta1.FrontierService/ListUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if handler.DisableUsersListing {
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
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateServiceUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateServiceUserRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.ServiceUserManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetServiceUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetServiceUserRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteServiceUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteServiceUserRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.ServiceUserManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListServiceUserKeys": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListServiceUserKeysRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateServiceUserKey": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateServiceUserKeyRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteServiceUserKey": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteServiceUserKeyRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateServiceUserSecret": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateServiceUserSecretRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListServiceUserSecrets": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListServiceUserSecretsRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteServiceUserSecret": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteServiceUserSecretRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},

	// org
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if handler.DisableOrgsListing {
			return status.Error(codes.Unavailable, ErrNotAvailable.Error())
		}
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationUsersRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationServiceUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationServiceUsersRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationProjects": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationProjectsRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.ProjectListPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/AddOrganizationUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.AddOrganizationUsersRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/RemoveOrganizationUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.RemoveOrganizationUserRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationInvitations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationInvitationsRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.InvitationListPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.InvitationCreatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, schema.InvitationNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/AcceptOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.AcceptOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, schema.InvitationNamespace, pbreq.GetId(), schema.AcceptPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, schema.InvitationNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/AddOrganizationDomain": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateOrganizationDomainRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/RemoveOrganizationDomain": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteOrganizationDomainRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationDomains": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationDomainsRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationsByDomain": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationDomain": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetOrganizationDomainRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/VerifyOrgDomain": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.VerifyOrganizationDomainRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if handler.DisableOrgOnCreate {
			return handler.IsSuperUser(ctx)
		}
		pbreq := req.(*frontierv1beta1.EnableOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DisableOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.DeletePermission)
	},

	// group
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationGroupsRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GroupListPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateGroupRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GroupCreatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListGroupUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListGroupUsersRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/AddGroupUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.AddGroupUsersRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/RemoveGroupUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.RemoveGroupUserRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.EnableGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DisableGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.DeletePermission)
	},

	// project
	"/raystack.frontier.v1beta1.FrontierService/CreateProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateProjectRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetBody().GetOrgId(), schema.ProjectCreatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectAdmins": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectAdminsRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectUsersRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectServiceUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectServiceUsersRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectGroupsRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/EnableProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.EnableProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DisableProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DisableProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.DeletePermission)
	},

	// roles
	"/raystack.frontier.v1beta1.FrontierService/ListRoles": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationRoles": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationRolesRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.RoleManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.RoleManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.RoleManagePermission)
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
			return handler.IsAuthorized(ctx, ns, id, schema.PolicyManagePermission)
		case schema.GroupNamespace:
			return handler.IsAuthorized(ctx, ns, id, group.AdminPermission)
		}
		return handler.IsAuthorized(ctx, ns, id, schema.DeletePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetPolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdatePolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/DeletePolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeletePolicyRequest)
		policyResp, err := handler.GetPolicy(ctx, &frontierv1beta1.GetPolicyRequest{Id: pbreq.Id})
		if err != nil {
			return err
		}
		ns, id, err := schema.SplitNamespaceAndResourceID(policyResp.GetPolicy().GetResource())
		if err != nil {
			return err
		}

		switch ns {
		case schema.OrganizationNamespace, schema.ProjectNamespace:
			return handler.IsAuthorized(ctx, ns, id, schema.PolicyManagePermission)
		case schema.GroupNamespace:
			return handler.IsAuthorized(ctx, ns, id, group.AdminPermission)
		}
		return handler.IsAuthorized(ctx, ns, id, schema.DeletePermission)
	},

	// relations
	"/raystack.frontier.v1beta1.FrontierService/CreateRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateRelationRequest)
		objNS, objID, err := schema.SplitNamespaceAndResourceID(pbreq.GetBody().GetObject())
		if err != nil {
			return err
		}
		subNS, subID, err := schema.SplitNamespaceAndResourceID(pbreq.GetBody().GetSubject())
		if err != nil {
			return err
		}

		if objNS == schema.OrganizationNamespace {
			return handler.IsAuthorized(ctx, schema.OrganizationNamespace, objID, schema.UpdatePermission)
		}
		if subNS == schema.OrganizationNamespace {
			return handler.IsAuthorized(ctx, schema.OrganizationNamespace, subID, schema.UpdatePermission)
		}
		if objNS == schema.ProjectNamespace {
			return handler.IsAuthorized(ctx, schema.ProjectNamespace, objID, schema.UpdatePermission)
		}
		if subNS == schema.ProjectNamespace {
			return handler.IsAuthorized(ctx, schema.ProjectNamespace, subID, schema.UpdatePermission)
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.frontier.v1beta1.FrontierService/GetRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateRelationRequest)
		objNS, objID, err := schema.SplitNamespaceAndResourceID(pbreq.GetBody().GetObject())
		if err != nil {
			return err
		}
		subNS, subID, err := schema.SplitNamespaceAndResourceID(pbreq.GetBody().GetSubject())
		if err != nil {
			return err
		}

		if objNS == schema.OrganizationNamespace {
			return handler.IsAuthorized(ctx, schema.OrganizationNamespace, objID, schema.UpdatePermission)
		}
		if subNS == schema.OrganizationNamespace {
			return handler.IsAuthorized(ctx, schema.OrganizationNamespace, subID, schema.UpdatePermission)
		}
		if objNS == schema.ProjectNamespace {
			return handler.IsAuthorized(ctx, schema.ProjectNamespace, objID, schema.UpdatePermission)
		}
		if subNS == schema.ProjectNamespace {
			return handler.IsAuthorized(ctx, schema.ProjectNamespace, subID, schema.UpdatePermission)
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},

	// resources
	"/raystack.frontier.v1beta1.FrontierService/ListProjectResources": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectResourcesRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetProjectId(), schema.ResourceListPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateProjectResourceRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetProjectId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetProjectResourceRequest)
		resp, err := handler.GetProjectResource(ctx, &frontierv1beta1.GetProjectResourceRequest{Id: pbreq.GetId()})
		if err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, resp.GetResource().GetNamespace(), resp.GetResource().GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/UpdateProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.UpdateProjectResourceRequest)
		return handler.IsAuthorized(ctx, pbreq.GetBody().GetNamespace(), pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/DeleteProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.DeleteProjectResourceRequest)
		resp, err := handler.GetProjectResource(ctx, &frontierv1beta1.GetProjectResourceRequest{Id: pbreq.GetId()})
		if err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, resp.GetResource().GetNamespace(), resp.GetResource().GetId(), schema.DeletePermission)
	},

	// audit logs
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationAuditLogs": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationAuditLogsRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.ManagePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationAuditLogs": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateOrganizationAuditLogsRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/GetOrganizationAuditLog": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.GetOrganizationAuditLogRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.ManagePermission)
	},

	// preferences
	"/raystack.frontier.v1beta1.FrontierService/CreateOrganizationPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateOrganizationPreferencesRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListOrganizationPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListOrganizationPreferencesRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateProjectPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateProjectPreferencesRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListProjectPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListProjectPreferencesRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/CreateGroupPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.CreateGroupPreferencesRequest)
		return handler.IsAuthorized(ctx, schema.GroupPrincipal, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.frontier.v1beta1.FrontierService/ListGroupPreferences": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*frontierv1beta1.ListGroupPreferencesRequest)
		return handler.IsAuthorized(ctx, schema.GroupPrincipal, pbreq.GetId(), schema.GetPermission)
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
	"/raystack.frontier.v1beta1.AdminService/ListPolicies": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
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
}
