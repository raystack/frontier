package interceptors

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/shield/pkg/server/health"

	"github.com/raystack/shield/internal/api/v1beta1"
	"github.com/raystack/shield/internal/bootstrap/schema"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
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
		// option (shield.v1beta1.auth_option) = {
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
	"/raystack.shield.v1beta1.ShieldService/GetJWKs":            true,
	"/raystack.shield.v1beta1.ShieldService/ListAuthStrategies": true,
	"/raystack.shield.v1beta1.ShieldService/Authenticate":       true,
	"/raystack.shield.v1beta1.ShieldService/AuthCallback":       true,
	"/raystack.shield.v1beta1.ShieldService/AuthToken":          true,
	"/raystack.shield.v1beta1.ShieldService/AuthLogout":         true,

	"/raystack.shield.v1beta1.ShieldService/ListPermissions": true,
	"/raystack.shield.v1beta1.ShieldService/GetPermission":   true,

	"/raystack.shield.v1beta1.ShieldService/ListNamespaces": true,
	"/raystack.shield.v1beta1.ShieldService/GetNamespace":   true,

	"/raystack.shield.v1beta1.ShieldService/ListMetaSchemas": true,
	"/raystack.shield.v1beta1.ShieldService/GetMetaSchema":   true,

	"/raystack.shield.v1beta1.ShieldService/ListCurrentUserGroups":         true,
	"/raystack.shield.v1beta1.ShieldService/GetCurrentUser":                true,
	"/raystack.shield.v1beta1.ShieldService/UpdateCurrentUser":             true,
	"/raystack.shield.v1beta1.ShieldService/GetOrganizationsByCurrentUser": true,
	"/raystack.shield.v1beta1.ShieldService/GetServiceUserKey":             true,

	"/raystack.shield.v1beta1.ShieldService/CreateOrganization": true,
}

