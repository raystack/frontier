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

export interface BillingAccountAddress {
  line1?: string;
  line2?: string;
  city?: string;
  state?: string;
  postal_code?: string;
  country?: string;
}

export interface BillingAccountBalance {
  /** @format int64 */
  amount?: string;
  currency?: string;
  /** @format date-time */
  updated_at?: string;
}

export interface BillingAccountTax {
  /**
   * tax type like "vat", "gst", "sales_tax" or if it's
   * provider specific us_ein, uk_vat, in_gst, etc
   */
  type?: string;
  /** unique identifier provided by the tax agency */
  id?: string;
}

export interface ChangeSubscriptionRequestPhaseChange {
  /** should the upcoming changes be cancelled */
  cancel_upcoming_changes?: boolean;
}

export interface ChangeSubscriptionRequestPlanChange {
  /** plan to change to */
  plan?: string;
  /** should the change be immediate or at the end of the current billing period */
  immediate?: boolean;
}

export interface ProductBehaviorConfig {
  /** @format int64 */
  credit_amount?: string;
  /** @format int64 */
  seat_limit?: string;
  /** @format int64 */
  min_quantity?: string;
  /** @format int64 */
  max_quantity?: string;
}

export interface SearchOrganizationInvoicesResponseOrganizationInvoice {
  id?: string;
  /** @format int64 */
  amount?: string;
  currency?: string;
  state?: string;
  invoice_link?: string;
  /** @format date-time */
  created_at?: string;
  org_id?: string;
}

export interface SearchOrganizationProjectsResponseOrganizationProject {
  id?: string;
  name?: string;
  title?: string;
  state?: string;
  /** @format int64 */
  member_count?: string;
  user_ids?: string[];
  /** @format date-time */
  created_at?: string;
  organization_id?: string;
}

export interface SearchOrganizationServiceUserCredentialsResponseOrganizationServiceUserCredential {
  title?: string;
  serviceuser_title?: string;
  /** @format date-time */
  created_at?: string;
  org_id?: string;
}

export interface SearchOrganizationTokensResponseOrganizationToken {
  /** @format int64 */
  amount?: string;
  type?: string;
  description?: string;
  user_id?: string;
  user_title?: string;
  user_avatar?: string;
  /** @format date-time */
  created_at?: string;
  org_id?: string;
}

export interface SearchOrganizationUsersResponseOrganizationUser {
  id?: string;
  name?: string;
  title?: string;
  email?: string;
  /** @format date-time */
  org_joined_at?: string;
  state?: string;
  avatar?: string;
  role_names?: string[];
  role_titles?: string[];
  role_ids?: string[];
  organization_id?: string;
}

export interface SearchOrganizationsResponseOrganizationResult {
  id?: string;
  name?: string;
  title?: string;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
  state?: string;
  avatar?: string;
  created_by?: string;
  plan_name?: string;
  payment_mode?: string;
  /** @format date-time */
  subscription_cycle_end_at?: string;
  country?: string;
  subscription_state?: string;
  plan_interval?: string;
  plan_id?: string;
}

export interface SearchProjectUsersResponseProjectUser {
  id?: string;
  name?: string;
  title?: string;
  email?: string;
  avatar?: string;
  role_names?: string[];
  role_titles?: string[];
  role_ids?: string[];
  project_id?: string;
}

export interface SubscriptionPhase {
  /** @format date-time */
  effective_at?: string;
  plan_id?: string;
  reason?: string;
}

export interface WebhookSecret {
  id?: string;
  value?: string;
}

/**
 * Message that represents an arbitrary HTTP body. It should only be used for
 * payload formats that can't be represented as JSON, such as raw binary or
 * an HTML page.
 *
 *
 * This message can be used both in streaming and non-streaming API methods in
 * the request as well as the response.
 *
 * It can be used as a top-level request field, which is convenient if one
 * wants to extract parameters from either the URL or HTTP template into the
 * request fields and also want access to the raw HTTP body.
 *
 * Example:
 *
 *     message GetResourceRequest {
 *       // A unique request id.
 *       string request_id = 1;
 *
 *       // The raw HTTP body is bound to this field.
 *       google.api.HttpBody http_body = 2;
 *
 *     }
 *
 *     service ResourceService {
 *       rpc GetResource(GetResourceRequest)
 *         returns (google.api.HttpBody);
 *       rpc UpdateResource(google.api.HttpBody)
 *         returns (google.protobuf.Empty);
 *
 *     }
 *
 * Example with streaming methods:
 *
 *     service CaldavService {
 *       rpc GetCalendar(stream google.api.HttpBody)
 *         returns (stream google.api.HttpBody);
 *       rpc UpdateCalendar(stream google.api.HttpBody)
 *         returns (stream google.api.HttpBody);
 *
 *     }
 *
 * Use of this type only changes how the request and response bodies are
 * handled, all other features will continue to work unchanged.
 */
export interface ApiHttpBody {
  /** The HTTP Content-Type header value specifying the content type of the body. */
  content_type?: string;
  /**
   * The HTTP request/response body as raw binary.
   * @format byte
   */
  data?: string;
  /**
   * Application specific response metadata. Must be set in the first response
   * for streaming APIs.
   */
  extensions?: ProtobufAny[];
}

export interface GooglerpcStatus {
  /** @format int32 */
  code?: number;
  message?: string;
  details?: ProtobufAny[];
}

/**
 * `Any` contains an arbitrary serialized protocol buffer message along with a
 * URL that describes the type of the serialized message.
 *
 * Protobuf library provides support to pack/unpack Any values in the form
 * of utility functions or additional generated methods of the Any type.
 *
 * Example 1: Pack and unpack a message in C++.
 *
 *     Foo foo = ...;
 *     Any any;
 *     any.PackFrom(foo);
 *     ...
 *     if (any.UnpackTo(&foo)) {
 *       ...
 *     }
 *
 * Example 2: Pack and unpack a message in Java.
 *
 *     Foo foo = ...;
 *     Any any = Any.pack(foo);
 *     ...
 *     if (any.is(Foo.class)) {
 *       foo = any.unpack(Foo.class);
 *     }
 *     // or ...
 *     if (any.isSameTypeAs(Foo.getDefaultInstance())) {
 *       foo = any.unpack(Foo.getDefaultInstance());
 *     }
 *
 *  Example 3: Pack and unpack a message in Python.
 *
 *     foo = Foo(...)
 *     any = Any()
 *     any.Pack(foo)
 *     ...
 *     if any.Is(Foo.DESCRIPTOR):
 *       any.Unpack(foo)
 *       ...
 *
 *  Example 4: Pack and unpack a message in Go
 *
 *      foo := &pb.Foo{...}
 *      any, err := anypb.New(foo)
 *      if err != nil {
 *        ...
 *      }
 *      ...
 *      foo := &pb.Foo{}
 *      if err := any.UnmarshalTo(foo); err != nil {
 *        ...
 *      }
 *
 * The pack methods provided by protobuf library will by default use
 * 'type.googleapis.com/full.type.name' as the type URL and the unpack
 * methods only use the fully qualified type name after the last '/'
 * in the type URL, for example "foo.bar.com/x/y.z" will yield type
 * name "y.z".
 *
 * JSON
 * ====
 * The JSON representation of an `Any` value uses the regular
 * representation of the deserialized, embedded message, with an
 * additional field `@type` which contains the type URL. Example:
 *
 *     package google.profile;
 *     message Person {
 *       string first_name = 1;
 *       string last_name = 2;
 *     }
 *
 *     {
 *       "@type": "type.googleapis.com/google.profile.Person",
 *       "firstName": <string>,
 *       "lastName": <string>
 *     }
 *
 * If the embedded message type is well-known and has a custom JSON
 * representation, that representation will be embedded adding a field
 * `value` which holds the custom JSON in addition to the `@type`
 * field. Example (for message [google.protobuf.Duration][]):
 *
 *     {
 *       "@type": "type.googleapis.com/google.protobuf.Duration",
 *       "value": "1.212s"
 *     }
 */
