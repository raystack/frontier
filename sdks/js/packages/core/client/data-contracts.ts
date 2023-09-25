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

export interface ProtobufAny {
  '@type'?: string;
  [key: string]: any;
}

/**
 * `NullValue` is a singleton enumeration to represent the null value for the
 * `Value` type union.
 *
 *  The JSON representation for `NullValue` is JSON `null`.
 *
 *  - NULL_VALUE: Null value.
 * @default "NULL_VALUE"
 */
export enum ProtobufNullValue {
  NULL_VALUE = 'NULL_VALUE'
}

export interface RpcStatus {
  /** @format int32 */
  code?: number;
  message?: string;
  details?: ProtobufAny[];
}

export type V1Beta1AcceptOrganizationInvitationResponse = object;

export type V1Beta1AddGroupUsersResponse = object;

export type V1Beta1AddOrganizationUsersResponse = object;

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
  createdAt?: string;
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
  strategyName?: string;
  /** for oidc & magic links */
  state?: string;
  code?: string;
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
  grantType?: string;
  /** client_id and client_secret are required for grant_type client_credentials */
  clientId?: string;
  clientSecret?: string;
  /** assertion is required for grant_type urn:ietf:params:oauth:grant-type:jwt-bearer */
  assertion?: string;
}

export interface V1Beta1AuthTokenResponse {
  accessToken?: string;
  tokenType?: string;
}

