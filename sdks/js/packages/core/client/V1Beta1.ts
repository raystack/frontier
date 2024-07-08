/* eslint-disable */
/* tslint:disable */
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

import {
  ChangeSubscriptionRequestPhaseChange,
  ChangeSubscriptionRequestPlanChange,
  RpcStatus,
  V1Beta1AcceptOrganizationInvitationResponse,
  V1Beta1AddGroupUsersResponse,
  V1Beta1AddOrganizationUsersResponse,
  V1Beta1AddPlatformUserRequest,
  V1Beta1AddPlatformUserResponse,
  V1Beta1AuditLog,
  V1Beta1AuthCallbackRequest,
  V1Beta1AuthCallbackResponse,
  V1Beta1AuthLogoutResponse,
  V1Beta1AuthTokenRequest,
  V1Beta1AuthTokenResponse,
  V1Beta1AuthenticateResponse,
  V1Beta1BatchCheckPermissionRequest,
  V1Beta1BatchCheckPermissionResponse,
  V1Beta1BillingAccountRequestBody,
  V1Beta1BillingWebhookCallbackResponse,
  V1Beta1CancelSubscriptionResponse,
  V1Beta1ChangeSubscriptionResponse,
  V1Beta1CheckFeatureEntitlementRequest,
  V1Beta1CheckFeatureEntitlementResponse,
  V1Beta1CheckFederatedResourcePermissionRequest,
  V1Beta1CheckFederatedResourcePermissionResponse,
  V1Beta1CheckResourcePermissionRequest,
  V1Beta1CheckResourcePermissionResponse,
  V1Beta1CheckoutProductBody,
  V1Beta1CheckoutSetupBody,
  V1Beta1CheckoutSubscriptionBody,
  V1Beta1CreateBillingAccountResponse,
  V1Beta1CreateBillingUsageRequest,
  V1Beta1CreateBillingUsageResponse,
  V1Beta1CreateCheckoutResponse,
  V1Beta1CreateCurrentUserPreferencesRequest,
  V1Beta1CreateCurrentUserPreferencesResponse,
  V1Beta1CreateFeatureRequest,
  V1Beta1CreateFeatureResponse,
  V1Beta1CreateGroupPreferencesResponse,
  V1Beta1CreateGroupResponse,
  V1Beta1CreateMetaSchemaResponse,
  V1Beta1CreateOrganizationAuditLogsResponse,
  V1Beta1CreateOrganizationDomainResponse,
  V1Beta1CreateOrganizationInvitationResponse,
  V1Beta1CreateOrganizationPreferencesResponse,
  V1Beta1CreateOrganizationResponse,
  V1Beta1CreateOrganizationRoleResponse,
  V1Beta1CreatePermissionRequest,
  V1Beta1CreatePermissionResponse,
  V1Beta1CreatePlanRequest,
  V1Beta1CreatePlanResponse,
  V1Beta1CreatePolicyForProjectBody,
  V1Beta1CreatePolicyForProjectResponse,
  V1Beta1CreatePolicyResponse,
  V1Beta1CreatePreferencesRequest,
  V1Beta1CreatePreferencesResponse,
  V1Beta1CreateProductRequest,
  V1Beta1CreateProductResponse,
  V1Beta1CreateProjectPreferencesResponse,
  V1Beta1CreateProjectResourceResponse,
  V1Beta1CreateProjectResponse,
  V1Beta1CreateRelationResponse,
  V1Beta1CreateRoleResponse,
  V1Beta1CreateServiceUserCredentialResponse,
  V1Beta1CreateServiceUserJWKResponse,
  V1Beta1CreateServiceUserRequest,
  V1Beta1CreateServiceUserResponse,
  V1Beta1CreateServiceUserTokenResponse,
  V1Beta1CreateUserPreferencesResponse,
  V1Beta1CreateUserResponse,
  V1Beta1CreateWebhookRequest,
  V1Beta1CreateWebhookResponse,
  V1Beta1DelegatedCheckoutResponse,
  V1Beta1DeleteBillingAccountResponse,
  V1Beta1DeleteGroupResponse,
  V1Beta1DeleteMetaSchemaResponse,
  V1Beta1DeleteOrganizationDomainResponse,
  V1Beta1DeleteOrganizationInvitationResponse,
  V1Beta1DeleteOrganizationResponse,
  V1Beta1DeleteOrganizationRoleResponse,
  V1Beta1DeletePermissionResponse,
  V1Beta1DeletePolicyResponse,
  V1Beta1DeleteProjectResourceResponse,
  V1Beta1DeleteProjectResponse,
  V1Beta1DeleteRelationResponse,
  V1Beta1DeleteRoleResponse,
  V1Beta1DeleteServiceUserCredentialResponse,
  V1Beta1DeleteServiceUserJWKResponse,
  V1Beta1DeleteServiceUserResponse,
  V1Beta1DeleteServiceUserTokenResponse,
  V1Beta1DeleteUserResponse,
  V1Beta1DeleteWebhookResponse,
  V1Beta1DescribePreferencesResponse,
  V1Beta1DisableBillingAccountResponse,
  V1Beta1DisableGroupResponse,
  V1Beta1DisableOrganizationResponse,
  V1Beta1DisableProjectResponse,
  V1Beta1DisableUserResponse,
  V1Beta1EnableBillingAccountResponse,
  V1Beta1EnableGroupResponse,
  V1Beta1EnableOrganizationResponse,
  V1Beta1EnableProjectResponse,
  V1Beta1EnableUserResponse,
  V1Beta1FeatureRequestBody,
  V1Beta1GetBillingAccountResponse,
  V1Beta1GetBillingBalanceResponse,
  V1Beta1GetCheckoutResponse,
  V1Beta1GetCurrentUserResponse,
  V1Beta1GetFeatureResponse,
  V1Beta1GetGroupResponse,
  V1Beta1GetJWKsResponse,
  V1Beta1GetMetaSchemaResponse,
  V1Beta1GetNamespaceResponse,
  V1Beta1GetOrganizationAuditLogResponse,
  V1Beta1GetOrganizationDomainResponse,
  V1Beta1GetOrganizationInvitationResponse,
  V1Beta1GetOrganizationResponse,
  V1Beta1GetOrganizationRoleResponse,
  V1Beta1GetPermissionResponse,
  V1Beta1GetPlanResponse,
  V1Beta1GetPolicyResponse,
  V1Beta1GetProductResponse,
  V1Beta1GetProjectResourceResponse,
  V1Beta1GetProjectResponse,
  V1Beta1GetRelationResponse,
  V1Beta1GetServiceUserJWKResponse,
  V1Beta1GetServiceUserResponse,
  V1Beta1GetSubscriptionResponse,
  V1Beta1GetUpcomingInvoiceResponse,
  V1Beta1GetUserResponse,
  V1Beta1GroupRequestBody,
  V1Beta1HasTrialedResponse,
  V1Beta1JoinOrganizationResponse,
  V1Beta1ListAllBillingAccountsResponse,
  V1Beta1ListAllInvoicesResponse,
  V1Beta1ListAllOrganizationsResponse,
  V1Beta1ListAllUsersResponse,
  V1Beta1ListAuthStrategiesResponse,
  V1Beta1ListBillingAccountsResponse,
  V1Beta1ListBillingTransactionsResponse,
  V1Beta1ListCheckoutsResponse,
  V1Beta1ListCurrentUserGroupsResponse,
  V1Beta1ListCurrentUserInvitationsResponse,
  V1Beta1ListCurrentUserPreferencesResponse,
  V1Beta1ListFeaturesResponse,
  V1Beta1ListGroupPreferencesResponse,
  V1Beta1ListGroupUsersResponse,
  V1Beta1ListGroupsResponse,
  V1Beta1ListInvoicesResponse,
  V1Beta1ListMetaSchemasResponse,
  V1Beta1ListNamespacesResponse,
  V1Beta1ListOrganizationAdminsResponse,
  V1Beta1ListOrganizationAuditLogsResponse,
  V1Beta1ListOrganizationDomainsResponse,
  V1Beta1ListOrganizationGroupsResponse,
  V1Beta1ListOrganizationInvitationsResponse,
  V1Beta1ListOrganizationPreferencesResponse,
  V1Beta1ListOrganizationProjectsResponse,
  V1Beta1ListOrganizationRolesResponse,
  V1Beta1ListOrganizationServiceUsersResponse,
  V1Beta1ListOrganizationUsersResponse,
  V1Beta1ListOrganizationsByCurrentUserResponse,
  V1Beta1ListOrganizationsByUserResponse,
  V1Beta1ListOrganizationsResponse,
  V1Beta1ListPermissionsResponse,
  V1Beta1ListPlansResponse,
  V1Beta1ListPlatformUsersResponse,
  V1Beta1ListPoliciesResponse,
  V1Beta1ListPreferencesResponse,
  V1Beta1ListProductsResponse,
  V1Beta1ListProjectAdminsResponse,
  V1Beta1ListProjectGroupsResponse,
  V1Beta1ListProjectPreferencesResponse,
  V1Beta1ListProjectResourcesResponse,
  V1Beta1ListProjectServiceUsersResponse,
  V1Beta1ListProjectUsersResponse,
  V1Beta1ListProjectsByCurrentUserResponse,
  V1Beta1ListProjectsByUserResponse,
  V1Beta1ListProjectsResponse,
  V1Beta1ListRelationsResponse,
  V1Beta1ListResourcesResponse,
  V1Beta1ListRolesResponse,
  V1Beta1ListServiceUserCredentialsResponse,
  V1Beta1ListServiceUserJWKsResponse,
  V1Beta1ListServiceUserTokensResponse,
  V1Beta1ListServiceUsersResponse,
  V1Beta1ListSubscriptionsResponse,
  V1Beta1ListUserGroupsResponse,
  V1Beta1ListUserInvitationsResponse,
  V1Beta1ListUserPreferencesResponse,
  V1Beta1ListUsersResponse,
  V1Beta1ListWebhooksResponse,
  V1Beta1MetaSchemaRequestBody,
  V1Beta1OrganizationRequestBody,
  V1Beta1PermissionRequestBody,
  V1Beta1PlanRequestBody,
  V1Beta1PolicyRequestBody,
  V1Beta1PreferenceRequestBody,
  V1Beta1ProductRequestBody,
  V1Beta1ProjectRequestBody,
  V1Beta1RegisterBillingAccountResponse,
  V1Beta1RelationRequestBody,
  V1Beta1RemoveGroupUserResponse,
  V1Beta1RemoveOrganizationUserResponse,
  V1Beta1RemovePlatformUserRequest,
  V1Beta1RemovePlatformUserResponse,
  V1Beta1ResourceRequestBody,
  V1Beta1RevertBillingUsageRequest,
  V1Beta1RevertBillingUsageResponse,
  V1Beta1RoleRequestBody,
  V1Beta1UpdateBillingAccountResponse,
  V1Beta1UpdateCurrentUserResponse,
  V1Beta1UpdateFeatureResponse,
  V1Beta1UpdateGroupResponse,
  V1Beta1UpdateMetaSchemaResponse,
  V1Beta1UpdateOrganizationResponse,
  V1Beta1UpdateOrganizationRoleResponse,
  V1Beta1UpdatePermissionResponse,
  V1Beta1UpdatePlanResponse,
  V1Beta1UpdatePolicyResponse,
  V1Beta1UpdateProductResponse,
  V1Beta1UpdateProjectResourceResponse,
  V1Beta1UpdateProjectResponse,
  V1Beta1UpdateRoleResponse,
  V1Beta1UpdateSubscriptionResponse,
  V1Beta1UpdateUserResponse,
  V1Beta1UpdateWebhookResponse,
  V1Beta1Usage,
  V1Beta1UserRequestBody,
  V1Beta1VerifyOrganizationDomainResponse,
  V1Beta1WebhookRequestBody
} from './data-contracts';
import { ContentType, HttpClient, RequestParams } from './http-client';

