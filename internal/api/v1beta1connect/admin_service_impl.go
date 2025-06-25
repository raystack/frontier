package v1beta1connect

import (
	"context"

	"github.com/bufbuild/connect-go"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
)

func (h *ConnectHandler) ListAllUsers(context.Context, *connect.Request[frontierv1beta1.ListAllUsersRequest]) (*connect.Response[frontierv1beta1.ListAllUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListGroups(context.Context, *connect.Request[frontierv1beta1.ListGroupsRequest]) (*connect.Response[frontierv1beta1.ListGroupsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListAllOrganizations(context.Context, *connect.Request[frontierv1beta1.ListAllOrganizationsRequest]) (*connect.Response[frontierv1beta1.ListAllOrganizationsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) AdminCreateOrganization(context.Context, *connect.Request[frontierv1beta1.AdminCreateOrganizationRequest]) (*connect.Response[frontierv1beta1.AdminCreateOrganizationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SearchOrganizations(context.Context, *connect.Request[frontierv1beta1.SearchOrganizationsRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SearchOrganizationUsers(context.Context, *connect.Request[frontierv1beta1.SearchOrganizationUsersRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SearchProjectUsers(context.Context, *connect.Request[frontierv1beta1.SearchProjectUsersRequest]) (*connect.Response[frontierv1beta1.SearchProjectUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SearchOrganizationProjects(context.Context, *connect.Request[frontierv1beta1.SearchOrganizationProjectsRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationProjectsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SearchOrganizationInvoices(context.Context, *connect.Request[frontierv1beta1.SearchOrganizationInvoicesRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationInvoicesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SearchOrganizationTokens(context.Context, *connect.Request[frontierv1beta1.SearchOrganizationTokensRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationTokensResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SearchOrganizationServiceUserCredentials(context.Context, *connect.Request[frontierv1beta1.SearchOrganizationServiceUserCredentialsRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationServiceUserCredentialsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ExportOrganizations(context.Context, *connect.Request[frontierv1beta1.ExportOrganizationsRequest], *connect.ServerStream[httpbody.HttpBody]) error {
	return connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ExportOrganizationUsers(context.Context, *connect.Request[frontierv1beta1.ExportOrganizationUsersRequest], *connect.ServerStream[httpbody.HttpBody]) error {
	return connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ExportOrganizationProjects(context.Context, *connect.Request[frontierv1beta1.ExportOrganizationProjectsRequest], *connect.ServerStream[httpbody.HttpBody]) error {
	return connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ExportOrganizationTokens(context.Context, *connect.Request[frontierv1beta1.ExportOrganizationTokensRequest], *connect.ServerStream[httpbody.HttpBody]) error {
	return connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ExportUsers(context.Context, *connect.Request[frontierv1beta1.ExportUsersRequest], *connect.ServerStream[httpbody.HttpBody]) error {
	return connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SearchUsers(context.Context, *connect.Request[frontierv1beta1.SearchUsersRequest]) (*connect.Response[frontierv1beta1.SearchUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SearchUserOrganizations(context.Context, *connect.Request[frontierv1beta1.SearchUserOrganizationsRequest]) (*connect.Response[frontierv1beta1.SearchUserOrganizationsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SearchUserProjects(context.Context, *connect.Request[frontierv1beta1.SearchUserProjectsRequest]) (*connect.Response[frontierv1beta1.SearchUserProjectsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SetOrganizationKyc(context.Context, *connect.Request[frontierv1beta1.SetOrganizationKycRequest]) (*connect.Response[frontierv1beta1.SetOrganizationKycResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListOrganizationsKyc(context.Context, *connect.Request[frontierv1beta1.ListOrganizationsKycRequest]) (*connect.Response[frontierv1beta1.ListOrganizationsKycResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListProjects(context.Context, *connect.Request[frontierv1beta1.ListProjectsRequest]) (*connect.Response[frontierv1beta1.ListProjectsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListRelations(context.Context, *connect.Request[frontierv1beta1.ListRelationsRequest]) (*connect.Response[frontierv1beta1.ListRelationsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListResources(context.Context, *connect.Request[frontierv1beta1.ListResourcesRequest]) (*connect.Response[frontierv1beta1.ListResourcesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateRole(context.Context, *connect.Request[frontierv1beta1.CreateRoleRequest]) (*connect.Response[frontierv1beta1.CreateRoleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateRole(context.Context, *connect.Request[frontierv1beta1.UpdateRoleRequest]) (*connect.Response[frontierv1beta1.UpdateRoleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteRole(context.Context, *connect.Request[frontierv1beta1.DeleteRoleRequest]) (*connect.Response[frontierv1beta1.DeleteRoleResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreatePermission(context.Context, *connect.Request[frontierv1beta1.CreatePermissionRequest]) (*connect.Response[frontierv1beta1.CreatePermissionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdatePermission(context.Context, *connect.Request[frontierv1beta1.UpdatePermissionRequest]) (*connect.Response[frontierv1beta1.UpdatePermissionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeletePermission(context.Context, *connect.Request[frontierv1beta1.DeletePermissionRequest]) (*connect.Response[frontierv1beta1.DeletePermissionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListPreferences(context.Context, *connect.Request[frontierv1beta1.ListPreferencesRequest]) (*connect.Response[frontierv1beta1.ListPreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreatePreferences(context.Context, *connect.Request[frontierv1beta1.CreatePreferencesRequest]) (*connect.Response[frontierv1beta1.CreatePreferencesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CheckFederatedResourcePermission(context.Context, *connect.Request[frontierv1beta1.CheckFederatedResourcePermissionRequest]) (*connect.Response[frontierv1beta1.CheckFederatedResourcePermissionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) AddPlatformUser(context.Context, *connect.Request[frontierv1beta1.AddPlatformUserRequest]) (*connect.Response[frontierv1beta1.AddPlatformUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListPlatformUsers(context.Context, *connect.Request[frontierv1beta1.ListPlatformUsersRequest]) (*connect.Response[frontierv1beta1.ListPlatformUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) RemovePlatformUser(context.Context, *connect.Request[frontierv1beta1.RemovePlatformUserRequest]) (*connect.Response[frontierv1beta1.RemovePlatformUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DelegatedCheckout(context.Context, *connect.Request[frontierv1beta1.DelegatedCheckoutRequest]) (*connect.Response[frontierv1beta1.DelegatedCheckoutResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListAllInvoices(context.Context, *connect.Request[frontierv1beta1.ListAllInvoicesRequest]) (*connect.Response[frontierv1beta1.ListAllInvoicesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GenerateInvoices(context.Context, *connect.Request[frontierv1beta1.GenerateInvoicesRequest]) (*connect.Response[frontierv1beta1.GenerateInvoicesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListAllBillingAccounts(context.Context, *connect.Request[frontierv1beta1.ListAllBillingAccountsRequest]) (*connect.Response[frontierv1beta1.ListAllBillingAccountsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) RevertBillingUsage(context.Context, *connect.Request[frontierv1beta1.RevertBillingUsageRequest]) (*connect.Response[frontierv1beta1.RevertBillingUsageResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateWebhook(context.Context, *connect.Request[frontierv1beta1.CreateWebhookRequest]) (*connect.Response[frontierv1beta1.CreateWebhookResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateWebhook(context.Context, *connect.Request[frontierv1beta1.UpdateWebhookRequest]) (*connect.Response[frontierv1beta1.UpdateWebhookResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteWebhook(context.Context, *connect.Request[frontierv1beta1.DeleteWebhookRequest]) (*connect.Response[frontierv1beta1.DeleteWebhookResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListWebhooks(context.Context, *connect.Request[frontierv1beta1.ListWebhooksRequest]) (*connect.Response[frontierv1beta1.ListWebhooksResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateBillingAccountLimits(context.Context, *connect.Request[frontierv1beta1.UpdateBillingAccountLimitsRequest]) (*connect.Response[frontierv1beta1.UpdateBillingAccountLimitsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetBillingAccountDetails(context.Context, *connect.Request[frontierv1beta1.GetBillingAccountDetailsRequest]) (*connect.Response[frontierv1beta1.GetBillingAccountDetailsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateBillingAccountDetails(context.Context, *connect.Request[frontierv1beta1.UpdateBillingAccountDetailsRequest]) (*connect.Response[frontierv1beta1.UpdateBillingAccountDetailsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) CreateProspect(context.Context, *connect.Request[frontierv1beta1.CreateProspectRequest]) (*connect.Response[frontierv1beta1.CreateProspectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) ListProspects(context.Context, *connect.Request[frontierv1beta1.ListProspectsRequest]) (*connect.Response[frontierv1beta1.ListProspectsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) GetProspect(context.Context, *connect.Request[frontierv1beta1.GetProspectRequest]) (*connect.Response[frontierv1beta1.GetProspectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) UpdateProspect(context.Context, *connect.Request[frontierv1beta1.UpdateProspectRequest]) (*connect.Response[frontierv1beta1.UpdateProspectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) DeleteProspect(context.Context, *connect.Request[frontierv1beta1.DeleteProspectRequest]) (*connect.Response[frontierv1beta1.DeleteProspectResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *ConnectHandler) SearchInvoices(context.Context, *connect.Request[frontierv1beta1.SearchInvoicesRequest]) (*connect.Response[frontierv1beta1.SearchInvoicesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