export interface V1Beta1AuthenticateResponse {
  /** endpoint specifies the url to redirect user for starting authentication flow */
  endpoint?: string;
  /** state is used for resuming authentication flow in applicable strategies */
  state?: string;
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

export interface V1Beta1CheckResourcePermissionRequest {
  /** Deprecated. Use `resource` field instead. */
  objectId?: string;
  /** Deprecated. Use `resource` field instead. */
  objectNamespace?: string;
  /** the permission name to check. <br/> *Example:* `get`, `list`, `compute.instance.create` */
  permission: string;
  /** `namespace:uuid` or `namespace:name` of the org or project, and `namespace:urn` of a resource under a project. In case of an org/project either provide the complete namespace (app/organization) or Frontier can also parse aliases for the same as `org` or `project`. <br/> *Example:* `organization:92f69c3a-334b-4f25-90b8-4d4f3be6b825` or `app/project:project-name` or `compute/instance:92f69c3a-334b-4f25-90b8-4d4f3be6b825` */
  resource?: string;
}

export interface V1Beta1CheckResourcePermissionResponse {
  status?: boolean;
}

export interface V1Beta1CreateCurrentUserPreferencesRequest {
  bodies?: V1Beta1PreferenceRequestBody[];
}

export interface V1Beta1CreateCurrentUserPreferencesResponse {
  preferences?: V1Beta1Preference[];
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

export interface V1Beta1CreatePolicyResponse {
  policy?: V1Beta1Policy;
}

export interface V1Beta1CreatePreferencesRequest {
  preferences?: V1Beta1PreferenceRequestBody[];
}

export interface V1Beta1CreatePreferencesResponse {
  preference?: V1Beta1Preference[];
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

export interface V1Beta1CreateRelationResponse {
  relation?: V1Beta1Relation;
}

export interface V1Beta1CreateRoleResponse {
  role?: V1Beta1Role;
}

export interface V1Beta1CreateServiceUserKeyResponse {
  key?: V1Beta1KeyCredential;
}

export interface V1Beta1CreateServiceUserRequest {
  body?: V1Beta1ServiceUserRequestBody;
  /** The organization ID to which the service user belongs to. */
  orgId: string;
}

export interface V1Beta1CreateServiceUserResponse {
  serviceuser?: V1Beta1ServiceUser;
}

export interface V1Beta1CreateServiceUserSecretResponse {
  secret?: V1Beta1SecretCredential;
}

export interface V1Beta1CreateUserPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1CreateUserResponse {
  user?: V1Beta1User;
}

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

export type V1Beta1DeleteRelationResponse = object;

export type V1Beta1DeleteRoleResponse = object;

export type V1Beta1DeleteServiceUserKeyResponse = object;

export type V1Beta1DeleteServiceUserResponse = object;

export type V1Beta1DeleteServiceUserSecretResponse = object;

export type V1Beta1DeleteUserResponse = object;

export interface V1Beta1DescribePreferencesResponse {
  traits?: V1Beta1PreferenceTrait[];
}

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
  orgId?: string;
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
  createdAt?: string;
  /**
   * The time the org domain was updated
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
}

export type V1Beta1EnableGroupResponse = object;

export type V1Beta1EnableOrganizationResponse = object;

export type V1Beta1EnableProjectResponse = object;

export type V1Beta1EnableUserResponse = object;

export interface V1Beta1GetCurrentUserResponse {
  user?: V1Beta1User;
  serviceuser?: V1Beta1ServiceUser;
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

export interface V1Beta1GetOrganizationResponse {
  organization?: V1Beta1Organization;
}

export interface V1Beta1GetOrganizationRoleResponse {
  role?: V1Beta1Role;
}

export interface V1Beta1GetPermissionResponse {
  permission?: V1Beta1Permission;
}

export interface V1Beta1GetPolicyResponse {
  policy?: V1Beta1Policy;
}

export interface V1Beta1GetProjectResourceResponse {
  resource?: V1Beta1Resource;
}

export interface V1Beta1GetProjectResponse {
  project?: V1Beta1Project;
}

export interface V1Beta1GetRelationResponse {
  relation?: V1Beta1Relation;
}

export interface V1Beta1GetServiceUserKeyResponse {
  keys?: V1Beta1JSONWebKey[];
}

export interface V1Beta1GetServiceUserResponse {
  serviceuser?: V1Beta1ServiceUser;
}

export interface V1Beta1GetUserResponse {
  user?: V1Beta1User;
}

export interface V1Beta1Group {
  id?: string;
  name?: string;
  title?: string;
  orgId?: string;
  metadata?: object;
  /**
   * The time the group was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  createdAt?: string;
  /**
   * The time the group was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
}

export interface V1Beta1GroupRequestBody {
  /** The name of the group. The name must be unique within the entire Frontier instance. The name can contain only alphanumeric characters, dashes and underscores. */
  name: string;
  /** The title can contain any UTF-8 character, used to provide a human-readable name for the group. Can also be left empty. */
  title?: string;
  /** Metadata object for groups that can hold key value pairs defined in Group Metaschema. The metadata object can be used to store arbitrary information about the group such as labels, descriptions etc. The default Group Metaschema contains labels and descripton fields. Update the Group Metaschema to add more fields.<br/>*Example:*`{"labels": {"key": "value"}, "description": "Group description"}` */
  metadata?: object;
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
  userId?: string;
  /**
   * The organization id to which the user is invited.
   * @example "b9c4f4e2-9b9a-4c1a-8f1a-2b9b9b9b9b9b"
   */
  orgId?: string;
  /**
   * The list of group ids to which the user is invited.
   * @example "c9c4f4e2-9b9a-4c1a-8f1a-2b9b9b9b9b9b"
   */
  groupIds?: string[];
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
  createdAt?: string;
  /**
   * The time when the invitation expires.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  expiresAt?: string;
  /**
   * The list of role ids to which the user is invited in an organization.
   * @example "d9c4f4e2-9b9a-4c1a-8f1a-2b9b9b9b9b9b"
   */
  roleIds?: string[];
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
  principalId?: string;
  /** RSA private key as string */
  privateKey?: string;
}

export interface V1Beta1ListAllOrganizationsResponse {
  organizations?: V1Beta1Organization[];
}

export interface V1Beta1ListAllUsersResponse {
  /** @format int32 */
  count?: number;
  users?: V1Beta1User[];
}

export interface V1Beta1ListAuthStrategiesResponse {
  strategies?: V1Beta1AuthStrategy[];
}

export interface V1Beta1ListCurrentUserGroupsResponse {
  groups?: V1Beta1Group[];
  accessPairs?: V1Beta1ListCurrentUserGroupsResponseAccessPair[];
}

export interface V1Beta1ListCurrentUserGroupsResponseAccessPair {
  groupId?: string;
  permissions?: string[];
}

export interface V1Beta1ListCurrentUserInvitationsResponse {
  invitations?: V1Beta1Invitation[];
  orgs?: V1Beta1Organization[];
}

export interface V1Beta1ListCurrentUserPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1ListGroupPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1ListGroupUsersResponse {
  users?: V1Beta1User[];
  rolePairs?: V1Beta1ListGroupUsersResponseRolePair[];
}

export interface V1Beta1ListGroupUsersResponseRolePair {
  userId?: string;
  roles?: V1Beta1Role[];
}

export interface V1Beta1ListGroupsResponse {
  groups?: V1Beta1Group[];
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
}

export interface V1Beta1ListOrganizationsByCurrentUserResponse {
  organizations?: V1Beta1Organization[];
  joinableViaDomain?: V1Beta1Organization[];
}

export interface V1Beta1ListOrganizationsByUserResponse {
  organizations?: V1Beta1Organization[];
  joinableViaDomain?: V1Beta1Organization[];
}

export interface V1Beta1ListOrganizationsResponse {
  organizations?: V1Beta1Organization[];
}

export interface V1Beta1ListPermissionsResponse {
  permissions?: V1Beta1Permission[];
}

export interface V1Beta1ListPoliciesResponse {
  policies?: V1Beta1Policy[];
}

export interface V1Beta1ListPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1ListProjectAdminsResponse {
  users?: V1Beta1User[];
}

export interface V1Beta1ListProjectGroupsResponse {
  groups?: V1Beta1Group[];
}

export interface V1Beta1ListProjectPreferencesResponse {
  preferences?: V1Beta1Preference[];
}

export interface V1Beta1ListProjectResourcesResponse {
  resources?: V1Beta1Resource[];
}

export interface V1Beta1ListProjectServiceUsersResponse {
  serviceusers?: V1Beta1ServiceUser[];
  rolePairs?: V1Beta1ListProjectServiceUsersResponseRolePair[];
}

export interface V1Beta1ListProjectServiceUsersResponseRolePair {
  serviceuserId?: string;
  roles?: V1Beta1Role[];
}

export interface V1Beta1ListProjectUsersResponse {
  users?: V1Beta1User[];
  rolePairs?: V1Beta1ListProjectUsersResponseRolePair[];
}

export interface V1Beta1ListProjectUsersResponseRolePair {
  userId?: string;
  roles?: V1Beta1Role[];
}

export interface V1Beta1ListProjectsByCurrentUserResponse {
  projects?: V1Beta1Project[];
  accessPairs?: V1Beta1ListProjectsByCurrentUserResponseAccessPair[];
}

export interface V1Beta1ListProjectsByCurrentUserResponseAccessPair {
  projectId?: string;
  permissions?: string[];
}

export interface V1Beta1ListProjectsByUserResponse {
  projects?: V1Beta1Project[];
}

export interface V1Beta1ListProjectsResponse {
  projects?: V1Beta1Project[];
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

export interface V1Beta1ListServiceUserKeysResponse {
  keys?: V1Beta1ServiceUserKey[];
}

export interface V1Beta1ListServiceUserSecretsResponse {
  /** secrets will be listed without the secret value */
  secrets?: V1Beta1SecretCredential[];
}

export interface V1Beta1ListServiceUsersResponse {
  serviceusers?: V1Beta1ServiceUser[];
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
  createdAt?: string;
  /**
   * The time when the metaschema was updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
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
  createdAt?: string;
  /**
   * The time the namespace was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
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
  createdAt?: string;
  /**
   * The time the organization was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
  /**
   * The state of the organization (enabled or disabled).
   * @example "enabled"
   */
  state?: string;
  /** The base64 encoded image string of the organization avatar. Should be less than 2MB. */
  avatar?: string;
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

export interface V1Beta1Permission {
  id?: string;
  name?: string;
  title?: string;
  /**
   * The time the permission was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  createdAt?: string;
  /**
   * The time the permission was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
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

export interface V1Beta1Policy {
  id?: string;
  title?: string;
  /**
   * The time the policy was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  createdAt?: string;
  /**
   * The time the policy was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
  roleId?: string;
  /** namespace:uuid */
  resource?: string;
  /** namespace:uuid */
  principal?: string;
  metadata?: object;
}

export interface V1Beta1PolicyRequestBody {
  /** unique id of the role to which policy is assigned */
  roleId: string;
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
  resourceId?: string;
  resourceType?: string;
  /**
   * The time when the preference was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  createdAt?: string;
  /**
   * The time when the preference was updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
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
  resourceType?: string;
  name?: string;
  title?: string;
  description?: string;
  longDescription?: string;
  heading?: string;
  subHeading?: string;
  breadcrumb?: string;
  inputHints?: string;
  text?: string;
  textarea?: string;
  select?: string;
  combobox?: string;
  checkbox?: string;
  multiselect?: string;
  number?: string;
}

export interface V1Beta1Project {
  id?: string;
  name?: string;
  title?: string;
  orgId?: string;
  metadata?: object;
  /**
   * The time the project was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  createdAt?: string;
  /**
   * The time the project was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
}

export interface V1Beta1ProjectRequestBody {
  /** The name of the project. The name must be unique within the entire Frontier instance. The name can contain only alphanumeric characters, dashes and underscores.<br/> *Example:* `frontier-playground` */
  name: string;
  /** The title can contain any UTF-8 character, used to provide a human-readable name for the project. Can also be left empty. <br/> *Example:* `Frontier Playground` */
  title?: string;
  /** Metadata object for projects that can hold key value pairs defined in Project Metaschema. */
  metadata?: object;
  /** unique id of the organization to which project belongs */
  orgId: string;
}

export interface V1Beta1Relation {
  id?: string;
  /**
   * The time the relation was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  createdAt?: string;
  /**
   * The time the relation was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
  subjectSubRelation?: string;
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
  subjectSubRelation?: string;
}

export type V1Beta1RemoveGroupUserResponse = object;

export type V1Beta1RemoveOrganizationUserResponse = object;

export interface V1Beta1Resource {
  id?: string;
  /** Name of the resource. Must be unique within the project. */
  name?: string;
  /**
   * The time the resource was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  createdAt?: string;
  /**
   * The time the resource was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
  urn?: string;
  projectId?: string;
  namespace?: string;
  principal?: string;
  metadata?: object;
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
  createdAt?: string;
  /**
   * The time the role was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
  orgId?: string;
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
  createdAt?: string;
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
  createdAt?: string;
  /**
   * The time the user was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
  state?: string;
  orgId?: string;
}

export interface V1Beta1ServiceUserKey {
  id?: string;
  title?: string;
  principalId?: string;
  publicKey?: string;
  /**
   * The time when the secret was created.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  createdAt?: string;
}

export interface V1Beta1ServiceUserRequestBody {
  /**
   * User friendly name of the service user.
   * @example "Order Service"
   */
  title?: string;
  metadata?: object;
}

export interface V1Beta1UpdateCurrentUserResponse {
  user?: V1Beta1User;
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

export interface V1Beta1UpdatePolicyResponse {
  policies?: V1Beta1Policy[];
}

export interface V1Beta1UpdateProjectResourceResponse {
  resource?: V1Beta1Resource;
}

export interface V1Beta1UpdateProjectResponse {
  project?: V1Beta1Project;
}

export interface V1Beta1UpdateRoleResponse {
  role?: V1Beta1Role;
}

export interface V1Beta1UpdateUserResponse {
  user?: V1Beta1User;
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
  createdAt?: string;
  /**
   * The time the user was last updated.
   * @format date-time
   * @example "2023-06-07T05:39:56.961Z"
   */
  updatedAt?: string;
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