export class V1Beta1<SecurityDataType = unknown> extends HttpClient<SecurityDataType> {
  /**
   * @description Lists all the billing accounts from all the organizations in a Frontier instance. It can be filtered by organization.
   *
   * @tags Billing
   * @name AdminServiceListAllBillingAccounts
   * @summary List all billing accounts
   * @request GET:/v1beta1/admin/billing/accounts
   * @secure
   */
  adminServiceListAllBillingAccounts = (
    query?: {
      org_id?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListAllBillingAccountsResponse, RpcStatus>({
      path: `/v1beta1/admin/billing/accounts`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Lists all the invoices from all the organizations in a Frontier instance. It can be filtered by organization.
   *
   * @tags Invoice
   * @name AdminServiceListAllInvoices
   * @summary List all invoices
   * @request GET:/v1beta1/admin/billing/invoices
   * @secure
   */
  adminServiceListAllInvoices = (
    query?: {
      org_id?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListAllInvoicesResponse, RpcStatus>({
      path: `/v1beta1/admin/billing/invoices`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns true if a principal has required permissions to access a resource and false otherwise.<br/> Note the principal can be a user, group or a service account.
   *
   * @tags Authz
   * @name AdminServiceCheckFederatedResourcePermission
   * @summary Check
   * @request POST:/v1beta1/admin/check
   * @secure
   */
  adminServiceCheckFederatedResourcePermission = (
    body: V1Beta1CheckFederatedResourcePermissionRequest,
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CheckFederatedResourcePermissionResponse, RpcStatus>({
      path: `/v1beta1/admin/check`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Lists all the groups from all the organizations in a Frontier instance. It can be filtered by organization and state.
   *
   * @tags Group
   * @name AdminServiceListGroups
   * @summary List all groups
   * @request GET:/v1beta1/admin/groups
   * @secure
   */
  adminServiceListGroups = (
    query?: {
      /** The organization id to filter by. */
      org_id?: string;
      /** The state to filter by. It can be enabled or disabled. */
      state?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListGroupsResponse, RpcStatus>({
      path: `/v1beta1/admin/groups`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Lists all the organizations in a Frontier instance. It can be filtered by user and state.
   *
   * @tags Organization
   * @name AdminServiceListAllOrganizations
   * @summary List all organizations
   * @request GET:/v1beta1/admin/organizations
   * @secure
   */
  adminServiceListAllOrganizations = (
    query?: {
      /** The user id to filter by. */
      user_id?: string;
      /** The state to filter by. It can be enabled or disabled. */
      state?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListAllOrganizationsResponse, RpcStatus>({
      path: `/v1beta1/admin/organizations`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Checkout a product to buy it one time or start a subscription plan on a billing account manually. It bypasses billing engine.
   *
   * @tags Checkout
   * @name AdminServiceDelegatedCheckout
   * @summary Checkout a product or subscription
   * @request POST:/v1beta1/admin/organizations/{org_id}/billing/{billing_id}/checkouts
   * @secure
   */
  adminServiceDelegatedCheckout = (
    orgId: string,
    billingId: string,
    body: {
      /** Subscription to create */
      subscription_body?: V1Beta1CheckoutSubscriptionBody;
      /** Product to buy */
      product_body?: V1Beta1CheckoutProductBody;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1DelegatedCheckoutResponse, RpcStatus>({
      path: `/v1beta1/admin/organizations/${orgId}/billing/${billingId}/checkouts`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Revert billing usage for a billing account.
   *
   * @tags Billing
   * @name AdminServiceRevertBillingUsage
   * @summary Revert billing usage
   * @request POST:/v1beta1/admin/organizations/{org_id}/billing/{billing_id}/usage/{usage_id}/revert
   * @secure
   */
  adminServiceRevertBillingUsage = (
    orgId: string,
    billingId: string,
    usageId: string,
    body: {
      project_id?: string;
      /**
       * amount should be equal or less than the usage amount
       * @format int64
       */
      amount?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1RevertBillingUsageResponse, RpcStatus>({
      path: `/v1beta1/admin/organizations/${orgId}/billing/${billingId}/usage/${usageId}/revert`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Checkout a product to buy it one time or start a subscription plan on a billing account manually. It bypasses billing engine.
   *
   * @tags Checkout
   * @name AdminServiceDelegatedCheckout2
   * @summary Checkout a product or subscription
   * @request POST:/v1beta1/admin/organizations/{org_id}/billing/checkouts
   * @secure
   */
  adminServiceDelegatedCheckout2 = (
    orgId: string,
    body: {
      /** ID of the billing account to update the subscription for */
      billing_id?: string;
      /** Subscription to create */
      subscription_body?: V1Beta1CheckoutSubscriptionBody;
      /** Product to buy */
      product_body?: V1Beta1CheckoutProductBody;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1DelegatedCheckoutResponse, RpcStatus>({
      path: `/v1beta1/admin/organizations/${orgId}/billing/checkouts`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Lists all the users added to the platform.
   *
   * @tags Platform
   * @name AdminServiceListPlatformUsers
   * @summary List platform users
   * @request GET:/v1beta1/admin/platform/users
   * @secure
   */
  adminServiceListPlatformUsers = (params: RequestParams = {}) =>
    this.request<V1Beta1ListPlatformUsersResponse, RpcStatus>({
      path: `/v1beta1/admin/platform/users`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Adds a user to the platform.
   *
   * @tags Platform
   * @name AdminServiceAddPlatformUser
   * @summary Add platform user
   * @request POST:/v1beta1/admin/platform/users
   * @secure
   */
  adminServiceAddPlatformUser = (body: V1Beta1AddPlatformUserRequest, params: RequestParams = {}) =>
    this.request<V1Beta1AddPlatformUserResponse, RpcStatus>({
      path: `/v1beta1/admin/platform/users`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Removes a user from the platform.
   *
   * @tags Platform
   * @name AdminServiceRemovePlatformUser
   * @summary Remove platform user
   * @request POST:/v1beta1/admin/platform/users/remove
   * @secure
   */
  adminServiceRemovePlatformUser = (body: V1Beta1RemovePlatformUserRequest, params: RequestParams = {}) =>
    this.request<V1Beta1RemovePlatformUserResponse, RpcStatus>({
      path: `/v1beta1/admin/platform/users/remove`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Lists all the projects from all the organizations in a Frontier instance. It can be filtered by organization and state.
   *
   * @tags Project
   * @name AdminServiceListProjects
   * @summary List all projects
   * @request GET:/v1beta1/admin/projects
   * @secure
   */
  adminServiceListProjects = (
    query?: {
      /** The organization id to filter by. */
      org_id?: string;
      /** The state to filter by. It can be enabled or disabled. */
      state?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListProjectsResponse, RpcStatus>({
      path: `/v1beta1/admin/projects`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Relation
   * @name AdminServiceListRelations
   * @summary List all relations
   * @request GET:/v1beta1/admin/relations
   * @secure
   */
  adminServiceListRelations = (
    query?: {
      /** The subject to filter by. */
      subject?: string;
      /** The object to filter by. */
      object?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListRelationsResponse, RpcStatus>({
      path: `/v1beta1/admin/relations`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Lists all the resources from all the organizations in a Frontier instance. It can be filtered by user, project, organization and namespace.
   *
   * @tags Resource
   * @name AdminServiceListResources
   * @summary List all resources
   * @request GET:/v1beta1/admin/resources
   * @secure
   */
  adminServiceListResources = (
    query?: {
      /** The user id to filter by. */
      user_id?: string;
      /** The project id to filter by. */
      project_id?: string;
      /** The organization id to filter by. */
      organization_id?: string;
      /** The namespace to filter by. */
      namespace?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListResourcesResponse, RpcStatus>({
      path: `/v1beta1/admin/resources`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Lists all the users from all the organizations in a Frontier instance. It can be filtered by keyword, organization, group and state.
   *
   * @tags User
   * @name AdminServiceListAllUsers
   * @summary List all users
   * @request GET:/v1beta1/admin/users
   * @secure
   */
  adminServiceListAllUsers = (
    query?: {
      /**
       * The maximum number of users to return per page. The default is 50.
       * @format int32
       */
      page_size?: number;
      /**
       * The page number to return. The default is 1.
       * @format int32
       */
      page_num?: number;
      /** The keyword to search for. It can be a user's name, email,metadata or id. */
      keyword?: string;
      /** The organization id to filter by. */
      org_id?: string;
      /** The group id to filter by. */
      group_id?: string;
      /** The state to filter by. It can be enabled or disabled. */
      state?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListAllUsersResponse, RpcStatus>({
      path: `/v1beta1/admin/users`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List all webhooks.
   *
   * @tags Webhook
   * @name AdminServiceListWebhooks
   * @summary List webhooks
   * @request GET:/v1beta1/admin/webhooks
   * @secure
   */
  adminServiceListWebhooks = (params: RequestParams = {}) =>
    this.request<V1Beta1ListWebhooksResponse, RpcStatus>({
      path: `/v1beta1/admin/webhooks`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new webhook.
   *
   * @tags Webhook
   * @name AdminServiceCreateWebhook
   * @summary Create webhook
   * @request POST:/v1beta1/admin/webhooks
   * @secure
   */
  adminServiceCreateWebhook = (body: V1Beta1CreateWebhookRequest, params: RequestParams = {}) =>
    this.request<V1Beta1CreateWebhookResponse, RpcStatus>({
      path: `/v1beta1/admin/webhooks`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a webhook.
   *
   * @tags Webhook
   * @name AdminServiceDeleteWebhook
   * @summary Delete webhook
   * @request DELETE:/v1beta1/admin/webhooks/{id}
   * @secure
   */
  adminServiceDeleteWebhook = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteWebhookResponse, RpcStatus>({
      path: `/v1beta1/admin/webhooks/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Update a webhook.
   *
   * @tags Webhook
   * @name AdminServiceUpdateWebhook
   * @summary Update webhook
   * @request PUT:/v1beta1/admin/webhooks/{id}
   * @secure
   */
  adminServiceUpdateWebhook = (
    id: string,
    body: {
      body?: V1Beta1WebhookRequestBody;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1UpdateWebhookResponse, RpcStatus>({
      path: `/v1beta1/admin/webhooks/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a list of identity providers configured on an instance level in Frontier. e.g Google, AzureAD, Github etc
   *
   * @tags Authn
   * @name FrontierServiceListAuthStrategies
   * @summary List authentication strategies
   * @request GET:/v1beta1/auth
   * @secure
   */
  frontierServiceListAuthStrategies = (params: RequestParams = {}) =>
    this.request<V1Beta1ListAuthStrategiesResponse, RpcStatus>({
      path: `/v1beta1/auth`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Callback from a strategy. This is the endpoint where the strategy will redirect the user after successful authentication. This endpoint will be called with the code and state query parameters. The code will be used to get the access token from the strategy and the state will be used to get the return_to url from the session and redirect the user to that url.
   *
   * @tags Authn
   * @name FrontierServiceAuthCallback
   * @summary Callback from a strategy
   * @request GET:/v1beta1/auth/callback
   * @secure
   */
  frontierServiceAuthCallback = (
    query?: {
      /** strategy_name will not be set for oidc but can be utilized for methods like email magic links */
      strategy_name?: string;
      /** for oidc & magic links */
      state?: string;
      code?: string;
      /**
       * state_options has additional configurations for the authentication flow at hand
       * for example, in case of passkey, it has challenge and public key
       */
      state_options?: object;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1AuthCallbackResponse, RpcStatus>({
      path: `/v1beta1/auth/callback`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Callback from a strategy. This is the endpoint where the strategy will redirect the user after successful authentication. This endpoint will be called with the code and state query parameters. The code will be used to get the access token from the strategy and the state will be used to get the return_to url from the session and redirect the user to that url.
   *
   * @tags Authn
   * @name FrontierServiceAuthCallback2
   * @summary Callback from a strategy
   * @request POST:/v1beta1/auth/callback
   * @secure
   */
  frontierServiceAuthCallback2 = (body: V1Beta1AuthCallbackRequest, params: RequestParams = {}) =>
    this.request<V1Beta1AuthCallbackResponse, RpcStatus>({
      path: `/v1beta1/auth/callback`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Authz
   * @name FrontierServiceGetJwKs
   * @summary Get well known JWKs
   * @request GET:/v1beta1/auth/jwks
   * @secure
   */
  frontierServiceGetJwKs = (params: RequestParams = {}) =>
    this.request<V1Beta1GetJWKsResponse, RpcStatus>({
      path: `/v1beta1/auth/jwks`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Authn
   * @name FrontierServiceAuthLogout
   * @summary Logout from a strategy
   * @request GET:/v1beta1/auth/logout
   * @secure
   */
  frontierServiceAuthLogout = (params: RequestParams = {}) =>
    this.request<V1Beta1AuthLogoutResponse, RpcStatus>({
      path: `/v1beta1/auth/logout`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Authn
   * @name FrontierServiceAuthLogout2
   * @summary Logout from a strategy
   * @request DELETE:/v1beta1/auth/logout
   * @secure
   */
  frontierServiceAuthLogout2 = (params: RequestParams = {}) =>
    this.request<V1Beta1AuthLogoutResponse, RpcStatus>({
      path: `/v1beta1/auth/logout`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Authenticate a user with a strategy. By default, after successful authentication no operation will be performed to apply redirection in case of browsers, provide a url in the authenticate request body that will be used for redirection after authentication. Also set redirect as true for redirecting the user request to the redirect_url after successful authentication.
   *
   * @tags Authn
   * @name FrontierServiceAuthenticate
   * @summary Authenticate with a strategy
   * @request GET:/v1beta1/auth/register/{strategy_name}
   * @secure
   */
  frontierServiceAuthenticate = (
    strategyName: string,
    query?: {
      /**
       * by default, location redirect header for starting authentication flow if applicable
       * will be skipped unless this is set to true, useful in browser, same value will
       * also be returned as endpoint in response anyway
       *
       * If set to true, location header will be set for redirect to start auth flow
       */
      redirect_onstart?: boolean;
      /**
       * by default, after successful authentication(flow completes) no operation will be performed,
       * to apply redirection in case of browsers, provide an url that will be used
       * after authentication where users are sent from frontier.
       * return_to should be one of the allowed urls configured at instance level
       *
       * URL to redirect after successful authentication.<br/> *Example:*`"https://frontier.example.com"`
       */
      return_to?: string;
      /**
       * email of the user for magic links
       *
       * Email of the user to authenticate. Used for magic links.<br/> *Example:*`example@acme.org`
       */
      email?: string;
      /**
       * callback_url will be used by strategy as last step to finish authentication flow
       * in OIDC this host will receive "state" and "code" query params, in case of magic links
       * this will be the url where user is redirected after clicking on magic link.
       * For most cases it could be host of frontier but in case of proxies, this will be proxy public endpoint.
       * callback_url should be one of the allowed urls configured at instance level
       *
       * Host which should handle the call to finish authentication flow, for most cases it could be host of frontier but in case of proxies, this will be proxy public endpoint.<br/> *Example:*`https://frontier.example.com/v1beta1/auth/callback`
       */
      callback_url?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1AuthenticateResponse, RpcStatus>({
      path: `/v1beta1/auth/register/${strategyName}`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Authenticate a user with a strategy. By default, after successful authentication no operation will be performed to apply redirection in case of browsers, provide a url in the authenticate request body that will be used for redirection after authentication. Also set redirect as true for redirecting the user request to the redirect_url after successful authentication.
   *
   * @tags Authn
   * @name FrontierServiceAuthenticate2
   * @summary Authenticate with a strategy
   * @request POST:/v1beta1/auth/register/{strategy_name}
   * @secure
   */
  frontierServiceAuthenticate2 = (
    strategyName: string,
    body: {
      /**
       * by default, location redirect header for starting authentication flow if applicable
       * will be skipped unless this is set to true, useful in browser, same value will
       * also be returned as endpoint in response anyway
       * If set to true, location header will be set for redirect to start auth flow
       */
      redirect_onstart?: boolean;
      /**
       * by default, after successful authentication(flow completes) no operation will be performed,
       * to apply redirection in case of browsers, provide an url that will be used
       * after authentication where users are sent from frontier.
       * return_to should be one of the allowed urls configured at instance level
       * URL to redirect after successful authentication.<br/> *Example:*`"https://frontier.example.com"`
       */
      return_to?: string;
      /**
       * email of the user for magic links
       * Email of the user to authenticate. Used for magic links.<br/> *Example:*`example@acme.org`
       */
      email?: string;
      /**
       * callback_url will be used by strategy as last step to finish authentication flow
       * in OIDC this host will receive "state" and "code" query params, in case of magic links
       * this will be the url where user is redirected after clicking on magic link.
       * For most cases it could be host of frontier but in case of proxies, this will be proxy public endpoint.
       * callback_url should be one of the allowed urls configured at instance level
       * Host which should handle the call to finish authentication flow, for most cases it could be host of frontier but in case of proxies, this will be proxy public endpoint.<br/> *Example:*`https://frontier.example.com/v1beta1/auth/callback`
       */
      callback_url?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1AuthenticateResponse, RpcStatus>({
      path: `/v1beta1/auth/register/${strategyName}`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Access token can be generated by providing the credentials in the request body/header. The credentials can be client id and secret or service account generated key jwt. Use the generated access token in Authorization header to access the frontier resources.
   *
   * @tags Authn
   * @name FrontierServiceAuthToken
   * @summary Generate access token by given credentials
   * @request POST:/v1beta1/auth/token
   * @secure
   */
  frontierServiceAuthToken = (body: V1Beta1AuthTokenRequest, params: RequestParams = {}) =>
    this.request<V1Beta1AuthTokenResponse, RpcStatus>({
      path: `/v1beta1/auth/token`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Returns true if a principal has required permissions to access a resource and false otherwise.<br/> Note the principal can be a user or a service account, and Frontier will the credentials from the current logged in principal from the session cookie (if any), or the client id and secret (in case of service users) or the access token (in case of human user accounts).
   *
   * @tags Authz
   * @name FrontierServiceBatchCheckPermission
   * @summary Batch check
   * @request POST:/v1beta1/batchcheck
   * @secure
   */
  frontierServiceBatchCheckPermission = (body: V1Beta1BatchCheckPermissionRequest, params: RequestParams = {}) =>
    this.request<V1Beta1BatchCheckPermissionResponse, RpcStatus>({
      path: `/v1beta1/batchcheck`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description List all features
   *
   * @tags Feature
   * @name FrontierServiceListFeatures
   * @summary List features
   * @request GET:/v1beta1/billing/features
   * @secure
   */
  frontierServiceListFeatures = (params: RequestParams = {}) =>
    this.request<V1Beta1ListFeaturesResponse, RpcStatus>({
      path: `/v1beta1/billing/features`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new feature for platform.
   *
   * @tags Feature
   * @name FrontierServiceCreateFeature
   * @summary Create feature
   * @request POST:/v1beta1/billing/features
   * @secure
   */
  frontierServiceCreateFeature = (body: V1Beta1CreateFeatureRequest, params: RequestParams = {}) =>
    this.request<V1Beta1CreateFeatureResponse, RpcStatus>({
      path: `/v1beta1/billing/features`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Get a feature by ID.
   *
   * @tags Feature
   * @name FrontierServiceGetFeature
   * @summary Get feature
   * @request GET:/v1beta1/billing/features/{id}
   * @secure
   */
  frontierServiceGetFeature = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetFeatureResponse, RpcStatus>({
      path: `/v1beta1/billing/features/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Update a feature by ID.
   *
   * @tags Feature
   * @name FrontierServiceUpdateFeature
   * @summary Update feature
   * @request PUT:/v1beta1/billing/features/{id}
   * @secure
   */
  frontierServiceUpdateFeature = (
    id: string,
    body: {
      /** Feature to update */
      body?: V1Beta1FeatureRequestBody;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1UpdateFeatureResponse, RpcStatus>({
      path: `/v1beta1/billing/features/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description List all plans.
   *
   * @tags Plan
   * @name FrontierServiceListPlans
   * @summary List plans
   * @request GET:/v1beta1/billing/plans
   * @secure
   */
  frontierServiceListPlans = (params: RequestParams = {}) =>
    this.request<V1Beta1ListPlansResponse, RpcStatus>({
      path: `/v1beta1/billing/plans`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new plan for platform.
   *
   * @tags Plan
   * @name FrontierServiceCreatePlan
   * @summary Create plan
   * @request POST:/v1beta1/billing/plans
   * @secure
   */
  frontierServiceCreatePlan = (body: V1Beta1CreatePlanRequest, params: RequestParams = {}) =>
    this.request<V1Beta1CreatePlanResponse, RpcStatus>({
      path: `/v1beta1/billing/plans`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Get a plan by ID.
   *
   * @tags Plan
   * @name FrontierServiceGetPlan
   * @summary Get plan
   * @request GET:/v1beta1/billing/plans/{id}
   * @secure
   */
  frontierServiceGetPlan = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetPlanResponse, RpcStatus>({
      path: `/v1beta1/billing/plans/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Update a plan by ID.
   *
   * @tags Plan
   * @name FrontierServiceUpdatePlan
   * @summary Update plan
   * @request PUT:/v1beta1/billing/plans/{id}
   * @secure
   */
  frontierServiceUpdatePlan = (
    id: string,
    body: {
      /** Plan to update */
      body?: V1Beta1PlanRequestBody;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1UpdatePlanResponse, RpcStatus>({
      path: `/v1beta1/billing/plans/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description List all products of a platform.
   *
   * @tags Product
   * @name FrontierServiceListProducts
   * @summary List products
   * @request GET:/v1beta1/billing/products
   * @secure
   */
  frontierServiceListProducts = (params: RequestParams = {}) =>
    this.request<V1Beta1ListProductsResponse, RpcStatus>({
      path: `/v1beta1/billing/products`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new product for platform.
   *
   * @tags Product
   * @name FrontierServiceCreateProduct
   * @summary Create product
   * @request POST:/v1beta1/billing/products
   * @secure
   */
  frontierServiceCreateProduct = (body: V1Beta1CreateProductRequest, params: RequestParams = {}) =>
    this.request<V1Beta1CreateProductResponse, RpcStatus>({
      path: `/v1beta1/billing/products`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Get a product by ID.
   *
   * @tags Product
   * @name FrontierServiceGetProduct
   * @summary Get product
   * @request GET:/v1beta1/billing/products/{id}
   * @secure
   */
  frontierServiceGetProduct = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetProductResponse, RpcStatus>({
      path: `/v1beta1/billing/products/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Update a product by ID.
   *
   * @tags Product
   * @name FrontierServiceUpdateProduct
   * @summary Update product
   * @request PUT:/v1beta1/billing/products/{id}
   * @secure
   */
  frontierServiceUpdateProduct = (
    id: string,
    body: {
      /** Product to update */
      body?: V1Beta1ProductRequestBody;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1UpdateProductResponse, RpcStatus>({
      path: `/v1beta1/billing/products/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Accepts a Billing webhook and processes it.
   *
   * @tags Webhook
   * @name FrontierServiceBillingWebhookCallback
   * @summary Accept Billing webhook
   * @request POST:/v1beta1/billing/webhooks/callback/{provider}
   * @secure
   */
  frontierServiceBillingWebhookCallback = (provider: string, body: string, params: RequestParams = {}) =>
    this.request<V1Beta1BillingWebhookCallbackResponse, RpcStatus>({
      path: `/v1beta1/billing/webhooks/callback/${provider}`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Returns true if a principal has required permissions to access a resource and false otherwise.<br/> Note the principal can be a user or a service account. Frontier will extract principal from the current logged in session cookie (if any), or the client id and secret (in case of service users) or the access token.
   *
   * @tags Authz
   * @name FrontierServiceCheckResourcePermission
   * @summary Check
   * @request POST:/v1beta1/check
   * @secure
   */
  frontierServiceCheckResourcePermission = (body: V1Beta1CheckResourcePermissionRequest, params: RequestParams = {}) =>
    this.request<V1Beta1CheckResourcePermissionResponse, RpcStatus>({
      path: `/v1beta1/check`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description List a group preferences by ID.
   *
   * @tags Preference
   * @name FrontierServiceListGroupPreferences
   * @summary List group preferences
   * @request GET:/v1beta1/groups/{id}/preferences
   * @secure
   */
  frontierServiceListGroupPreferences = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListGroupPreferencesResponse, RpcStatus>({
      path: `/v1beta1/groups/${id}/preferences`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new group preferences. The group preferences **name** must be unique within the group and can contain only alphanumeric characters, dashes and underscores.
   *
   * @tags Preference
   * @name FrontierServiceCreateGroupPreferences
   * @summary Create group preferences
   * @request POST:/v1beta1/groups/{id}/preferences
   * @secure
   */
  frontierServiceCreateGroupPreferences = (
    id: string,
    body: {
      bodies?: V1Beta1PreferenceRequestBody[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateGroupPreferencesResponse, RpcStatus>({
      path: `/v1beta1/groups/${id}/preferences`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a list of all metaschemas configured on an instance level in Frontier. e.g user, project, organization etc
   *
   * @tags MetaSchema
   * @name FrontierServiceListMetaSchemas
   * @summary List metaschemas
   * @request GET:/v1beta1/meta/schemas
   * @secure
   */
  frontierServiceListMetaSchemas = (params: RequestParams = {}) =>
    this.request<V1Beta1ListMetaSchemasResponse, RpcStatus>({
      path: `/v1beta1/meta/schemas`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new metadata schema. The metaschema **name** must be unique within the entire Frontier instance and can contain only alphanumeric characters, dashes and underscores. The metaschema **schema** must be a valid JSON schema.Please refer to https://json-schema.org/ to know more about json schema. <br/>*Example:* `{name:"user",schema:{"type":"object","properties":{"label":{"type":"object","additionalProperties":{"type":"string"}},"description":{"type":"string"}}}}`
   *
   * @tags MetaSchema
   * @name FrontierServiceCreateMetaSchema
   * @summary Create metaschema
   * @request POST:/v1beta1/meta/schemas
   * @secure
   */
  frontierServiceCreateMetaSchema = (body: V1Beta1MetaSchemaRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1CreateMetaSchemaResponse, RpcStatus>({
      path: `/v1beta1/meta/schemas`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Get a metadata schema by ID.
   *
   * @tags MetaSchema
   * @name FrontierServiceGetMetaSchema
   * @summary Get metaschema
   * @request GET:/v1beta1/meta/schemas/{id}
   * @secure
   */
  frontierServiceGetMetaSchema = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetMetaSchemaResponse, RpcStatus>({
      path: `/v1beta1/meta/schemas/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a metadata schema permanently. Once deleted the metaschema won't be used to validate the metadata. For example, if a metaschema(with `label` and `description` fields) is used to validate the metadata of a user, then deleting the metaschema will not validate the metadata of the user and metadata field can contain any key-value pair(and say another field called `foo` can be inserted in a user's metadata).
   *
   * @tags MetaSchema
   * @name FrontierServiceDeleteMetaSchema
   * @summary Delete metaschema
   * @request DELETE:/v1beta1/meta/schemas/{id}
   * @secure
   */
  frontierServiceDeleteMetaSchema = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteMetaSchemaResponse, RpcStatus>({
      path: `/v1beta1/meta/schemas/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Update a metadata schema. Only `schema` field of a metaschema can be updated. The metaschema `schema` must be a valid JSON schema.Please refer to https://json-schema.org/ to know more about json schema. <br/>*Example:* `{name:"user",schema:{"type":"object","properties":{"label":{"type":"object","additionalProperties":{"type":"string"}},"description":{"type":"string"}}}}`
   *
   * @tags MetaSchema
   * @name FrontierServiceUpdateMetaSchema
   * @summary Update metaschema
   * @request PUT:/v1beta1/meta/schemas/{id}
   * @secure
   */
  frontierServiceUpdateMetaSchema = (id: string, body: V1Beta1MetaSchemaRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1UpdateMetaSchemaResponse, RpcStatus>({
      path: `/v1beta1/meta/schemas/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns the list of all namespaces in a Frontier instance
   *
   * @tags Namespace
   * @name FrontierServiceListNamespaces
   * @summary Get all namespaces
   * @request GET:/v1beta1/namespaces
   * @secure
   */
  frontierServiceListNamespaces = (params: RequestParams = {}) =>
    this.request<V1Beta1ListNamespacesResponse, RpcStatus>({
      path: `/v1beta1/namespaces`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a namespace by ID
   *
   * @tags Namespace
   * @name FrontierServiceGetNamespace
   * @summary Get namespace
   * @request GET:/v1beta1/namespaces/{id}
   * @secure
   */
  frontierServiceGetNamespace = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetNamespaceResponse, RpcStatus>({
      path: `/v1beta1/namespaces/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a list of organizations. It can be filtered by userID or organization state.
   *
   * @tags Organization
   * @name FrontierServiceListOrganizations
   * @summary List organizations
   * @request GET:/v1beta1/organizations
   * @secure
   */
  frontierServiceListOrganizations = (
    query?: {
      /** The user ID to filter by. It can be used to list all the organizations that the user is a member of. Otherwise, all the organizations in the Frontier instance will be listed. */
      user_id?: string;
      /** The state to filter by. It can be `enabled` or `disabled`. */
      state?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListOrganizationsResponse, RpcStatus>({
      path: `/v1beta1/organizations`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Organization
   * @name FrontierServiceCreateOrganization
   * @summary Create organization
   * @request POST:/v1beta1/organizations
   * @secure
   */
  frontierServiceCreateOrganization = (body: V1Beta1OrganizationRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1CreateOrganizationResponse, RpcStatus>({
      path: `/v1beta1/organizations`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Get organization by ID or name
   *
   * @tags Organization
   * @name FrontierServiceGetOrganization
   * @summary Get organization
   * @request GET:/v1beta1/organizations/{id}
   * @secure
   */
  frontierServiceGetOrganization = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetOrganizationResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete an organization and all of its relations permanently. The organization users will not be deleted from Frontier.
   *
   * @tags Organization
   * @name FrontierServiceDeleteOrganization
   * @summary Delete organization
   * @request DELETE:/v1beta1/organizations/{id}
   * @secure
   */
  frontierServiceDeleteOrganization = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteOrganizationResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Update organization by ID
   *
   * @tags Organization
   * @name FrontierServiceUpdateOrganization
   * @summary Update organization
   * @request PUT:/v1beta1/organizations/{id}
   * @secure
   */
  frontierServiceUpdateOrganization = (id: string, body: V1Beta1OrganizationRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1UpdateOrganizationResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a list admins of an organization
   *
   * @tags Organization
   * @name FrontierServiceListOrganizationAdmins
   * @summary List organization admins
   * @request GET:/v1beta1/organizations/{id}/admins
   * @secure
   */
  frontierServiceListOrganizationAdmins = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListOrganizationAdminsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}/admins`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Sets the state of the organization as disabled. The existing users in the org will not be able to access any organization resources.
   *
   * @tags Organization
   * @name FrontierServiceDisableOrganization
   * @summary Disable organization
   * @request POST:/v1beta1/organizations/{id}/disable
   * @secure
   */
  frontierServiceDisableOrganization = (id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1DisableOrganizationResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}/disable`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Sets the state of the organization as enabled. The existing users in the org will be able to access any organization resources.
   *
   * @tags Organization
   * @name FrontierServiceEnableOrganization
   * @summary Enable organization
   * @request POST:/v1beta1/organizations/{id}/enable
   * @secure
   */
  frontierServiceEnableOrganization = (id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1EnableOrganizationResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}/enable`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List an organization preferences by ID.
   *
   * @tags Preference
   * @name FrontierServiceListOrganizationPreferences
   * @summary List organization preferences
   * @request GET:/v1beta1/organizations/{id}/preferences
   * @secure
   */
  frontierServiceListOrganizationPreferences = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListOrganizationPreferencesResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}/preferences`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new organization preferences. The organization preferences **name** must be unique within the organization and can contain only alphanumeric characters, dashes and underscores.
   *
   * @tags Preference
   * @name FrontierServiceCreateOrganizationPreferences
   * @summary Create organization preferences
   * @request POST:/v1beta1/organizations/{id}/preferences
   * @secure
   */
  frontierServiceCreateOrganizationPreferences = (
    id: string,
    body: {
      bodies?: V1Beta1PreferenceRequestBody[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateOrganizationPreferencesResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}/preferences`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Get all projects that belong to an organization
   *
   * @tags Organization
   * @name FrontierServiceListOrganizationProjects
   * @summary Get organization projects
   * @request GET:/v1beta1/organizations/{id}/projects
   * @secure
   */
  frontierServiceListOrganizationProjects = (
    id: string,
    query?: {
      /** Filter projects by state. If not specified, all projects are returned. <br/> *Possible values:* `enabled` or `disabled` */
      state?: string;
      with_member_count?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListOrganizationProjectsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}/projects`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Organization
   * @name FrontierServiceListOrganizationServiceUsers
   * @summary List organization service users
   * @request GET:/v1beta1/organizations/{id}/serviceusers
   * @secure
   */
  frontierServiceListOrganizationServiceUsers = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListOrganizationServiceUsersResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}/serviceusers`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Organization
   * @name FrontierServiceListOrganizationUsers
   * @summary List organization users
   * @request GET:/v1beta1/organizations/{id}/users
   * @secure
   */
  frontierServiceListOrganizationUsers = (
    id: string,
    query?: {
      permission_filter?: string;
      with_roles?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListOrganizationUsersResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}/users`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Add a user to an organization. A user must exists in Frontier before adding it to an org. This request will fail if the user doesn't exists
   *
   * @tags Organization
   * @name FrontierServiceAddOrganizationUsers
   * @summary Add organization user
   * @request POST:/v1beta1/organizations/{id}/users
   * @secure
   */
  frontierServiceAddOrganizationUsers = (
    id: string,
    body: {
      /** List of user IDs to be added to the organization. */
      user_ids?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1AddOrganizationUsersResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}/users`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Remove a user from an organization
   *
   * @tags Organization
   * @name FrontierServiceRemoveOrganizationUser
   * @summary Remove organization user
   * @request DELETE:/v1beta1/organizations/{id}/users/{user_id}
   * @secure
   */
  frontierServiceRemoveOrganizationUser = (id: string, userId: string, params: RequestParams = {}) =>
    this.request<V1Beta1RemoveOrganizationUserResponse, RpcStatus>({
      path: `/v1beta1/organizations/${id}/users/${userId}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a list of audit logs of an organization in Frontier.
   *
   * @tags AuditLog
   * @name FrontierServiceListOrganizationAuditLogs
   * @summary List audit logs
   * @request GET:/v1beta1/organizations/{org_id}/auditlogs
   * @secure
   */
  frontierServiceListOrganizationAuditLogs = (
    orgId: string,
    query?: {
      source?: string;
      action?: string;
      /**
       * start_time and end_time are inclusive
       * @format date-time
       */
      start_time?: string;
      /** @format date-time */
      end_time?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListOrganizationAuditLogsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/auditlogs`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create new audit logs in a batch.
   *
   * @tags AuditLog
   * @name FrontierServiceCreateOrganizationAuditLogs
   * @summary Create audit log
   * @request POST:/v1beta1/organizations/{org_id}/auditlogs
   * @secure
   */
  frontierServiceCreateOrganizationAuditLogs = (
    orgId: string,
    body: {
      logs?: V1Beta1AuditLog[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateOrganizationAuditLogsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/auditlogs`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Get an audit log by ID.
   *
   * @tags AuditLog
   * @name FrontierServiceGetOrganizationAuditLog
   * @summary Get audit log
   * @request GET:/v1beta1/organizations/{org_id}/auditlogs/{id}
   * @secure
   */
  frontierServiceGetOrganizationAuditLog = (orgId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetOrganizationAuditLogResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/auditlogs/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List billing accounts of an organization.
   *
   * @tags Billing
   * @name FrontierServiceListBillingAccounts
   * @summary List billing accounts
   * @request GET:/v1beta1/organizations/{org_id}/billing
   * @secure
   */
  frontierServiceListBillingAccounts = (
    orgId: string,
    query?: {
      expand?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListBillingAccountsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new billing account for an organization.
   *
   * @tags Billing
   * @name FrontierServiceCreateBillingAccount
   * @summary Create billing account
   * @request POST:/v1beta1/organizations/{org_id}/billing
   * @secure
   */
  frontierServiceCreateBillingAccount = (
    orgId: string,
    body: {
      /** Billing account to create. */
      body?: V1Beta1BillingAccountRequestBody;
      /** Offline billing account don't get registered with billing provider */
      offline?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateBillingAccountResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Check if a billing account is entitled to a feature.
   *
   * @tags Entitlement
   * @name FrontierServiceCheckFeatureEntitlement
   * @summary Check entitlement
   * @request POST:/v1beta1/organizations/{org_id}/billing/{billing_id}/check
   * @secure
   */
  frontierServiceCheckFeatureEntitlement = (
    orgId: string,
    billingId: string,
    body: {
      /**
       * either provide billing_id of the org or API can infer the default
       * billing ID from either org_id or project_id, not both
       */
      project_id?: string;
      /** feature or product name */
      feature?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CheckFeatureEntitlementResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/check`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description List all checkouts of a billing account.
   *
   * @tags Checkout
   * @name FrontierServiceListCheckouts
   * @summary List checkouts
   * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/checkouts
   * @secure
   */
  frontierServiceListCheckouts = (orgId: string, billingId: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListCheckoutsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/checkouts`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Checkout a product to buy it one time or start a subscription plan on a billing account.
   *
   * @tags Checkout
   * @name FrontierServiceCreateCheckout
   * @summary Checkout a product or subscription
   * @request POST:/v1beta1/organizations/{org_id}/billing/{billing_id}/checkouts
   * @secure
   */
  frontierServiceCreateCheckout = (
    orgId: string,
    billingId: string,
    body: {
      success_url?: string;
      cancel_url?: string;
      /** Subscription to create */
      subscription_body?: V1Beta1CheckoutSubscriptionBody;
      /** Product to buy */
      product_body?: V1Beta1CheckoutProductBody;
      /** Payment method setup */
      setup_body?: V1Beta1CheckoutSetupBody;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateCheckoutResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/checkouts`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Get a checkout by ID.
   *
   * @tags Checkout
   * @name FrontierServiceGetCheckout
   * @summary Get checkout
   * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/checkouts/{id}
   * @secure
   */
  frontierServiceGetCheckout = (orgId: string, billingId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetCheckoutResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/checkouts/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List all invoices of a billing account.
   *
   * @tags Invoice
   * @name FrontierServiceListInvoices
   * @summary List invoices
   * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/invoices
   * @secure
   */
  frontierServiceListInvoices = (
    orgId: string,
    billingId: string,
    query?: {
      nonzero_amount_only?: boolean;
      expand?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListInvoicesResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/invoices`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Get the upcoming invoice of a billing account.
   *
   * @tags Invoice
   * @name FrontierServiceGetUpcomingInvoice
   * @summary Get upcoming invoice
   * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/invoices/upcoming
   * @secure
   */
  frontierServiceGetUpcomingInvoice = (orgId: string, billingId: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetUpcomingInvoiceResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/invoices/upcoming`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List subscriptions of a billing account.
   *
   * @tags Subscription
   * @name FrontierServiceListSubscriptions
   * @summary List subscriptions
   * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/subscriptions
   * @secure
   */
  frontierServiceListSubscriptions = (
    orgId: string,
    billingId: string,
    query?: {
      /** Filter subscriptions by state */
      state?: string;
      plan?: string;
      expand?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListSubscriptionsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/subscriptions`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Get a subscription by ID.
   *
   * @tags Subscription
   * @name FrontierServiceGetSubscription
   * @summary Get subscription
   * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/subscriptions/{id}
   * @secure
   */
  frontierServiceGetSubscription = (
    orgId: string,
    billingId: string,
    id: string,
    query?: {
      expand?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1GetSubscriptionResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/subscriptions/${id}`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Update a subscription by ID.
   *
   * @tags Subscription
   * @name FrontierServiceUpdateSubscription
   * @summary Update subscription
   * @request PUT:/v1beta1/organizations/{org_id}/billing/{billing_id}/subscriptions/{id}
   * @secure
   */
  frontierServiceUpdateSubscription = (
    orgId: string,
    billingId: string,
    id: string,
    body: {
      metadata?: object;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1UpdateSubscriptionResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/subscriptions/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Cancel a subscription by ID.
   *
   * @tags Subscription
   * @name FrontierServiceCancelSubscription
   * @summary Cancel subscription
   * @request POST:/v1beta1/organizations/{org_id}/billing/{billing_id}/subscriptions/{id}/cancel
   * @secure
   */
  frontierServiceCancelSubscription = (
    orgId: string,
    billingId: string,
    id: string,
    body: {
      immediate?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CancelSubscriptionResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/subscriptions/${id}/cancel`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Change a subscription plan by ID.
   *
   * @tags Subscription
   * @name FrontierServiceChangeSubscription
   * @summary Change subscription plan
   * @request POST:/v1beta1/organizations/{org_id}/billing/{billing_id}/subscriptions/{id}/change
   * @secure
   */
  frontierServiceChangeSubscription = (
    orgId: string,
    billingId: string,
    id: string,
    body: {
      /**
       * plan to change to
       * deprecated in favor of plan_change
       */
      plan?: string;
      /**
       * should the change be immediate or at the end of the current billing period
       * deprecated in favor of plan_change
       */
      immediate?: boolean;
      plan_change?: ChangeSubscriptionRequestPlanChange;
      phase_change?: ChangeSubscriptionRequestPhaseChange;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ChangeSubscriptionResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/subscriptions/${id}/change`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description List all transactions of a billing account.
   *
   * @tags Transaction
   * @name FrontierServiceListBillingTransactions
   * @summary List billing transactions
   * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/transactions
   * @secure
   */
  frontierServiceListBillingTransactions = (
    orgId: string,
    billingId: string,
    query?: {
      /** @format date-time */
      since?: string;
      /** @format date-time */
      start_range?: string;
      /** @format date-time */
      end_range?: string;
      expand?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListBillingTransactionsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/transactions`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Report a new billing usage for a billing account.
   *
   * @tags Usage
   * @name FrontierServiceCreateBillingUsage
   * @summary Create billing usage
   * @request POST:/v1beta1/organizations/{org_id}/billing/{billing_id}/usages
   * @secure
   */
  frontierServiceCreateBillingUsage = (
    orgId: string,
    billingId: string,
    body: {
      /**
       * either provide billing_id of the org or API can infer the default
       * billing ID from either org_id or project_id, not both
       */
      project_id?: string;
      /** Usage to create */
      usages?: V1Beta1Usage[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateBillingUsageResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${billingId}/usages`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Get a billing account by ID.
   *
   * @tags Billing
   * @name FrontierServiceGetBillingAccount
   * @summary Get billing account
   * @request GET:/v1beta1/organizations/{org_id}/billing/{id}
   * @secure
   */
  frontierServiceGetBillingAccount = (
    orgId: string,
    id: string,
    query?: {
      with_payment_methods?: boolean;
      expand?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1GetBillingAccountResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${id}`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a billing account by ID.
   *
   * @tags Billing
   * @name FrontierServiceDeleteBillingAccount
   * @summary Delete billing account
   * @request DELETE:/v1beta1/organizations/{org_id}/billing/{id}
   * @secure
   */
  frontierServiceDeleteBillingAccount = (orgId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteBillingAccountResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Update a billing account by ID.
   *
   * @tags Billing
   * @name FrontierServiceUpdateBillingAccount
   * @summary Update billing account
   * @request PUT:/v1beta1/organizations/{org_id}/billing/{id}
   * @secure
   */
  frontierServiceUpdateBillingAccount = (
    orgId: string,
    id: string,
    body: {
      /** Billing account to update. */
      body?: V1Beta1BillingAccountRequestBody;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1UpdateBillingAccountResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Get the balance of a billing account by ID.
   *
   * @tags Billing
   * @name FrontierServiceGetBillingBalance
   * @summary Get billing balance
   * @request GET:/v1beta1/organizations/{org_id}/billing/{id}/balance
   * @secure
   */
  frontierServiceGetBillingBalance = (orgId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetBillingBalanceResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${id}/balance`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Disable a billing account by ID. Disabling a billing account doesn't automatically disable it's active subscriptions.
   *
   * @tags Billing
   * @name FrontierServiceDisableBillingAccount
   * @summary Disable billing account
   * @request POST:/v1beta1/organizations/{org_id}/billing/{id}/disable
   * @secure
   */
  frontierServiceDisableBillingAccount = (orgId: string, id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1DisableBillingAccountResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${id}/disable`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Enable a billing account by ID.
   *
   * @tags Billing
   * @name FrontierServiceEnableBillingAccount
   * @summary Enable billing account
   * @request POST:/v1beta1/organizations/{org_id}/billing/{id}/enable
   * @secure
   */
  frontierServiceEnableBillingAccount = (orgId: string, id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1EnableBillingAccountResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${id}/enable`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Check if a billing account has trialed.
   *
   * @tags Billing
   * @name FrontierServiceHasTrialed
   * @summary Has trialed
   * @request GET:/v1beta1/organizations/{org_id}/billing/{id}/plans/{plan_id}/trialed
   * @secure
   */
  frontierServiceHasTrialed = (orgId: string, id: string, planId: string, params: RequestParams = {}) =>
    this.request<V1Beta1HasTrialedResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${id}/plans/${planId}/trialed`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Register a billing account to a provider if it's not already.
   *
   * @tags Billing
   * @name FrontierServiceRegisterBillingAccount
   * @summary Register billing account to provider
   * @request POST:/v1beta1/organizations/{org_id}/billing/{id}/register
   * @secure
   */
  frontierServiceRegisterBillingAccount = (orgId: string, id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1RegisterBillingAccountResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/${id}/register`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List all checkouts of a billing account.
   *
   * @tags Checkout
   * @name FrontierServiceListCheckouts2
   * @summary List checkouts
   * @request GET:/v1beta1/organizations/{org_id}/billing/checkouts
   * @secure
   */
  frontierServiceListCheckouts2 = (
    orgId: string,
    query?: {
      /** ID of the billing account to get the subscriptions for */
      billing_id?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListCheckoutsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/checkouts`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Checkout a product to buy it one time or start a subscription plan on a billing account.
   *
   * @tags Checkout
   * @name FrontierServiceCreateCheckout2
   * @summary Checkout a product or subscription
   * @request POST:/v1beta1/organizations/{org_id}/billing/checkouts
   * @secure
   */
  frontierServiceCreateCheckout2 = (
    orgId: string,
    body: {
      /** ID of the billing account to update the subscription for */
      billing_id?: string;
      success_url?: string;
      cancel_url?: string;
      /** Subscription to create */
      subscription_body?: V1Beta1CheckoutSubscriptionBody;
      /** Product to buy */
      product_body?: V1Beta1CheckoutProductBody;
      /** Payment method setup */
      setup_body?: V1Beta1CheckoutSetupBody;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateCheckoutResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/checkouts`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Get a checkout by ID.
   *
   * @tags Checkout
   * @name FrontierServiceGetCheckout2
   * @summary Get checkout
   * @request GET:/v1beta1/organizations/{org_id}/billing/checkouts/{id}
   * @secure
   */
  frontierServiceGetCheckout2 = (
    orgId: string,
    id: string,
    query?: {
      /** ID of the billing account to get the subscriptions for */
      billing_id?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1GetCheckoutResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/checkouts/${id}`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List all invoices of a billing account.
   *
   * @tags Invoice
   * @name FrontierServiceListInvoices2
   * @summary List invoices
   * @request GET:/v1beta1/organizations/{org_id}/billing/invoices
   * @secure
   */
  frontierServiceListInvoices2 = (
    orgId: string,
    query?: {
      /** ID of the billing account to list invoices for */
      billing_id?: string;
      nonzero_amount_only?: boolean;
      expand?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListInvoicesResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/invoices`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Get the upcoming invoice of a billing account.
   *
   * @tags Invoice
   * @name FrontierServiceGetUpcomingInvoice2
   * @summary Get upcoming invoice
   * @request GET:/v1beta1/organizations/{org_id}/billing/invoices/upcoming
   * @secure
   */
  frontierServiceGetUpcomingInvoice2 = (
    orgId: string,
    query?: {
      /** ID of the billing account to get the upcoming invoice for */
      billing_id?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1GetUpcomingInvoiceResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/invoices/upcoming`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List subscriptions of a billing account.
   *
   * @tags Subscription
   * @name FrontierServiceListSubscriptions2
   * @summary List subscriptions
   * @request GET:/v1beta1/organizations/{org_id}/billing/subscriptions
   * @secure
   */
  frontierServiceListSubscriptions2 = (
    orgId: string,
    query?: {
      /** ID of the billing account to list subscriptions for */
      billing_id?: string;
      /** Filter subscriptions by state */
      state?: string;
      plan?: string;
      expand?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListSubscriptionsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/subscriptions`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Get a subscription by ID.
   *
   * @tags Subscription
   * @name FrontierServiceGetSubscription2
   * @summary Get subscription
   * @request GET:/v1beta1/organizations/{org_id}/billing/subscriptions/{id}
   * @secure
   */
  frontierServiceGetSubscription2 = (
    orgId: string,
    id: string,
    query?: {
      /** ID of the billing account to get the subscription for */
      billing_id?: string;
      expand?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1GetSubscriptionResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/subscriptions/${id}`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Update a subscription by ID.
   *
   * @tags Subscription
   * @name FrontierServiceUpdateSubscription2
   * @summary Update subscription
   * @request POST:/v1beta1/organizations/{org_id}/billing/subscriptions/{id}
   * @secure
   */
  frontierServiceUpdateSubscription2 = (
    orgId: string,
    id: string,
    body: {
      /** ID of the billing account to update the subscription for */
      billing_id?: string;
      metadata?: object;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1UpdateSubscriptionResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/subscriptions/${id}`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Cancel a subscription by ID.
   *
   * @tags Subscription
   * @name FrontierServiceCancelSubscription2
   * @summary Cancel subscription
   * @request POST:/v1beta1/organizations/{org_id}/billing/subscriptions/{id}/cancel
   * @secure
   */
  frontierServiceCancelSubscription2 = (
    orgId: string,
    id: string,
    body: {
      /** ID of the billing account to update the subscription for */
      billing_id?: string;
      immediate?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CancelSubscriptionResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/subscriptions/${id}/cancel`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Change a subscription plan by ID.
   *
   * @tags Subscription
   * @name FrontierServiceChangeSubscription2
   * @summary Change subscription plan
   * @request POST:/v1beta1/organizations/{org_id}/billing/subscriptions/{id}/change
   * @secure
   */
  frontierServiceChangeSubscription2 = (
    orgId: string,
    id: string,
    body: {
      /** ID of the billing account to update the subscription for */
      billing_id?: string;
      /**
       * plan to change to
       * deprecated in favor of plan_change
       */
      plan?: string;
      /**
       * should the change be immediate or at the end of the current billing period
       * deprecated in favor of plan_change
       */
      immediate?: boolean;
      plan_change?: ChangeSubscriptionRequestPlanChange;
      phase_change?: ChangeSubscriptionRequestPhaseChange;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ChangeSubscriptionResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/subscriptions/${id}/change`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description List all transactions of a billing account.
   *
   * @tags Transaction
   * @name FrontierServiceListBillingTransactions2
   * @summary List billing transactions
   * @request GET:/v1beta1/organizations/{org_id}/billing/transactions
   * @secure
   */
  frontierServiceListBillingTransactions2 = (
    orgId: string,
    query?: {
      /** ID of the billing account to update the subscription for */
      billing_id?: string;
      /** @format date-time */
      since?: string;
      /** @format date-time */
      start_range?: string;
      /** @format date-time */
      end_range?: string;
      expand?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListBillingTransactionsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/billing/transactions`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns all domains whitelisted for an organization (both pending and verified if no filters are provided for the state). The verified domains allow users email with the org's whitelisted domain to join the organization without invitation.
   *
   * @tags Organization
   * @name FrontierServiceListOrganizationDomains
   * @summary List org domains
   * @request GET:/v1beta1/organizations/{org_id}/domains
   * @secure
   */
  frontierServiceListOrganizationDomains = (
    orgId: string,
    query?: {
      /** filter to list domains by their state (pending/verified). If not provided, all domains for an org will be listed */
      state?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListOrganizationDomainsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/domains`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Add a domain to an organization which if verified allows all users of the same domain to be signed up to the organization without invitation. This API generates a verification token for a domain which must be added to your domain's DNS provider as a TXT record should be verified with Frontier VerifyOrganizationDomain API before it can be used as an Organization's trusted domain to sign up users.
   *
   * @tags Organization
   * @name FrontierServiceCreateOrganizationDomain
   * @summary Create org domain
   * @request POST:/v1beta1/organizations/{org_id}/domains
   * @secure
   */
  frontierServiceCreateOrganizationDomain = (
    orgId: string,
    body: {
      /** domain name to be added to the trusted domain list */
      domain: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateOrganizationDomainResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/domains`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Get a domain from the list of an organization's whitelisted domains. Returns both verified and unverified domains by their ID
   *
   * @tags Organization
   * @name FrontierServiceGetOrganizationDomain
   * @summary Get org domain
   * @request GET:/v1beta1/organizations/{org_id}/domains/{id}
   * @secure
   */
  frontierServiceGetOrganizationDomain = (orgId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetOrganizationDomainResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/domains/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Remove a domain from the list of an organization's trusted domains list
   *
   * @tags Organization
   * @name FrontierServiceDeleteOrganizationDomain
   * @summary Delete org domain
   * @request DELETE:/v1beta1/organizations/{org_id}/domains/{id}
   * @secure
   */
  frontierServiceDeleteOrganizationDomain = (orgId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteOrganizationDomainResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/domains/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Verify a domain for an organization with a verification token generated by Frontier GenerateDomainVerificationToken API. The token must be added to your domain's DNS provider as a TXT record before it can be verified. This API returns the state of the domain (pending/verified) after verification.
   *
   * @tags Organization
   * @name FrontierServiceVerifyOrganizationDomain
   * @summary Verify org domain
   * @request POST:/v1beta1/organizations/{org_id}/domains/{id}/verify
   * @secure
   */
  frontierServiceVerifyOrganizationDomain = (orgId: string, id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1VerifyOrganizationDomainResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/domains/${id}/verify`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Get all groups that belong to an organization. The results can be filtered by state which can be either be enabled or disabled.
   *
   * @tags Group
   * @name FrontierServiceListOrganizationGroups
   * @summary List organization groups
   * @request GET:/v1beta1/organizations/{org_id}/groups
   * @secure
   */
  frontierServiceListOrganizationGroups = (
    orgId: string,
    query?: {
      /** The state of the group to filter by. It can be enabled or disabled. */
      state?: string;
      group_ids?: string[];
      with_members?: boolean;
      with_member_count?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListOrganizationGroupsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/groups`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new group in an organization which serves as a container for users. The group can be assigned roles and permissions and can be used to manage access to resources. Also a group can also be assigned to other groups.
   *
   * @tags Group
   * @name FrontierServiceCreateGroup
   * @summary Create group
   * @request POST:/v1beta1/organizations/{org_id}/groups
   * @secure
   */
  frontierServiceCreateGroup = (orgId: string, body: V1Beta1GroupRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1CreateGroupResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/groups`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Group
   * @name FrontierServiceGetGroup
   * @summary Get group
   * @request GET:/v1beta1/organizations/{org_id}/groups/{id}
   * @secure
   */
  frontierServiceGetGroup = (
    orgId: string,
    id: string,
    query?: {
      with_members?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1GetGroupResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/groups/${id}`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete an organization group permanently and all of its relations
   *
   * @tags Group
   * @name FrontierServiceDeleteGroup
   * @summary Delete group
   * @request DELETE:/v1beta1/organizations/{org_id}/groups/{id}
   * @secure
   */
  frontierServiceDeleteGroup = (orgId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteGroupResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/groups/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Group
   * @name FrontierServiceUpdateGroup
   * @summary Update group
   * @request PUT:/v1beta1/organizations/{org_id}/groups/{id}
   * @secure
   */
  frontierServiceUpdateGroup = (orgId: string, id: string, body: V1Beta1GroupRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1UpdateGroupResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/groups/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Sets the state of the group as disabled. The group will not be available for access control and the existing users in the group will not be able to access any resources that are assigned to the group. No other users can be added to the group while it is disabled.
   *
   * @tags Group
   * @name FrontierServiceDisableGroup
   * @summary Disable group
   * @request POST:/v1beta1/organizations/{org_id}/groups/{id}/disable
   * @secure
   */
  frontierServiceDisableGroup = (orgId: string, id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1DisableGroupResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/groups/${id}/disable`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Sets the state of the group as enabled. The `enabled` flag is used to determine if the group can be used for access control.
   *
   * @tags Group
   * @name FrontierServiceEnableGroup
   * @summary Enable group
   * @request POST:/v1beta1/organizations/{org_id}/groups/{id}/enable
   * @secure
   */
  frontierServiceEnableGroup = (orgId: string, id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1EnableGroupResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/groups/${id}/enable`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a list of users that belong to a group.
   *
   * @tags Group
   * @name FrontierServiceListGroupUsers
   * @summary List group users
   * @request GET:/v1beta1/organizations/{org_id}/groups/{id}/users
   * @secure
   */
  frontierServiceListGroupUsers = (
    orgId: string,
    id: string,
    query?: {
      with_roles?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListGroupUsersResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/groups/${id}/users`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Adds a principle(user and service-users) to a group, existing users in the group will be ignored. A group can't have nested groups as members.
   *
   * @tags Group
   * @name FrontierServiceAddGroupUsers
   * @summary Add group user
   * @request POST:/v1beta1/organizations/{org_id}/groups/{id}/users
   * @secure
   */
  frontierServiceAddGroupUsers = (
    orgId: string,
    id: string,
    body: {
      user_ids?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1AddGroupUsersResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/groups/${id}/users`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Removes a principle(user and service-users) from a group. If the user is not in the group, the request will succeed but no changes will be made.
   *
   * @tags Group
   * @name FrontierServiceRemoveGroupUser
   * @summary Remove group user
   * @request DELETE:/v1beta1/organizations/{org_id}/groups/{id}/users/{user_id}
   * @secure
   */
  frontierServiceRemoveGroupUser = (orgId: string, id: string, userId: string, params: RequestParams = {}) =>
    this.request<V1Beta1RemoveGroupUserResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/groups/${id}/users/${userId}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns all pending invitations queued for an organization
   *
   * @tags Organization
   * @name FrontierServiceListOrganizationInvitations
   * @summary List pending invitations
   * @request GET:/v1beta1/organizations/{org_id}/invitations
   * @secure
   */
  frontierServiceListOrganizationInvitations = (
    orgId: string,
    query?: {
      /** user_id filter is the email id of user who are invited inside the organization. */
      user_id?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListOrganizationInvitationsResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/invitations`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Invite users to an organization, if user is not registered on the platform, it will be notified. Invitations expire in 7 days
   *
   * @tags Organization
   * @name FrontierServiceCreateOrganizationInvitation
   * @summary Invite user
   * @request POST:/v1beta1/organizations/{org_id}/invitations
   * @secure
   */
  frontierServiceCreateOrganizationInvitation = (
    orgId: string,
    body: {
      /** user_id is email id of user who are invited inside the organization. If user is not registered on the platform, it will be notified */
      user_ids: string[];
      /** list of group ids to which user needs to be added as a member. */
      group_ids?: string[];
      /** list of role ids to which user needs to be added as a member. Roles are binded at organization level by default. */
      role_ids?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateOrganizationInvitationResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/invitations`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a pending invitation queued for an organization
   *
   * @tags Organization
   * @name FrontierServiceGetOrganizationInvitation
   * @summary Get pending invitation
   * @request GET:/v1beta1/organizations/{org_id}/invitations/{id}
   * @secure
   */
  frontierServiceGetOrganizationInvitation = (orgId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetOrganizationInvitationResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/invitations/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a pending invitation queued for an organization
   *
   * @tags Organization
   * @name FrontierServiceDeleteOrganizationInvitation
   * @summary Delete pending invitation
   * @request DELETE:/v1beta1/organizations/{org_id}/invitations/{id}
   * @secure
   */
  frontierServiceDeleteOrganizationInvitation = (orgId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteOrganizationInvitationResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/invitations/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Accept pending invitation queued for an organization. The user will be added to the organization and groups defined in the invitation
   *
   * @tags Organization
   * @name FrontierServiceAcceptOrganizationInvitation
   * @summary Accept pending invitation
   * @request POST:/v1beta1/organizations/{org_id}/invitations/{id}/accept
   * @secure
   */
  frontierServiceAcceptOrganizationInvitation = (orgId: string, id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1AcceptOrganizationInvitationResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/invitations/${id}/accept`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Allows the current logged in user to join the Org if one is not a part of it. The user will only be able to join when the user email's domain matches the organization's whitelisted domains.
   *
   * @tags Organization
   * @name FrontierServiceJoinOrganization
   * @summary Join organization
   * @request POST:/v1beta1/organizations/{org_id}/join
   * @secure
   */
  frontierServiceJoinOrganization = (orgId: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1JoinOrganizationResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/join`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a list of custom roles created under an organization with their associated permissions
   *
   * @tags Role
   * @name FrontierServiceListOrganizationRoles
   * @summary List organization roles
   * @request GET:/v1beta1/organizations/{org_id}/roles
   * @secure
   */
  frontierServiceListOrganizationRoles = (
    orgId: string,
    query?: {
      state?: string;
      scopes?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListOrganizationRolesResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/roles`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a custom role under an organization. This custom role will only be available for assignment to the principles within the organization.
   *
   * @tags Role
   * @name FrontierServiceCreateOrganizationRole
   * @summary Create organization role
   * @request POST:/v1beta1/organizations/{org_id}/roles
   * @secure
   */
  frontierServiceCreateOrganizationRole = (orgId: string, body: V1Beta1RoleRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1CreateOrganizationRoleResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/roles`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a custom role under an organization along with its associated permissions
   *
   * @tags Role
   * @name FrontierServiceGetOrganizationRole
   * @summary Get organization role
   * @request GET:/v1beta1/organizations/{org_id}/roles/{id}
   * @secure
   */
  frontierServiceGetOrganizationRole = (orgId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetOrganizationRoleResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/roles/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a custom role and all of its relations under an organization permanently.
   *
   * @tags Role
   * @name FrontierServiceDeleteOrganizationRole
   * @summary Delete organization role
   * @request DELETE:/v1beta1/organizations/{org_id}/roles/{id}
   * @secure
   */
  frontierServiceDeleteOrganizationRole = (orgId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteOrganizationRoleResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/roles/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Update a custom role under an organization. This custom role will only be available for assignment to the principles within the organization.
   *
   * @tags Role
   * @name FrontierServiceUpdateOrganizationRole
   * @summary Update organization role
   * @request PUT:/v1beta1/organizations/{org_id}/roles/{id}
   * @secure
   */
  frontierServiceUpdateOrganizationRole = (
    orgId: string,
    id: string,
    body: V1Beta1RoleRequestBody,
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1UpdateOrganizationRoleResponse, RpcStatus>({
      path: `/v1beta1/organizations/${orgId}/roles/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Check if a billing account is entitled to a feature.
   *
   * @tags Entitlement
   * @name FrontierServiceCheckFeatureEntitlement2
   * @summary Check entitlement
   * @request POST:/v1beta1/organizations/billing/check
   * @secure
   */
  frontierServiceCheckFeatureEntitlement2 = (body: V1Beta1CheckFeatureEntitlementRequest, params: RequestParams = {}) =>
    this.request<V1Beta1CheckFeatureEntitlementResponse, RpcStatus>({
      path: `/v1beta1/organizations/billing/check`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Report a new billing usage for a billing account.
   *
   * @tags Usage
   * @name FrontierServiceCreateBillingUsage2
   * @summary Create billing usage
   * @request POST:/v1beta1/organizations/billing/usages
   * @secure
   */
  frontierServiceCreateBillingUsage2 = (body: V1Beta1CreateBillingUsageRequest, params: RequestParams = {}) =>
    this.request<V1Beta1CreateBillingUsageResponse, RpcStatus>({
      path: `/v1beta1/organizations/billing/usages`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Revert billing usage for a billing account.
   *
   * @tags Billing
   * @name AdminServiceRevertBillingUsage2
   * @summary Revert billing usage
   * @request POST:/v1beta1/organizations/billing/usages/revert
   * @secure
   */
  adminServiceRevertBillingUsage2 = (body: V1Beta1RevertBillingUsageRequest, params: RequestParams = {}) =>
    this.request<V1Beta1RevertBillingUsageResponse, RpcStatus>({
      path: `/v1beta1/organizations/billing/usages/revert`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Permission
   * @name FrontierServiceListPermissions
   * @summary Get all permissions
   * @request GET:/v1beta1/permissions
   * @secure
   */
  frontierServiceListPermissions = (params: RequestParams = {}) =>
    this.request<V1Beta1ListPermissionsResponse, RpcStatus>({
      path: `/v1beta1/permissions`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Creates a permission. It can be used to grant permissions to all the resources in a Frontier instance.
   *
   * @tags Permission
   * @name AdminServiceCreatePermission
   * @summary Create platform permission
   * @request POST:/v1beta1/permissions
   * @secure
   */
  adminServiceCreatePermission = (body: V1Beta1CreatePermissionRequest, params: RequestParams = {}) =>
    this.request<V1Beta1CreatePermissionResponse, RpcStatus>({
      path: `/v1beta1/permissions`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a permission by ID
   *
   * @tags Permission
   * @name FrontierServiceGetPermission
   * @summary Get permission
   * @request GET:/v1beta1/permissions/{id}
   * @secure
   */
  frontierServiceGetPermission = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetPermissionResponse, RpcStatus>({
      path: `/v1beta1/permissions/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Permission
   * @name AdminServiceDeletePermission
   * @summary Delete platform permission
   * @request DELETE:/v1beta1/permissions/{id}
   * @secure
   */
  adminServiceDeletePermission = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeletePermissionResponse, RpcStatus>({
      path: `/v1beta1/permissions/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Updates a permission by ID. It can be used to grant permissions to all the resources in a Frontier instance.
   *
   * @tags Permission
   * @name AdminServiceUpdatePermission
   * @summary Update platform permission
   * @request PUT:/v1beta1/permissions/{id}
   * @secure
   */
  adminServiceUpdatePermission = (id: string, body: V1Beta1PermissionRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1UpdatePermissionResponse, RpcStatus>({
      path: `/v1beta1/permissions/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Lists all the policies from all the organizations in a Frontier instance. It can be filtered by organization, project, user, role and group.
   *
   * @tags Policy
   * @name FrontierServiceListPolicies
   * @summary List all policies
   * @request GET:/v1beta1/policies
   * @secure
   */
  frontierServiceListPolicies = (
    query?: {
      /** The organization id to filter by. */
      org_id?: string;
      /** The project id to filter by. */
      project_id?: string;
      /** The user id to filter by. */
      user_id?: string;
      /** The role id to filter by. */
      role_id?: string;
      /** The group id to filter by. */
      group_id?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListPoliciesResponse, RpcStatus>({
      path: `/v1beta1/policies`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Creates a policy
   *
   * @tags Policy
   * @name FrontierServiceCreatePolicy
   * @summary Create policy
   * @request POST:/v1beta1/policies
   * @secure
   */
  frontierServiceCreatePolicy = (body: V1Beta1PolicyRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1CreatePolicyResponse, RpcStatus>({
      path: `/v1beta1/policies`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a policy by ID
   *
   * @tags Policy
   * @name FrontierServiceGetPolicy
   * @summary Get policy
   * @request GET:/v1beta1/policies/{id}
   * @secure
   */
  frontierServiceGetPolicy = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetPolicyResponse, RpcStatus>({
      path: `/v1beta1/policies/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a policy all of its relations permanently.
   *
   * @tags Policy
   * @name FrontierServiceDeletePolicy
   * @summary Delete Policy
   * @request DELETE:/v1beta1/policies/{id}
   * @secure
   */
  frontierServiceDeletePolicy = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeletePolicyResponse, RpcStatus>({
      path: `/v1beta1/policies/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Updates a policy by ID
   *
   * @tags Policy
   * @name FrontierServiceUpdatePolicy
   * @summary Update policy
   * @request PUT:/v1beta1/policies/{id}
   * @secure
   */
  frontierServiceUpdatePolicy = (id: string, body: V1Beta1PolicyRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1UpdatePolicyResponse, RpcStatus>({
      path: `/v1beta1/policies/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a policy for a project
   *
   * @tags Policy
   * @name FrontierServiceCreatePolicyForProject
   * @summary Create Policy for Project
   * @request POST:/v1beta1/policies/projects/{project_id}
   * @secure
   */
  frontierServiceCreatePolicyForProject = (
    projectId: string,
    body: V1Beta1CreatePolicyForProjectBody,
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreatePolicyForProjectResponse, RpcStatus>({
      path: `/v1beta1/policies/projects/${projectId}`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a list of all preferences configured on an instance in Frontier. e.g user, project, organization etc
   *
   * @tags Preference
   * @name AdminServiceListPreferences
   * @summary List platform preferences
   * @request GET:/v1beta1/preferences
   * @secure
   */
  adminServiceListPreferences = (params: RequestParams = {}) =>
    this.request<V1Beta1ListPreferencesResponse, RpcStatus>({
      path: `/v1beta1/preferences`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create new platform preferences. The platform preferences **name** must be unique within the platform and can contain only alphanumeric characters, dashes and underscores.
   *
   * @tags Preference
   * @name AdminServiceCreatePreferences
   * @summary Create platform preferences
   * @request POST:/v1beta1/preferences
   * @secure
   */
  adminServiceCreatePreferences = (body: V1Beta1CreatePreferencesRequest, params: RequestParams = {}) =>
    this.request<V1Beta1CreatePreferencesResponse, RpcStatus>({
      path: `/v1beta1/preferences`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a list of all preferences supported by Frontier.
   *
   * @tags Preference
   * @name FrontierServiceDescribePreferences
   * @summary Describe preferences
   * @request GET:/v1beta1/preferences/traits
   * @secure
   */
  frontierServiceDescribePreferences = (params: RequestParams = {}) =>
    this.request<V1Beta1DescribePreferencesResponse, RpcStatus>({
      path: `/v1beta1/preferences/traits`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Project
   * @name FrontierServiceCreateProject
   * @summary Create project
   * @request POST:/v1beta1/projects
   * @secure
   */
  frontierServiceCreateProject = (body: V1Beta1ProjectRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1CreateProjectResponse, RpcStatus>({
      path: `/v1beta1/projects`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a project by ID
   *
   * @tags Project
   * @name FrontierServiceGetProject
   * @summary Get project
   * @request GET:/v1beta1/projects/{id}
   * @secure
   */
  frontierServiceGetProject = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetProjectResponse, RpcStatus>({
      path: `/v1beta1/projects/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a project all of its relations permanently.
   *
   * @tags Project
   * @name FrontierServiceDeleteProject
   * @summary Delete Project
   * @request DELETE:/v1beta1/projects/{id}
   * @secure
   */
  frontierServiceDeleteProject = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteProjectResponse, RpcStatus>({
      path: `/v1beta1/projects/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Updates a project by ID
   *
   * @tags Project
   * @name FrontierServiceUpdateProject
   * @summary Update project
   * @request PUT:/v1beta1/projects/{id}
   * @secure
   */
  frontierServiceUpdateProject = (id: string, body: V1Beta1ProjectRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1UpdateProjectResponse, RpcStatus>({
      path: `/v1beta1/projects/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a collection of admins of a project
   *
   * @tags Project
   * @name FrontierServiceListProjectAdmins
   * @summary List project admins
   * @request GET:/v1beta1/projects/{id}/admins
   * @secure
   */
  frontierServiceListProjectAdmins = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListProjectAdminsResponse, RpcStatus>({
      path: `/v1beta1/projects/${id}/admins`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Project
   * @name FrontierServiceDisableProject
   * @summary Disable project
   * @request POST:/v1beta1/projects/{id}/disable
   * @secure
   */
  frontierServiceDisableProject = (id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1DisableProjectResponse, RpcStatus>({
      path: `/v1beta1/projects/${id}/disable`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Project
   * @name FrontierServiceEnableProject
   * @summary Enable project
   * @request POST:/v1beta1/projects/{id}/enable
   * @secure
   */
  frontierServiceEnableProject = (id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1EnableProjectResponse, RpcStatus>({
      path: `/v1beta1/projects/${id}/enable`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a collection of groups of a project.
   *
   * @tags Project
   * @name FrontierServiceListProjectGroups
   * @summary List project groups
   * @request GET:/v1beta1/projects/{id}/groups
   * @secure
   */
  frontierServiceListProjectGroups = (
    id: string,
    query?: {
      with_roles?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListProjectGroupsResponse, RpcStatus>({
      path: `/v1beta1/projects/${id}/groups`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List a project preferences by ID.
   *
   * @tags Preference
   * @name FrontierServiceListProjectPreferences
   * @summary List project preferences
   * @request GET:/v1beta1/projects/{id}/preferences
   * @secure
   */
  frontierServiceListProjectPreferences = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListProjectPreferencesResponse, RpcStatus>({
      path: `/v1beta1/projects/${id}/preferences`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new project preferences. The project preferences **name** must be unique within the project and can contain only alphanumeric characters, dashes and underscores.
   *
   * @tags Preference
   * @name FrontierServiceCreateProjectPreferences
   * @summary Create project preferences
   * @request POST:/v1beta1/projects/{id}/preferences
   * @secure
   */
  frontierServiceCreateProjectPreferences = (
    id: string,
    body: {
      bodies?: V1Beta1PreferenceRequestBody[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateProjectPreferencesResponse, RpcStatus>({
      path: `/v1beta1/projects/${id}/preferences`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a collection of users of a project.
   *
   * @tags Project
   * @name FrontierServiceListProjectServiceUsers
   * @summary List project serviceusers
   * @request GET:/v1beta1/projects/{id}/serviceusers
   * @secure
   */
  frontierServiceListProjectServiceUsers = (
    id: string,
    query?: {
      with_roles?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListProjectServiceUsersResponse, RpcStatus>({
      path: `/v1beta1/projects/${id}/serviceusers`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a collection of users of a project. Filter by user permissions is supported.
   *
   * @tags Project
   * @name FrontierServiceListProjectUsers
   * @summary List project users
   * @request GET:/v1beta1/projects/{id}/users
   * @secure
   */
  frontierServiceListProjectUsers = (
    id: string,
    query?: {
      permission_filter?: string;
      with_roles?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListProjectUsersResponse, RpcStatus>({
      path: `/v1beta1/projects/${id}/users`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Resource
   * @name FrontierServiceListProjectResources
   * @summary Get all resources
   * @request GET:/v1beta1/projects/{project_id}/resources
   * @secure
   */
  frontierServiceListProjectResources = (
    projectId: string,
    query?: {
      namespace?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListProjectResourcesResponse, RpcStatus>({
      path: `/v1beta1/projects/${projectId}/resources`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Creates a resource in a project
   *
   * @tags Resource
   * @name FrontierServiceCreateProjectResource
   * @summary Create resource
   * @request POST:/v1beta1/projects/{project_id}/resources
   * @secure
   */
  frontierServiceCreateProjectResource = (
    projectId: string,
    body: V1Beta1ResourceRequestBody,
    query?: {
      /** Autogenerated if skipped. */
      id?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateProjectResourceResponse, RpcStatus>({
      path: `/v1beta1/projects/${projectId}/resources`,
      method: 'POST',
      query: query,
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a project resource by ID
   *
   * @tags Resource
   * @name FrontierServiceGetProjectResource
   * @summary Get resource
   * @request GET:/v1beta1/projects/{project_id}/resources/{id}
   * @secure
   */
  frontierServiceGetProjectResource = (projectId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetProjectResourceResponse, RpcStatus>({
      path: `/v1beta1/projects/${projectId}/resources/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Deletes a resource from a project permanently
   *
   * @tags Resource
   * @name FrontierServiceDeleteProjectResource
   * @summary Delete resource
   * @request DELETE:/v1beta1/projects/{project_id}/resources/{id}
   * @secure
   */
  frontierServiceDeleteProjectResource = (projectId: string, id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteProjectResourceResponse, RpcStatus>({
      path: `/v1beta1/projects/${projectId}/resources/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Updates a resource in a project
   *
   * @tags Resource
   * @name FrontierServiceUpdateProjectResource
   * @summary Update resource
   * @request PUT:/v1beta1/projects/{project_id}/resources/{id}
   * @secure
   */
  frontierServiceUpdateProjectResource = (
    projectId: string,
    id: string,
    body: V1Beta1ResourceRequestBody,
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1UpdateProjectResourceResponse, RpcStatus>({
      path: `/v1beta1/projects/${projectId}/resources/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags Relation
   * @name FrontierServiceCreateRelation
   * @summary Create relation
   * @request POST:/v1beta1/relations
   * @secure
   */
  frontierServiceCreateRelation = (body: V1Beta1RelationRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1CreateRelationResponse, RpcStatus>({
      path: `/v1beta1/relations`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a relation by ID
   *
   * @tags Relation
   * @name FrontierServiceGetRelation
   * @summary Get relation
   * @request GET:/v1beta1/relations/{id}
   * @secure
   */
  frontierServiceGetRelation = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetRelationResponse, RpcStatus>({
      path: `/v1beta1/relations/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Remove a subject having a relation from an object
   *
   * @tags Relation
   * @name FrontierServiceDeleteRelation
   * @summary Delete relation
   * @request DELETE:/v1beta1/relations/{relation}/object/{object}/subject/{subject}
   * @secure
   */
  frontierServiceDeleteRelation = (relation: string, object: string, subject: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteRelationResponse, RpcStatus>({
      path: `/v1beta1/relations/${relation}/object/${object}/subject/${subject}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns a list of platform wide roles available in enitre Frontier instance along with their associated permissions
   *
   * @tags Role
   * @name FrontierServiceListRoles
   * @summary List platform roles
   * @request GET:/v1beta1/roles
   * @secure
   */
  frontierServiceListRoles = (
    query?: {
      state?: string;
      scopes?: string[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListRolesResponse, RpcStatus>({
      path: `/v1beta1/roles`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Creates a platform wide role. It can be used to grant permissions to all the resources in a Frontier instance.
   *
   * @tags Role
   * @name AdminServiceCreateRole
   * @summary Create platform role
   * @request POST:/v1beta1/roles
   * @secure
   */
  adminServiceCreateRole = (body: V1Beta1RoleRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1CreateRoleResponse, RpcStatus>({
      path: `/v1beta1/roles`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a platform wide role and all of its relations.
   *
   * @tags Role
   * @name AdminServiceDeleteRole
   * @summary Delete platform role
   * @request DELETE:/v1beta1/roles/{id}
   * @secure
   */
  adminServiceDeleteRole = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteRoleResponse, RpcStatus>({
      path: `/v1beta1/roles/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Update a role title, description and permissions.
   *
   * @tags Role
   * @name AdminServiceUpdateRole
   * @summary Update role
   * @request PUT:/v1beta1/roles/{id}
   * @secure
   */
  adminServiceUpdateRole = (id: string, body: V1Beta1RoleRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1UpdateRoleResponse, RpcStatus>({
      path: `/v1beta1/roles/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns the service user of an organization in a Frontier instance. It can be filter by it's state
   *
   * @tags ServiceUser
   * @name FrontierServiceListServiceUsers
   * @summary List org service users
   * @request GET:/v1beta1/serviceusers
   * @secure
   */
  frontierServiceListServiceUsers = (
    query: {
      /** The organization ID to filter service users by. */
      org_id: string;
      /** The state to filter by. It can be enabled or disabled. */
      state?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListServiceUsersResponse, RpcStatus>({
      path: `/v1beta1/serviceusers`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a service user.
   *
   * @tags ServiceUser
   * @name FrontierServiceCreateServiceUser
   * @summary Create service user
   * @request POST:/v1beta1/serviceusers
   * @secure
   */
  frontierServiceCreateServiceUser = (body: V1Beta1CreateServiceUserRequest, params: RequestParams = {}) =>
    this.request<V1Beta1CreateServiceUserResponse, RpcStatus>({
      path: `/v1beta1/serviceusers`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Get service user details by its id.
   *
   * @tags ServiceUser
   * @name FrontierServiceGetServiceUser
   * @summary Get service user
   * @request GET:/v1beta1/serviceusers/{id}
   * @secure
   */
  frontierServiceGetServiceUser = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetServiceUserResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a service user permanently and all of its relations (keys, organizations, roles, etc)
   *
   * @tags ServiceUser
   * @name FrontierServiceDeleteServiceUser
   * @summary Delete service user
   * @request DELETE:/v1beta1/serviceusers/{id}
   * @secure
   */
  frontierServiceDeleteServiceUser = (
    id: string,
    query?: {
      org_id?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1DeleteServiceUserResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}`,
      method: 'DELETE',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List all the keys of a service user with its details except jwk.
   *
   * @tags ServiceUser
   * @name FrontierServiceListServiceUserJwKs
   * @summary List service user keys
   * @request GET:/v1beta1/serviceusers/{id}/keys
   * @secure
   */
  frontierServiceListServiceUserJwKs = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListServiceUserJWKsResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}/keys`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Generate a service user key and return it, the private key part of the response will not be persisted and should be kept securely by client.
   *
   * @tags ServiceUser
   * @name FrontierServiceCreateServiceUserJwk
   * @summary Create service user public/private key pair
   * @request POST:/v1beta1/serviceusers/{id}/keys
   * @secure
   */
  frontierServiceCreateServiceUserJwk = (
    id: string,
    body: {
      title?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateServiceUserJWKResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}/keys`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Get a service user public RSA JWK set.
   *
   * @tags ServiceUser
   * @name FrontierServiceGetServiceUserJwk
   * @summary Get service user key
   * @request GET:/v1beta1/serviceusers/{id}/keys/{key_id}
   * @secure
   */
  frontierServiceGetServiceUserJwk = (id: string, keyId: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetServiceUserJWKResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}/keys/${keyId}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a service user key permanently.
   *
   * @tags ServiceUser
   * @name FrontierServiceDeleteServiceUserJwk
   * @summary Delete service user key
   * @request DELETE:/v1beta1/serviceusers/{id}/keys/{key_id}
   * @secure
   */
  frontierServiceDeleteServiceUserJwk = (id: string, keyId: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteServiceUserJWKResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}/keys/${keyId}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List all the credentials of a service user.
   *
   * @tags ServiceUser
   * @name FrontierServiceListServiceUserCredentials
   * @summary List service user credentials
   * @request GET:/v1beta1/serviceusers/{id}/secrets
   * @secure
   */
  frontierServiceListServiceUserCredentials = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListServiceUserCredentialsResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}/secrets`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Generate a service user credential and return it. The credential value will not be persisted and should be securely stored by client.
   *
   * @tags ServiceUser
   * @name FrontierServiceCreateServiceUserCredential
   * @summary Create service user client credentials
   * @request POST:/v1beta1/serviceusers/{id}/secrets
   * @secure
   */
  frontierServiceCreateServiceUserCredential = (
    id: string,
    body: {
      title?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateServiceUserCredentialResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}/secrets`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a service user credential.
   *
   * @tags ServiceUser
   * @name FrontierServiceDeleteServiceUserCredential
   * @summary Delete service user credentials
   * @request DELETE:/v1beta1/serviceusers/{id}/secrets/{secret_id}
   * @secure
   */
  frontierServiceDeleteServiceUserCredential = (id: string, secretId: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteServiceUserCredentialResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}/secrets/${secretId}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List all the tokens of a service user.
   *
   * @tags ServiceUser
   * @name FrontierServiceListServiceUserTokens
   * @summary List service user tokens
   * @request GET:/v1beta1/serviceusers/{id}/tokens
   * @secure
   */
  frontierServiceListServiceUserTokens = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListServiceUserTokensResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}/tokens`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Generate a service user token and return it. The token value will not be persisted and should be securely stored by client.
   *
   * @tags ServiceUser
   * @name FrontierServiceCreateServiceUserToken
   * @summary Create service user token
   * @request POST:/v1beta1/serviceusers/{id}/tokens
   * @secure
   */
  frontierServiceCreateServiceUserToken = (
    id: string,
    body: {
      title?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateServiceUserTokenResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}/tokens`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete a service user token.
   *
   * @tags ServiceUser
   * @name FrontierServiceDeleteServiceUserToken
   * @summary Delete service user token
   * @request DELETE:/v1beta1/serviceusers/{id}/tokens/{token_id}
   * @secure
   */
  frontierServiceDeleteServiceUserToken = (id: string, tokenId: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteServiceUserTokenResponse, RpcStatus>({
      path: `/v1beta1/serviceusers/${id}/tokens/${tokenId}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Returns the users from all the organizations in a Frontier instance. It can be filtered by keyword, organization, group and state. Additionally you can include page number and page size for pagination.
   *
   * @tags User
   * @name FrontierServiceListUsers
   * @summary List public users
   * @request GET:/v1beta1/users
   * @secure
   */
  frontierServiceListUsers = (
    query?: {
      /**
       * The maximum number of users to return per page. The default is 50.
       * @format int32
       */
      page_size?: number;
      /**
       * The page number to return. The default is 1.
       * @format int32
       */
      page_num?: number;
      /** The keyword to search for in name or email. */
      keyword?: string;
      /** The organization ID to filter users by. */
      org_id?: string;
      /** The group id to filter by. */
      group_id?: string;
      /** The state to filter by. It can be enabled or disabled. */
      state?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListUsersResponse, RpcStatus>({
      path: `/v1beta1/users`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a user with the given details. A user is not attached to an organization or a group by default,and can be invited to the org/group. The name of the user must be unique within the entire Frontier instance. If a user name is not provided, Frontier automatically generates a name from the user email. The user metadata is validated against the user metaschema. By default the user metaschema contains `labels` and `descriptions` for the user. The `title` field can be optionally added for a user-friendly name. <br/><br/>*Example:*`{"email":"john.doe@raystack.org","title":"John Doe",metadata:{"label": {"key1": "value1"}, "description": "User Description"}}`
   *
   * @tags User
   * @name FrontierServiceCreateUser
   * @summary Create user
   * @request POST:/v1beta1/users
   */
  frontierServiceCreateUser = (body: V1Beta1UserRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1CreateUserResponse, RpcStatus>({
      path: `/v1beta1/users`,
      method: 'POST',
      body: body,
      format: 'json',
      ...params
    });
  /**
   * @description Get a user by id searched over all organizations in Frontier.
   *
   * @tags User
   * @name FrontierServiceGetUser
   * @summary Get user
   * @request GET:/v1beta1/users/{id}
   * @secure
   */
  frontierServiceGetUser = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1GetUserResponse, RpcStatus>({
      path: `/v1beta1/users/${id}`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Delete an user permanently forever and all of its relations (organizations, groups, etc)
   *
   * @tags User
   * @name FrontierServiceDeleteUser
   * @summary Delete user
   * @request DELETE:/v1beta1/users/{id}
   * @secure
   */
  frontierServiceDeleteUser = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1DeleteUserResponse, RpcStatus>({
      path: `/v1beta1/users/${id}`,
      method: 'DELETE',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags User
   * @name FrontierServiceUpdateUser
   * @summary Update user
   * @request PUT:/v1beta1/users/{id}
   * @secure
   */
  frontierServiceUpdateUser = (id: string, body: V1Beta1UserRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1UpdateUserResponse, RpcStatus>({
      path: `/v1beta1/users/${id}`,
      method: 'PUT',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Sets the state of the user as diabled.The user's membership to groups and organizations will still exist along with all it's roles for access control, but the user will not be able to log in and access the Frontier instance.
   *
   * @tags User
   * @name FrontierServiceDisableUser
   * @summary Disable user
   * @request POST:/v1beta1/users/{id}/disable
   * @secure
   */
  frontierServiceDisableUser = (id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1DisableUserResponse, RpcStatus>({
      path: `/v1beta1/users/${id}/disable`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Sets the state of the user as enabled. The user will be able to log in and access the Frontier instance.
   *
   * @tags User
   * @name FrontierServiceEnableUser
   * @summary Enable user
   * @request POST:/v1beta1/users/{id}/enable
   * @secure
   */
  frontierServiceEnableUser = (id: string, body: object, params: RequestParams = {}) =>
    this.request<V1Beta1EnableUserResponse, RpcStatus>({
      path: `/v1beta1/users/${id}/enable`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Lists all the groups a user belongs to across all organization in Frontier. To get the groups of a user in a specific organization, use the org_id filter in the query parameter.
   *
   * @tags User
   * @name FrontierServiceListUserGroups
   * @summary List user groups
   * @request GET:/v1beta1/users/{id}/groups
   * @secure
   */
  frontierServiceListUserGroups = (
    id: string,
    query?: {
      /** The organization ID to filter groups by. If not provided, groups from all organizations are returned. */
      org_id?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListUserGroupsResponse, RpcStatus>({
      path: `/v1beta1/users/${id}/groups`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List all the invitations sent to a user.
   *
   * @tags User
   * @name FrontierServiceListUserInvitations
   * @summary List user invitations
   * @request GET:/v1beta1/users/{id}/invitations
   * @secure
   */
  frontierServiceListUserInvitations = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListUserInvitationsResponse, RpcStatus>({
      path: `/v1beta1/users/${id}/invitations`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description This API returns two list of organizations for the user. i) The list of orgs which the current user is already a part of ii) The list of organizations the user can join directly (based on domain whitelisted and verified by the org). This list will also contain orgs of which user is already a part of. Note: the domain needs to be verified by the org before the it is returned as one of the joinable orgs by domain
   *
   * @tags User
   * @name FrontierServiceListOrganizationsByUser
   * @summary Get user organizations
   * @request GET:/v1beta1/users/{id}/organizations
   * @secure
   */
  frontierServiceListOrganizationsByUser = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListOrganizationsByUserResponse, RpcStatus>({
      path: `/v1beta1/users/${id}/organizations`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List a user preferences by ID.
   *
   * @tags Preference
   * @name FrontierServiceListUserPreferences
   * @summary List user preferences
   * @request GET:/v1beta1/users/{id}/preferences
   * @secure
   */
  frontierServiceListUserPreferences = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListUserPreferencesResponse, RpcStatus>({
      path: `/v1beta1/users/${id}/preferences`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new user preferences. The user preferences **name** must be unique within the user and can contain only alphanumeric characters, dashes and underscores.
   *
   * @tags Preference
   * @name FrontierServiceCreateUserPreferences
   * @summary Create user preferences
   * @request POST:/v1beta1/users/{id}/preferences
   * @secure
   */
  frontierServiceCreateUserPreferences = (
    id: string,
    body: {
      bodies?: V1Beta1PreferenceRequestBody[];
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateUserPreferencesResponse, RpcStatus>({
      path: `/v1beta1/users/${id}/preferences`,
      method: 'POST',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Get all the projects a user belongs to.
   *
   * @tags User
   * @name FrontierServiceListProjectsByUser
   * @summary Get user projects
   * @request GET:/v1beta1/users/{id}/projects
   * @secure
   */
  frontierServiceListProjectsByUser = (id: string, params: RequestParams = {}) =>
    this.request<V1Beta1ListProjectsByUserResponse, RpcStatus>({
      path: `/v1beta1/users/${id}/projects`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags User
   * @name FrontierServiceGetCurrentUser
   * @summary Get current user
   * @request GET:/v1beta1/users/self
   * @secure
   */
  frontierServiceGetCurrentUser = (params: RequestParams = {}) =>
    this.request<V1Beta1GetCurrentUserResponse, RpcStatus>({
      path: `/v1beta1/users/self`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags User
   * @name FrontierServiceUpdateCurrentUser
   * @summary Update current user
   * @request PUT:/v1beta1/users/self
   * @secure
   */
  frontierServiceUpdateCurrentUser = (body: V1Beta1UserRequestBody, params: RequestParams = {}) =>
    this.request<V1Beta1UpdateCurrentUserResponse, RpcStatus>({
      path: `/v1beta1/users/self`,
      method: 'PUT',
      body: body,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * No description
   *
   * @tags User
   * @name FrontierServiceListCurrentUserGroups
   * @summary List my groups
   * @request GET:/v1beta1/users/self/groups
   * @secure
   */
  frontierServiceListCurrentUserGroups = (
    query?: {
      /** org_id is optional filter over an organization */
      org_id?: string;
      with_permissions?: string[];
      with_member_count?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListCurrentUserGroupsResponse, RpcStatus>({
      path: `/v1beta1/users/self/groups`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List all the invitations sent to current user.
   *
   * @tags User
   * @name FrontierServiceListCurrentUserInvitations
   * @summary List user invitations
   * @request GET:/v1beta1/users/self/invitations
   * @secure
   */
  frontierServiceListCurrentUserInvitations = (params: RequestParams = {}) =>
    this.request<V1Beta1ListCurrentUserInvitationsResponse, RpcStatus>({
      path: `/v1beta1/users/self/invitations`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description This API returns two list of organizations for the current logged in user. i) The list of orgs which the current user is already a part of ii) The list of organizations the user can join directly (based on domain whitelisted and verified by the org). This list will also contain orgs of which user is already a part of. Note: the domain needs to be verified by the org before the it is returned as one of the joinable orgs by domain
   *
   * @tags User
   * @name FrontierServiceListOrganizationsByCurrentUser
   * @summary Get my organizations
   * @request GET:/v1beta1/users/self/organizations
   * @secure
   */
  frontierServiceListOrganizationsByCurrentUser = (
    query?: {
      state?: string;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListOrganizationsByCurrentUserResponse, RpcStatus>({
      path: `/v1beta1/users/self/organizations`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description List a user preferences by ID.
   *
   * @tags Preference
   * @name FrontierServiceListCurrentUserPreferences
   * @summary List current user preferences
   * @request GET:/v1beta1/users/self/preferences
   * @secure
   */
  frontierServiceListCurrentUserPreferences = (params: RequestParams = {}) =>
    this.request<V1Beta1ListCurrentUserPreferencesResponse, RpcStatus>({
      path: `/v1beta1/users/self/preferences`,
      method: 'GET',
      secure: true,
      format: 'json',
      ...params
    });
  /**
   * @description Create a new user preferences. The user preferences **name** must be unique within the user and can contain only alphanumeric characters, dashes and underscores.
   *
   * @tags Preference
   * @name FrontierServiceCreateCurrentUserPreferences
   * @summary Create current user preferences
   * @request POST:/v1beta1/users/self/preferences
   * @secure
   */
  frontierServiceCreateCurrentUserPreferences = (
    body: V1Beta1CreateCurrentUserPreferencesRequest,
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1CreateCurrentUserPreferencesResponse, RpcStatus>({
      path: `/v1beta1/users/self/preferences`,
      method: 'POST',
      body: body,
      secure: true,
      type: ContentType.Json,
      format: 'json',
      ...params
    });
  /**
   * @description Get all projects the current user belongs to
   *
   * @tags User
   * @name FrontierServiceListProjectsByCurrentUser
   * @summary Get my projects
   * @request GET:/v1beta1/users/self/projects
   * @secure
   */
  frontierServiceListProjectsByCurrentUser = (
    query?: {
      /** org_id is optional and filter projects by org */
      org_id?: string;
      /**
       * list of permissions needs to be checked against each project
       * query params are set as with_permissions=get&with_permissions=delete
       * to be represented as array
       */
      with_permissions?: string[];
      /**
       * Note: this is a bad design and would recommend against using this filter
       * It is used to list only projects which are explicitly given permission
       * to user. A user could get permission to access a project either via getting
       * access from organization level role or a group. But for some reason we want
       * only users who could have inherited these permissions from top but we only
       * want explictly added ones.
       */
      non_inherited?: boolean;
      with_member_count?: boolean;
    },
    params: RequestParams = {}
  ) =>
    this.request<V1Beta1ListProjectsByCurrentUserResponse, RpcStatus>({
      path: `/v1beta1/users/self/projects`,
      method: 'GET',
      query: query,
      secure: true,
      format: 'json',
      ...params
    });
}