export interface ProtobufAny {
  /**
   * A URL/resource name that uniquely identifies the type of the serialized
   * protocol buffer message. This string must contain at least
   * one "/" character. The last segment of the URL's path must represent
   * the fully qualified name of the type (as in
   * `path/google.protobuf.Duration`). The name should be in a canonical form
   * (e.g., leading "." is not accepted).
   *
   * In practice, teams usually precompile into the binary all types that they
   * expect it to use in the context of Any. However, for URLs which use the
   * scheme `http`, `https`, or no scheme, one can optionally set up a type
   * server that maps type URLs to message definitions as follows:
   *
   * * If no scheme is provided, `https` is assumed.
   * * An HTTP GET on the URL must yield a [google.protobuf.Type][]
   *   value in binary format, or produce an error.
   * * Applications are allowed to cache lookup results based on the
   *   URL, or have them precompiled into a binary to avoid any
   *   lookup. Therefore, binary compatibility needs to be preserved
   *   on changes to types. (Use versioned type names to manage
   *   breaking changes.)
   *
   * Note: this functionality is not currently available in the official
   * protobuf release, and it is not used for type URLs beginning with
   * type.googleapis.com. As of May 2023, there are no widely used type server
   * implementations and no plans to implement one.
   *
   * Schemes other than `http`, `https` (or the empty scheme) might be
   * used with implementation specific semantics.
   */
  '@type'?: string;
  [key: string]: any;
}

/**
 * `NullValue` is a singleton enumeration to represent the null value for the
 * `Value` type union.
 *
 * The JSON representation for `NullValue` is JSON `null`.
 *
 *  - NULL_VALUE: Null value.
 * @default "NULL_VALUE"
 */
export enum ProtobufNullValue {
  NULL_VALUE = 'NULL_VALUE'
}

export type V1Beta1AcceptOrganizationInvitationResponse = object;

export type V1Beta1AddGroupUsersResponse = object;

export type V1Beta1AddOrganizationUsersResponse = object;

export interface V1Beta1AddPlatformUserRequest {
  /** The user id to add to the platform. */
  user_id?: string;
  /** The service user id to add to the platform. */
  serviceuser_id?: string;
  /** The relation to add as in the platform. It can be admin or member. */
  relation: string;
}

export type V1Beta1AddPlatformUserResponse = object;

export interface V1Beta1AuditLog {
  /** A unique identifier of the audit log if not supplied will be autogenerated */
  id?: string;
  /**
   * Source service generating the event
   * The source service generating the event.
   * @example "auth"
   */
  source: string;
  /** Action performed, e.g. project.create, user.update, serviceuser.delete */
  action: string;
  /** Actor performing the action */
  actor?: V1Beta1AuditLogActor;
  /** Target of the action */
  target?: V1Beta1AuditLogTarget;
  /** Context contains additional information about the event */
  context?: Record<string, string>;
  /**
   * The time the log was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
}

export interface V1Beta1AuditLogActor {
  id?: string;
  type?: string;
  name?: string;
}

export interface V1Beta1AuditLogTarget {
  id?: string;
  type?: string;
  name?: string;
}

export interface V1Beta1AuthCallbackRequest {
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
}

export type V1Beta1AuthCallbackResponse = object;

export type V1Beta1AuthLogoutResponse = object;

export interface V1Beta1AuthStrategy {
  name?: string;
  params?: object;
}

export interface V1Beta1AuthTokenRequest {
  /**
   * grant_type can be one of the following:
   * - client_credentials
   * - urn:ietf:params:oauth:grant-type:jwt-bearer
   */
  grant_type?: string;
  /** client_id and client_secret are required for grant_type client_credentials */
  client_id?: string;
  client_secret?: string;
  /** assertion is required for grant_type urn:ietf:params:oauth:grant-type:jwt-bearer */
  assertion?: string;
}

export interface V1Beta1AuthTokenResponse {
  access_token?: string;
  token_type?: string;
}

export interface V1Beta1AuthenticateResponse {
  /** endpoint specifies the url to redirect user for starting authentication flow */
  endpoint?: string;
  /** state is used for resuming authentication flow in applicable strategies */
  state?: string;
  /** state_options has additional configurations for the authentication flow at hand */
  state_options?: object;
}

export interface V1Beta1BatchCheckPermissionBody {
  /** the permission name to check. <br/> *Example:* `get` or `list` */
  permission: string;
  /** `namespace:uuid` or `namespace:name` of the org or project, and `namespace:urn` of a resource under a project. In case of an org/project either provide the complete namespace (app/organization) or Frontier can also parse aliases for the same as `org` or `project`. <br/> *Example:* `organization:92f69c3a-334b-4f25-90b8-4d4f3be6b825` or `app/project:project-name` or `compute/instance:92f69c3a-334b-4f25-90b8-4d4f3be6b825` */
  resource?: string;
}

export interface V1Beta1BatchCheckPermissionRequest {
  bodies?: V1Beta1BatchCheckPermissionBody[];
}

export interface V1Beta1BatchCheckPermissionResponse {
  pairs?: V1Beta1BatchCheckPermissionResponsePair[];
}

export interface V1Beta1BatchCheckPermissionResponsePair {
  body?: V1Beta1BatchCheckPermissionBody;
  status?: boolean;
}

export interface V1Beta1BillingAccount {
  id?: string;
  org_id?: string;
  name?: string;
  email?: string;
  phone?: string;
  address?: BillingAccountAddress;
  provider_id?: string;
  provider?: string;
  currency?: string;
  state?: string;
  tax_data?: BillingAccountTax[];
  metadata?: object;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
  organization?: V1Beta1Organization;
}

export interface V1Beta1BillingAccountRequestBody {
  name?: string;
  email?: string;
  phone?: string;
  address?: BillingAccountAddress;
  currency?: string;
  tax_data?: BillingAccountTax[];
  metadata?: object;
}

export interface V1Beta1BillingTransaction {
  id?: string;
  customer_id?: string;
  /** additional metadata for storing event/service that triggered this usage */
  source?: string;
  /** @format int64 */
  amount?: string;
  type?: string;
  description?: string;
  /** user_id is the user that triggered this usage */
  user_id?: string;
  metadata?: object;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
  user?: V1Beta1User;
  customer?: V1Beta1BillingAccount;
}

export type V1Beta1BillingWebhookCallbackResponse = object;

export type V1Beta1CancelSubscriptionResponse = object;

export interface V1Beta1ChangeSubscriptionResponse {
  phase?: SubscriptionPhase;
}

export interface V1Beta1CheckFeatureEntitlementRequest {
  org_id?: string;
  /** ID of the billing account to update the subscription for */
  billing_id?: string;
  /**
   * either provide billing_id of the org or API can infer the default
   * billing ID from either org_id or project_id, not both
   */
  project_id?: string;
  /** feature or product name */
  feature?: string;
}

export interface V1Beta1CheckFeatureEntitlementResponse {
  status?: boolean;
}

export interface V1Beta1CheckFederatedResourcePermissionRequest {
  /** the subject to check. <br/> *Example:* `user:..uuidofuser..` */
  subject: string;
  /** `namespace:uuid` or `namespace:name` of the org or project, and `namespace:urn` of a resource under a project. In case of an org/project either provide the complete namespace (app/organization) or Frontier can also parse aliases for the same as `org` or `project`. <br/> *Example:* `organization:92f69c3a-334b-4f25-90b8-4d4f3be6b825` or `app/project:project-name` or `compute/instance:92f69c3a-334b-4f25-90b8-4d4f3be6b825` */
  resource: string;
  /** the permission name to check. <br/> *Example:* `get`, `list`, `compute.instance.create` */
  permission: string;
}

export interface V1Beta1CheckFederatedResourcePermissionResponse {
  status?: boolean;
}

export interface V1Beta1CheckResourcePermissionRequest {
  /** Deprecated. Use `resource` field instead. */
  object_id?: string;
  /** Deprecated. Use `resource` field instead. */
  object_namespace?: string;
  /** the permission name to check. <br/> *Example:* `get`, `list`, `compute.instance.create` */
  permission: string;
  /** `namespace:uuid` or `namespace:name` of the org or project, and `namespace:urn` of a resource under a project. In case of an org/project either provide the complete namespace (app/organization) or Frontier can also parse aliases for the same as `org` or `project`. <br/> *Example:* `organization:92f69c3a-334b-4f25-90b8-4d4f3be6b825` or `app/project:project-name` or `compute/instance:92f69c3a-334b-4f25-90b8-4d4f3be6b825` */
  resource?: string;
}

