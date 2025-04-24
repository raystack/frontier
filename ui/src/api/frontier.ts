/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

/**
 * subscription status
 * @default "STATUS_UNSPECIFIED"
 */
export enum V1Beta1ProspectStatus {
  STATUS_UNSPECIFIED = "STATUS_UNSPECIFIED",
  STATUS_UNSUBSCRIBED = "STATUS_UNSUBSCRIBED",
  STATUS_SUBSCRIBED = "STATUS_SUBSCRIBED",
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
  NULL_VALUE = "NULL_VALUE",
}

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

export interface SearchUserOrganizationsResponseUserOrganization {
  org_id?: string;
  org_title?: string;
  org_name?: string;
  org_avatar?: string;
  /** @format int64 */
  project_count?: string;
  role_names?: string[];
  role_titles?: string[];
  role_ids?: string[];
  /** @format date-time */
  org_joined_on?: string;
  user_id?: string;
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

export interface Frontierv1Beta1OrganizationRequestBody {
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
  "@type"?: string;
  [key: string]: any;
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

export interface V1Beta1AdminCreateOrganizationRequestOrganizationRequestBody {
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
  /**
   * this user will be invited to org as owner.
   * if this user doesn't exist, it will be created first
   * The email of the user. The email must be unique within the entire Frontier instance.<br/>*Example:*`"john.doe@raystack.org"`
   */
  org_owner_email: string;
}

export interface V1Beta1AdminCreateOrganizationResponse {
  organization?: V1Beta1Organization;
}

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

export interface V1Beta1SearchUserOrganizationsResponse {
  user_organizations?: SearchUserOrganizationsResponseUserOrganization[];
  pagination?: V1Beta1RQLQueryPaginationResponse;
  group?: V1Beta1RQLQueryGroupResponse;
}

export interface V1Beta1SearchUsersResponse {
  users?: V1Beta1User[];
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

import type {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  HeadersDefaults,
  ResponseType,
} from "axios";
import axios from "axios";

export type QueryParamsType = Record<string | number, any>;

export interface FullRequestParams
  extends Omit<AxiosRequestConfig, "data" | "params" | "url" | "responseType"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseType;
  /** request body */
  body?: unknown;
}

export type RequestParams = Omit<
  FullRequestParams,
  "body" | "method" | "query" | "path"
>;

export interface ApiConfig<SecurityDataType = unknown>
  extends Omit<AxiosRequestConfig, "data" | "cancelToken"> {
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<AxiosRequestConfig | void> | AxiosRequestConfig | void;
  secure?: boolean;
  format?: ResponseType;
}

export enum ContentType {
  Json = "application/json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
  Text = "text/plain",
}

export class HttpClient<SecurityDataType = unknown> {
  public instance: AxiosInstance;
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private secure?: boolean;
  private format?: ResponseType;

