package v1beta1connect

import (
	"context"

	"github.com/bufbuild/connect-go"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

func (h *ConnectHandler) ListUsers(context.Context, *connect.Request[frontierv1beta1.ListUsersRequest]) (*connect.Response[frontierv1beta1.ListUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateUser(context.Context, *connect.Request[frontierv1beta1.CreateUserRequest]) (*connect.Response[frontierv1beta1.CreateUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetUser(context.Context, *connect.Request[frontierv1beta1.GetUserRequest]) (*connect.Response[frontierv1beta1.GetUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListUserGroups(context.Context, *connect.Request[frontierv1beta1.ListUserGroupsRequest]) (*connect.Response[frontierv1beta1.ListUserGroupsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListCurrentUserGroups(context.Context, *connect.Request[frontierv1beta1.ListCurrentUserGroupsRequest]) (*connect.Response[frontierv1beta1.ListCurrentUserGroupsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetCurrentUser(context.Context, *connect.Request[frontierv1beta1.GetCurrentUserRequest]) (*connect.Response[frontierv1beta1.GetCurrentUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateUser(context.Context, *connect.Request[frontierv1beta1.UpdateUserRequest]) (*connect.Response[frontierv1beta1.UpdateUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateCurrentUser(context.Context, *connect.Request[frontierv1beta1.UpdateCurrentUserRequest]) (*connect.Response[frontierv1beta1.UpdateCurrentUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) EnableUser(context.Context, *connect.Request[frontierv1beta1.EnableUserRequest]) (*connect.Response[frontierv1beta1.EnableUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DisableUser(context.Context, *connect.Request[frontierv1beta1.DisableUserRequest]) (*connect.Response[frontierv1beta1.DisableUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteUser(context.Context, *connect.Request[frontierv1beta1.DeleteUserRequest]) (*connect.Response[frontierv1beta1.DeleteUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationsByUser(context.Context, *connect.Request[frontierv1beta1.ListOrganizationsByUserRequest]) (*connect.Response[frontierv1beta1.ListOrganizationsByUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationsByCurrentUser(context.Context, *connect.Request[frontierv1beta1.ListOrganizationsByCurrentUserRequest]) (*connect.Response[frontierv1beta1.ListOrganizationsByCurrentUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListProjectsByUser(context.Context, *connect.Request[frontierv1beta1.ListProjectsByUserRequest]) (*connect.Response[frontierv1beta1.ListProjectsByUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListProjectsByCurrentUser(context.Context, *connect.Request[frontierv1beta1.ListProjectsByCurrentUserRequest]) (*connect.Response[frontierv1beta1.ListProjectsByCurrentUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListUserInvitations(context.Context, *connect.Request[frontierv1beta1.ListUserInvitationsRequest]) (*connect.Response[frontierv1beta1.ListUserInvitationsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListCurrentUserInvitations(context.Context, *connect.Request[frontierv1beta1.ListCurrentUserInvitationsRequest]) (*connect.Response[frontierv1beta1.ListCurrentUserInvitationsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListServiceUsers(context.Context, *connect.Request[frontierv1beta1.ListServiceUsersRequest]) (*connect.Response[frontierv1beta1.ListServiceUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateServiceUser(context.Context, *connect.Request[frontierv1beta1.CreateServiceUserRequest]) (*connect.Response[frontierv1beta1.CreateServiceUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetServiceUser(context.Context, *connect.Request[frontierv1beta1.GetServiceUserRequest]) (*connect.Response[frontierv1beta1.GetServiceUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteServiceUser(context.Context, *connect.Request[frontierv1beta1.DeleteServiceUserRequest]) (*connect.Response[frontierv1beta1.DeleteServiceUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateServiceUserJWK(context.Context, *connect.Request[frontierv1beta1.CreateServiceUserJWKRequest]) (*connect.Response[frontierv1beta1.CreateServiceUserJWKResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListServiceUserJWKs(context.Context, *connect.Request[frontierv1beta1.ListServiceUserJWKsRequest]) (*connect.Response[frontierv1beta1.ListServiceUserJWKsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetServiceUserJWK(context.Context, *connect.Request[frontierv1beta1.GetServiceUserJWKRequest]) (*connect.Response[frontierv1beta1.GetServiceUserJWKResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteServiceUserJWK(context.Context, *connect.Request[frontierv1beta1.DeleteServiceUserJWKRequest]) (*connect.Response[frontierv1beta1.DeleteServiceUserJWKResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateServiceUserCredential(context.Context, *connect.Request[frontierv1beta1.CreateServiceUserCredentialRequest]) (*connect.Response[frontierv1beta1.CreateServiceUserCredentialResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListServiceUserCredentials(context.Context, *connect.Request[frontierv1beta1.ListServiceUserCredentialsRequest]) (*connect.Response[frontierv1beta1.ListServiceUserCredentialsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteServiceUserCredential(context.Context, *connect.Request[frontierv1beta1.DeleteServiceUserCredentialRequest]) (*connect.Response[frontierv1beta1.DeleteServiceUserCredentialResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateServiceUserToken(context.Context, *connect.Request[frontierv1beta1.CreateServiceUserTokenRequest]) (*connect.Response[frontierv1beta1.CreateServiceUserTokenResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListServiceUserTokens(context.Context, *connect.Request[frontierv1beta1.ListServiceUserTokensRequest]) (*connect.Response[frontierv1beta1.ListServiceUserTokensResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteServiceUserToken(context.Context, *connect.Request[frontierv1beta1.DeleteServiceUserTokenRequest]) (*connect.Response[frontierv1beta1.DeleteServiceUserTokenResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListServiceUserProjects(context.Context, *connect.Request[frontierv1beta1.ListServiceUserProjectsRequest]) (*connect.Response[frontierv1beta1.ListServiceUserProjectsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationGroups(context.Context, *connect.Request[frontierv1beta1.ListOrganizationGroupsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationGroupsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateGroup(context.Context, *connect.Request[frontierv1beta1.CreateGroupRequest]) (*connect.Response[frontierv1beta1.CreateGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetGroup(context.Context, *connect.Request[frontierv1beta1.GetGroupRequest]) (*connect.Response[frontierv1beta1.GetGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateGroup(context.Context, *connect.Request[frontierv1beta1.UpdateGroupRequest]) (*connect.Response[frontierv1beta1.UpdateGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListGroupUsers(context.Context, *connect.Request[frontierv1beta1.ListGroupUsersRequest]) (*connect.Response[frontierv1beta1.ListGroupUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) AddGroupUsers(context.Context, *connect.Request[frontierv1beta1.AddGroupUsersRequest]) (*connect.Response[frontierv1beta1.AddGroupUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) RemoveGroupUser(context.Context, *connect.Request[frontierv1beta1.RemoveGroupUserRequest]) (*connect.Response[frontierv1beta1.RemoveGroupUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) EnableGroup(context.Context, *connect.Request[frontierv1beta1.EnableGroupRequest]) (*connect.Response[frontierv1beta1.EnableGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DisableGroup(context.Context, *connect.Request[frontierv1beta1.DisableGroupRequest]) (*connect.Response[frontierv1beta1.DisableGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteGroup(context.Context, *connect.Request[frontierv1beta1.DeleteGroupRequest]) (*connect.Response[frontierv1beta1.DeleteGroupResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListRoles(context.Context, *connect.Request[frontierv1beta1.ListRolesRequest]) (*connect.Response[frontierv1beta1.ListRolesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationRoles(context.Context, *connect.Request[frontierv1beta1.ListOrganizationRolesRequest]) (*connect.Response[frontierv1beta1.ListOrganizationRolesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateOrganizationRole(context.Context, *connect.Request[frontierv1beta1.CreateOrganizationRoleRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationRoleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetOrganizationRole(context.Context, *connect.Request[frontierv1beta1.GetOrganizationRoleRequest]) (*connect.Response[frontierv1beta1.GetOrganizationRoleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateOrganizationRole(context.Context, *connect.Request[frontierv1beta1.UpdateOrganizationRoleRequest]) (*connect.Response[frontierv1beta1.UpdateOrganizationRoleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteOrganizationRole(context.Context, *connect.Request[frontierv1beta1.DeleteOrganizationRoleRequest]) (*connect.Response[frontierv1beta1.DeleteOrganizationRoleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizations(context.Context, *connect.Request[frontierv1beta1.ListOrganizationsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateOrganization(context.Context, *connect.Request[frontierv1beta1.CreateOrganizationRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetOrganization(context.Context, *connect.Request[frontierv1beta1.GetOrganizationRequest]) (*connect.Response[frontierv1beta1.GetOrganizationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateOrganization(context.Context, *connect.Request[frontierv1beta1.UpdateOrganizationRequest]) (*connect.Response[frontierv1beta1.UpdateOrganizationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationProjects(context.Context, *connect.Request[frontierv1beta1.ListOrganizationProjectsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationProjectsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationAdmins(context.Context, *connect.Request[frontierv1beta1.ListOrganizationAdminsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationAdminsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationUsers(context.Context, *connect.Request[frontierv1beta1.ListOrganizationUsersRequest]) (*connect.Response[frontierv1beta1.ListOrganizationUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) AddOrganizationUsers(context.Context, *connect.Request[frontierv1beta1.AddOrganizationUsersRequest]) (*connect.Response[frontierv1beta1.AddOrganizationUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) RemoveOrganizationUser(context.Context, *connect.Request[frontierv1beta1.RemoveOrganizationUserRequest]) (*connect.Response[frontierv1beta1.RemoveOrganizationUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetOrganizationKyc(context.Context, *connect.Request[frontierv1beta1.GetOrganizationKycRequest]) (*connect.Response[frontierv1beta1.GetOrganizationKycResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationServiceUsers(context.Context, *connect.Request[frontierv1beta1.ListOrganizationServiceUsersRequest]) (*connect.Response[frontierv1beta1.ListOrganizationServiceUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationInvitations(context.Context, *connect.Request[frontierv1beta1.ListOrganizationInvitationsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationInvitationsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateOrganizationInvitation(context.Context, *connect.Request[frontierv1beta1.CreateOrganizationInvitationRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationInvitationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetOrganizationInvitation(context.Context, *connect.Request[frontierv1beta1.GetOrganizationInvitationRequest]) (*connect.Response[frontierv1beta1.GetOrganizationInvitationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) AcceptOrganizationInvitation(context.Context, *connect.Request[frontierv1beta1.AcceptOrganizationInvitationRequest]) (*connect.Response[frontierv1beta1.AcceptOrganizationInvitationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteOrganizationInvitation(context.Context, *connect.Request[frontierv1beta1.DeleteOrganizationInvitationRequest]) (*connect.Response[frontierv1beta1.DeleteOrganizationInvitationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationDomains(context.Context, *connect.Request[frontierv1beta1.ListOrganizationDomainsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationDomainsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateOrganizationDomain(context.Context, *connect.Request[frontierv1beta1.CreateOrganizationDomainRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationDomainResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteOrganizationDomain(context.Context, *connect.Request[frontierv1beta1.DeleteOrganizationDomainRequest]) (*connect.Response[frontierv1beta1.DeleteOrganizationDomainResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetOrganizationDomain(context.Context, *connect.Request[frontierv1beta1.GetOrganizationDomainRequest]) (*connect.Response[frontierv1beta1.GetOrganizationDomainResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) VerifyOrganizationDomain(context.Context, *connect.Request[frontierv1beta1.VerifyOrganizationDomainRequest]) (*connect.Response[frontierv1beta1.VerifyOrganizationDomainResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) JoinOrganization(context.Context, *connect.Request[frontierv1beta1.JoinOrganizationRequest]) (*connect.Response[frontierv1beta1.JoinOrganizationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) EnableOrganization(context.Context, *connect.Request[frontierv1beta1.EnableOrganizationRequest]) (*connect.Response[frontierv1beta1.EnableOrganizationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DisableOrganization(context.Context, *connect.Request[frontierv1beta1.DisableOrganizationRequest]) (*connect.Response[frontierv1beta1.DisableOrganizationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteOrganization(context.Context, *connect.Request[frontierv1beta1.DeleteOrganizationRequest]) (*connect.Response[frontierv1beta1.DeleteOrganizationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateProject(context.Context, *connect.Request[frontierv1beta1.CreateProjectRequest]) (*connect.Response[frontierv1beta1.CreateProjectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetProject(context.Context, *connect.Request[frontierv1beta1.GetProjectRequest]) (*connect.Response[frontierv1beta1.GetProjectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateProject(context.Context, *connect.Request[frontierv1beta1.UpdateProjectRequest]) (*connect.Response[frontierv1beta1.UpdateProjectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListProjectAdmins(context.Context, *connect.Request[frontierv1beta1.ListProjectAdminsRequest]) (*connect.Response[frontierv1beta1.ListProjectAdminsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListProjectUsers(context.Context, *connect.Request[frontierv1beta1.ListProjectUsersRequest]) (*connect.Response[frontierv1beta1.ListProjectUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListProjectServiceUsers(context.Context, *connect.Request[frontierv1beta1.ListProjectServiceUsersRequest]) (*connect.Response[frontierv1beta1.ListProjectServiceUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListProjectGroups(context.Context, *connect.Request[frontierv1beta1.ListProjectGroupsRequest]) (*connect.Response[frontierv1beta1.ListProjectGroupsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) EnableProject(context.Context, *connect.Request[frontierv1beta1.EnableProjectRequest]) (*connect.Response[frontierv1beta1.EnableProjectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DisableProject(context.Context, *connect.Request[frontierv1beta1.DisableProjectRequest]) (*connect.Response[frontierv1beta1.DisableProjectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteProject(context.Context, *connect.Request[frontierv1beta1.DeleteProjectRequest]) (*connect.Response[frontierv1beta1.DeleteProjectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreatePolicy(context.Context, *connect.Request[frontierv1beta1.CreatePolicyRequest]) (*connect.Response[frontierv1beta1.CreatePolicyResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetPolicy(context.Context, *connect.Request[frontierv1beta1.GetPolicyRequest]) (*connect.Response[frontierv1beta1.GetPolicyResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListPolicies(context.Context, *connect.Request[frontierv1beta1.ListPoliciesRequest]) (*connect.Response[frontierv1beta1.ListPoliciesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdatePolicy(context.Context, *connect.Request[frontierv1beta1.UpdatePolicyRequest]) (*connect.Response[frontierv1beta1.UpdatePolicyResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeletePolicy(context.Context, *connect.Request[frontierv1beta1.DeletePolicyRequest]) (*connect.Response[frontierv1beta1.DeletePolicyResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreatePolicyForProject(context.Context, *connect.Request[frontierv1beta1.CreatePolicyForProjectRequest]) (*connect.Response[frontierv1beta1.CreatePolicyForProjectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateRelation(context.Context, *connect.Request[frontierv1beta1.CreateRelationRequest]) (*connect.Response[frontierv1beta1.CreateRelationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetRelation(context.Context, *connect.Request[frontierv1beta1.GetRelationRequest]) (*connect.Response[frontierv1beta1.GetRelationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteRelation(context.Context, *connect.Request[frontierv1beta1.DeleteRelationRequest]) (*connect.Response[frontierv1beta1.DeleteRelationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListPermissions(context.Context, *connect.Request[frontierv1beta1.ListPermissionsRequest]) (*connect.Response[frontierv1beta1.ListPermissionsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetPermission(context.Context, *connect.Request[frontierv1beta1.GetPermissionRequest]) (*connect.Response[frontierv1beta1.GetPermissionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListNamespaces(context.Context, *connect.Request[frontierv1beta1.ListNamespacesRequest]) (*connect.Response[frontierv1beta1.ListNamespacesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetNamespace(context.Context, *connect.Request[frontierv1beta1.GetNamespaceRequest]) (*connect.Response[frontierv1beta1.GetNamespaceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListProjectResources(context.Context, *connect.Request[frontierv1beta1.ListProjectResourcesRequest]) (*connect.Response[frontierv1beta1.ListProjectResourcesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateProjectResource(context.Context, *connect.Request[frontierv1beta1.CreateProjectResourceRequest]) (*connect.Response[frontierv1beta1.CreateProjectResourceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetProjectResource(context.Context, *connect.Request[frontierv1beta1.GetProjectResourceRequest]) (*connect.Response[frontierv1beta1.GetProjectResourceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateProjectResource(context.Context, *connect.Request[frontierv1beta1.UpdateProjectResourceRequest]) (*connect.Response[frontierv1beta1.UpdateProjectResourceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteProjectResource(context.Context, *connect.Request[frontierv1beta1.DeleteProjectResourceRequest]) (*connect.Response[frontierv1beta1.DeleteProjectResourceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CheckResourcePermission(context.Context, *connect.Request[frontierv1beta1.CheckResourcePermissionRequest]) (*connect.Response[frontierv1beta1.CheckResourcePermissionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) BatchCheckPermission(context.Context, *connect.Request[frontierv1beta1.BatchCheckPermissionRequest]) (*connect.Response[frontierv1beta1.BatchCheckPermissionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetJWKs(context.Context, *connect.Request[frontierv1beta1.GetJWKsRequest]) (*connect.Response[frontierv1beta1.GetJWKsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListAuthStrategies(context.Context, *connect.Request[frontierv1beta1.ListAuthStrategiesRequest]) (*connect.Response[frontierv1beta1.ListAuthStrategiesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) Authenticate(context.Context, *connect.Request[frontierv1beta1.AuthenticateRequest]) (*connect.Response[frontierv1beta1.AuthenticateResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) AuthCallback(context.Context, *connect.Request[frontierv1beta1.AuthCallbackRequest]) (*connect.Response[frontierv1beta1.AuthCallbackResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) AuthToken(context.Context, *connect.Request[frontierv1beta1.AuthTokenRequest]) (*connect.Response[frontierv1beta1.AuthTokenResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) AuthLogout(context.Context, *connect.Request[frontierv1beta1.AuthLogoutRequest]) (*connect.Response[frontierv1beta1.AuthLogoutResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListMetaSchemas(context.Context, *connect.Request[frontierv1beta1.ListMetaSchemasRequest]) (*connect.Response[frontierv1beta1.ListMetaSchemasResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateMetaSchema(context.Context, *connect.Request[frontierv1beta1.CreateMetaSchemaRequest]) (*connect.Response[frontierv1beta1.CreateMetaSchemaResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetMetaSchema(context.Context, *connect.Request[frontierv1beta1.GetMetaSchemaRequest]) (*connect.Response[frontierv1beta1.GetMetaSchemaResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateMetaSchema(context.Context, *connect.Request[frontierv1beta1.UpdateMetaSchemaRequest]) (*connect.Response[frontierv1beta1.UpdateMetaSchemaResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteMetaSchema(context.Context, *connect.Request[frontierv1beta1.DeleteMetaSchemaRequest]) (*connect.Response[frontierv1beta1.DeleteMetaSchemaResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationAuditLogs(context.Context, *connect.Request[frontierv1beta1.ListOrganizationAuditLogsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationAuditLogsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateOrganizationAuditLogs(context.Context, *connect.Request[frontierv1beta1.CreateOrganizationAuditLogsRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationAuditLogsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetOrganizationAuditLog(context.Context, *connect.Request[frontierv1beta1.GetOrganizationAuditLogRequest]) (*connect.Response[frontierv1beta1.GetOrganizationAuditLogResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DescribePreferences(context.Context, *connect.Request[frontierv1beta1.DescribePreferencesRequest]) (*connect.Response[frontierv1beta1.DescribePreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateOrganizationPreferences(context.Context, *connect.Request[frontierv1beta1.CreateOrganizationPreferencesRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationPreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationPreferences(context.Context, *connect.Request[frontierv1beta1.ListOrganizationPreferencesRequest]) (*connect.Response[frontierv1beta1.ListOrganizationPreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateProjectPreferences(context.Context, *connect.Request[frontierv1beta1.CreateProjectPreferencesRequest]) (*connect.Response[frontierv1beta1.CreateProjectPreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListProjectPreferences(context.Context, *connect.Request[frontierv1beta1.ListProjectPreferencesRequest]) (*connect.Response[frontierv1beta1.ListProjectPreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateGroupPreferences(context.Context, *connect.Request[frontierv1beta1.CreateGroupPreferencesRequest]) (*connect.Response[frontierv1beta1.CreateGroupPreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListGroupPreferences(context.Context, *connect.Request[frontierv1beta1.ListGroupPreferencesRequest]) (*connect.Response[frontierv1beta1.ListGroupPreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateUserPreferences(context.Context, *connect.Request[frontierv1beta1.CreateUserPreferencesRequest]) (*connect.Response[frontierv1beta1.CreateUserPreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListUserPreferences(context.Context, *connect.Request[frontierv1beta1.ListUserPreferencesRequest]) (*connect.Response[frontierv1beta1.ListUserPreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateCurrentUserPreferences(context.Context, *connect.Request[frontierv1beta1.CreateCurrentUserPreferencesRequest]) (*connect.Response[frontierv1beta1.CreateCurrentUserPreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListCurrentUserPreferences(context.Context, *connect.Request[frontierv1beta1.ListCurrentUserPreferencesRequest]) (*connect.Response[frontierv1beta1.ListCurrentUserPreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateBillingAccount(context.Context, *connect.Request[frontierv1beta1.CreateBillingAccountRequest]) (*connect.Response[frontierv1beta1.CreateBillingAccountResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetBillingAccount(context.Context, *connect.Request[frontierv1beta1.GetBillingAccountRequest]) (*connect.Response[frontierv1beta1.GetBillingAccountResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateBillingAccount(context.Context, *connect.Request[frontierv1beta1.UpdateBillingAccountRequest]) (*connect.Response[frontierv1beta1.UpdateBillingAccountResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) RegisterBillingAccount(context.Context, *connect.Request[frontierv1beta1.RegisterBillingAccountRequest]) (*connect.Response[frontierv1beta1.RegisterBillingAccountResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListBillingAccounts(context.Context, *connect.Request[frontierv1beta1.ListBillingAccountsRequest]) (*connect.Response[frontierv1beta1.ListBillingAccountsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteBillingAccount(context.Context, *connect.Request[frontierv1beta1.DeleteBillingAccountRequest]) (*connect.Response[frontierv1beta1.DeleteBillingAccountResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) EnableBillingAccount(context.Context, *connect.Request[frontierv1beta1.EnableBillingAccountRequest]) (*connect.Response[frontierv1beta1.EnableBillingAccountResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DisableBillingAccount(context.Context, *connect.Request[frontierv1beta1.DisableBillingAccountRequest]) (*connect.Response[frontierv1beta1.DisableBillingAccountResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetBillingBalance(context.Context, *connect.Request[frontierv1beta1.GetBillingBalanceRequest]) (*connect.Response[frontierv1beta1.GetBillingBalanceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) HasTrialed(context.Context, *connect.Request[frontierv1beta1.HasTrialedRequest]) (*connect.Response[frontierv1beta1.HasTrialedResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetSubscription(context.Context, *connect.Request[frontierv1beta1.GetSubscriptionRequest]) (*connect.Response[frontierv1beta1.GetSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CancelSubscription(context.Context, *connect.Request[frontierv1beta1.CancelSubscriptionRequest]) (*connect.Response[frontierv1beta1.CancelSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListSubscriptions(context.Context, *connect.Request[frontierv1beta1.ListSubscriptionsRequest]) (*connect.Response[frontierv1beta1.ListSubscriptionsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ChangeSubscription(context.Context, *connect.Request[frontierv1beta1.ChangeSubscriptionRequest]) (*connect.Response[frontierv1beta1.ChangeSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateSubscription(context.Context, *connect.Request[frontierv1beta1.UpdateSubscriptionRequest]) (*connect.Response[frontierv1beta1.UpdateSubscriptionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateProduct(context.Context, *connect.Request[frontierv1beta1.CreateProductRequest]) (*connect.Response[frontierv1beta1.CreateProductResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetProduct(context.Context, *connect.Request[frontierv1beta1.GetProductRequest]) (*connect.Response[frontierv1beta1.GetProductResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListProducts(context.Context, *connect.Request[frontierv1beta1.ListProductsRequest]) (*connect.Response[frontierv1beta1.ListProductsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateProduct(context.Context, *connect.Request[frontierv1beta1.UpdateProductRequest]) (*connect.Response[frontierv1beta1.UpdateProductResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateFeature(context.Context, *connect.Request[frontierv1beta1.CreateFeatureRequest]) (*connect.Response[frontierv1beta1.CreateFeatureResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetFeature(context.Context, *connect.Request[frontierv1beta1.GetFeatureRequest]) (*connect.Response[frontierv1beta1.GetFeatureResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateFeature(context.Context, *connect.Request[frontierv1beta1.UpdateFeatureRequest]) (*connect.Response[frontierv1beta1.UpdateFeatureResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListFeatures(context.Context, *connect.Request[frontierv1beta1.ListFeaturesRequest]) (*connect.Response[frontierv1beta1.ListFeaturesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreatePlan(context.Context, *connect.Request[frontierv1beta1.CreatePlanRequest]) (*connect.Response[frontierv1beta1.CreatePlanResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListPlans(context.Context, *connect.Request[frontierv1beta1.ListPlansRequest]) (*connect.Response[frontierv1beta1.ListPlansResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetPlan(context.Context, *connect.Request[frontierv1beta1.GetPlanRequest]) (*connect.Response[frontierv1beta1.GetPlanResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdatePlan(context.Context, *connect.Request[frontierv1beta1.UpdatePlanRequest]) (*connect.Response[frontierv1beta1.UpdatePlanResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateCheckout(context.Context, *connect.Request[frontierv1beta1.CreateCheckoutRequest]) (*connect.Response[frontierv1beta1.CreateCheckoutResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListCheckouts(context.Context, *connect.Request[frontierv1beta1.ListCheckoutsRequest]) (*connect.Response[frontierv1beta1.ListCheckoutsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetCheckout(context.Context, *connect.Request[frontierv1beta1.GetCheckoutRequest]) (*connect.Response[frontierv1beta1.GetCheckoutResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CheckFeatureEntitlement(context.Context, *connect.Request[frontierv1beta1.CheckFeatureEntitlementRequest]) (*connect.Response[frontierv1beta1.CheckFeatureEntitlementResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateBillingUsage(context.Context, *connect.Request[frontierv1beta1.CreateBillingUsageRequest]) (*connect.Response[frontierv1beta1.CreateBillingUsageResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListBillingTransactions(context.Context, *connect.Request[frontierv1beta1.ListBillingTransactionsRequest]) (*connect.Response[frontierv1beta1.ListBillingTransactionsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) TotalDebitedTransactions(context.Context, *connect.Request[frontierv1beta1.TotalDebitedTransactionsRequest]) (*connect.Response[frontierv1beta1.TotalDebitedTransactionsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListInvoices(context.Context, *connect.Request[frontierv1beta1.ListInvoicesRequest]) (*connect.Response[frontierv1beta1.ListInvoicesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetUpcomingInvoice(context.Context, *connect.Request[frontierv1beta1.GetUpcomingInvoiceRequest]) (*connect.Response[frontierv1beta1.GetUpcomingInvoiceResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) BillingWebhookCallback(context.Context, *connect.Request[frontierv1beta1.BillingWebhookCallbackRequest]) (*connect.Response[frontierv1beta1.BillingWebhookCallbackResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateProspectPublic(context.Context, *connect.Request[frontierv1beta1.CreateProspectPublicRequest]) (*connect.Response[frontierv1beta1.CreateProspectPublicResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