export interface V1Beta1CheckResourcePermissionResponse {
  status?: boolean;
}

export interface V1Beta1CheckoutProductBody {
  product?: string;
  /** @format int64 */
  quantity?: string;
}

export interface V1Beta1CheckoutSession {
  id?: string;
  checkout_url?: string;
  success_url?: string;
  cancel_url?: string;
  state?: string;
  plan?: string;
  product?: string;
  metadata?: object;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
  /** @format date-time */
  expire_at?: string;
}

export interface V1Beta1CheckoutSetupBody {
  payment_method?: boolean;
  customer_portal?: boolean;
}

export interface V1Beta1CheckoutSubscriptionBody {
  plan?: string;
  skip_trial?: boolean;
  cancel_after_trial?: boolean;
  /**
   * provider_coupon_id is the coupon code that will be applied to the subscription
   * generated by the billing provider, this will be deprecated once coupons are
   * managed by the platform
   */
  provider_coupon_id?: string;
}

export interface V1Beta1CreateBillingAccountResponse {
  /** Created billing account */
  billing_account?: V1Beta1BillingAccount;
}

export interface V1Beta1CreateBillingUsageRequest {
  org_id?: string;
  /** ID of the billing account to update the subscription for */
  billing_id?: string;
  /**
   * either provide billing_id of the org or API can infer the default
   * billing ID from either org_id or project_id, not both
   */
  project_id?: string;
  /** Usage to create */
  usages?: V1Beta1Usage[];
}

export type V1Beta1CreateBillingUsageResponse = object;

export interface V1Beta1CreateCheckoutResponse {
  /** Checkout session */
  checkout_session?: V1Beta1CheckoutSession;
}

export interface V1Beta1CreateCurrentUserPreferencesRequest {
  bodies?: V1Beta1PreferenceRequestBody[];
}

export interface V1Beta1CreateCurrentUserPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1CreateFeatureRequest {
  /** Feature to create */
  body?: V1Beta1FeatureRequestBody;
}

export interface V1Beta1CreateFeatureResponse {
  /** Created feature */
  feature?: V1Beta1Feature;
}

export interface V1Beta1CreateGroupPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1CreateGroupResponse {
  group?: V1Beta1Group;
}

export interface V1Beta1CreateMetaSchemaResponse {
  metaschema?: V1Beta1MetaSchema;
}

export type V1Beta1CreateOrganizationAuditLogsResponse = object;

export interface V1Beta1CreateOrganizationDomainResponse {
  domain?: V1Beta1Domain;
}

export interface V1Beta1CreateOrganizationInvitationResponse {
  invitations?: V1Beta1Invitation[];
}

export interface V1Beta1CreateOrganizationPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1CreateOrganizationResponse {
  organization?: V1Beta1Organization;
}

export interface V1Beta1CreateOrganizationRoleResponse {
  role?: V1Beta1Role;
}

export interface V1Beta1CreatePermissionRequest {
  bodies?: V1Beta1PermissionRequestBody[];
}

export interface V1Beta1CreatePermissionResponse {
  permissions?: V1Beta1Permission[];
}

export interface V1Beta1CreatePlanRequest {
  /** Plan to create */
  body?: V1Beta1PlanRequestBody;
}

export interface V1Beta1CreatePlanResponse {
  /** Created plan */
  plan?: V1Beta1Plan;
}

export interface V1Beta1CreatePolicyForProjectBody {
  /** unique id of the role to which policy is assigned */
  role_id: string;
  /** The title can contain any UTF-8 character, used to provide a human-readable name for the policy. Can also be left empty. <br/> *Example:* `Policy title` */
  title?: string;
  /** principal is the user or group to which policy is assigned. The principal id must be prefixed with its namespace id in this format `namespace:uuid`. The namespace can be `app/user`, `app/group` or `app/serviceuser` (coming up!) and uuid is the unique id of the principal. <br/> *Example:* `app/user:92f69c3a-334b-4f25-90b8-4d4f3be6b825` */
  principal: string;
}

export type V1Beta1CreatePolicyForProjectResponse = object;

export interface V1Beta1CreatePolicyResponse {
  policy?: V1Beta1Policy;
}

export interface V1Beta1CreatePreferencesRequest {
  preferences?: V1Beta1PreferenceRequestBody[];
}

export interface V1Beta1CreatePreferencesResponse {
  preference?: V1Beta1Preference[];
}

export interface V1Beta1CreateProductRequest {
  /** Product to create */
  body?: V1Beta1ProductRequestBody;
}

export interface V1Beta1CreateProductResponse {
  /** Created product */
  product?: V1Beta1Product;
}

export interface V1Beta1CreateProjectPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1CreateProjectResourceResponse {
  resource?: V1Beta1Resource;
}

export interface V1Beta1CreateProjectResponse {
  project?: V1Beta1Project;
}

export interface V1Beta1CreateProspectPublicRequest {
  /** full name of the user */
  name?: string;
  /** email of the user */
  email: string;
  /** phone number of the user */
  phone?: string;
  /** activity for which user is subscribing, e.g. newsletter, waitlist, blogs etc */
  activity: string;
  /** source of this user addition e.g. platform, website, admin etc */
  source?: string;
  /** additional info as key value pair */
  metadata?: object;
}

export type V1Beta1CreateProspectPublicResponse = object;

export interface V1Beta1CreateProspectRequest {
  /** full name of the user */
  name?: string;
  /** email of the user */
  email: string;
  /** phone number of the user */
  phone?: string;
  /** activity for which user is subscribing, e.g. newsletter, waitlist, blogs etc */
  activity: string;
  /** subscription status for this activity. Allowed values - 1 for UNSUBSCRIBED; 2 for SUBSCRIBED; */
  status: V1Beta1ProspectStatus;
  /** source of this user addition e.g. platform, website, admin etc */
  source?: string;
  /** verification status of this user addition e.g. true, false */
  verified?: boolean;
  /** additional info as key value pair */
  metadata?: object;
}

export interface V1Beta1CreateProspectResponse {
  prospect?: V1Beta1Prospect;
}

export interface V1Beta1CreateRelationResponse {
  relation?: V1Beta1Relation;
}

export interface V1Beta1CreateRoleResponse {
  role?: V1Beta1Role;
}

export interface V1Beta1CreateServiceUserCredentialResponse {
  secret?: V1Beta1SecretCredential;
}

export interface V1Beta1CreateServiceUserJWKResponse {
  key?: V1Beta1KeyCredential;
}

export interface V1Beta1CreateServiceUserResponse {
  serviceuser?: V1Beta1ServiceUser;
}

export interface V1Beta1CreateServiceUserTokenResponse {
  token?: V1Beta1ServiceUserToken;
}

export interface V1Beta1CreateUserPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1CreateUserResponse {
  user?: V1Beta1User;
}

export interface V1Beta1CreateWebhookRequest {
  body?: V1Beta1WebhookRequestBody;
}

export interface V1Beta1CreateWebhookResponse {
  webhook?: V1Beta1Webhook;
}

export interface V1Beta1DelegatedCheckoutResponse {
  /** subscription if created */
  subscription?: V1Beta1Subscription;
  /** product if bought */
  product?: V1Beta1Product;
}

export type V1Beta1DeleteBillingAccountResponse = object;

export type V1Beta1DeleteGroupResponse = object;

export type V1Beta1DeleteMetaSchemaResponse = object;

export type V1Beta1DeleteOrganizationDomainResponse = object;

export type V1Beta1DeleteOrganizationInvitationResponse = object;

export type V1Beta1DeleteOrganizationResponse = object;

export type V1Beta1DeleteOrganizationRoleResponse = object;

export type V1Beta1DeletePermissionResponse = object;

export type V1Beta1DeletePolicyResponse = object;

export type V1Beta1DeleteProjectResourceResponse = object;

export type V1Beta1DeleteProjectResponse = object;

export type V1Beta1DeleteProspectResponse = object;

export type V1Beta1DeleteRelationResponse = object;

export type V1Beta1DeleteRoleResponse = object;

export type V1Beta1DeleteServiceUserCredentialResponse = object;

export type V1Beta1DeleteServiceUserJWKResponse = object;

export type V1Beta1DeleteServiceUserResponse = object;