  constructor({
    securityWorker,
    secure,
    format,
    ...axiosConfig
  }: ApiConfig<SecurityDataType> = {}) {
    this.instance = axios.create({
      ...axiosConfig,
      baseURL: axiosConfig.baseURL || "http://127.0.0.1:7400",
    });
    this.secure = secure;
    this.format = format;
    this.securityWorker = securityWorker;
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected mergeRequestParams(
    params1: AxiosRequestConfig,
    params2?: AxiosRequestConfig,
  ): AxiosRequestConfig {
    const method = params1.method || (params2 && params2.method);

    return {
      ...this.instance.defaults,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...((method &&
          this.instance.defaults.headers[
            method.toLowerCase() as keyof HeadersDefaults
          ]) ||
          {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected stringifyFormItem(formItem: unknown) {
    if (typeof formItem === "object" && formItem !== null) {
      return JSON.stringify(formItem);
    } else {
      return `${formItem}`;
    }
  }

  protected createFormData(input: Record<string, unknown>): FormData {
    if (input instanceof FormData) {
      return input;
    }
    return Object.keys(input || {}).reduce((formData, key) => {
      const property = input[key];
      const propertyContent: any[] =
        property instanceof Array ? property : [property];

      for (const formItem of propertyContent) {
        const isFileType = formItem instanceof Blob || formItem instanceof File;
        formData.append(
          key,
          isFileType ? formItem : this.stringifyFormItem(formItem),
        );
      }

      return formData;
    }, new FormData());
  }

  public request = async <T = any, _E = any>({
    secure,
    path,
    type,
    query,
    format,
    body,
    ...params
  }: FullRequestParams): Promise<AxiosResponse<T>> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const responseFormat = format || this.format || undefined;

    if (
      type === ContentType.FormData &&
      body &&
      body !== null &&
      typeof body === "object"
    ) {
      body = this.createFormData(body as Record<string, unknown>);
    }

    if (
      type === ContentType.Text &&
      body &&
      body !== null &&
      typeof body !== "string"
    ) {
      body = JSON.stringify(body);
    }

    return this.instance.request({
      ...requestParams,
      headers: {
        ...(requestParams.headers || {}),
        ...(type ? { "Content-Type": type } : {}),
      },
      params: query,
      responseType: responseFormat,
      data: body,
      url: path,
    });
  };
}

/**
 * @title Frontier Administration API
 * @version 0.2.0
 * @license Apache 2.0 (https://github.com/raystack/frontier/blob/main/LICENSE)
 * @baseUrl http://127.0.0.1:7400
 * @contact Raystack Foundation <hello@raystack.org> (https://raystack.org/)
 *
 * The Frontier APIs adhere to the OpenAPI specification, also known as Swagger, which provides a standardized approach for designing, documenting, and consuming RESTful APIs. With OpenAPI, you gain a clear understanding of the API endpoints, request/response structures, and authentication mechanisms supported by the Frontier APIs. By leveraging the OpenAPI specification, developers can easily explore and interact with the Frontier APIs using a variety of tools and libraries. The OpenAPI specification enables automatic code generation, interactive API documentation, and seamless integration with API testing frameworks, making it easier than ever to integrate Frontier into your existing applications and workflows.
 */
export class Api<
  SecurityDataType extends unknown,
> extends HttpClient<SecurityDataType> {
  wellKnown = {
    /**
     * No description
     *
     * @tags Authz
     * @name FrontierServiceGetJwKs2
     * @summary Get well known JWKs
     * @request GET:/.well-known/jwks.json
     * @secure
     */
    frontierServiceGetJwKs2: (params: RequestParams = {}) =>
      this.request<V1Beta1GetJWKsResponse, GooglerpcStatus>({
        path: `/.well-known/jwks.json`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),
  };
  v1Beta1 = {
    /**
     * @description Lists all the billing accounts from all the organizations in a Frontier instance. It can be filtered by organization.
     *
     * @tags Billing
     * @name AdminServiceListAllBillingAccounts
     * @summary List all billing accounts
     * @request GET:/v1beta1/admin/billing/accounts
     * @secure
     */
    adminServiceListAllBillingAccounts: (
      query?: {
        org_id?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListAllBillingAccountsResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/billing/accounts`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Lists all the invoices from all the organizations in a Frontier instance. It can be filtered by organization.
     *
     * @tags Invoice
     * @name AdminServiceListAllInvoices
     * @summary List all invoices
     * @request GET:/v1beta1/admin/billing/invoices
     * @secure
     */
    adminServiceListAllInvoices: (
      query?: {
        org_id?: string;
        /**
         * The maximum number of invoices to return per page. The default is 50.
         * @format int32
         */
        page_size?: number;
        /**
         * The page number to return. The default is 1.
         * @format int32
         */
        page_num?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListAllInvoicesResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/billing/invoices`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Triggers the creation of credit overdraft invoices for all billing accounts.
     *
     * @tags Invoice
     * @name AdminServiceGenerateInvoices
     * @summary Trigger invoice generation
     * @request POST:/v1beta1/admin/billing/invoices/generate
     * @secure
     */
    adminServiceGenerateInvoices: (
      body: V1Beta1GenerateInvoicesRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GenerateInvoicesResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/billing/invoices/generate`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns true if a principal has required permissions to access a resource and false otherwise.<br/> Note the principal can be a user, group or a service account.
     *
     * @tags Authz
     * @name AdminServiceCheckFederatedResourcePermission
     * @summary Check
     * @request POST:/v1beta1/admin/check
     * @secure
     */
    adminServiceCheckFederatedResourcePermission: (
      body: V1Beta1CheckFederatedResourcePermissionRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        V1Beta1CheckFederatedResourcePermissionResponse,
        GooglerpcStatus
      >({
        path: `/v1beta1/admin/check`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Lists all the groups from all the organizations in a Frontier instance. It can be filtered by organization and state.
     *
     * @tags Group
     * @name AdminServiceListGroups
     * @summary List all groups
     * @request GET:/v1beta1/admin/groups
     * @secure
     */
    adminServiceListGroups: (
      query?: {
        /** The organization id to filter by. */
        org_id?: string;
        /** The state to filter by. It can be enabled or disabled. */
        state?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListGroupsResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/groups`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Lists all the organizations in a Frontier instance. It can be filtered by user and state.
     *
     * @tags Organization
     * @name AdminServiceListAllOrganizations
     * @summary List all organizations
     * @request GET:/v1beta1/admin/organizations
     * @secure
     */
    adminServiceListAllOrganizations: (
      query?: {
        /** The user id to filter by. */
        user_id?: string;
        /** The state to filter by. It can be enabled or disabled. */
        state?: string;
        /**
         * The maximum number of organizations to return per page. The default is 50.
         * @format int32
         */
        page_size?: number;
        /**
         * The page number to return. The default is 1.
         * @format int32
         */
        page_num?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListAllOrganizationsResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Organization
     * @name AdminServiceAdminCreateOrganization
     * @summary Create organization as superadmin
     * @request POST:/v1beta1/admin/organizations
     * @secure
     */
    adminServiceAdminCreateOrganization: (
      body: V1Beta1AdminCreateOrganizationRequestOrganizationRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1AdminCreateOrganizationResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Organization
     * @name AdminServiceSearchOrganizationInvoices
     * @summary Search organization invoices
     * @request POST:/v1beta1/admin/organizations/{id}/invoices/search
     * @secure
     */
    adminServiceSearchOrganizationInvoices: (
      id: string,
      query: V1Beta1RQLRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1SearchOrganizationInvoicesResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${id}/invoices/search`,
        method: "POST",
        body: query,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Export organization projects with user IDs
     *
     * @tags Organization
     * @name AdminServiceExportOrganizationProjects
     * @summary Export organization projects
     * @request GET:/v1beta1/admin/organizations/{id}/projects/export
     * @secure
     */
    adminServiceExportOrganizationProjects: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<File, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${id}/projects/export`,
        method: "GET",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags Organization
     * @name AdminServiceSearchOrganizationProjects
     * @summary Search organization projects
     * @request POST:/v1beta1/admin/organizations/{id}/projects/search
     * @secure
     */
    adminServiceSearchOrganizationProjects: (
      id: string,
      query: V1Beta1RQLRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1SearchOrganizationProjectsResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${id}/projects/search`,
        method: "POST",
        body: query,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Organization
     * @name AdminServiceSearchOrganizationServiceUserCredentials
     * @summary Search organization service user credentials
     * @request POST:/v1beta1/admin/organizations/{id}/serviceuser_credentials/search
     * @secure
     */
    adminServiceSearchOrganizationServiceUserCredentials: (
      id: string,
      query: V1Beta1RQLRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        V1Beta1SearchOrganizationServiceUserCredentialsResponse,
        GooglerpcStatus
      >({
        path: `/v1beta1/admin/organizations/${id}/serviceuser_credentials/search`,
        method: "POST",
        body: query,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Export organization tokens
     *
     * @tags Organization
     * @name AdminServiceExportOrganizationTokens
     * @summary Export organization tokens
     * @request GET:/v1beta1/admin/organizations/{id}/tokens/export
     * @secure
     */
    adminServiceExportOrganizationTokens: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<File, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${id}/tokens/export`,
        method: "GET",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags Organization
     * @name AdminServiceSearchOrganizationTokens
     * @summary Search organization tokens
     * @request POST:/v1beta1/admin/organizations/{id}/tokens/search
     * @secure
     */
    adminServiceSearchOrganizationTokens: (
      id: string,
      query: V1Beta1RQLRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1SearchOrganizationTokensResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${id}/tokens/search`,
        method: "POST",
        body: query,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Export organization user their role details
     *
     * @tags Organization
     * @name AdminServiceExportOrganizationUsers
     * @summary Export organization users
     * @request GET:/v1beta1/admin/organizations/{id}/users/export
     * @secure
     */
    adminServiceExportOrganizationUsers: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<File, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${id}/users/export`,
        method: "GET",
        secure: true,
        ...params,
      }),

    /**
     * No description
     *
     * @tags Organization
     * @name AdminServiceSearchOrganizationUsers
     * @summary Search organization users
     * @request POST:/v1beta1/admin/organizations/{id}/users/search
     * @secure
     */
    adminServiceSearchOrganizationUsers: (
      id: string,
      query: V1Beta1RQLRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1SearchOrganizationUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${id}/users/search`,
        method: "POST",
        body: query,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Checkout a product to buy it one time or start a subscription plan on a billing account manually. It bypasses billing engine.
     *
     * @tags Checkout
     * @name AdminServiceDelegatedCheckout
     * @summary Checkout a product or subscription
     * @request POST:/v1beta1/admin/organizations/{org_id}/billing/{billing_id}/checkouts
     * @secure
     */
    adminServiceDelegatedCheckout: (
      orgId: string,
      billingId: string,
      body: {
        /** Subscription to create */
        subscription_body?: V1Beta1CheckoutSubscriptionBody;
        /** Product to buy */
        product_body?: V1Beta1CheckoutProductBody;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DelegatedCheckoutResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${orgId}/billing/${billingId}/checkouts`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Revert billing usage for a billing account.
     *
     * @tags Billing
     * @name AdminServiceRevertBillingUsage
     * @summary Revert billing usage
     * @request POST:/v1beta1/admin/organizations/{org_id}/billing/{billing_id}/usage/{usage_id}/revert
     * @secure
     */
    adminServiceRevertBillingUsage: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1RevertBillingUsageResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${orgId}/billing/${billingId}/usage/${usageId}/revert`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Update billing account limits.
     *
     * @tags Billing
     * @name AdminServiceUpdateBillingAccountLimits
     * @summary Update billing account limits
     * @request PUT:/v1beta1/admin/organizations/{org_id}/billing/{id}/limits
     * @secure
     */
    adminServiceUpdateBillingAccountLimits: (
      orgId: string,
      id: string,
      body: {
        /**
         * credit_min is the minimum credit limit for the billing account
         * default is 0, negative numbers work as overdraft, positive
         * numbers work as minimum purchase limit
         * @format int64
         */
        credit_min?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateBillingAccountLimitsResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${orgId}/billing/${id}/limits`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Checkout a product to buy it one time or start a subscription plan on a billing account manually. It bypasses billing engine.
     *
     * @tags Checkout
     * @name AdminServiceDelegatedCheckout2
     * @summary Checkout a product or subscription
     * @request POST:/v1beta1/admin/organizations/{org_id}/billing/checkouts
     * @secure
     */
    adminServiceDelegatedCheckout2: (
      orgId: string,
      body: {
        /** ID of the billing account to update the subscription for */
        billing_id?: string;
        /** Subscription to create */
        subscription_body?: V1Beta1CheckoutSubscriptionBody;
        /** Product to buy */
        product_body?: V1Beta1CheckoutProductBody;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DelegatedCheckoutResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${orgId}/billing/checkouts`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Set KYC information of an organization using ID or name
     *
     * @tags OrganizationKyc
     * @name AdminServiceSetOrganizationKyc
     * @summary Set KYC information of an organization
     * @request PUT:/v1beta1/admin/organizations/{org_id}/kyc
     * @secure
     */
    adminServiceSetOrganizationKyc: (
      orgId: string,
      body: {
        status?: boolean;
        link?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1SetOrganizationKycResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/${orgId}/kyc`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Export organization with demographic properties and billing plan details
     *
     * @tags Organization
     * @name AdminServiceExportOrganizations
     * @summary Export organizations
     * @request GET:/v1beta1/admin/organizations/export
     * @secure
     */
    adminServiceExportOrganizations: (params: RequestParams = {}) =>
      this.request<File, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/export`,
        method: "GET",
        secure: true,
        ...params,
      }),

    /**
     * @description List KYC information of all organization
     *
     * @tags OrganizationKyc
     * @name AdminServiceListOrganizationsKyc
     * @summary List KYC information of all organizations
     * @request GET:/v1beta1/admin/organizations/kyc
     * @secure
     */
    adminServiceListOrganizationsKyc: (params: RequestParams = {}) =>
      this.request<V1Beta1ListOrganizationsKycResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/kyc`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Search organization based on org demographic properties and billing plan details
     *
     * @tags Organization
     * @name AdminServiceSearchOrganizations
     * @summary Search organizations
     * @request POST:/v1beta1/admin/organizations/search
     * @secure
     */
    adminServiceSearchOrganizations: (
      query: V1Beta1RQLRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1SearchOrganizationsResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/organizations/search`,
        method: "POST",
        body: query,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Lists all the users added to the platform.
     *
     * @tags Platform
     * @name AdminServiceListPlatformUsers
     * @summary List platform users
     * @request GET:/v1beta1/admin/platform/users
     * @secure
     */
    adminServiceListPlatformUsers: (params: RequestParams = {}) =>
      this.request<V1Beta1ListPlatformUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/platform/users`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Adds a user to the platform.
     *
     * @tags Platform
     * @name AdminServiceAddPlatformUser
     * @summary Add platform user
     * @request POST:/v1beta1/admin/platform/users
     * @secure
     */
    adminServiceAddPlatformUser: (
      body: V1Beta1AddPlatformUserRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1AddPlatformUserResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/platform/users`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Removes a user from the platform.
     *
     * @tags Platform
     * @name AdminServiceRemovePlatformUser
     * @summary Remove platform user
     * @request POST:/v1beta1/admin/platform/users/remove
     * @secure
     */
    adminServiceRemovePlatformUser: (
      body: V1Beta1RemovePlatformUserRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1RemovePlatformUserResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/platform/users/remove`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Lists all the projects from all the organizations in a Frontier instance. It can be filtered by organization and state.
     *
     * @tags Project
     * @name AdminServiceListProjects
     * @summary List all projects
     * @request GET:/v1beta1/admin/projects
     * @secure
     */
    adminServiceListProjects: (
      query?: {
        /** The organization id to filter by. */
        org_id?: string;
        /** The state to filter by. It can be enabled or disabled. */
        state?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListProjectsResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/projects`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Project
     * @name AdminServiceSearchProjectUsers
     * @summary Search Project users
     * @request POST:/v1beta1/admin/projects/{id}/users/search
     * @secure
     */
    adminServiceSearchProjectUsers: (
      id: string,
      query: V1Beta1RQLRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1SearchProjectUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/projects/${id}/users/search`,
        method: "POST",
        body: query,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Create prospect for given email and activity.
     *
     * @tags Prospect
     * @name AdminServiceCreateProspect
     * @summary Create prospect
     * @request POST:/v1beta1/admin/prospects
     * @secure
     */
    adminServiceCreateProspect: (
      body: V1Beta1CreateProspectRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateProspectResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/prospects`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get prospect by ID
     *
     * @tags Prospect
     * @name AdminServiceGetProspect
     * @summary Get prospect
     * @request GET:/v1beta1/admin/prospects/{id}
     * @secure
     */
    adminServiceGetProspect: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetProspectResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/prospects/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete prospect for given ID
     *
     * @tags Prospect
     * @name AdminServiceDeleteProspect
     * @summary Delete prospect
     * @request DELETE:/v1beta1/admin/prospects/{id}
     * @secure
     */
    adminServiceDeleteProspect: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1DeleteProspectResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/prospects/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update prospect for given ID
     *
     * @tags Prospect
     * @name AdminServiceUpdateProspect
     * @summary Update prospect
     * @request PUT:/v1beta1/admin/prospects/{id}
     * @secure
     */
    adminServiceUpdateProspect: (
      id: string,
      body: {
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
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateProspectResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/prospects/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description List prospects and supports filters, sorting and pagination.
     *
     * @tags Prospect
     * @name AdminServiceListProspects
     * @summary List prospects
     * @request POST:/v1beta1/admin/prospects/list
     * @secure
     */
    adminServiceListProspects: (
      query: V1Beta1RQLRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListProspectsResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/prospects/list`,
        method: "POST",
        body: query,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Relation
     * @name AdminServiceListRelations
     * @summary List all relations
     * @request GET:/v1beta1/admin/relations
     * @secure
     */
    adminServiceListRelations: (
      query?: {
        /** The subject to filter by. */
        subject?: string;
        /** The object to filter by. */
        object?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListRelationsResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/relations`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Lists all the resources from all the organizations in a Frontier instance. It can be filtered by user, project, organization and namespace.
     *
     * @tags Resource
     * @name AdminServiceListResources
     * @summary List all resources
     * @request GET:/v1beta1/admin/resources
     * @secure
     */
    adminServiceListResources: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListResourcesResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/resources`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Lists all the users from all the organizations in a Frontier instance. It can be filtered by keyword, organization, group and state.
     *
     * @tags User
     * @name AdminServiceListAllUsers
     * @summary List all users
     * @request GET:/v1beta1/admin/users
     * @secure
     */
    adminServiceListAllUsers: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListAllUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/users`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags User
     * @name AdminServiceSearchUserOrganizations
     * @summary Search users organizations
     * @request POST:/v1beta1/admin/users/{id}/organizations/search
     * @secure
     */
    adminServiceSearchUserOrganizations: (
      id: string,
      query: V1Beta1RQLRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1SearchUserOrganizationsResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/users/${id}/organizations/search`,
        method: "POST",
        body: query,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Export user with demographic properties and billing plan details
     *
     * @tags User
     * @name AdminServiceExportUsers
     * @summary Export users
     * @request GET:/v1beta1/admin/users/export
     * @secure
     */
    adminServiceExportUsers: (params: RequestParams = {}) =>
      this.request<File, GooglerpcStatus>({
        path: `/v1beta1/admin/users/export`,
        method: "GET",
        secure: true,
        ...params,
      }),

    /**
     * @description Search users based on their properties
     *
     * @tags User
     * @name AdminServiceSearchUsers
     * @summary Search users
     * @request POST:/v1beta1/admin/users/search
     * @secure
     */
    adminServiceSearchUsers: (
      query: V1Beta1RQLRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1SearchUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/users/search`,
        method: "POST",
        body: query,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description List all webhooks.
     *
     * @tags Webhook
     * @name AdminServiceListWebhooks
     * @summary List webhooks
     * @request GET:/v1beta1/admin/webhooks
     * @secure
     */
    adminServiceListWebhooks: (params: RequestParams = {}) =>
      this.request<V1Beta1ListWebhooksResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/webhooks`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new webhook.
     *
     * @tags Webhook
     * @name AdminServiceCreateWebhook
     * @summary Create webhook
     * @request POST:/v1beta1/admin/webhooks
     * @secure
     */
    adminServiceCreateWebhook: (
      body: V1Beta1CreateWebhookRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateWebhookResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/webhooks`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a webhook.
     *
     * @tags Webhook
     * @name AdminServiceDeleteWebhook
     * @summary Delete webhook
     * @request DELETE:/v1beta1/admin/webhooks/{id}
     * @secure
     */
    adminServiceDeleteWebhook: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1DeleteWebhookResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/webhooks/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update a webhook.
     *
     * @tags Webhook
     * @name AdminServiceUpdateWebhook
     * @summary Update webhook
     * @request PUT:/v1beta1/admin/webhooks/{id}
     * @secure
     */
    adminServiceUpdateWebhook: (
      id: string,
      body: {
        body?: V1Beta1WebhookRequestBody;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateWebhookResponse, GooglerpcStatus>({
        path: `/v1beta1/admin/webhooks/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a list of identity providers configured on an instance level in Frontier. e.g Google, AzureAD, Github etc
     *
     * @tags Authn
     * @name FrontierServiceListAuthStrategies
     * @summary List authentication strategies
     * @request GET:/v1beta1/auth
     * @secure
     */
    frontierServiceListAuthStrategies: (params: RequestParams = {}) =>
      this.request<V1Beta1ListAuthStrategiesResponse, GooglerpcStatus>({
        path: `/v1beta1/auth`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Callback from a strategy. This is the endpoint where the strategy will redirect the user after successful authentication. This endpoint will be called with the code and state query parameters. The code will be used to get the access token from the strategy and the state will be used to get the return_to url from the session and redirect the user to that url.
     *
     * @tags Authn
     * @name FrontierServiceAuthCallback
     * @summary Callback from a strategy
     * @request GET:/v1beta1/auth/callback
     * @secure
     */
    frontierServiceAuthCallback: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1AuthCallbackResponse, GooglerpcStatus>({
        path: `/v1beta1/auth/callback`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Callback from a strategy. This is the endpoint where the strategy will redirect the user after successful authentication. This endpoint will be called with the code and state query parameters. The code will be used to get the access token from the strategy and the state will be used to get the return_to url from the session and redirect the user to that url.
     *
     * @tags Authn
     * @name FrontierServiceAuthCallback2
     * @summary Callback from a strategy
     * @request POST:/v1beta1/auth/callback
     * @secure
     */
    frontierServiceAuthCallback2: (
      body: V1Beta1AuthCallbackRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1AuthCallbackResponse, GooglerpcStatus>({
        path: `/v1beta1/auth/callback`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Authz
     * @name FrontierServiceGetJwKs
     * @summary Get well known JWKs
     * @request GET:/v1beta1/auth/jwks
     * @secure
     */
    frontierServiceGetJwKs: (params: RequestParams = {}) =>
      this.request<V1Beta1GetJWKsResponse, GooglerpcStatus>({
        path: `/v1beta1/auth/jwks`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Authn
     * @name FrontierServiceAuthLogout
     * @summary Logout from a strategy
     * @request GET:/v1beta1/auth/logout
     * @secure
     */
    frontierServiceAuthLogout: (params: RequestParams = {}) =>
      this.request<V1Beta1AuthLogoutResponse, GooglerpcStatus>({
        path: `/v1beta1/auth/logout`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Authn
     * @name FrontierServiceAuthLogout2
     * @summary Logout from a strategy
     * @request DELETE:/v1beta1/auth/logout
     * @secure
     */
    frontierServiceAuthLogout2: (params: RequestParams = {}) =>
      this.request<V1Beta1AuthLogoutResponse, GooglerpcStatus>({
        path: `/v1beta1/auth/logout`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Authenticate a user with a strategy. By default, after successful authentication no operation will be performed to apply redirection in case of browsers, provide a url in the authenticate request body that will be used for redirection after authentication. Also set redirect as true for redirecting the user request to the redirect_url after successful authentication.
     *
     * @tags Authn
     * @name FrontierServiceAuthenticate
     * @summary Authenticate with a strategy
     * @request GET:/v1beta1/auth/register/{strategy_name}
     * @secure
     */
    frontierServiceAuthenticate: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1AuthenticateResponse, GooglerpcStatus>({
        path: `/v1beta1/auth/register/${strategyName}`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Authenticate a user with a strategy. By default, after successful authentication no operation will be performed to apply redirection in case of browsers, provide a url in the authenticate request body that will be used for redirection after authentication. Also set redirect as true for redirecting the user request to the redirect_url after successful authentication.
     *
     * @tags Authn
     * @name FrontierServiceAuthenticate2
     * @summary Authenticate with a strategy
     * @request POST:/v1beta1/auth/register/{strategy_name}
     * @secure
     */
    frontierServiceAuthenticate2: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1AuthenticateResponse, GooglerpcStatus>({
        path: `/v1beta1/auth/register/${strategyName}`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Access token can be generated by providing the credentials in the request body/header. The credentials can be client id and secret or service account generated key jwt. Use the generated access token in Authorization header to access the frontier resources.
     *
     * @tags Authn
     * @name FrontierServiceAuthToken
     * @summary Generate access token by given credentials
     * @request POST:/v1beta1/auth/token
     * @secure
     */
    frontierServiceAuthToken: (
      body: V1Beta1AuthTokenRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1AuthTokenResponse, GooglerpcStatus>({
        path: `/v1beta1/auth/token`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns true if a principal has required permissions to access a resource and false otherwise.<br/> Note the principal can be a user or a service account, and Frontier will the credentials from the current logged in principal from the session cookie (if any), or the client id and secret (in case of service users) or the access token (in case of human user accounts).
     *
     * @tags Authz
     * @name FrontierServiceBatchCheckPermission
     * @summary Batch check
     * @request POST:/v1beta1/batchcheck
     * @secure
     */
    frontierServiceBatchCheckPermission: (
      body: V1Beta1BatchCheckPermissionRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1BatchCheckPermissionResponse, GooglerpcStatus>({
        path: `/v1beta1/batchcheck`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description List all features
     *
     * @tags Feature
     * @name FrontierServiceListFeatures
     * @summary List features
     * @request GET:/v1beta1/billing/features
     * @secure
     */
    frontierServiceListFeatures: (params: RequestParams = {}) =>
      this.request<V1Beta1ListFeaturesResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/features`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new feature for platform.
     *
     * @tags Feature
     * @name FrontierServiceCreateFeature
     * @summary Create feature
     * @request POST:/v1beta1/billing/features
     * @secure
     */
    frontierServiceCreateFeature: (
      body: V1Beta1CreateFeatureRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateFeatureResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/features`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a feature by ID.
     *
     * @tags Feature
     * @name FrontierServiceGetFeature
     * @summary Get feature
     * @request GET:/v1beta1/billing/features/{id}
     * @secure
     */
    frontierServiceGetFeature: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetFeatureResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/features/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update a feature by ID.
     *
     * @tags Feature
     * @name FrontierServiceUpdateFeature
     * @summary Update feature
     * @request PUT:/v1beta1/billing/features/{id}
     * @secure
     */
    frontierServiceUpdateFeature: (
      id: string,
      body: {
        /** Feature to update */
        body?: V1Beta1FeatureRequestBody;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateFeatureResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/features/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description List all plans.
     *
     * @tags Plan
     * @name FrontierServiceListPlans
     * @summary List plans
     * @request GET:/v1beta1/billing/plans
     * @secure
     */
    frontierServiceListPlans: (params: RequestParams = {}) =>
      this.request<V1Beta1ListPlansResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/plans`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new plan for platform.
     *
     * @tags Plan
     * @name FrontierServiceCreatePlan
     * @summary Create plan
     * @request POST:/v1beta1/billing/plans
     * @secure
     */
    frontierServiceCreatePlan: (
      body: V1Beta1CreatePlanRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreatePlanResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/plans`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a plan by ID.
     *
     * @tags Plan
     * @name FrontierServiceGetPlan
     * @summary Get plan
     * @request GET:/v1beta1/billing/plans/{id}
     * @secure
     */
    frontierServiceGetPlan: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetPlanResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/plans/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update a plan by ID.
     *
     * @tags Plan
     * @name FrontierServiceUpdatePlan
     * @summary Update plan
     * @request PUT:/v1beta1/billing/plans/{id}
     * @secure
     */
    frontierServiceUpdatePlan: (
      id: string,
      body: {
        /** Plan to update */
        body?: V1Beta1PlanRequestBody;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdatePlanResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/plans/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description List all products of a platform.
     *
     * @tags Product
     * @name FrontierServiceListProducts
     * @summary List products
     * @request GET:/v1beta1/billing/products
     * @secure
     */
    frontierServiceListProducts: (params: RequestParams = {}) =>
      this.request<V1Beta1ListProductsResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/products`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new product for platform.
     *
     * @tags Product
     * @name FrontierServiceCreateProduct
     * @summary Create product
     * @request POST:/v1beta1/billing/products
     * @secure
     */
    frontierServiceCreateProduct: (
      body: V1Beta1CreateProductRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateProductResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/products`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a product by ID.
     *
     * @tags Product
     * @name FrontierServiceGetProduct
     * @summary Get product
     * @request GET:/v1beta1/billing/products/{id}
     * @secure
     */
    frontierServiceGetProduct: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetProductResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/products/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update a product by ID.
     *
     * @tags Product
     * @name FrontierServiceUpdateProduct
     * @summary Update product
     * @request PUT:/v1beta1/billing/products/{id}
     * @secure
     */
    frontierServiceUpdateProduct: (
      id: string,
      body: {
        /** Product to update */
        body?: V1Beta1ProductRequestBody;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateProductResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/products/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Accepts a Billing webhook and processes it.
     *
     * @tags Webhook
     * @name FrontierServiceBillingWebhookCallback
     * @summary Accept Billing webhook
     * @request POST:/v1beta1/billing/webhooks/callback/{provider}
     * @secure
     */
    frontierServiceBillingWebhookCallback: (
      provider: string,
      body: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1BillingWebhookCallbackResponse, GooglerpcStatus>({
        path: `/v1beta1/billing/webhooks/callback/${provider}`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns true if a principal has required permissions to access a resource and false otherwise.<br/> Note the principal can be a user or a service account. Frontier will extract principal from the current logged in session cookie (if any), or the client id and secret (in case of service users) or the access token.
     *
     * @tags Authz
     * @name FrontierServiceCheckResourcePermission
     * @summary Check
     * @request POST:/v1beta1/check
     * @secure
     */
    frontierServiceCheckResourcePermission: (
      body: V1Beta1CheckResourcePermissionRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CheckResourcePermissionResponse, GooglerpcStatus>({
        path: `/v1beta1/check`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description List a group preferences by ID.
     *
     * @tags Preference
     * @name FrontierServiceListGroupPreferences
     * @summary List group preferences
     * @request GET:/v1beta1/groups/{id}/preferences
     * @secure
     */
    frontierServiceListGroupPreferences: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListGroupPreferencesResponse, GooglerpcStatus>({
        path: `/v1beta1/groups/${id}/preferences`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new group preferences. The group preferences **name** must be unique within the group and can contain only alphanumeric characters, dashes and underscores.
     *
     * @tags Preference
     * @name FrontierServiceCreateGroupPreferences
     * @summary Create group preferences
     * @request POST:/v1beta1/groups/{id}/preferences
     * @secure
     */
    frontierServiceCreateGroupPreferences: (
      id: string,
      body: {
        bodies?: V1Beta1PreferenceRequestBody[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateGroupPreferencesResponse, GooglerpcStatus>({
        path: `/v1beta1/groups/${id}/preferences`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a list of all metaschemas configured on an instance level in Frontier. e.g user, project, organization etc
     *
     * @tags MetaSchema
     * @name FrontierServiceListMetaSchemas
     * @summary List metaschemas
     * @request GET:/v1beta1/meta/schemas
     * @secure
     */
    frontierServiceListMetaSchemas: (params: RequestParams = {}) =>
      this.request<V1Beta1ListMetaSchemasResponse, GooglerpcStatus>({
        path: `/v1beta1/meta/schemas`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new metadata schema. The metaschema **name** must be unique within the entire Frontier instance and can contain only alphanumeric characters, dashes and underscores. The metaschema **schema** must be a valid JSON schema.Please refer to https://json-schema.org/ to know more about json schema. <br/>*Example:* `{name:"user",schema:{"type":"object","properties":{"label":{"type":"object","additionalProperties":{"type":"string"}},"description":{"type":"string"}}}}`
     *
     * @tags MetaSchema
     * @name FrontierServiceCreateMetaSchema
     * @summary Create metaschema
     * @request POST:/v1beta1/meta/schemas
     * @secure
     */
    frontierServiceCreateMetaSchema: (
      body: V1Beta1MetaSchemaRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateMetaSchemaResponse, GooglerpcStatus>({
        path: `/v1beta1/meta/schemas`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a metadata schema by ID.
     *
     * @tags MetaSchema
     * @name FrontierServiceGetMetaSchema
     * @summary Get metaschema
     * @request GET:/v1beta1/meta/schemas/{id}
     * @secure
     */
    frontierServiceGetMetaSchema: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetMetaSchemaResponse, GooglerpcStatus>({
        path: `/v1beta1/meta/schemas/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a metadata schema permanently. Once deleted the metaschema won't be used to validate the metadata. For example, if a metaschema(with `label` and `description` fields) is used to validate the metadata of a user, then deleting the metaschema will not validate the metadata of the user and metadata field can contain any key-value pair(and say another field called `foo` can be inserted in a user's metadata).
     *
     * @tags MetaSchema
     * @name FrontierServiceDeleteMetaSchema
     * @summary Delete metaschema
     * @request DELETE:/v1beta1/meta/schemas/{id}
     * @secure
     */
    frontierServiceDeleteMetaSchema: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1DeleteMetaSchemaResponse, GooglerpcStatus>({
        path: `/v1beta1/meta/schemas/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update a metadata schema. Only `schema` field of a metaschema can be updated. The metaschema `schema` must be a valid JSON schema.Please refer to https://json-schema.org/ to know more about json schema. <br/>*Example:* `{name:"user",schema:{"type":"object","properties":{"label":{"type":"object","additionalProperties":{"type":"string"}},"description":{"type":"string"}}}}`
     *
     * @tags MetaSchema
     * @name FrontierServiceUpdateMetaSchema
     * @summary Update metaschema
     * @request PUT:/v1beta1/meta/schemas/{id}
     * @secure
     */
    frontierServiceUpdateMetaSchema: (
      id: string,
      body: V1Beta1MetaSchemaRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateMetaSchemaResponse, GooglerpcStatus>({
        path: `/v1beta1/meta/schemas/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns the list of all namespaces in a Frontier instance
     *
     * @tags Namespace
     * @name FrontierServiceListNamespaces
     * @summary Get all namespaces
     * @request GET:/v1beta1/namespaces
     * @secure
     */
    frontierServiceListNamespaces: (params: RequestParams = {}) =>
      this.request<V1Beta1ListNamespacesResponse, GooglerpcStatus>({
        path: `/v1beta1/namespaces`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a namespace by ID
     *
     * @tags Namespace
     * @name FrontierServiceGetNamespace
     * @summary Get namespace
     * @request GET:/v1beta1/namespaces/{id}
     * @secure
     */
    frontierServiceGetNamespace: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetNamespaceResponse, GooglerpcStatus>({
        path: `/v1beta1/namespaces/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a list of organizations. It can be filtered by userID or organization state.
     *
     * @tags Organization
     * @name FrontierServiceListOrganizations
     * @summary List organizations
     * @request GET:/v1beta1/organizations
     * @secure
     */
    frontierServiceListOrganizations: (
      query?: {
        /** The user ID to filter by. It can be used to list all the organizations that the user is a member of. Otherwise, all the organizations in the Frontier instance will be listed. */
        user_id?: string;
        /** The state to filter by. It can be `enabled` or `disabled`. */
        state?: string;
        /**
         * The maximum number of users to return per page. The default is 1000.
         * @format int32
         */
        page_size?: number;
        /**
         * The page number to return. The default is 1.
         * @format int32
         */
        page_num?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListOrganizationsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Organization
     * @name FrontierServiceCreateOrganization
     * @summary Create organization
     * @request POST:/v1beta1/organizations
     * @secure
     */
    frontierServiceCreateOrganization: (
      body: Frontierv1Beta1OrganizationRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateOrganizationResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get organization by ID or name
     *
     * @tags Organization
     * @name FrontierServiceGetOrganization
     * @summary Get organization
     * @request GET:/v1beta1/organizations/{id}
     * @secure
     */
    frontierServiceGetOrganization: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetOrganizationResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete an organization and all of its relations permanently. The organization users will not be deleted from Frontier.
     *
     * @tags Organization
     * @name FrontierServiceDeleteOrganization
     * @summary Delete organization
     * @request DELETE:/v1beta1/organizations/{id}
     * @secure
     */
    frontierServiceDeleteOrganization: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DeleteOrganizationResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update organization by ID
     *
     * @tags Organization
     * @name FrontierServiceUpdateOrganization
     * @summary Update organization
     * @request PUT:/v1beta1/organizations/{id}
     * @secure
     */
    frontierServiceUpdateOrganization: (
      id: string,
      body: Frontierv1Beta1OrganizationRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateOrganizationResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a list admins of an organization
     *
     * @tags Organization
     * @name FrontierServiceListOrganizationAdmins
     * @summary List organization admins
     * @request GET:/v1beta1/organizations/{id}/admins
     * @secure
     */
    frontierServiceListOrganizationAdmins: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListOrganizationAdminsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${id}/admins`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Sets the state of the organization as disabled. The existing users in the org will not be able to access any organization resources.
     *
     * @tags Organization
     * @name FrontierServiceDisableOrganization
     * @summary Disable organization
     * @request POST:/v1beta1/organizations/{id}/disable
     * @secure
     */
    frontierServiceDisableOrganization: (
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DisableOrganizationResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${id}/disable`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Sets the state of the organization as enabled. The existing users in the org will be able to access any organization resources.
     *
     * @tags Organization
     * @name FrontierServiceEnableOrganization
     * @summary Enable organization
     * @request POST:/v1beta1/organizations/{id}/enable
     * @secure
     */
    frontierServiceEnableOrganization: (
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1EnableOrganizationResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${id}/enable`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description List an organization preferences by ID.
     *
     * @tags Preference
     * @name FrontierServiceListOrganizationPreferences
     * @summary List organization preferences
     * @request GET:/v1beta1/organizations/{id}/preferences
     * @secure
     */
    frontierServiceListOrganizationPreferences: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListOrganizationPreferencesResponse, GooglerpcStatus>(
        {
          path: `/v1beta1/organizations/${id}/preferences`,
          method: "GET",
          secure: true,
          format: "json",
          ...params,
        },
      ),

    /**
     * @description Create a new organization preferences. The organization preferences **name** must be unique within the organization and can contain only alphanumeric characters, dashes and underscores.
     *
     * @tags Preference
     * @name FrontierServiceCreateOrganizationPreferences
     * @summary Create organization preferences
     * @request POST:/v1beta1/organizations/{id}/preferences
     * @secure
     */
    frontierServiceCreateOrganizationPreferences: (
      id: string,
      body: {
        bodies?: V1Beta1PreferenceRequestBody[];
      },
      params: RequestParams = {},
    ) =>
      this.request<
        V1Beta1CreateOrganizationPreferencesResponse,
        GooglerpcStatus
      >({
        path: `/v1beta1/organizations/${id}/preferences`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get all projects that belong to an organization
     *
     * @tags Organization
     * @name FrontierServiceListOrganizationProjects
     * @summary Get organization projects
     * @request GET:/v1beta1/organizations/{id}/projects
     * @secure
     */
    frontierServiceListOrganizationProjects: (
      id: string,
      query?: {
        /** Filter projects by state. If not specified, all projects are returned. <br/> *Possible values:* `enabled` or `disabled` */
        state?: string;
        with_member_count?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListOrganizationProjectsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${id}/projects`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Organization
     * @name FrontierServiceListOrganizationServiceUsers
     * @summary List organization service users
     * @request GET:/v1beta1/organizations/{id}/serviceusers
     * @secure
     */
    frontierServiceListOrganizationServiceUsers: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<
        V1Beta1ListOrganizationServiceUsersResponse,
        GooglerpcStatus
      >({
        path: `/v1beta1/organizations/${id}/serviceusers`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Organization
     * @name FrontierServiceListOrganizationUsers
     * @summary List organization users
     * @request GET:/v1beta1/organizations/{id}/users
     * @secure
     */
    frontierServiceListOrganizationUsers: (
      id: string,
      query?: {
        permission_filter?: string;
        with_roles?: boolean;
        role_filters?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListOrganizationUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${id}/users`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Add a user to an organization. A user must exists in Frontier before adding it to an org. This request will fail if the user doesn't exists
     *
     * @tags Organization
     * @name FrontierServiceAddOrganizationUsers
     * @summary Add organization user
     * @request POST:/v1beta1/organizations/{id}/users
     * @secure
     */
    frontierServiceAddOrganizationUsers: (
      id: string,
      body: {
        /** List of user IDs to be added to the organization. */
        user_ids?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1AddOrganizationUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${id}/users`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Remove a user from an organization
     *
     * @tags Organization
     * @name FrontierServiceRemoveOrganizationUser
     * @summary Remove organization user
     * @request DELETE:/v1beta1/organizations/{id}/users/{user_id}
     * @secure
     */
    frontierServiceRemoveOrganizationUser: (
      id: string,
      userId: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1RemoveOrganizationUserResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${id}/users/${userId}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a list of audit logs of an organization in Frontier.
     *
     * @tags AuditLog
     * @name FrontierServiceListOrganizationAuditLogs
     * @summary List audit logs
     * @request GET:/v1beta1/organizations/{org_id}/auditlogs
     * @secure
     */
    frontierServiceListOrganizationAuditLogs: (
      orgId: string,
      query?: {
        source?: string;
        action?: string;
        ignore_system?: boolean;
        /**
         * start_time and end_time are inclusive
         * @format date-time
         */
        start_time?: string;
        /** @format date-time */
        end_time?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListOrganizationAuditLogsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/auditlogs`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create new audit logs in a batch.
     *
     * @tags AuditLog
     * @name FrontierServiceCreateOrganizationAuditLogs
     * @summary Create audit log
     * @request POST:/v1beta1/organizations/{org_id}/auditlogs
     * @secure
     */
    frontierServiceCreateOrganizationAuditLogs: (
      orgId: string,
      body: {
        logs?: V1Beta1AuditLog[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateOrganizationAuditLogsResponse, GooglerpcStatus>(
        {
          path: `/v1beta1/organizations/${orgId}/auditlogs`,
          method: "POST",
          body: body,
          secure: true,
          type: ContentType.Json,
          format: "json",
          ...params,
        },
      ),

    /**
     * @description Get an audit log by ID.
     *
     * @tags AuditLog
     * @name FrontierServiceGetOrganizationAuditLog
     * @summary Get audit log
     * @request GET:/v1beta1/organizations/{org_id}/auditlogs/{id}
     * @secure
     */
    frontierServiceGetOrganizationAuditLog: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetOrganizationAuditLogResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/auditlogs/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List billing accounts of an organization.
     *
     * @tags Billing
     * @name FrontierServiceListBillingAccounts
     * @summary List billing accounts
     * @request GET:/v1beta1/organizations/{org_id}/billing
     * @secure
     */
    frontierServiceListBillingAccounts: (
      orgId: string,
      query?: {
        expand?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListBillingAccountsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new billing account for an organization.
     *
     * @tags Billing
     * @name FrontierServiceCreateBillingAccount
     * @summary Create billing account
     * @request POST:/v1beta1/organizations/{org_id}/billing
     * @secure
     */
    frontierServiceCreateBillingAccount: (
      orgId: string,
      body: {
        /** Billing account to create. */
        body?: V1Beta1BillingAccountRequestBody;
        /** Offline billing account don't get registered with billing provider */
        offline?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateBillingAccountResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Check if a billing account is entitled to a feature.
     *
     * @tags Entitlement
     * @name FrontierServiceCheckFeatureEntitlement
     * @summary Check entitlement
     * @request POST:/v1beta1/organizations/{org_id}/billing/{billing_id}/check
     * @secure
     */
    frontierServiceCheckFeatureEntitlement: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CheckFeatureEntitlementResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/check`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description List all checkouts of a billing account.
     *
     * @tags Checkout
     * @name FrontierServiceListCheckouts
     * @summary List checkouts
     * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/checkouts
     * @secure
     */
    frontierServiceListCheckouts: (
      orgId: string,
      billingId: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListCheckoutsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/checkouts`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Checkout a product to buy it one time or start a subscription plan on a billing account.
     *
     * @tags Checkout
     * @name FrontierServiceCreateCheckout
     * @summary Checkout a product or subscription
     * @request POST:/v1beta1/organizations/{org_id}/billing/{billing_id}/checkouts
     * @secure
     */
    frontierServiceCreateCheckout: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateCheckoutResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/checkouts`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a checkout by ID.
     *
     * @tags Checkout
     * @name FrontierServiceGetCheckout
     * @summary Get checkout
     * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/checkouts/{id}
     * @secure
     */
    frontierServiceGetCheckout: (
      orgId: string,
      billingId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetCheckoutResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/checkouts/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Sum of amount of debited transactions including refunds
     *
     * @tags Transaction
     * @name FrontierServiceTotalDebitedTransactions
     * @summary Sum of amount of debited transactions including refunds
     * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/debited_transactions_total
     * @secure
     */
    frontierServiceTotalDebitedTransactions: (
      orgId: string,
      billingId: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1TotalDebitedTransactionsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/debited_transactions_total`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List all invoices of a billing account.
     *
     * @tags Invoice
     * @name FrontierServiceListInvoices
     * @summary List invoices
     * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/invoices
     * @secure
     */
    frontierServiceListInvoices: (
      orgId: string,
      billingId: string,
      query?: {
        nonzero_amount_only?: boolean;
        expand?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListInvoicesResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/invoices`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Get the upcoming invoice of a billing account.
     *
     * @tags Invoice
     * @name FrontierServiceGetUpcomingInvoice
     * @summary Get upcoming invoice
     * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/invoices/upcoming
     * @secure
     */
    frontierServiceGetUpcomingInvoice: (
      orgId: string,
      billingId: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetUpcomingInvoiceResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/invoices/upcoming`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List subscriptions of a billing account.
     *
     * @tags Subscription
     * @name FrontierServiceListSubscriptions
     * @summary List subscriptions
     * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/subscriptions
     * @secure
     */
    frontierServiceListSubscriptions: (
      orgId: string,
      billingId: string,
      query?: {
        /** Filter subscriptions by state */
        state?: string;
        plan?: string;
        expand?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListSubscriptionsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/subscriptions`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a subscription by ID.
     *
     * @tags Subscription
     * @name FrontierServiceGetSubscription
     * @summary Get subscription
     * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/subscriptions/{id}
     * @secure
     */
    frontierServiceGetSubscription: (
      orgId: string,
      billingId: string,
      id: string,
      query?: {
        expand?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetSubscriptionResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/subscriptions/${id}`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update a subscription by ID.
     *
     * @tags Subscription
     * @name FrontierServiceUpdateSubscription
     * @summary Update subscription
     * @request PUT:/v1beta1/organizations/{org_id}/billing/{billing_id}/subscriptions/{id}
     * @secure
     */
    frontierServiceUpdateSubscription: (
      orgId: string,
      billingId: string,
      id: string,
      body: {
        metadata?: object;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateSubscriptionResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/subscriptions/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Cancel a subscription by ID.
     *
     * @tags Subscription
     * @name FrontierServiceCancelSubscription
     * @summary Cancel subscription
     * @request POST:/v1beta1/organizations/{org_id}/billing/{billing_id}/subscriptions/{id}/cancel
     * @secure
     */
    frontierServiceCancelSubscription: (
      orgId: string,
      billingId: string,
      id: string,
      body: {
        immediate?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CancelSubscriptionResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/subscriptions/${id}/cancel`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Change a subscription plan by ID.
     *
     * @tags Subscription
     * @name FrontierServiceChangeSubscription
     * @summary Change subscription plan
     * @request POST:/v1beta1/organizations/{org_id}/billing/{billing_id}/subscriptions/{id}/change
     * @secure
     */
    frontierServiceChangeSubscription: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ChangeSubscriptionResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/subscriptions/${id}/change`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description List all transactions of a billing account.
     *
     * @tags Transaction
     * @name FrontierServiceListBillingTransactions
     * @summary List billing transactions
     * @request GET:/v1beta1/organizations/{org_id}/billing/{billing_id}/transactions
     * @secure
     */
    frontierServiceListBillingTransactions: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListBillingTransactionsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/transactions`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Report a new billing usage for a billing account.
     *
     * @tags Usage
     * @name FrontierServiceCreateBillingUsage
     * @summary Create billing usage
     * @request POST:/v1beta1/organizations/{org_id}/billing/{billing_id}/usages
     * @secure
     */
    frontierServiceCreateBillingUsage: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateBillingUsageResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${billingId}/usages`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a billing account by ID.
     *
     * @tags Billing
     * @name FrontierServiceGetBillingAccount
     * @summary Get billing account
     * @request GET:/v1beta1/organizations/{org_id}/billing/{id}
     * @secure
     */
    frontierServiceGetBillingAccount: (
      orgId: string,
      id: string,
      query?: {
        with_payment_methods?: boolean;
        expand?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetBillingAccountResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${id}`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a billing account by ID.
     *
     * @tags Billing
     * @name FrontierServiceDeleteBillingAccount
     * @summary Delete billing account
     * @request DELETE:/v1beta1/organizations/{org_id}/billing/{id}
     * @secure
     */
    frontierServiceDeleteBillingAccount: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DeleteBillingAccountResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update a billing account by ID.
     *
     * @tags Billing
     * @name FrontierServiceUpdateBillingAccount
     * @summary Update billing account
     * @request PUT:/v1beta1/organizations/{org_id}/billing/{id}
     * @secure
     */
    frontierServiceUpdateBillingAccount: (
      orgId: string,
      id: string,
      body: {
        /** Billing account to update. */
        body?: V1Beta1BillingAccountRequestBody;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateBillingAccountResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get the balance of a billing account by ID.
     *
     * @tags Billing
     * @name FrontierServiceGetBillingBalance
     * @summary Get billing balance
     * @request GET:/v1beta1/organizations/{org_id}/billing/{id}/balance
     * @secure
     */
    frontierServiceGetBillingBalance: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetBillingBalanceResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${id}/balance`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Disable a billing account by ID. Disabling a billing account doesn't automatically disable it's active subscriptions.
     *
     * @tags Billing
     * @name FrontierServiceDisableBillingAccount
     * @summary Disable billing account
     * @request POST:/v1beta1/organizations/{org_id}/billing/{id}/disable
     * @secure
     */
    frontierServiceDisableBillingAccount: (
      orgId: string,
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DisableBillingAccountResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${id}/disable`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Enable a billing account by ID.
     *
     * @tags Billing
     * @name FrontierServiceEnableBillingAccount
     * @summary Enable billing account
     * @request POST:/v1beta1/organizations/{org_id}/billing/{id}/enable
     * @secure
     */
    frontierServiceEnableBillingAccount: (
      orgId: string,
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1EnableBillingAccountResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${id}/enable`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Check if a billing account has trialed.
     *
     * @tags Billing
     * @name FrontierServiceHasTrialed
     * @summary Has trialed
     * @request GET:/v1beta1/organizations/{org_id}/billing/{id}/plans/{plan_id}/trialed
     * @secure
     */
    frontierServiceHasTrialed: (
      orgId: string,
      id: string,
      planId: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1HasTrialedResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${id}/plans/${planId}/trialed`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Register a billing account to a provider if it's not already.
     *
     * @tags Billing
     * @name FrontierServiceRegisterBillingAccount
     * @summary Register billing account to provider
     * @request POST:/v1beta1/organizations/{org_id}/billing/{id}/register
     * @secure
     */
    frontierServiceRegisterBillingAccount: (
      orgId: string,
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1RegisterBillingAccountResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/${id}/register`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description List all checkouts of a billing account.
     *
     * @tags Checkout
     * @name FrontierServiceListCheckouts2
     * @summary List checkouts
     * @request GET:/v1beta1/organizations/{org_id}/billing/checkouts
     * @secure
     */
    frontierServiceListCheckouts2: (
      orgId: string,
      query?: {
        /** ID of the billing account to get the subscriptions for */
        billing_id?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListCheckoutsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/checkouts`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Checkout a product to buy it one time or start a subscription plan on a billing account.
     *
     * @tags Checkout
     * @name FrontierServiceCreateCheckout2
     * @summary Checkout a product or subscription
     * @request POST:/v1beta1/organizations/{org_id}/billing/checkouts
     * @secure
     */
    frontierServiceCreateCheckout2: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateCheckoutResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/checkouts`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a checkout by ID.
     *
     * @tags Checkout
     * @name FrontierServiceGetCheckout2
     * @summary Get checkout
     * @request GET:/v1beta1/organizations/{org_id}/billing/checkouts/{id}
     * @secure
     */
    frontierServiceGetCheckout2: (
      orgId: string,
      id: string,
      query?: {
        /** ID of the billing account to get the subscriptions for */
        billing_id?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetCheckoutResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/checkouts/${id}`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List all invoices of a billing account.
     *
     * @tags Invoice
     * @name FrontierServiceListInvoices2
     * @summary List invoices
     * @request GET:/v1beta1/organizations/{org_id}/billing/invoices
     * @secure
     */
    frontierServiceListInvoices2: (
      orgId: string,
      query?: {
        /** ID of the billing account to list invoices for */
        billing_id?: string;
        nonzero_amount_only?: boolean;
        expand?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListInvoicesResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/invoices`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Get the upcoming invoice of a billing account.
     *
     * @tags Invoice
     * @name FrontierServiceGetUpcomingInvoice2
     * @summary Get upcoming invoice
     * @request GET:/v1beta1/organizations/{org_id}/billing/invoices/upcoming
     * @secure
     */
    frontierServiceGetUpcomingInvoice2: (
      orgId: string,
      query?: {
        /** ID of the billing account to get the upcoming invoice for */
        billing_id?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetUpcomingInvoiceResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/invoices/upcoming`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List subscriptions of a billing account.
     *
     * @tags Subscription
     * @name FrontierServiceListSubscriptions2
     * @summary List subscriptions
     * @request GET:/v1beta1/organizations/{org_id}/billing/subscriptions
     * @secure
     */
    frontierServiceListSubscriptions2: (
      orgId: string,
      query?: {
        /** ID of the billing account to list subscriptions for */
        billing_id?: string;
        /** Filter subscriptions by state */
        state?: string;
        plan?: string;
        expand?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListSubscriptionsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/subscriptions`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a subscription by ID.
     *
     * @tags Subscription
     * @name FrontierServiceGetSubscription2
     * @summary Get subscription
     * @request GET:/v1beta1/organizations/{org_id}/billing/subscriptions/{id}
     * @secure
     */
    frontierServiceGetSubscription2: (
      orgId: string,
      id: string,
      query?: {
        /** ID of the billing account to get the subscription for */
        billing_id?: string;
        expand?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetSubscriptionResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/subscriptions/${id}`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update a subscription by ID.
     *
     * @tags Subscription
     * @name FrontierServiceUpdateSubscription2
     * @summary Update subscription
     * @request POST:/v1beta1/organizations/{org_id}/billing/subscriptions/{id}
     * @secure
     */
    frontierServiceUpdateSubscription2: (
      orgId: string,
      id: string,
      body: {
        /** ID of the billing account to update the subscription for */
        billing_id?: string;
        metadata?: object;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateSubscriptionResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/subscriptions/${id}`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Cancel a subscription by ID.
     *
     * @tags Subscription
     * @name FrontierServiceCancelSubscription2
     * @summary Cancel subscription
     * @request POST:/v1beta1/organizations/{org_id}/billing/subscriptions/{id}/cancel
     * @secure
     */
    frontierServiceCancelSubscription2: (
      orgId: string,
      id: string,
      body: {
        /** ID of the billing account to update the subscription for */
        billing_id?: string;
        immediate?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CancelSubscriptionResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/subscriptions/${id}/cancel`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Change a subscription plan by ID.
     *
     * @tags Subscription
     * @name FrontierServiceChangeSubscription2
     * @summary Change subscription plan
     * @request POST:/v1beta1/organizations/{org_id}/billing/subscriptions/{id}/change
     * @secure
     */
    frontierServiceChangeSubscription2: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ChangeSubscriptionResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/subscriptions/${id}/change`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description List all transactions of a billing account.
     *
     * @tags Transaction
     * @name FrontierServiceListBillingTransactions2
     * @summary List billing transactions
     * @request GET:/v1beta1/organizations/{org_id}/billing/transactions
     * @secure
     */
    frontierServiceListBillingTransactions2: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListBillingTransactionsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/billing/transactions`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns all domains whitelisted for an organization (both pending and verified if no filters are provided for the state). The verified domains allow users email with the org's whitelisted domain to join the organization without invitation.
     *
     * @tags Organization
     * @name FrontierServiceListOrganizationDomains
     * @summary List org domains
     * @request GET:/v1beta1/organizations/{org_id}/domains
     * @secure
     */
    frontierServiceListOrganizationDomains: (
      orgId: string,
      query?: {
        /** filter to list domains by their state (pending/verified). If not provided, all domains for an org will be listed */
        state?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListOrganizationDomainsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/domains`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Add a domain to an organization which if verified allows all users of the same domain to be signed up to the organization without invitation. This API generates a verification token for a domain which must be added to your domain's DNS provider as a TXT record should be verified with Frontier VerifyOrganizationDomain API before it can be used as an Organization's trusted domain to sign up users.
     *
     * @tags Organization
     * @name FrontierServiceCreateOrganizationDomain
     * @summary Create org domain
     * @request POST:/v1beta1/organizations/{org_id}/domains
     * @secure
     */
    frontierServiceCreateOrganizationDomain: (
      orgId: string,
      body: {
        /** domain name to be added to the trusted domain list */
        domain: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateOrganizationDomainResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/domains`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a domain from the list of an organization's whitelisted domains. Returns both verified and unverified domains by their ID
     *
     * @tags Organization
     * @name FrontierServiceGetOrganizationDomain
     * @summary Get org domain
     * @request GET:/v1beta1/organizations/{org_id}/domains/{id}
     * @secure
     */
    frontierServiceGetOrganizationDomain: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetOrganizationDomainResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/domains/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Remove a domain from the list of an organization's trusted domains list
     *
     * @tags Organization
     * @name FrontierServiceDeleteOrganizationDomain
     * @summary Delete org domain
     * @request DELETE:/v1beta1/organizations/{org_id}/domains/{id}
     * @secure
     */
    frontierServiceDeleteOrganizationDomain: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DeleteOrganizationDomainResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/domains/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Verify a domain for an organization with a verification token generated by Frontier GenerateDomainVerificationToken API. The token must be added to your domain's DNS provider as a TXT record before it can be verified. This API returns the state of the domain (pending/verified) after verification.
     *
     * @tags Organization
     * @name FrontierServiceVerifyOrganizationDomain
     * @summary Verify org domain
     * @request POST:/v1beta1/organizations/{org_id}/domains/{id}/verify
     * @secure
     */
    frontierServiceVerifyOrganizationDomain: (
      orgId: string,
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1VerifyOrganizationDomainResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/domains/${id}/verify`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get all groups that belong to an organization. The results can be filtered by state which can be either be enabled or disabled.
     *
     * @tags Group
     * @name FrontierServiceListOrganizationGroups
     * @summary List organization groups
     * @request GET:/v1beta1/organizations/{org_id}/groups
     * @secure
     */
    frontierServiceListOrganizationGroups: (
      orgId: string,
      query?: {
        /** The state of the group to filter by. It can be enabled or disabled. */
        state?: string;
        group_ids?: string[];
        with_members?: boolean;
        with_member_count?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListOrganizationGroupsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/groups`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new group in an organization which serves as a container for users. The group can be assigned roles and permissions and can be used to manage access to resources. Also a group can also be assigned to other groups.
     *
     * @tags Group
     * @name FrontierServiceCreateGroup
     * @summary Create group
     * @request POST:/v1beta1/organizations/{org_id}/groups
     * @secure
     */
    frontierServiceCreateGroup: (
      orgId: string,
      body: V1Beta1GroupRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateGroupResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/groups`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Group
     * @name FrontierServiceGetGroup
     * @summary Get group
     * @request GET:/v1beta1/organizations/{org_id}/groups/{id}
     * @secure
     */
    frontierServiceGetGroup: (
      orgId: string,
      id: string,
      query?: {
        with_members?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetGroupResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/groups/${id}`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete an organization group permanently and all of its relations
     *
     * @tags Group
     * @name FrontierServiceDeleteGroup
     * @summary Delete group
     * @request DELETE:/v1beta1/organizations/{org_id}/groups/{id}
     * @secure
     */
    frontierServiceDeleteGroup: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DeleteGroupResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/groups/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Group
     * @name FrontierServiceUpdateGroup
     * @summary Update group
     * @request PUT:/v1beta1/organizations/{org_id}/groups/{id}
     * @secure
     */
    frontierServiceUpdateGroup: (
      orgId: string,
      id: string,
      body: V1Beta1GroupRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateGroupResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/groups/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Sets the state of the group as disabled. The group will not be available for access control and the existing users in the group will not be able to access any resources that are assigned to the group. No other users can be added to the group while it is disabled.
     *
     * @tags Group
     * @name FrontierServiceDisableGroup
     * @summary Disable group
     * @request POST:/v1beta1/organizations/{org_id}/groups/{id}/disable
     * @secure
     */
    frontierServiceDisableGroup: (
      orgId: string,
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DisableGroupResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/groups/${id}/disable`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Sets the state of the group as enabled. The `enabled` flag is used to determine if the group can be used for access control.
     *
     * @tags Group
     * @name FrontierServiceEnableGroup
     * @summary Enable group
     * @request POST:/v1beta1/organizations/{org_id}/groups/{id}/enable
     * @secure
     */
    frontierServiceEnableGroup: (
      orgId: string,
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1EnableGroupResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/groups/${id}/enable`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a list of users that belong to a group.
     *
     * @tags Group
     * @name FrontierServiceListGroupUsers
     * @summary List group users
     * @request GET:/v1beta1/organizations/{org_id}/groups/{id}/users
     * @secure
     */
    frontierServiceListGroupUsers: (
      orgId: string,
      id: string,
      query?: {
        with_roles?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListGroupUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/groups/${id}/users`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Adds a principle(user and service-users) to a group, existing users in the group will be ignored. A group can't have nested groups as members.
     *
     * @tags Group
     * @name FrontierServiceAddGroupUsers
     * @summary Add group user
     * @request POST:/v1beta1/organizations/{org_id}/groups/{id}/users
     * @secure
     */
    frontierServiceAddGroupUsers: (
      orgId: string,
      id: string,
      body: {
        user_ids?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1AddGroupUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/groups/${id}/users`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Removes a principle(user and service-users) from a group. If the user is not in the group, the request will succeed but no changes will be made.
     *
     * @tags Group
     * @name FrontierServiceRemoveGroupUser
     * @summary Remove group user
     * @request DELETE:/v1beta1/organizations/{org_id}/groups/{id}/users/{user_id}
     * @secure
     */
    frontierServiceRemoveGroupUser: (
      orgId: string,
      id: string,
      userId: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1RemoveGroupUserResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/groups/${id}/users/${userId}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns all pending invitations queued for an organization
     *
     * @tags Organization
     * @name FrontierServiceListOrganizationInvitations
     * @summary List pending invitations
     * @request GET:/v1beta1/organizations/{org_id}/invitations
     * @secure
     */
    frontierServiceListOrganizationInvitations: (
      orgId: string,
      query?: {
        /** user_id filter is the email id of user who are invited inside the organization. */
        user_id?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListOrganizationInvitationsResponse, GooglerpcStatus>(
        {
          path: `/v1beta1/organizations/${orgId}/invitations`,
          method: "GET",
          query: query,
          secure: true,
          format: "json",
          ...params,
        },
      ),

    /**
     * @description Invite users to an organization, if user is not registered on the platform, it will be notified. Invitations expire in 7 days
     *
     * @tags Organization
     * @name FrontierServiceCreateOrganizationInvitation
     * @summary Invite user
     * @request POST:/v1beta1/organizations/{org_id}/invitations
     * @secure
     */
    frontierServiceCreateOrganizationInvitation: (
      orgId: string,
      body: {
        /** user_id is email id of user who are invited inside the organization. If user is not registered on the platform, it will be notified */
        user_ids: string[];
        /** list of group ids to which user needs to be added as a member. */
        group_ids?: string[];
        /** list of role ids to which user needs to be added as a member. Roles are binded at organization level by default. */
        role_ids?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<
        V1Beta1CreateOrganizationInvitationResponse,
        GooglerpcStatus
      >({
        path: `/v1beta1/organizations/${orgId}/invitations`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a pending invitation queued for an organization
     *
     * @tags Organization
     * @name FrontierServiceGetOrganizationInvitation
     * @summary Get pending invitation
     * @request GET:/v1beta1/organizations/{org_id}/invitations/{id}
     * @secure
     */
    frontierServiceGetOrganizationInvitation: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetOrganizationInvitationResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/invitations/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a pending invitation queued for an organization
     *
     * @tags Organization
     * @name FrontierServiceDeleteOrganizationInvitation
     * @summary Delete pending invitation
     * @request DELETE:/v1beta1/organizations/{org_id}/invitations/{id}
     * @secure
     */
    frontierServiceDeleteOrganizationInvitation: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<
        V1Beta1DeleteOrganizationInvitationResponse,
        GooglerpcStatus
      >({
        path: `/v1beta1/organizations/${orgId}/invitations/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Accept pending invitation queued for an organization. The user will be added to the organization and groups defined in the invitation
     *
     * @tags Organization
     * @name FrontierServiceAcceptOrganizationInvitation
     * @summary Accept pending invitation
     * @request POST:/v1beta1/organizations/{org_id}/invitations/{id}/accept
     * @secure
     */
    frontierServiceAcceptOrganizationInvitation: (
      orgId: string,
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<
        V1Beta1AcceptOrganizationInvitationResponse,
        GooglerpcStatus
      >({
        path: `/v1beta1/organizations/${orgId}/invitations/${id}/accept`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Allows the current logged in user to join the Org if one is not a part of it. The user will only be able to join when the user email's domain matches the organization's whitelisted domains.
     *
     * @tags Organization
     * @name FrontierServiceJoinOrganization
     * @summary Join organization
     * @request POST:/v1beta1/organizations/{org_id}/join
     * @secure
     */
    frontierServiceJoinOrganization: (
      orgId: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1JoinOrganizationResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/join`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get KYC info of an organization using ID or name
     *
     * @tags OrganizationKyc
     * @name FrontierServiceGetOrganizationKyc
     * @summary Get KYC info of an organization
     * @request GET:/v1beta1/organizations/{org_id}/kyc
     * @secure
     */
    frontierServiceGetOrganizationKyc: (
      orgId: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetOrganizationKycResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/kyc`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a list of custom roles created under an organization with their associated permissions
     *
     * @tags Role
     * @name FrontierServiceListOrganizationRoles
     * @summary List organization roles
     * @request GET:/v1beta1/organizations/{org_id}/roles
     * @secure
     */
    frontierServiceListOrganizationRoles: (
      orgId: string,
      query?: {
        state?: string;
        scopes?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListOrganizationRolesResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/roles`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a custom role under an organization. This custom role will only be available for assignment to the principles within the organization.
     *
     * @tags Role
     * @name FrontierServiceCreateOrganizationRole
     * @summary Create organization role
     * @request POST:/v1beta1/organizations/{org_id}/roles
     * @secure
     */
    frontierServiceCreateOrganizationRole: (
      orgId: string,
      body: V1Beta1RoleRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateOrganizationRoleResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/roles`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a custom role under an organization along with its associated permissions
     *
     * @tags Role
     * @name FrontierServiceGetOrganizationRole
     * @summary Get organization role
     * @request GET:/v1beta1/organizations/{org_id}/roles/{id}
     * @secure
     */
    frontierServiceGetOrganizationRole: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetOrganizationRoleResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/roles/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a custom role and all of its relations under an organization permanently.
     *
     * @tags Role
     * @name FrontierServiceDeleteOrganizationRole
     * @summary Delete organization role
     * @request DELETE:/v1beta1/organizations/{org_id}/roles/{id}
     * @secure
     */
    frontierServiceDeleteOrganizationRole: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DeleteOrganizationRoleResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/roles/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update a custom role under an organization. This custom role will only be available for assignment to the principles within the organization.
     *
     * @tags Role
     * @name FrontierServiceUpdateOrganizationRole
     * @summary Update organization role
     * @request PUT:/v1beta1/organizations/{org_id}/roles/{id}
     * @secure
     */
    frontierServiceUpdateOrganizationRole: (
      orgId: string,
      id: string,
      body: V1Beta1RoleRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateOrganizationRoleResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/roles/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a service user.
     *
     * @tags ServiceUser
     * @name FrontierServiceCreateServiceUser
     * @summary Create service user
     * @request POST:/v1beta1/organizations/{org_id}/serviceusers
     * @secure
     */
    frontierServiceCreateServiceUser: (
      orgId: string,
      body: {
        body?: V1Beta1ServiceUserRequestBody;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateServiceUserResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get service user details by its id.
     *
     * @tags ServiceUser
     * @name FrontierServiceGetServiceUser
     * @summary Get service user
     * @request GET:/v1beta1/organizations/{org_id}/serviceusers/{id}
     * @secure
     */
    frontierServiceGetServiceUser: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetServiceUserResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a service user permanently and all of its relations (keys, organizations, roles, etc)
     *
     * @tags ServiceUser
     * @name FrontierServiceDeleteServiceUser
     * @summary Delete service user
     * @request DELETE:/v1beta1/organizations/{org_id}/serviceusers/{id}
     * @secure
     */
    frontierServiceDeleteServiceUser: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DeleteServiceUserResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List all the keys of a service user with its details except jwk.
     *
     * @tags ServiceUser
     * @name FrontierServiceListServiceUserJwKs
     * @summary List service user keys
     * @request GET:/v1beta1/organizations/{org_id}/serviceusers/{id}/keys
     * @secure
     */
    frontierServiceListServiceUserJwKs: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListServiceUserJWKsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers/${id}/keys`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Generate a service user key and return it, the private key part of the response will not be persisted and should be kept securely by client.
     *
     * @tags ServiceUser
     * @name FrontierServiceCreateServiceUserJwk
     * @summary Create service user public/private key pair
     * @request POST:/v1beta1/organizations/{org_id}/serviceusers/{id}/keys
     * @secure
     */
    frontierServiceCreateServiceUserJwk: (
      orgId: string,
      id: string,
      body: {
        title?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateServiceUserJWKResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers/${id}/keys`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a service user public RSA JWK set.
     *
     * @tags ServiceUser
     * @name FrontierServiceGetServiceUserJwk
     * @summary Get service user key
     * @request GET:/v1beta1/organizations/{org_id}/serviceusers/{id}/keys/{key_id}
     * @secure
     */
    frontierServiceGetServiceUserJwk: (
      orgId: string,
      id: string,
      keyId: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetServiceUserJWKResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers/${id}/keys/${keyId}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a service user key permanently.
     *
     * @tags ServiceUser
     * @name FrontierServiceDeleteServiceUserJwk
     * @summary Delete service user key
     * @request DELETE:/v1beta1/organizations/{org_id}/serviceusers/{id}/keys/{key_id}
     * @secure
     */
    frontierServiceDeleteServiceUserJwk: (
      orgId: string,
      id: string,
      keyId: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DeleteServiceUserJWKResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers/${id}/keys/${keyId}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List all projects the service user belongs to
     *
     * @tags ServiceUser
     * @name FrontierServiceListServiceUserProjects
     * @summary List service sser projects
     * @request GET:/v1beta1/organizations/{org_id}/serviceusers/{id}/projects
     * @secure
     */
    frontierServiceListServiceUserProjects: (
      orgId: string,
      id: string,
      query?: {
        /**
         * list of permissions needs to be checked against each project
         * query params are set as with_permissions=get&with_permissions=delete
         * to be represented as array
         */
        with_permissions?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListServiceUserProjectsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers/${id}/projects`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List all the credentials of a service user.
     *
     * @tags ServiceUser
     * @name FrontierServiceListServiceUserCredentials
     * @summary List service user credentials
     * @request GET:/v1beta1/organizations/{org_id}/serviceusers/{id}/secrets
     * @secure
     */
    frontierServiceListServiceUserCredentials: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListServiceUserCredentialsResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers/${id}/secrets`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Generate a service user credential and return it. The credential value will not be persisted and should be securely stored by client.
     *
     * @tags ServiceUser
     * @name FrontierServiceCreateServiceUserCredential
     * @summary Create service user client credentials
     * @request POST:/v1beta1/organizations/{org_id}/serviceusers/{id}/secrets
     * @secure
     */
    frontierServiceCreateServiceUserCredential: (
      orgId: string,
      id: string,
      body: {
        title?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateServiceUserCredentialResponse, GooglerpcStatus>(
        {
          path: `/v1beta1/organizations/${orgId}/serviceusers/${id}/secrets`,
          method: "POST",
          body: body,
          secure: true,
          type: ContentType.Json,
          format: "json",
          ...params,
        },
      ),

    /**
     * @description Delete a service user credential.
     *
     * @tags ServiceUser
     * @name FrontierServiceDeleteServiceUserCredential
     * @summary Delete service user credentials
     * @request DELETE:/v1beta1/organizations/{org_id}/serviceusers/{id}/secrets/{secret_id}
     * @secure
     */
    frontierServiceDeleteServiceUserCredential: (
      orgId: string,
      id: string,
      secretId: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DeleteServiceUserCredentialResponse, GooglerpcStatus>(
        {
          path: `/v1beta1/organizations/${orgId}/serviceusers/${id}/secrets/${secretId}`,
          method: "DELETE",
          secure: true,
          format: "json",
          ...params,
        },
      ),

    /**
     * @description List all the tokens of a service user.
     *
     * @tags ServiceUser
     * @name FrontierServiceListServiceUserTokens
     * @summary List service user tokens
     * @request GET:/v1beta1/organizations/{org_id}/serviceusers/{id}/tokens
     * @secure
     */
    frontierServiceListServiceUserTokens: (
      orgId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListServiceUserTokensResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers/${id}/tokens`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Generate a service user token and return it. The token value will not be persisted and should be securely stored by client.
     *
     * @tags ServiceUser
     * @name FrontierServiceCreateServiceUserToken
     * @summary Create service user token
     * @request POST:/v1beta1/organizations/{org_id}/serviceusers/{id}/tokens
     * @secure
     */
    frontierServiceCreateServiceUserToken: (
      orgId: string,
      id: string,
      body: {
        title?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateServiceUserTokenResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers/${id}/tokens`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a service user token.
     *
     * @tags ServiceUser
     * @name FrontierServiceDeleteServiceUserToken
     * @summary Delete service user token
     * @request DELETE:/v1beta1/organizations/{org_id}/serviceusers/{id}/tokens/{token_id}
     * @secure
     */
    frontierServiceDeleteServiceUserToken: (
      orgId: string,
      id: string,
      tokenId: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DeleteServiceUserTokenResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/${orgId}/serviceusers/${id}/tokens/${tokenId}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Check if a billing account is entitled to a feature.
     *
     * @tags Entitlement
     * @name FrontierServiceCheckFeatureEntitlement2
     * @summary Check entitlement
     * @request POST:/v1beta1/organizations/billing/check
     * @secure
     */
    frontierServiceCheckFeatureEntitlement2: (
      body: V1Beta1CheckFeatureEntitlementRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CheckFeatureEntitlementResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/billing/check`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Report a new billing usage for a billing account.
     *
     * @tags Usage
     * @name FrontierServiceCreateBillingUsage2
     * @summary Create billing usage
     * @request POST:/v1beta1/organizations/billing/usages
     * @secure
     */
    frontierServiceCreateBillingUsage2: (
      body: V1Beta1CreateBillingUsageRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateBillingUsageResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/billing/usages`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Revert billing usage for a billing account.
     *
     * @tags Billing
     * @name AdminServiceRevertBillingUsage2
     * @summary Revert billing usage
     * @request POST:/v1beta1/organizations/billing/usages/revert
     * @secure
     */
    adminServiceRevertBillingUsage2: (
      body: V1Beta1RevertBillingUsageRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1RevertBillingUsageResponse, GooglerpcStatus>({
        path: `/v1beta1/organizations/billing/usages/revert`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Permission
     * @name FrontierServiceListPermissions
     * @summary Get all permissions
     * @request GET:/v1beta1/permissions
     * @secure
     */
    frontierServiceListPermissions: (params: RequestParams = {}) =>
      this.request<V1Beta1ListPermissionsResponse, GooglerpcStatus>({
        path: `/v1beta1/permissions`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Creates a permission. It can be used to grant permissions to all the resources in a Frontier instance.
     *
     * @tags Permission
     * @name AdminServiceCreatePermission
     * @summary Create platform permission
     * @request POST:/v1beta1/permissions
     * @secure
     */
    adminServiceCreatePermission: (
      body: V1Beta1CreatePermissionRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreatePermissionResponse, GooglerpcStatus>({
        path: `/v1beta1/permissions`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a permission by ID
     *
     * @tags Permission
     * @name FrontierServiceGetPermission
     * @summary Get permission
     * @request GET:/v1beta1/permissions/{id}
     * @secure
     */
    frontierServiceGetPermission: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetPermissionResponse, GooglerpcStatus>({
        path: `/v1beta1/permissions/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Permission
     * @name AdminServiceDeletePermission
     * @summary Delete platform permission
     * @request DELETE:/v1beta1/permissions/{id}
     * @secure
     */
    adminServiceDeletePermission: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1DeletePermissionResponse, GooglerpcStatus>({
        path: `/v1beta1/permissions/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Updates a permission by ID. It can be used to grant permissions to all the resources in a Frontier instance.
     *
     * @tags Permission
     * @name AdminServiceUpdatePermission
     * @summary Update platform permission
     * @request PUT:/v1beta1/permissions/{id}
     * @secure
     */
    adminServiceUpdatePermission: (
      id: string,
      body: V1Beta1PermissionRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdatePermissionResponse, GooglerpcStatus>({
        path: `/v1beta1/permissions/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Lists all the policies from all the organizations in a Frontier instance. It can be filtered by organization, project, user, role and group.
     *
     * @tags Policy
     * @name FrontierServiceListPolicies
     * @summary List all policies
     * @request GET:/v1beta1/policies
     * @secure
     */
    frontierServiceListPolicies: (
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListPoliciesResponse, GooglerpcStatus>({
        path: `/v1beta1/policies`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Creates a policy
     *
     * @tags Policy
     * @name FrontierServiceCreatePolicy
     * @summary Create policy
     * @request POST:/v1beta1/policies
     * @secure
     */
    frontierServiceCreatePolicy: (
      body: V1Beta1PolicyRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreatePolicyResponse, GooglerpcStatus>({
        path: `/v1beta1/policies`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a policy by ID
     *
     * @tags Policy
     * @name FrontierServiceGetPolicy
     * @summary Get policy
     * @request GET:/v1beta1/policies/{id}
     * @secure
     */
    frontierServiceGetPolicy: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetPolicyResponse, GooglerpcStatus>({
        path: `/v1beta1/policies/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a policy all of its relations permanently.
     *
     * @tags Policy
     * @name FrontierServiceDeletePolicy
     * @summary Delete Policy
     * @request DELETE:/v1beta1/policies/{id}
     * @secure
     */
    frontierServiceDeletePolicy: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1DeletePolicyResponse, GooglerpcStatus>({
        path: `/v1beta1/policies/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Updates a policy by ID
     *
     * @tags Policy
     * @name FrontierServiceUpdatePolicy
     * @summary Update policy
     * @request PUT:/v1beta1/policies/{id}
     * @secure
     */
    frontierServiceUpdatePolicy: (
      id: string,
      body: V1Beta1PolicyRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdatePolicyResponse, GooglerpcStatus>({
        path: `/v1beta1/policies/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a policy for a project
     *
     * @tags Policy
     * @name FrontierServiceCreatePolicyForProject
     * @summary Create Policy for Project
     * @request POST:/v1beta1/policies/projects/{project_id}
     * @secure
     */
    frontierServiceCreatePolicyForProject: (
      projectId: string,
      body: V1Beta1CreatePolicyForProjectBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreatePolicyForProjectResponse, GooglerpcStatus>({
        path: `/v1beta1/policies/projects/${projectId}`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a list of all preferences configured on an instance in Frontier. e.g user, project, organization etc
     *
     * @tags Preference
     * @name AdminServiceListPreferences
     * @summary List platform preferences
     * @request GET:/v1beta1/preferences
     * @secure
     */
    adminServiceListPreferences: (params: RequestParams = {}) =>
      this.request<V1Beta1ListPreferencesResponse, GooglerpcStatus>({
        path: `/v1beta1/preferences`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create new platform preferences. The platform preferences **name** must be unique within the platform and can contain only alphanumeric characters, dashes and underscores.
     *
     * @tags Preference
     * @name AdminServiceCreatePreferences
     * @summary Create platform preferences
     * @request POST:/v1beta1/preferences
     * @secure
     */
    adminServiceCreatePreferences: (
      body: V1Beta1CreatePreferencesRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreatePreferencesResponse, GooglerpcStatus>({
        path: `/v1beta1/preferences`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a list of all preferences supported by Frontier.
     *
     * @tags Preference
     * @name FrontierServiceDescribePreferences
     * @summary Describe preferences
     * @request GET:/v1beta1/preferences/traits
     * @secure
     */
    frontierServiceDescribePreferences: (params: RequestParams = {}) =>
      this.request<V1Beta1DescribePreferencesResponse, GooglerpcStatus>({
        path: `/v1beta1/preferences/traits`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Project
     * @name FrontierServiceCreateProject
     * @summary Create project
     * @request POST:/v1beta1/projects
     * @secure
     */
    frontierServiceCreateProject: (
      body: V1Beta1ProjectRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateProjectResponse, GooglerpcStatus>({
        path: `/v1beta1/projects`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a project by ID
     *
     * @tags Project
     * @name FrontierServiceGetProject
     * @summary Get project
     * @request GET:/v1beta1/projects/{id}
     * @secure
     */
    frontierServiceGetProject: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetProjectResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a project all of its relations permanently.
     *
     * @tags Project
     * @name FrontierServiceDeleteProject
     * @summary Delete Project
     * @request DELETE:/v1beta1/projects/{id}
     * @secure
     */
    frontierServiceDeleteProject: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1DeleteProjectResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Updates a project by ID
     *
     * @tags Project
     * @name FrontierServiceUpdateProject
     * @summary Update project
     * @request PUT:/v1beta1/projects/{id}
     * @secure
     */
    frontierServiceUpdateProject: (
      id: string,
      body: V1Beta1ProjectRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateProjectResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a collection of admins of a project
     *
     * @tags Project
     * @name FrontierServiceListProjectAdmins
     * @summary List project admins
     * @request GET:/v1beta1/projects/{id}/admins
     * @secure
     */
    frontierServiceListProjectAdmins: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListProjectAdminsResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${id}/admins`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Project
     * @name FrontierServiceDisableProject
     * @summary Disable project
     * @request POST:/v1beta1/projects/{id}/disable
     * @secure
     */
    frontierServiceDisableProject: (
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DisableProjectResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${id}/disable`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Project
     * @name FrontierServiceEnableProject
     * @summary Enable project
     * @request POST:/v1beta1/projects/{id}/enable
     * @secure
     */
    frontierServiceEnableProject: (
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1EnableProjectResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${id}/enable`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a collection of groups of a project.
     *
     * @tags Project
     * @name FrontierServiceListProjectGroups
     * @summary List project groups
     * @request GET:/v1beta1/projects/{id}/groups
     * @secure
     */
    frontierServiceListProjectGroups: (
      id: string,
      query?: {
        with_roles?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListProjectGroupsResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${id}/groups`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List a project preferences by ID.
     *
     * @tags Preference
     * @name FrontierServiceListProjectPreferences
     * @summary List project preferences
     * @request GET:/v1beta1/projects/{id}/preferences
     * @secure
     */
    frontierServiceListProjectPreferences: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListProjectPreferencesResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${id}/preferences`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new project preferences. The project preferences **name** must be unique within the project and can contain only alphanumeric characters, dashes and underscores.
     *
     * @tags Preference
     * @name FrontierServiceCreateProjectPreferences
     * @summary Create project preferences
     * @request POST:/v1beta1/projects/{id}/preferences
     * @secure
     */
    frontierServiceCreateProjectPreferences: (
      id: string,
      body: {
        bodies?: V1Beta1PreferenceRequestBody[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateProjectPreferencesResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${id}/preferences`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a collection of users of a project.
     *
     * @tags Project
     * @name FrontierServiceListProjectServiceUsers
     * @summary List project serviceusers
     * @request GET:/v1beta1/projects/{id}/serviceusers
     * @secure
     */
    frontierServiceListProjectServiceUsers: (
      id: string,
      query?: {
        with_roles?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListProjectServiceUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${id}/serviceusers`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a collection of users of a project. Filter by user permissions is supported.
     *
     * @tags Project
     * @name FrontierServiceListProjectUsers
     * @summary List project users
     * @request GET:/v1beta1/projects/{id}/users
     * @secure
     */
    frontierServiceListProjectUsers: (
      id: string,
      query?: {
        permission_filter?: string;
        with_roles?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListProjectUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${id}/users`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Resource
     * @name FrontierServiceListProjectResources
     * @summary Get all resources
     * @request GET:/v1beta1/projects/{project_id}/resources
     * @secure
     */
    frontierServiceListProjectResources: (
      projectId: string,
      query?: {
        namespace?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListProjectResourcesResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${projectId}/resources`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Creates a resource in a project
     *
     * @tags Resource
     * @name FrontierServiceCreateProjectResource
     * @summary Create resource
     * @request POST:/v1beta1/projects/{project_id}/resources
     * @secure
     */
    frontierServiceCreateProjectResource: (
      projectId: string,
      body: V1Beta1ResourceRequestBody,
      query?: {
        /** Autogenerated if skipped. */
        id?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateProjectResourceResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${projectId}/resources`,
        method: "POST",
        query: query,
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a project resource by ID
     *
     * @tags Resource
     * @name FrontierServiceGetProjectResource
     * @summary Get resource
     * @request GET:/v1beta1/projects/{project_id}/resources/{id}
     * @secure
     */
    frontierServiceGetProjectResource: (
      projectId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1GetProjectResourceResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${projectId}/resources/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Deletes a resource from a project permanently
     *
     * @tags Resource
     * @name FrontierServiceDeleteProjectResource
     * @summary Delete resource
     * @request DELETE:/v1beta1/projects/{project_id}/resources/{id}
     * @secure
     */
    frontierServiceDeleteProjectResource: (
      projectId: string,
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DeleteProjectResourceResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${projectId}/resources/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Updates a resource in a project
     *
     * @tags Resource
     * @name FrontierServiceUpdateProjectResource
     * @summary Update resource
     * @request PUT:/v1beta1/projects/{project_id}/resources/{id}
     * @secure
     */
    frontierServiceUpdateProjectResource: (
      projectId: string,
      id: string,
      body: V1Beta1ResourceRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateProjectResourceResponse, GooglerpcStatus>({
        path: `/v1beta1/projects/${projectId}/resources/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Create prospect for given email and activity. Available for public access.
     *
     * @tags Prospect
     * @name FrontierServiceCreateProspectPublic
     * @summary Create prospect
     * @request POST:/v1beta1/prospects
     * @secure
     */
    frontierServiceCreateProspectPublic: (
      body: V1Beta1CreateProspectPublicRequest,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateProspectPublicResponse, GooglerpcStatus>({
        path: `/v1beta1/prospects`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Relation
     * @name FrontierServiceCreateRelation
     * @summary Create relation
     * @request POST:/v1beta1/relations
     * @secure
     */
    frontierServiceCreateRelation: (
      body: V1Beta1RelationRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateRelationResponse, GooglerpcStatus>({
        path: `/v1beta1/relations`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a relation by ID
     *
     * @tags Relation
     * @name FrontierServiceGetRelation
     * @summary Get relation
     * @request GET:/v1beta1/relations/{id}
     * @secure
     */
    frontierServiceGetRelation: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetRelationResponse, GooglerpcStatus>({
        path: `/v1beta1/relations/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Remove a subject having a relation from an object
     *
     * @tags Relation
     * @name FrontierServiceDeleteRelation
     * @summary Delete relation
     * @request DELETE:/v1beta1/relations/{relation}/object/{object}/subject/{subject}
     * @secure
     */
    frontierServiceDeleteRelation: (
      relation: string,
      object: string,
      subject: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DeleteRelationResponse, GooglerpcStatus>({
        path: `/v1beta1/relations/${relation}/object/${object}/subject/${subject}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns a list of platform wide roles available in enitre Frontier instance along with their associated permissions
     *
     * @tags Role
     * @name FrontierServiceListRoles
     * @summary List platform roles
     * @request GET:/v1beta1/roles
     * @secure
     */
    frontierServiceListRoles: (
      query?: {
        state?: string;
        scopes?: string[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListRolesResponse, GooglerpcStatus>({
        path: `/v1beta1/roles`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Creates a platform wide role. It can be used to grant permissions to all the resources in a Frontier instance.
     *
     * @tags Role
     * @name AdminServiceCreateRole
     * @summary Create platform role
     * @request POST:/v1beta1/roles
     * @secure
     */
    adminServiceCreateRole: (
      body: V1Beta1RoleRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateRoleResponse, GooglerpcStatus>({
        path: `/v1beta1/roles`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete a platform wide role and all of its relations.
     *
     * @tags Role
     * @name AdminServiceDeleteRole
     * @summary Delete platform role
     * @request DELETE:/v1beta1/roles/{id}
     * @secure
     */
    adminServiceDeleteRole: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1DeleteRoleResponse, GooglerpcStatus>({
        path: `/v1beta1/roles/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Update a role title, description and permissions.
     *
     * @tags Role
     * @name AdminServiceUpdateRole
     * @summary Update role
     * @request PUT:/v1beta1/roles/{id}
     * @secure
     */
    adminServiceUpdateRole: (
      id: string,
      body: V1Beta1RoleRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateRoleResponse, GooglerpcStatus>({
        path: `/v1beta1/roles/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns the service user of an organization in a Frontier instance. It can be filter by it's state
     *
     * @tags ServiceUser
     * @name FrontierServiceListServiceUsers
     * @summary List org service users
     * @request GET:/v1beta1/serviceusers
     * @secure
     */
    frontierServiceListServiceUsers: (
      query: {
        /** The organization ID to filter service users by. */
        org_id: string;
        /** The state to filter by. It can be enabled or disabled. */
        state?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListServiceUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/serviceusers`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns the users from all the organizations in a Frontier instance. It can be filtered by keyword, organization, group and state. Additionally you can include page number and page size for pagination.
     *
     * @tags User
     * @name FrontierServiceListUsers
     * @summary List public users
     * @request GET:/v1beta1/users
     * @secure
     */
    frontierServiceListUsers: (
      query?: {
        /**
         * The maximum number of users to return per page. The default is 1000.
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
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListUsersResponse, GooglerpcStatus>({
        path: `/v1beta1/users`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a user with the given details. A user is not attached to an organization or a group by default,and can be invited to the org/group. The name of the user must be unique within the entire Frontier instance. If a user name is not provided, Frontier automatically generates a name from the user email. The user metadata is validated against the user metaschema. By default the user metaschema contains `labels` and `descriptions` for the user. The `title` field can be optionally added for a user-friendly name. <br/><br/>*Example:*`{"email":"john.doe@raystack.org","title":"John Doe",metadata:{"label": {"key1": "value1"}, "description": "User Description"}}`
     *
     * @tags User
     * @name FrontierServiceCreateUser
     * @summary Create user
     * @request POST:/v1beta1/users
     */
    frontierServiceCreateUser: (
      body: V1Beta1UserRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateUserResponse, GooglerpcStatus>({
        path: `/v1beta1/users`,
        method: "POST",
        body: body,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get a user by id searched over all organizations in Frontier.
     *
     * @tags User
     * @name FrontierServiceGetUser
     * @summary Get user
     * @request GET:/v1beta1/users/{id}
     * @secure
     */
    frontierServiceGetUser: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1GetUserResponse, GooglerpcStatus>({
        path: `/v1beta1/users/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Delete an user permanently forever and all of its relations (organizations, groups, etc)
     *
     * @tags User
     * @name FrontierServiceDeleteUser
     * @summary Delete user
     * @request DELETE:/v1beta1/users/{id}
     * @secure
     */
    frontierServiceDeleteUser: (id: string, params: RequestParams = {}) =>
      this.request<V1Beta1DeleteUserResponse, GooglerpcStatus>({
        path: `/v1beta1/users/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags User
     * @name FrontierServiceUpdateUser
     * @summary Update user
     * @request PUT:/v1beta1/users/{id}
     * @secure
     */
    frontierServiceUpdateUser: (
      id: string,
      body: V1Beta1UserRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateUserResponse, GooglerpcStatus>({
        path: `/v1beta1/users/${id}`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Sets the state of the user as diabled.The user's membership to groups and organizations will still exist along with all it's roles for access control, but the user will not be able to log in and access the Frontier instance.
     *
     * @tags User
     * @name FrontierServiceDisableUser
     * @summary Disable user
     * @request POST:/v1beta1/users/{id}/disable
     * @secure
     */
    frontierServiceDisableUser: (
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1DisableUserResponse, GooglerpcStatus>({
        path: `/v1beta1/users/${id}/disable`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Sets the state of the user as enabled. The user will be able to log in and access the Frontier instance.
     *
     * @tags User
     * @name FrontierServiceEnableUser
     * @summary Enable user
     * @request POST:/v1beta1/users/{id}/enable
     * @secure
     */
    frontierServiceEnableUser: (
      id: string,
      body: V1Beta1AcceptOrganizationInvitationResponse,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1EnableUserResponse, GooglerpcStatus>({
        path: `/v1beta1/users/${id}/enable`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Lists all the groups a user belongs to across all organization in Frontier. To get the groups of a user in a specific organization, use the org_id filter in the query parameter.
     *
     * @tags User
     * @name FrontierServiceListUserGroups
     * @summary List user groups
     * @request GET:/v1beta1/users/{id}/groups
     * @secure
     */
    frontierServiceListUserGroups: (
      id: string,
      query?: {
        /** The organization ID to filter groups by. If not provided, groups from all organizations are returned. */
        org_id?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListUserGroupsResponse, GooglerpcStatus>({
        path: `/v1beta1/users/${id}/groups`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List all the invitations sent to a user.
     *
     * @tags User
     * @name FrontierServiceListUserInvitations
     * @summary List user invitations
     * @request GET:/v1beta1/users/{id}/invitations
     * @secure
     */
    frontierServiceListUserInvitations: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListUserInvitationsResponse, GooglerpcStatus>({
        path: `/v1beta1/users/${id}/invitations`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description This API returns two list of organizations for the user. i) The list of orgs which the current user is already a part of ii) The list of organizations the user can join directly (based on domain whitelisted and verified by the org). This list will also contain orgs of which user is already a part of. Note: the domain needs to be verified by the org before the it is returned as one of the joinable orgs by domain
     *
     * @tags User
     * @name FrontierServiceListOrganizationsByUser
     * @summary Get user organizations
     * @request GET:/v1beta1/users/{id}/organizations
     * @secure
     */
    frontierServiceListOrganizationsByUser: (
      id: string,
      query?: {
        state?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListOrganizationsByUserResponse, GooglerpcStatus>({
        path: `/v1beta1/users/${id}/organizations`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List a user preferences by ID.
     *
     * @tags Preference
     * @name FrontierServiceListUserPreferences
     * @summary List user preferences
     * @request GET:/v1beta1/users/{id}/preferences
     * @secure
     */
    frontierServiceListUserPreferences: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListUserPreferencesResponse, GooglerpcStatus>({
        path: `/v1beta1/users/${id}/preferences`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new user preferences. The user preferences **name** must be unique within the user and can contain only alphanumeric characters, dashes and underscores.
     *
     * @tags Preference
     * @name FrontierServiceCreateUserPreferences
     * @summary Create user preferences
     * @request POST:/v1beta1/users/{id}/preferences
     * @secure
     */
    frontierServiceCreateUserPreferences: (
      id: string,
      body: {
        bodies?: V1Beta1PreferenceRequestBody[];
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1CreateUserPreferencesResponse, GooglerpcStatus>({
        path: `/v1beta1/users/${id}/preferences`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get all the projects a user belongs to.
     *
     * @tags User
     * @name FrontierServiceListProjectsByUser
     * @summary Get user projects
     * @request GET:/v1beta1/users/{id}/projects
     * @secure
     */
    frontierServiceListProjectsByUser: (
      id: string,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListProjectsByUserResponse, GooglerpcStatus>({
        path: `/v1beta1/users/${id}/projects`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags User
     * @name FrontierServiceGetCurrentUser
     * @summary Get current user
     * @request GET:/v1beta1/users/self
     * @secure
     */
    frontierServiceGetCurrentUser: (params: RequestParams = {}) =>
      this.request<V1Beta1GetCurrentUserResponse, GooglerpcStatus>({
        path: `/v1beta1/users/self`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags User
     * @name FrontierServiceUpdateCurrentUser
     * @summary Update current user
     * @request PUT:/v1beta1/users/self
     * @secure
     */
    frontierServiceUpdateCurrentUser: (
      body: V1Beta1UserRequestBody,
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1UpdateCurrentUserResponse, GooglerpcStatus>({
        path: `/v1beta1/users/self`,
        method: "PUT",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags User
     * @name FrontierServiceListCurrentUserGroups
     * @summary List my groups
     * @request GET:/v1beta1/users/self/groups
     * @secure
     */
    frontierServiceListCurrentUserGroups: (
      query?: {
        /** org_id is optional filter over an organization */
        org_id?: string;
        with_permissions?: string[];
        with_member_count?: boolean;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListCurrentUserGroupsResponse, GooglerpcStatus>({
        path: `/v1beta1/users/self/groups`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List all the invitations sent to current user.
     *
     * @tags User
     * @name FrontierServiceListCurrentUserInvitations
     * @summary List user invitations
     * @request GET:/v1beta1/users/self/invitations
     * @secure
     */
    frontierServiceListCurrentUserInvitations: (params: RequestParams = {}) =>
      this.request<V1Beta1ListCurrentUserInvitationsResponse, GooglerpcStatus>({
        path: `/v1beta1/users/self/invitations`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description This API returns two list of organizations for the current logged in user. i) The list of orgs which the current user is already a part of ii) The list of organizations the user can join directly (based on domain whitelisted and verified by the org). This list will also contain orgs of which user is already a part of. Note: the domain needs to be verified by the org before the it is returned as one of the joinable orgs by domain
     *
     * @tags User
     * @name FrontierServiceListOrganizationsByCurrentUser
     * @summary Get my organizations
     * @request GET:/v1beta1/users/self/organizations
     * @secure
     */
    frontierServiceListOrganizationsByCurrentUser: (
      query?: {
        state?: string;
      },
      params: RequestParams = {},
    ) =>
      this.request<
        V1Beta1ListOrganizationsByCurrentUserResponse,
        GooglerpcStatus
      >({
        path: `/v1beta1/users/self/organizations`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description List a user preferences by ID.
     *
     * @tags Preference
     * @name FrontierServiceListCurrentUserPreferences
     * @summary List current user preferences
     * @request GET:/v1beta1/users/self/preferences
     * @secure
     */
    frontierServiceListCurrentUserPreferences: (params: RequestParams = {}) =>
      this.request<V1Beta1ListCurrentUserPreferencesResponse, GooglerpcStatus>({
        path: `/v1beta1/users/self/preferences`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Create a new user preferences. The user preferences **name** must be unique within the user and can contain only alphanumeric characters, dashes and underscores.
     *
     * @tags Preference
     * @name FrontierServiceCreateCurrentUserPreferences
     * @summary Create current user preferences
     * @request POST:/v1beta1/users/self/preferences
     * @secure
     */
    frontierServiceCreateCurrentUserPreferences: (
      body: V1Beta1CreateCurrentUserPreferencesRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        V1Beta1CreateCurrentUserPreferencesResponse,
        GooglerpcStatus
      >({
        path: `/v1beta1/users/self/preferences`,
        method: "POST",
        body: body,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Get all projects the current user belongs to
     *
     * @tags User
     * @name FrontierServiceListProjectsByCurrentUser
     * @summary Get my projects
     * @request GET:/v1beta1/users/self/projects
     * @secure
     */
    frontierServiceListProjectsByCurrentUser: (
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
        /**
         * The maximum number of users to return per page. The default is 1000.
         * @format int32
         */
        page_size?: number;
        /**
         * The page number to return. The default is 1.
         * @format int32
         */
        page_num?: number;
      },
      params: RequestParams = {},
    ) =>
      this.request<V1Beta1ListProjectsByCurrentUserResponse, GooglerpcStatus>({
        path: `/v1beta1/users/self/projects`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),
  };
}
