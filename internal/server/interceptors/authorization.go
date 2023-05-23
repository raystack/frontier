package interceptors

import (
	"context"
	"fmt"

	"github.com/odpf/shield/internal/api/v1beta1"
	"github.com/odpf/shield/internal/bootstrap/schema"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotAvailable = fmt.Errorf("function not available at the moment")
)

func UnaryAuthorizationCheck(identityHeader string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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
		if serverHandler, ok := info.Server.(*v1beta1.Handler); ok {
			if azFunc, azVerifier := authorizationValidationMap[info.FullMethod]; azVerifier {
				if err = azFunc(ctx, serverHandler, req); err != nil {
					return nil, err
				}
			}
		}
		return handler(ctx, req)
	}
}

// authorizationValidationMap stores path to validation function
var authorizationValidationMap = map[string]func(ctx context.Context, handler *v1beta1.Handler, req any) error{
	// user
	"/odpf.shield.v1beta1.ShieldService/ListUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if handler.DisableUsersListing {
			return status.Error(codes.Unavailable, ErrNotAvailable.Error())
		}
		return nil
	},
	"/odpf.shield.v1beta1.ShieldService/CreateUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/odpf.shield.v1beta1.ShieldService/GetUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/odpf.shield.v1beta1.ShieldService/ListUserGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/odpf.shield.v1beta1.ShieldService/ListCurrentUserGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/odpf.shield.v1beta1.ShieldService/GetCurrentUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/odpf.shield.v1beta1.ShieldService/UpdateUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/odpf.shield.v1beta1.ShieldService/UpdateCurrentUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/odpf.shield.v1beta1.ShieldService/EnableUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/odpf.shield.v1beta1.ShieldService/DisableUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/odpf.shield.v1beta1.ShieldService/DeleteUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/odpf.shield.v1beta1.ShieldService/GetOrganizationsByUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if err := handler.IsSuperUser(ctx); err == nil {
			return nil
		}
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/odpf.shield.v1beta1.ShieldService/GetOrganizationsByCurrentUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},

	// org
	"/odpf.shield.v1beta1.ShieldService/ListOrganizations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		if handler.DisableOrgsListing {
			return status.Error(codes.Unavailable, ErrNotAvailable.Error())
		}
		return nil
	},
	"/odpf.shield.v1beta1.ShieldService/CreateOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/odpf.shield.v1beta1.ShieldService/GetOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.GetOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/UpdateOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.UpdateOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/ListOrganizationUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListOrganizationUsersRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/ListOrganizationProjects": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListOrganizationProjectsRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.ProjectListPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/AddOrganizationUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.AddOrganizationUsersRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/RemoveOrganizationUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.RemoveOrganizationUserRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/EnableOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.EnableOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/DisableOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DisableOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/DeleteOrganization": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteOrganizationRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetId(), schema.DeletePermission)
	},

	// group
	"/odpf.shield.v1beta1.ShieldService/ListOrganizationGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListOrganizationGroupsRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GroupListPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/CreateGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateGroupRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetBody().GetOrgId(), schema.GroupCreatePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/GetGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.GetGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/UpdateGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.UpdateGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/ListGroupUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListGroupUsersRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/AddGroupUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.AddGroupUsersRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/RemoveGroupUser": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.RemoveGroupUserRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/EnableGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.EnableGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/DisableGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DisableGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/DeleteGroup": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteGroupRequest)
		return handler.IsAuthorized(ctx, schema.GroupNamespace, pbreq.GetId(), schema.DeletePermission)
	},

	// project
	"/odpf.shield.v1beta1.ShieldService/CreateProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateProjectRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetBody().GetOrgId(), schema.ProjectCreatePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/GetProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.GetProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/UpdateProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.UpdateProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.UpdatePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/ListProjectAdmins": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListProjectAdminsRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/ListProjectUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListProjectUsersRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.GetPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/EnableProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.EnableProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/DisableProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DisableProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.DeletePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/DeleteProject": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteProjectRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetId(), schema.DeletePermission)
	},

	// roles
	"/odpf.shield.v1beta1.ShieldService/ListRoles": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/odpf.shield.v1beta1.ShieldService/ListOrganizationRoles": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListOrganizationRolesRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GetPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/CreateOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetBody().GetOrgId(), schema.RoleManagePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/GetOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.GetOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.GetPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/UpdateOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.UpdateOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetBody().GetOrgId(), schema.RoleManagePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/DeleteOrganizationRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteOrganizationRoleRequest)
		return handler.IsAuthorized(ctx, schema.OrganizationNamespace, pbreq.GetOrgId(), schema.RoleManagePermission)
	},

	// permissions
	"/odpf.shield.v1beta1.ShieldService/ListPermissions": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/odpf.shield.v1beta1.ShieldService/GetPermission": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},

	// namespaces
	"/odpf.shield.v1beta1.ShieldService/ListNamespaces": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/odpf.shield.v1beta1.ShieldService/GetNamespace": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},

	// policies
	"/odpf.shield.v1beta1.ShieldService/CreatePolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
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
	"/odpf.shield.v1beta1.ShieldService/GetPolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/odpf.shield.v1beta1.ShieldService/UpdatePolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
	"/odpf.shield.v1beta1.ShieldService/DeletePolicy": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
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
	"/odpf.shield.v1beta1.ShieldService/CreateRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
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
	"/odpf.shield.v1beta1.ShieldService/GetRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return nil
	},
	"/odpf.shield.v1beta1.ShieldService/DeleteRelation": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
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
	"/odpf.shield.v1beta1.ShieldService/ListProjectResources": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.ListProjectResourcesRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetProjectId(), schema.ResourceListPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/CreateProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.CreateProjectResourceRequest)
		return handler.IsAuthorized(ctx, schema.ProjectNamespace, pbreq.GetBody().GetProjectId(), schema.GetPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/GetProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.GetProjectResourceRequest)
		resp, err := handler.GetProjectResource(ctx, &shieldv1beta1.GetProjectResourceRequest{Id: pbreq.GetId()})
		if err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, resp.GetResource().GetNamespace(), resp.GetResource().GetId(), schema.GetPermission)
	},
	"/odpf.shield.v1beta1.ShieldService/UpdateProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.UpdateProjectResourceRequest)
		return handler.IsAuthorized(ctx, pbreq.GetBody().GetNamespace(), pbreq.GetId(), schema.UpdatePermission)
	},
	"/odpf.shield.v1beta1.ShieldService/DeleteProjectResource": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		pbreq := req.(*shieldv1beta1.DeleteProjectResourceRequest)
		resp, err := handler.GetProjectResource(ctx, &shieldv1beta1.GetProjectResourceRequest{Id: pbreq.GetId()})
		if err != nil {
			return err
		}
		return handler.IsAuthorized(ctx, resp.GetResource().GetNamespace(), resp.GetResource().GetId(), schema.DeletePermission)
	},

	// admin APIs
	"/odpf.shield.v1beta1.AdminService/ListAllUsers": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/odpf.shield.v1beta1.AdminService/ListGroups": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/odpf.shield.v1beta1.AdminService/ListAllOrganizations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/odpf.shield.v1beta1.AdminService/ListProjects": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/odpf.shield.v1beta1.AdminService/ListRelations": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/odpf.shield.v1beta1.AdminService/ListResources": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/odpf.shield.v1beta1.AdminService/ListPolicies": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/odpf.shield.v1beta1.AdminService/CreateRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/odpf.shield.v1beta1.AdminService/DeleteRole": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/odpf.shield.v1beta1.AdminService/CreatePermission": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/odpf.shield.v1beta1.AdminService/UpdatePermission": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return handler.IsSuperUser(ctx)
	},
	"/odpf.shield.v1beta1.AdminService/DeletePermission": func(ctx context.Context, handler *v1beta1.Handler, req any) error {
		return status.Error(codes.Unavailable, ErrNotAvailable.Error())
	},
}