export type V1Beta1DeleteServiceUserTokenResponse = object;

export type V1Beta1DeleteUserResponse = object;

export type V1Beta1DeleteWebhookResponse = object;

export interface V1Beta1DescribePreferencesResponse {
  traits?: V1Beta1PreferenceTrait[];
}

export type V1Beta1DisableBillingAccountResponse = object;

export type V1Beta1DisableGroupResponse = object;

export type V1Beta1DisableOrganizationResponse = object;

export type V1Beta1DisableProjectResponse = object;

export type V1Beta1DisableUserResponse = object;

export interface V1Beta1Domain {
  /**
   * The domain id
   * @example "943e4567-e89b-12d3-a456-426655440000"
   */
  id?: string;
  /**
   * The domain name
   * @example "raystack.org"
   */
  name?: string;
  /**
   * The organization id
   * @example "123e4567-e89b-12d3-a456-426655440000"
   */
  org_id?: string;
  /**
   * The dns TXT record token to verify the domain
   * @example "_frontier-domain-verification=LB6U2lSQgGS55HOy6kpWFqkngRC8TMEjyrakfmYC2D0s+nfy/WkFSg=="
   */
  token?: string;
  /**
   * The domain state either pending or verified
   * @example "pending"
   */
  state?: string;
  /**
   * The time the domain whitelist request was created
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the org domain was updated
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
}

export type V1Beta1EnableBillingAccountResponse = object;

export type V1Beta1EnableGroupResponse = object;

export type V1Beta1EnableOrganizationResponse = object;

export type V1Beta1EnableProjectResponse = object;

export type V1Beta1EnableUserResponse = object;

export interface V1Beta1Feature {
  id?: string;
  /** machine friendly name */
  name?: string;
  product_ids?: string[];
  /** human friendly name */
  title?: string;
  metadata?: object;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
}

export interface V1Beta1FeatureRequestBody {
  /** machine friendly name */
  name?: string;
  /** human friendly title */
  title?: string;
  product_ids?: string[];
  metadata?: object;
}

export type V1Beta1GenerateInvoicesRequest = object;

export type V1Beta1GenerateInvoicesResponse = object;

export interface V1Beta1GetBillingAccountResponse {
  /** Billing account */
  billing_account?: V1Beta1BillingAccount;
  /** List of payment methods */
  payment_methods?: V1Beta1PaymentMethod[];
}

export interface V1Beta1GetBillingBalanceResponse {
  /** Balance of the billing account */
  balance?: BillingAccountBalance;
}

export interface V1Beta1GetCheckoutResponse {
  /** Checkout session */
  checkout_session?: V1Beta1CheckoutSession;
}

export interface V1Beta1GetCurrentUserResponse {
  user?: V1Beta1User;
  serviceuser?: V1Beta1ServiceUser;
}

export interface V1Beta1GetFeatureResponse {
  /** Feature */
  feature?: V1Beta1Feature;
}

export interface V1Beta1GetGroupResponse {
  group?: V1Beta1Group;
}

/** GetJWKsResponse is a valid JSON Web Key Set as specififed in rfc 7517 */
export interface V1Beta1GetJWKsResponse {
  keys?: V1Beta1JSONWebKey[];
}

export interface V1Beta1GetMetaSchemaResponse {
  metaschema?: V1Beta1MetaSchema;
}

export interface V1Beta1GetNamespaceResponse {
  namespace?: V1Beta1Namespace;
}

export interface V1Beta1GetOrganizationAuditLogResponse {
  log?: V1Beta1AuditLog;
}

export interface V1Beta1GetOrganizationDomainResponse {
  domain?: V1Beta1Domain;
}

export interface V1Beta1GetOrganizationInvitationResponse {
  invitation?: V1Beta1Invitation;
}

export interface V1Beta1GetOrganizationKycResponse {
  organization_kyc?: V1Beta1OrganizationKyc;
}

export interface V1Beta1GetOrganizationResponse {
  organization?: V1Beta1Organization;
}

export interface V1Beta1GetOrganizationRoleResponse {
  role?: V1Beta1Role;
}

export interface V1Beta1GetPermissionResponse {
  permission?: V1Beta1Permission;
}

export interface V1Beta1GetPlanResponse {
  /** Plan */
  plan?: V1Beta1Plan;
}

export interface V1Beta1GetPolicyResponse {
  policy?: V1Beta1Policy;
}

export interface V1Beta1GetProductResponse {
  /** Product */
  product?: V1Beta1Product;
}

export interface V1Beta1GetProjectResourceResponse {
  resource?: V1Beta1Resource;
}

export interface V1Beta1GetProjectResponse {
  project?: V1Beta1Project;
}

export interface V1Beta1GetProspectResponse {
  prospect?: V1Beta1Prospect;
}

export interface V1Beta1GetRelationResponse {
  relation?: V1Beta1Relation;
}

export interface V1Beta1GetServiceUserJWKResponse {
  keys?: V1Beta1JSONWebKey[];
}

export interface V1Beta1GetServiceUserResponse {
  serviceuser?: V1Beta1ServiceUser;
}

export interface V1Beta1GetSubscriptionResponse {
  subscription?: V1Beta1Subscription;
}

export interface V1Beta1GetUpcomingInvoiceResponse {
  /** Upcoming invoice */
  invoice?: V1Beta1Invoice;
}

export interface V1Beta1GetUserResponse {
  user?: V1Beta1User;
}

export interface V1Beta1Group {
  id?: string;
  name?: string;
  title?: string;
  org_id?: string;
  metadata?: object;
  /**
   * The time the group was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the group was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
  users?: V1Beta1User[];
  /**
   * The number of members explicitly added in the project.
   * @format int32
   * @example 2
   */
  members_count?: number;
}

export interface V1Beta1GroupRequestBody {
  /** The name of the group. The name must be unique within the entire Frontier instance. The name can contain only alphanumeric characters, dashes and underscores. */
  name: string;
  /** The title can contain any UTF-8 character, used to provide a human-readable name for the group. Can also be left empty. */
  title?: string;
  /** Metadata object for groups that can hold key value pairs defined in Group Metaschema. The metadata object can be used to store arbitrary information about the group such as labels, descriptions etc. The default Group Metaschema contains labels and descripton fields. Update the Group Metaschema to add more fields.<br/>*Example:*`{"labels": {"key": "value"}, "description": "Group description"}` */
  metadata?: object;
}

export interface V1Beta1HasTrialedResponse {
  /** Has the billing account trialed the plan */
  trialed?: boolean;
}

export interface V1Beta1Invitation {
  /**
   * The unique invitation identifier.
   * @example "k9c4f4e2-9b9a-4c1a-8f1a-2b9b9b9b9b9b"
   */
  id?: string;
  /**
   * The user email of the invited user.
   * @example "john.doe@raystack.org"
   */
  user_id?: string;
  /**
   * The organization id to which the user is invited.
   * @example "b9c4f4e2-9b9a-4c1a-8f1a-2b9b9b9b9b9b"
   */
  org_id?: string;
  /**
   * The list of group ids to which the user is invited.
   * @example "c9c4f4e2-9b9a-4c1a-8f1a-2b9b9b9b9b9b"
   */
  group_ids?: string[];
  /**
   * The metadata of the invitation.
   * @example {"key":"value"}
   */
  metadata?: object;
  /**
   * The time when the invitation was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time when the invitation expires.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  expires_at?: string;
  /**
   * The list of role ids to which the user is invited in an organization.
   * @example "d9c4f4e2-9b9a-4c1a-8f1a-2b9b9b9b9b9b"
   */
  role_ids?: string[];
}

export interface V1Beta1Invoice {
  id?: string;
  customer_id?: string;
  provider_id?: string;
  state?: string;
  currency?: string;
  /** @format int64 */
  amount?: string;
  hosted_url?: string;
  /**
   * The date on which payment for this invoice is due
   * @format date-time
   */
  due_date?: string;
  /**
   * The date when this invoice is in effect.
   * @format date-time
   */
  effective_at?: string;
  /** @format date-time */
  period_start_at?: string;
  /** @format date-time */
  period_end_at?: string;
  metadata?: object;
  /** @format date-time */
  created_at?: string;
  customer?: V1Beta1BillingAccount;
}