// authorizationValidationMap stores path to validation function
var authorizationValidationMap = map[string]func(ctx context.Context, handler *v1beta1.Handler, req any) error{
	// user
	"/raystack.shield.v1beta1.ShieldService/ListUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if handler.DisableUsersListing {
			return status.Error(codes.Unavailable, ErrNotAvailable.Error())
		}
		return nil
	},
	"/raystack.shield.v1beta1.ShieldService/CreateUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.shield.v1beta1.ShieldService/GetUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.shield.v1beta1.ShieldService/ListUserGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.shield.v1beta1.ShieldService/UpdateUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.shield.v1beta1.ShieldService/EnableUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.shield.v1beta1.ShieldService/DisableUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.shield.v1beta1.ShieldService/DeleteUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.shield.v1beta1.ShieldService/GetOrganizationsByUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.shield.v1beta1.ShieldService/GetOrganizationsByCurrentUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},

	// serviceuser
	"/raystack.shield.v1beta1.ShieldService/ListServiceUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListServiceUsersRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/CreateServiceUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateServiceUserRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.ServiceUserManagePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/GetServiceUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateServiceUserKeyRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DeleteServiceUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateServiceUserRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.ServiceUserManagePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/ListServiceUserKeys": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListServiceUserKeysRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/CreateServiceUserKey": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateServiceUserKeyRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DeleteServiceUserKey": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteServiceUserKeyRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/CreateServiceUserSecret": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateServiceUserSecretRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/ListServiceUserSecrets": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListServiceUserSecretsRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DeleteServiceUserSecret": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteServiceUserSecretRequest)
		return handler.IsAuthorized(ctx, schema.ServiceUserPrincipal, pbreq.GetId(), schema.ManagePermission)
	},

	// org
	"/raystack.shield.v1beta1.ShieldService/ListOrganizations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if handler.DisableOrgsListing {
			return status.Error(codes.Unavailable, ErrNotAvailable.Error())
		}
		return nil
	},
	"/raystack.shield.v1beta1.ShieldService/GetOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.GetOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/UpdateOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.UpdateOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/ListOrganizationUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListOrganizationUsersRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/ListOrganizationProjects": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListOrganizationProjectsRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.ProjectListPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/AddOrganizationUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.AddOrganizationUsersRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/RemoveOrganizationUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.RemoveOrganizationUserRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/ListOrganizationInvitations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListOrganizationInvitationsRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.InvitationListPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/CreateOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.InvitationCreatePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/GetOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.GetOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, schema.InvitationNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/AcceptOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.AcceptOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, schema.InvitationNamespace, pbreq.GetId(), schema.AcceptPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DeleteOrganizationInvitation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteOrganizationInvitationRequest)
		return handler.IsAuthorized(ctx, schema.InvitationNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/EnableOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.EnableOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DisableOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DisableOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DeleteOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.DeletePermission)
	},

	// group
	"/raystack.shield.v1beta1.ShieldService/ListOrganizationGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListOrganizationGroupsRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GroupListPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/CreateGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateGroupRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GroupCreatePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/GetGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.GetGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/UpdateGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.UpdateGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/ListGroupUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListGroupUsersRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/AddGroupUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.AddGroupUsersRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/RemoveGroupUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.RemoveGroupUserRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/EnableGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.EnableGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DisableGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DisableGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DeleteGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.DeletePermission)
	},

	// project
	"/raystack.shield.v1beta1.ShieldService/CreateProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateProjectRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetBody().GetOrgId(), schema.ProjectCreatePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/GetProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.GetProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/UpdateProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.UpdateProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/ListProjectAdmins": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListProjectAdminsRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/ListProjectUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListProjectUsersRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/EnableProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.EnableProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DisableProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DisableProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DeleteProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.DeletePermission)
	},

	// roles
	"/raystack.shield.v1beta1.ShieldService/ListRoles": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/raystack.shield.v1beta1.ShieldService/ListOrganizationRoles": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListOrganizationRolesRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/CreateOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.RoleManagePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/GetOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.GetOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/UpdateOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.UpdateOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.RoleManagePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DeleteOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.RoleManagePermission)
	},

	// policies
	"/raystack.shield.v1beta1.ShieldService/CreatePolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreatePolicyRequest)
		ns, id, err := schema.SplitNamespaceAndResourceID(pbreq.GetBody().GetResource())
		if err != nil {
			return err
		}

		if ns == schema.OrganizationNamespace {
			return handler.IsAuthorized(ctx, schema.OrganizationNamespace, id, schema.PolicyManagePermission)
		}
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, id, schema.PolicyManagePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/GetPolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/raystack.shield.v1beta1.ShieldService/UpdatePolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/raystack.shield.v1beta1.ShieldService/DeletePolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeletePolicyRequest)
		policyResp, err := handler.GetPolicy(ctx, &shieldv1beta1.GetPolicyRequest{Id: pbreq.Id})
		if err != nil {
			return err
		}
		ns, id, err := schema.SplitNamespaceAndResourceID(policyResp.GetPolicy().GetResource())
		if err != nil {
			return err
		}

		if ns == schema.OrganizationNamespace {
			return handler.IsAuthorized(ctx, schema.OrganizationNamespace, id, schema.PolicyManagePermission)
		}
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, id, schema.PolicyManagePermission)
	},

	// relations
	"/raystack.shield.v1beta1.ShieldService/CreateRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateRelationRequest)
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
	"/raystack.shield.v1beta1.ShieldService/GetRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.ShieldService/DeleteRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateRelationRequest)
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
	"/raystack.shield.v1beta1.ShieldService/ListProjectResources": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListProjectResourcesRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetProjectId(), schema.ResourceListPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/CreateProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateProjectResourceRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetProjectId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/GetProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.GetProjectResourceRequest)
		resp, err := handler.GetProjectResource(ctx, &shieldv1beta1.GetProjectResourceRequest{Id: pbreq.GetId()})
		if err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, resp.GetResource().GetNamespace(), resp.GetResource().GetId(), schema.GetPermission)
	},
	"/raystack.shield.v1beta1.ShieldService/UpdateProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.UpdateProjectResourceRequest)
		return handler.IsAuthorized(ctx, pbreq.GetBody().GetNamespace(), pbreq.GetId(), schema.UpdatePermission)
	},
	"/raystack.shield.v1beta1.ShieldService/DeleteProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteProjectResourceRequest)
		resp, err := handler.GetProjectResource(ctx, &shieldv1beta1.GetProjectResourceRequest{Id: pbreq.GetId()})
		if err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, resp.GetResource().GetNamespace(), resp.GetResource().GetId(), schema.DeletePermission)
	},

	// admin APIs
	"/raystack.shield.v1beta1.AdminService/ListAllUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.AdminService/ListGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.AdminService/ListAllOrganizations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.AdminService/ListProjects": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.AdminService/ListRelations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.AdminService/ListResources": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.AdminService/ListPolicies": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.AdminService/CreateRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.AdminService/DeleteRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.AdminService/CreatePermission": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.AdminService/UpdatePermission": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/raystack.shield.v1beta1.AdminService/DeletePermission": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
}