/** JSON Web Key as specified in RFC 7517 */
export interface V1Beta1JSONWebKey {
  /** Key Type. */
  kty?: string;
  /** Algorithm. */
  alg?: string;
  /** Permitted uses for the public keys. */
  use?: string;
  /** Key ID. */
  kid?: string;
  /** Used for RSA keys. */
  n?: string;
  /** Used for RSA keys. */
  e?: string;
  /** Used for ECDSA keys. */
  x?: string;
  /** Used for ECDSA keys. */
  y?: string;
  /** Used for ECDSA keys. */
  crv?: string;
}

export type V1Beta1JoinOrganizationResponse = object;

export interface V1Beta1KeyCredential {
  type?: string;
  kid?: string;
  principal_id?: string;
  /** RSA private key as string */
  private_key?: string;
}

export interface V1Beta1ListAllBillingAccountsResponse {
  billing_accounts?: V1Beta1BillingAccount[];
}

export interface V1Beta1ListAllInvoicesResponse {
  invoices?: V1Beta1Invoice[];
  /**
   * Total number of records present
   * @format int32
   */
  count?: number;
}

export interface V1Beta1ListAllOrganizationsResponse {
  organizations?: V1Beta1Organization[];
  /**
   * Total number of records present
   * @format int32
   */
  count?: number;
}

export interface V1Beta1ListAllUsersResponse {
  /** @format int32 */
  count?: number;
  users?: V1Beta1User[];
}

export interface V1Beta1ListAuthStrategiesResponse {
  strategies?: V1Beta1AuthStrategy[];
}

export interface V1Beta1ListBillingAccountsResponse {
  /** List of billing accounts */
  billing_accounts?: V1Beta1BillingAccount[];
}

export interface V1Beta1ListBillingTransactionsResponse {
  /** List of transactions */
  transactions?: V1Beta1BillingTransaction[];
}

export interface V1Beta1ListCheckoutsResponse {
  /** List of checkouts */
  checkout_sessions?: V1Beta1CheckoutSession[];
}

export interface V1Beta1ListCurrentUserGroupsResponse {
  groups?: V1Beta1Group[];
  access_pairs?: V1Beta1ListCurrentUserGroupsResponseAccessPair[];
}

export interface V1Beta1ListCurrentUserGroupsResponseAccessPair {
  group_id?: string;
  permissions?: string[];
}

export interface V1Beta1ListCurrentUserInvitationsResponse {
  invitations?: V1Beta1Invitation[];
  orgs?: V1Beta1Organization[];
}

export interface V1Beta1ListCurrentUserPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1ListFeaturesResponse {
  features?: V1Beta1Feature[];
}

export interface V1Beta1ListGroupPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1ListGroupUsersResponse {
  users?: V1Beta1User[];
  role_pairs?: V1Beta1ListGroupUsersResponseRolePair[];
}

export interface V1Beta1ListGroupUsersResponseRolePair {
  user_id?: string;
  roles?: V1Beta1Role[];
}

export interface V1Beta1ListGroupsResponse {
  groups?: V1Beta1Group[];
}

export interface V1Beta1ListInvoicesResponse {
  /** List of invoices */
  invoices?: V1Beta1Invoice[];
}

export interface V1Beta1ListMetaSchemasResponse {
  metaschemas?: V1Beta1MetaSchema[];
}

export interface V1Beta1ListNamespacesResponse {
  namespaces?: V1Beta1Namespace[];
}

export interface V1Beta1ListOrganizationAdminsResponse {
  users?: V1Beta1User[];
}

export interface V1Beta1ListOrganizationAuditLogsResponse {
  logs?: V1Beta1AuditLog[];
}

export interface V1Beta1ListOrganizationDomainsResponse {
  domains?: V1Beta1Domain[];
}

export interface V1Beta1ListOrganizationGroupsResponse {
  groups?: V1Beta1Group[];
}

export interface V1Beta1ListOrganizationInvitationsResponse {
  invitations?: V1Beta1Invitation[];
}

export interface V1Beta1ListOrganizationPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1ListOrganizationProjectsResponse {
  projects?: V1Beta1Project[];
}

export interface V1Beta1ListOrganizationRolesResponse {
  roles?: V1Beta1Role[];
}

export interface V1Beta1ListOrganizationServiceUsersResponse {
  serviceusers?: V1Beta1ServiceUser[];
}

export interface V1Beta1ListOrganizationUsersResponse {
  users?: V1Beta1User[];
  role_pairs?: V1Beta1ListOrganizationUsersResponseRolePair[];
}

export interface V1Beta1ListOrganizationUsersResponseRolePair {
  user_id?: string;
  roles?: V1Beta1Role[];
}

export interface V1Beta1ListOrganizationsByCurrentUserResponse {
  organizations?: V1Beta1Organization[];
  joinable_via_domain?: V1Beta1Organization[];
}

export interface V1Beta1ListOrganizationsByUserResponse {
  organizations?: V1Beta1Organization[];
  joinable_via_domain?: V1Beta1Organization[];
}

export interface V1Beta1ListOrganizationsKycResponse {
  organizations_kyc?: V1Beta1OrganizationKyc[];
}

export interface V1Beta1ListOrganizationsResponse {
  organizations?: V1Beta1Organization[];
}

export interface V1Beta1ListPermissionsResponse {
  permissions?: V1Beta1Permission[];
}

export interface V1Beta1ListPlansResponse {
  /** List of plans */
  plans?: V1Beta1Plan[];
}

export interface V1Beta1ListPlatformUsersResponse {
  users?: V1Beta1User[];
  serviceusers?: V1Beta1ServiceUser[];
}

export interface V1Beta1ListPoliciesResponse {
  policies?: V1Beta1Policy[];
}

export interface V1Beta1ListPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1ListProductsResponse {
  /** List of products */
  products?: V1Beta1Product[];
}

export interface V1Beta1ListProjectAdminsResponse {
  users?: V1Beta1User[];
}

export interface V1Beta1ListProjectGroupsResponse {
  groups?: V1Beta1Group[];
  role_pairs?: V1Beta1ListProjectGroupsResponseRolePair[];
}

export interface V1Beta1ListProjectGroupsResponseRolePair {
  group_id?: string;
  roles?: V1Beta1Role[];
}

export interface V1Beta1ListProjectPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1ListProjectResourcesResponse {
  resources?: V1Beta1Resource[];
}

export interface V1Beta1ListProjectServiceUsersResponse {
  serviceusers?: V1Beta1ServiceUser[];
  role_pairs?: V1Beta1ListProjectServiceUsersResponseRolePair[];
}

export interface V1Beta1ListProjectServiceUsersResponseRolePair {
  serviceuser_id?: string;
  roles?: V1Beta1Role[];
}

export interface V1Beta1ListProjectUsersResponse {
  users?: V1Beta1User[];
  role_pairs?: V1Beta1ListProjectUsersResponseRolePair[];
}

export interface V1Beta1ListProjectUsersResponseRolePair {
  user_id?: string;
  roles?: V1Beta1Role[];
}

export interface V1Beta1ListProjectsByCurrentUserResponse {
  projects?: V1Beta1Project[];
  access_pairs?: V1Beta1ListProjectsByCurrentUserResponseAccessPair[];
  /** @format int32 */
  count?: number;
}

export interface V1Beta1ListProjectsByCurrentUserResponseAccessPair {
  project_id?: string;
  permissions?: string[];
}

export interface V1Beta1ListProjectsByUserResponse {
  projects?: V1Beta1Project[];
}

export interface V1Beta1ListProjectsResponse {
  projects?: V1Beta1Project[];
}

export interface V1Beta1ListProspectsResponse {
  prospects?: V1Beta1Prospect[];
  pagination?: V1Beta1RQLQueryPaginationResponse;
  group?: V1Beta1RQLQueryGroupResponse;
}

export interface V1Beta1ListRelationsResponse {
  relations?: V1Beta1Relation[];
}

export interface V1Beta1ListResourcesResponse {
  resources?: V1Beta1Resource[];
}

export interface V1Beta1ListRolesResponse {
  roles?: V1Beta1Role[];
}

export interface V1Beta1ListServiceUserCredentialsResponse {
  /** secrets will be listed without the secret value */
  secrets?: V1Beta1SecretCredential[];
}

export interface V1Beta1ListServiceUserJWKsResponse {
  keys?: V1Beta1ServiceUserJWK[];
}

export interface V1Beta1ListServiceUserProjectsResponse {
  projects?: V1Beta1Project[];
  access_pairs?: V1Beta1ListServiceUserProjectsResponseAccessPair[];
}

export interface V1Beta1ListServiceUserProjectsResponseAccessPair {
  project_id?: string;
  permissions?: string[];
}

export interface V1Beta1ListServiceUserTokensResponse {
  tokens?: V1Beta1ServiceUserToken[];
}

export interface V1Beta1ListServiceUsersResponse {
  serviceusers?: V1Beta1ServiceUser[];
}

export interface V1Beta1ListSubscriptionsResponse {
  /** List of subscriptions */
  subscriptions?: V1Beta1Subscription[];
}

export interface V1Beta1ListUserGroupsResponse {
  groups?: V1Beta1Group[];
}

export interface V1Beta1ListUserInvitationsResponse {
  invitations?: V1Beta1Invitation[];
}

export interface V1Beta1ListUserPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1ListUsersResponse {
  /** @format int32 */
  count?: number;
  users?: V1Beta1User[];
}

export interface V1Beta1ListWebhooksResponse {
  webhooks?: V1Beta1Webhook[];
}

export interface V1Beta1MetaSchema {
  /**
   * The unique metaschema uuid.
   * @example "a9c4f4e2-9b9a-4c1a-8f1a-2b9b9b9b9b9b"
   */
  id?: string;
  name?: string;
  /**
   * The metaschema json schema.
   * @example {"type":"object"}
   */
  schema?: string;
  /**
   * The time when the metaschema was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time when the metaschema was updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
}

export interface V1Beta1MetaSchemaRequestBody {
  /** The name of the metaschema. The name must be unique within the entire Frontier instance. The name can contain only alphanumeric characters, dashes and underscores. */
  name: string;
  /** The schema of the metaschema. The schema must be a valid JSON schema.Please refer to https://json-schema.org/ to know more about json schema. */
  schema: string;
}

export interface V1Beta1Namespace {
  id?: string;
  /** name could be in a format like: app/organization */
  name?: string;
  metadata?: object;
  /**
   * The time the namespace was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the namespace was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
}

export interface V1Beta1Organization {
  id?: string;
  name?: string;
  title?: string;
  metadata?: object;
  /**
   * The time the organization was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the organization was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
  /**
   * The state of the organization (enabled or disabled).
   * @example "enabled"
   */
  state?: string;
  /** The base64 encoded image string of the organization avatar. Should be less than 2MB. */
  avatar?: string;
}

export interface V1Beta1OrganizationKyc {
  org_id?: string;
  status?: boolean;
  link?: string;
  /**
   * The time the organization kyc was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the organization kyc was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
}

export interface V1Beta1OrganizationRequestBody {
  /** The name of the organization. The name must be unique within the entire Frontier instance. The name can contain only alphanumeric characters, dashes and underscores.<br/>*Example:*`"frontier-org1-acme"` */
  name: string;
  /** The title can contain any UTF-8 character, used to provide a human-readable name for the organization. Can also be left empty.<br/> *Example*: `"Acme Inc"` */
  title?: string;
  /** Metadata object for organizations that can hold key value pairs defined in Organization Metaschema. The metadata object can be used to store arbitrary information about the organization such as labels, descriptions etc. The default Organization Metaschema contains labels and descripton fields. Update the Organization Metaschema to add more fields. <br/>*Example*:`{"labels": {"key": "value"}, "description": "Organization description"}` */
  metadata?: object;
  /**
   * The avatar is base64 encoded image data of the user. Can also be left empty. The image should be less than 200KB. Should follow the regex pattern `^data:image/(png|jpg|jpeg|gif);base64,([a-zA-Z0-9+/]+={0,2})+$`.
   * @example "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAA"
   */
  avatar?: string;
}

export interface V1Beta1PaymentMethod {
  id?: string;
  customer_id?: string;
  provider_id?: string;
  type?: string;
  card_brand?: string;
  card_last4?: string;
  /** @format int64 */
  card_expiry_month?: string;
  /** @format int64 */
  card_expiry_year?: string;
  metadata?: object;
  /** @format date-time */
  created_at?: string;
}

export interface V1Beta1Permission {
  id?: string;
  name?: string;
  title?: string;
  /**
   * The time the permission was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the permission was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
  namespace?: string;
  metadata?: object;
  /**
   * Permission path key is composed of three parts, 'service.resource.verb'. Where 'service.resource' works as a namespace for the 'verb'.
   * @example "compute.instance.get"
   */
  key?: string;
}

export interface V1Beta1PermissionRequestBody {
  /** The name of the permission. It should be unique across a Frontier instance and can contain only alphanumeric characters. */
  name?: string;
  /**
   * namespace should be in service/resource format
   * The namespace of the permission. The namespace should be in service/resource format.<br/>*Example:*`compute/guardian`
   */
  namespace?: string;
  /** The metadata object for permissions that can hold key value pairs. */
  metadata?: object;
  /** The title can contain any UTF-8 character, used to provide a human-readable name for the permissions. Can also be left empty. */
  title?: string;
  /**
   * key is composed of three parts, 'service.resource.verb'. Where 'service.resource' works as a namespace for the 'verb'.
   * Use this instead of using name and namespace fields
   * Permission path key is composed of three parts, 'service.resource.verb'. Where 'service.resource' works as a namespace for the 'verb'. Namespace name cannot be `app` as it's reserved for core permissions.
   * @example "compute.instance.get"
   */
  key?: string;
}

export interface V1Beta1Plan {
  id?: string;
  name?: string;
  title?: string;
  description?: string;
  products?: V1Beta1Product[];
  /** known intervals are "day", "week", "month", and "year" */
  interval?: string;
  /** @format int64 */
  on_start_credits?: string;
  /** @format int64 */
  trial_days?: string;
  metadata?: object;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
}

export interface V1Beta1PlanRequestBody {
  name?: string;
  title?: string;
  description?: string;
  products?: V1Beta1Product[];
  /** known intervals are "day", "week", "month", and "year" */
  interval?: string;
  /** @format int64 */
  on_start_credits?: string;
  /** @format int64 */
  trial_days?: string;
  metadata?: object;
}

export interface V1Beta1Policy {
  id?: string;
  title?: string;
  /**
   * The time the policy was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the policy was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
  role_id?: string;
  /** namespace:uuid */
  resource?: string;
  /** namespace:uuid */
  principal?: string;
  metadata?: object;
}

export interface V1Beta1PolicyRequestBody {
  /** unique id of the role to which policy is assigned */
  role_id: string;
  /** The title can contain any UTF-8 character, used to provide a human-readable name for the policy. Can also be left empty. <br/> *Example:* `Policy title` */
  title?: string;
  /** The resource to which policy is assigned in this format `namespace:uuid`. <br/> *Example:* `app/guardian:70f69c3a-334b-4f25-90b8-4d4f3be6b8e2` */
  resource: string;
  /** principal is the user or group to which policy is assigned. The principal id must be prefixed with its namespace id in this format `namespace:uuid`. The namespace can be `app/user`, `app/group` or `app/serviceuser` (coming up!) and uuid is the unique id of the principal. <br/> *Example:* `app/user:92f69c3a-334b-4f25-90b8-4d4f3be6b825` */
  principal: string;
  /** Metadata object for policies that can hold key value pairs defined in Policy Metaschema.<br/> *Example:* `{"labels": {"key": "value"}, "description": "Policy description"}` */
  metadata?: object;
}

export interface V1Beta1Preference {
  id?: string;
  name?: string;
  value?: string;
  resource_id?: string;
  resource_type?: string;
  /**
   * The time when the preference was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time when the preference was updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
}

export interface V1Beta1PreferenceRequestBody {
  name?: string;
  value?: string;
}

/**
 * PreferenceTrait is a trait that can be used to add preferences to a resource
 * it explains what preferences are available for a resource
 */
export interface V1Beta1PreferenceTrait {
  resource_type?: string;
  name?: string;
  title?: string;
  description?: string;
  long_description?: string;
  heading?: string;
  sub_heading?: string;
  breadcrumb?: string;
  default?: string;
  input_hints?: string;
  text?: string;
  textarea?: string;
  select?: string;
  combobox?: string;
  checkbox?: string;
  multiselect?: string;
  number?: string;
}

export interface V1Beta1Price {
  id?: string;
  product_id?: string;
  provider_id?: string;
  name?: string;
  /** known intervals are "day", "week", "month", and "year" */
  interval?: string;
  /** usage_type known types are "licensed" and "metered" */
  usage_type?: string;
  /** billing_scheme known schemes are "tiered" and "flat" */
  billing_scheme?: string;
  state?: string;
  /** currency like "usd", "eur", "gbp" */
  currency?: string;
  /** @format int64 */
  amount?: string;
  /** metered_aggregate known aggregations are "sum", "last_during_period" and "max" */
  metered_aggregate?: string;
  /** tier_mode known modes are "graduated" and "volume" */
  tier_mode?: string;
  metadata?: object;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
}

export interface V1Beta1Product {
  id?: string;
  name?: string;
  title?: string;
  description?: string;
  plan_ids?: string[];
  state?: string;
  prices?: V1Beta1Price[];
  behavior?: string;
  features?: V1Beta1Feature[];
  behavior_config?: ProductBehaviorConfig;
  metadata?: object;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
}

export interface V1Beta1ProductRequestBody {
  name?: string;
  title?: string;
  description?: string;
  plan_id?: string;
  prices?: V1Beta1Price[];
  behavior?: string;
  features?: V1Beta1Feature[];
  behavior_config?: ProductBehaviorConfig;
  metadata?: object;
}

export interface V1Beta1Project {
  id?: string;
  name?: string;
  title?: string;
  org_id?: string;
  metadata?: object;
  /**
   * The time the project was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the project was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
  /**
   * The number of members explicitly added in the project.
   * @format int32
   * @example 2
   */
  members_count?: number;
}

export interface V1Beta1ProjectRequestBody {
  /** The name of the project. The name must be unique within the entire Frontier instance. The name can contain only alphanumeric characters, dashes and underscores.<br/> *Example:* `frontier-playground` */
  name: string;
  /** The title can contain any UTF-8 character, used to provide a human-readable name for the project. Can also be left empty. <br/> *Example:* `Frontier Playground` */
  title?: string;
  /** Metadata object for projects that can hold key value pairs defined in Project Metaschema. */
  metadata?: object;
  /** unique id of the organization to which project belongs */
  org_id: string;
}

export interface V1Beta1Prospect {
  id?: string;
  name?: string;
  email?: string;
  phone?: string;
  activity?: string;
  status?: V1Beta1ProspectStatus;
  /** @format date-time */
  changed_at?: string;
  source?: string;
  verified?: boolean;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
  metadata?: object;
}

/**
 * subscription status
 * @default "STATUS_UNSPECIFIED"
 */
export enum V1Beta1ProspectStatus {
  STATUS_UNSPECIFIED = 'STATUS_UNSPECIFIED',
  STATUS_UNSUBSCRIBED = 'STATUS_UNSUBSCRIBED',
  STATUS_SUBSCRIBED = 'STATUS_SUBSCRIBED'
}

export interface V1Beta1RQLFilter {
  name?: string;
  operator?: string;
  bool_value?: boolean;
  string_value?: string;
  /** @format float */
  number_value?: number;
}

export interface V1Beta1RQLQueryGroupData {
  name?: string;
  /** @format int64 */
  count?: number;
}

export interface V1Beta1RQLQueryGroupResponse {
  name?: string;
  data?: V1Beta1RQLQueryGroupData[];
}

export interface V1Beta1RQLQueryPaginationResponse {
  /** @format int64 */
  offset?: number;
  /** @format int64 */
  limit?: number;
  /** @format int64 */
  total_count?: number;
}

export interface V1Beta1RQLRequest {
  filters?: V1Beta1RQLFilter[];
  group_by?: string[];
  /** @format int64 */
  offset?: number;
  /** @format int64 */
  limit?: number;
  search?: string;
  sort?: V1Beta1RQLSort[];
}

export interface V1Beta1RQLSort {
  name?: string;
  order?: string;
}

export type V1Beta1RegisterBillingAccountResponse = object;

export interface V1Beta1Relation {
  id?: string;
  /**
   * The time the relation was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the relation was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
  subject_sub_relation?: string;
  relation?: string;
  /** objectnamespace:id */
  object?: string;
  /** subjectnamespace:id */
  subject?: string;
}

export interface V1Beta1RelationRequestBody {
  /** objectnamespace:uuid */
  object?: string;
  /** subjectnamespace:uuid */
  subject?: string;
  relation?: string;
  subject_sub_relation?: string;
}

export type V1Beta1RemoveGroupUserResponse = object;

export type V1Beta1RemoveOrganizationUserResponse = object;

export interface V1Beta1RemovePlatformUserRequest {
  /** The user id to remove from the the platform. */
  user_id?: string;
  /** The service user id to remove from the the platform. */
  serviceuser_id?: string;
}

export type V1Beta1RemovePlatformUserResponse = object;

export interface V1Beta1Resource {
  id?: string;
  /** Name of the resource. Must be unique within the project. */
  name?: string;
  /**
   * The time the resource was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the resource was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
  urn?: string;
  project_id?: string;
  namespace?: string;
  principal?: string;
  metadata?: object;
  title?: string;
}

export interface V1Beta1ResourceRequestBody {
  /** The name of the resource.  Must be unique within the project. <br/> *Example:* `my-resource` */
  name: string;
  /** The title can contain any UTF-8 character, used to provide a human-readable name for the resource. Can also be left empty. */
  title?: string;
  /** The namespace of the resource. The resource namespace are created when permissions for that resource is created in Frontier. If namespace doesn't exists the request will fail. <br/> *Example:* `compute/instance` */
  namespace: string;
  /**
   * format namespace:uuid or just uuid for user
   * UserID or ServiceUserID that should be marked as owner of the resource. If not provided, the current logged in user will be made the resource owner. <br/> *Example:* `user:92f69c3a-334b-4f25-90b8-4d4f3be6b825`
   */
  principal?: string;
  metadata?: object;
}

export interface V1Beta1RevertBillingUsageRequest {
  /** either provide org_id or infer org from project_id */
  org_id?: string;
  project_id?: string;
  /**
   * either provide billing_id of the org or API can infer the default
   * billing ID from either org_id or project_id, not both
   */
  billing_id?: string;
  /** usage id to revert, a usage can only be allowed to revert once */
  usage_id?: string;
  /**
   * amount should be equal or less than the usage amount
   * @format int64
   */
  amount?: string;
}

export type V1Beta1RevertBillingUsageResponse = object;

export interface V1Beta1Role {
  id?: string;
  name?: string;
  permissions?: string[];
  title?: string;
  metadata?: object;
  /**
   * The time the role was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the role was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
  org_id?: string;
  state?: string;
  scopes?: string[];
}

export interface V1Beta1RoleRequestBody {
  name?: string;
  permissions?: string[];
  metadata?: object;
  title?: string;
  scopes?: string[];
}

export interface V1Beta1SearchOrganizationInvoicesResponse {
  organization_invoices?: SearchOrganizationInvoicesResponseOrganizationInvoice[];
  pagination?: V1Beta1RQLQueryPaginationResponse;
  group?: V1Beta1RQLQueryGroupResponse;
}

export interface V1Beta1SearchOrganizationProjectsResponse {
  org_projects?: SearchOrganizationProjectsResponseOrganizationProject[];
  pagination?: V1Beta1RQLQueryPaginationResponse;
  group?: V1Beta1RQLQueryGroupResponse;
}

export interface V1Beta1SearchOrganizationServiceUserCredentialsResponse {
  organization_serviceuser_credentials?: SearchOrganizationServiceUserCredentialsResponseOrganizationServiceUserCredential[];
  pagination?: V1Beta1RQLQueryPaginationResponse;
  group?: V1Beta1RQLQueryGroupResponse;
}

export interface V1Beta1SearchOrganizationTokensResponse {
  organization_tokens?: SearchOrganizationTokensResponseOrganizationToken[];
  pagination?: V1Beta1RQLQueryPaginationResponse;
  group?: V1Beta1RQLQueryGroupResponse;
}

export interface V1Beta1SearchOrganizationUsersResponse {
  org_users?: SearchOrganizationUsersResponseOrganizationUser[];
  pagination?: V1Beta1RQLQueryPaginationResponse;
  group?: V1Beta1RQLQueryGroupResponse;
}

export interface V1Beta1SearchOrganizationsResponse {
  organizations?: SearchOrganizationsResponseOrganizationResult[];
  pagination?: V1Beta1RQLQueryPaginationResponse;
  group?: V1Beta1RQLQueryGroupResponse;
}

export interface V1Beta1SearchProjectUsersResponse {
  project_users?: SearchProjectUsersResponseProjectUser[];
  pagination?: V1Beta1RQLQueryPaginationResponse;
  group?: V1Beta1RQLQueryGroupResponse;
}

export interface V1Beta1SecretCredential {
  id?: string;
  title?: string;
  /**
   * secret will only be returned once as part of the create process
   * this value is never persisted in the system so if lost, can't be recovered
   */
  secret?: string;
  /**
   * The time when the secret was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
}

export interface V1Beta1ServiceUser {
  id?: string;
  /**
   * User friendly name of the service user.
   * @example "Order Service"
   */
  title?: string;
  metadata?: object;
  /**
   * The time the user was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the user was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
  state?: string;
  org_id?: string;
}

export interface V1Beta1ServiceUserJWK {
  id?: string;
  title?: string;
  principal_id?: string;
  public_key?: string;
  /**
   * The time when the secret was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
}

export interface V1Beta1ServiceUserRequestBody {
  /**
   * User friendly name of the service user.
   * @example "Order Service"
   */
  title?: string;
  metadata?: object;
}

export interface V1Beta1ServiceUserToken {
  id?: string;
  title?: string;
  /**
   * token will only be returned once as part of the create process
   * this value is never persisted in the system so if lost, can't be recovered
   */
  token?: string;
  /**
   * The time when the token was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
}

export interface V1Beta1SetOrganizationKycResponse {
  organization_kyc?: V1Beta1OrganizationKyc;
}

export interface V1Beta1Subscription {
  id?: string;
  customer_id?: string;
  provider_id?: string;
  plan_id?: string;
  state?: string;
  metadata?: object;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
  /** @format date-time */
  canceled_at?: string;
  /** @format date-time */
  ended_at?: string;
  /** @format date-time */
  trial_ends_at?: string;
  /** @format date-time */
  current_period_start_at?: string;
  /** @format date-time */
  current_period_end_at?: string;
  /** @format date-time */
  billing_cycle_anchor_at?: string;
  phases?: SubscriptionPhase[];
  customer?: V1Beta1BillingAccount;
  plan?: V1Beta1Plan;
}

export interface V1Beta1TotalDebitedTransactionsResponse {
  /** Total debited amount in the billing account */
  debited?: BillingAccountBalance;
}

export type V1Beta1UpdateBillingAccountLimitsResponse = object;

export interface V1Beta1UpdateBillingAccountResponse {
  /** Updated billing account */
  billing_account?: V1Beta1BillingAccount;
}

export interface V1Beta1UpdateCurrentUserResponse {
  user?: V1Beta1User;
}

export interface V1Beta1UpdateFeatureResponse {
  /** Updated feature */
  feature?: V1Beta1Feature;
}

export interface V1Beta1UpdateGroupResponse {
  group?: V1Beta1Group;
}

export interface V1Beta1UpdateMetaSchemaResponse {
  metaschema?: V1Beta1MetaSchema;
}

export interface V1Beta1UpdateOrganizationResponse {
  organization?: V1Beta1Organization;
}

export interface V1Beta1UpdateOrganizationRoleResponse {
  role?: V1Beta1Role;
}

export interface V1Beta1UpdatePermissionResponse {
  permission?: V1Beta1Permission;
}

export interface V1Beta1UpdatePlanResponse {
  /** Updated plan */
  plan?: V1Beta1Plan;
}

export interface V1Beta1UpdatePolicyResponse {
  policies?: V1Beta1Policy[];
}

export interface V1Beta1UpdateProductResponse {
  /** Updated product */
  product?: V1Beta1Product;
}

export interface V1Beta1UpdateProjectResourceResponse {
  resource?: V1Beta1Resource;
}

export interface V1Beta1UpdateProjectResponse {
  project?: V1Beta1Project;
}

export interface V1Beta1UpdateProspectResponse {
  prospect?: V1Beta1Prospect;
}

export interface V1Beta1UpdateRoleResponse {
  role?: V1Beta1Role;
}

export interface V1Beta1UpdateSubscriptionResponse {
  /** Updated subscription */
  subscription?: V1Beta1Subscription;
}

export interface V1Beta1UpdateUserResponse {
  user?: V1Beta1User;
}

export interface V1Beta1UpdateWebhookResponse {
  webhook?: V1Beta1Webhook;
}

export interface V1Beta1Usage {
  /** uuid used as an idempotent key */
  id?: string;
  customer_id?: string;
  /** additional metadata for storing event/service that triggered this usage */
  source?: string;
  description?: string;
  /**
   * Type is the type of usage, it can be credit or feature
   * if credit, the amount is the amount of credits that were consumed
   * if feature, the amount is the amount of features that were used
   */
  type?: string;
  /** @format int64 */
  amount?: string;
  /** user_id is the user that triggered this usage */
  user_id?: string;
  metadata?: object;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
}

export interface V1Beta1User {
  id?: string;
  /**
   * can either be empty or must start with a character
   * Unique name of the user.
   * @example "johndoe"
   */
  name?: string;
  /**
   * Name of the user.
   * @example "John Doe"
   */
  title?: string;
  email?: string;
  metadata?: object;
  /**
   * The time the user was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  created_at?: string;
  /**
   * The time the user was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updated_at?: string;
  /**
   * The state of the user (enabled or disabled).
   * @example "enabled"
   */
  state?: string;
  /** The base64 encoded image string of the user avatar. Should be less than 2MB. */
  avatar?: string;
}

export interface V1Beta1UserRequestBody {
  /** The name of the user. The name must be unique within the entire Frontier instance. The name can contain only alphanumeric characters, dashes and underscores and must start with a letter. If not provided, Frontier automatically generates a name from the user email.  */
  name?: string;
  /** The email of the user. The email must be unique within the entire Frontier instance.<br/>*Example:*`"john.doe@raystack.org"` */
  email: string;
  /** Metadata object for users that can hold key value pairs pre-defined in User Metaschema. The metadata object can be used to store arbitrary information about the user such as label, description etc. By default the user metaschema contains labels and descriptions for the user. Update the same to add more fields to the user metadata object. <br/>*Example:*`{"label": {"key1": "value1"}, "description": "User Description"}` */
  metadata?: object;
  /** The title can contain any UTF-8 character, used to provide a human-readable name for the user. Can also be left empty. <br/>*Example:*`"John Doe"` */
  title?: string;
  /**
   * The avatar is base64 encoded image data of the user. Can also be left empty. The image should be less than 200KB. Should follow the regex pattern `^data:image/(png|jpg|jpeg|gif);base64,([a-zA-Z0-9+/]+={0,2})+$`.
   * @example "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAA"
   */
  avatar?: string;
}

export interface V1Beta1VerifyOrganizationDomainResponse {
  state?: string;
}

export interface V1Beta1Webhook {
  id?: string;
  description?: string;
  url?: string;
  subscribed_events?: string[];
  /** headers to be sent with the webhook */
  headers?: Record<string, string>;
  /**
   * secret to sign the payload, there could be multiple enabled secrets for a webhook
   * but only one will be used to sign the payload, this is useful for rotating secrets
   */
  secrets?: WebhookSecret[];
  state?: string;
  metadata?: object;
  /** @format date-time */
  created_at?: string;
  /** @format date-time */
  updated_at?: string;
}

export interface V1Beta1WebhookRequestBody {
  /** URL to send the webhook to (valid absolute URI via RFC 3986) */
  url?: string;
  description?: string;
  /** events to subscribe to, if empty all events are subscribed */
  subscribed_events?: string[];
  /** headers to be sent with the webhook */
  headers?: Record<string, string>;
  state?: string;
  metadata?: object;
}
