# Shield API

## Version: 0.1.0

### /v1beta1/actions

#### GET
##### Summary

Get all Actions

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListActionsResponse](#v1beta1listactionsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Action

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1ActionRequestBody](#v1beta1actionrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateActionResponse](#v1beta1createactionresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/actions/{id}

#### GET
##### Summary

Get Action by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetActionResponse](#v1beta1getactionresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Action by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1ActionRequestBody](#v1beta1actionrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateActionResponse](#v1beta1updateactionresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/check

#### POST
##### Summary

check permission for action on a resource by an user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1CheckResourcePermissionRequest](#v1beta1checkresourcepermissionrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CheckResourcePermissionResponse](#v1beta1checkresourcepermissionresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/groups

#### GET
##### Summary

Get all Groups

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| userId | query |  | No | string |
| orgId | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListGroupsResponse](#v1beta1listgroupsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1GroupRequestBody](#v1beta1grouprequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateGroupResponse](#v1beta1creategroupresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/groups/{id}

#### GET
##### Summary

Get Group by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetGroupResponse](#v1beta1getgroupresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Group by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1GroupRequestBody](#v1beta1grouprequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateGroupResponse](#v1beta1updategroupresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/groups/{id}/admins

#### GET
##### Summary

Get all Admins of a Group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListGroupAdminsResponse](#v1beta1listgroupadminsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Add Admin to Group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1AddGroupAdminRequestBody](#v1beta1addgroupadminrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AddGroupAdminResponse](#v1beta1addgroupadminresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/groups/{id}/admins/{userId}

#### DELETE
##### Summary

Remove Admin from Group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| userId | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1RemoveGroupAdminResponse](#v1beta1removegroupadminresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/groups/{id}/relations

#### GET
##### Summary

Get all relations for a group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| subjectType | query |  | No | string |
| role | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListGroupRelationsResponse](#v1beta1listgrouprelationsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/groups/{id}/users

#### GET
##### Summary

Get all Users in a Group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListGroupUsersResponse](#v1beta1listgroupusersresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Add User to Group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1AddGroupUserRequestBody](#v1beta1addgroupuserrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AddGroupUserResponse](#v1beta1addgroupuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/groups/{id}/users/{userId}

#### DELETE
##### Summary

Remove User from Group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| userId | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1RemoveGroupUserResponse](#v1beta1removegroupuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/metadatakey

#### POST
##### Summary

Create Metadata Key

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1MetadataKeyRequestBody](#v1beta1metadatakeyrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateMetadataKeyResponse](#v1beta1createmetadatakeyresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/namespaces

#### GET
##### Summary

Get all Namespaces

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListNamespacesResponse](#v1beta1listnamespacesresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Namespace

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1NamespaceRequestBody](#v1beta1namespacerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateNamespaceResponse](#v1beta1createnamespaceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/namespaces/{id}

#### GET
##### Summary

Get Namespace by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetNamespaceResponse](#v1beta1getnamespaceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Namespace by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1NamespaceRequestBody](#v1beta1namespacerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateNamespaceResponse](#v1beta1updatenamespaceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/object/{objectId}/subject/{subjectId}/role/{role}

#### DELETE
##### Summary

Remove a subject having a role from an object

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| objectId | path |  | Yes | string |
| subjectId | path |  | Yes | string |
| role | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteRelationResponse](#v1beta1deleterelationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations

#### GET
##### Summary

Get all Organization

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationsResponse](#v1beta1listorganizationsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1OrganizationRequestBody](#v1beta1organizationrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateOrganizationResponse](#v1beta1createorganizationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}

#### GET
##### Summary

Get Organization by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetOrganizationResponse](#v1beta1getorganizationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Organization by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1OrganizationRequestBody](#v1beta1organizationrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateOrganizationResponse](#v1beta1updateorganizationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}/admins

#### GET
##### Summary

Get all Admins of an Organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationAdminsResponse](#v1beta1listorganizationadminsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Add Admin to Organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1AddOrganizationAdminRequestBody](#v1beta1addorganizationadminrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AddOrganizationAdminResponse](#v1beta1addorganizationadminresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}/admins/{userId}

#### DELETE
##### Summary

Remove Admin from Organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| userId | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1RemoveOrganizationAdminResponse](#v1beta1removeorganizationadminresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/policies

#### GET
##### Summary

Get all Policy

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListPoliciesResponse](#v1beta1listpoliciesresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Policy

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1PolicyRequestBody](#v1beta1policyrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreatePolicyResponse](#v1beta1createpolicyresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/policies/{id}

#### GET
##### Summary

Get Policy by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetPolicyResponse](#v1beta1getpolicyresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Policy by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1PolicyRequestBody](#v1beta1policyrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdatePolicyResponse](#v1beta1updatepolicyresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects

#### GET
##### Summary

Get all Project

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListProjectsResponse](#v1beta1listprojectsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Project

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1ProjectRequestBody](#v1beta1projectrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateProjectResponse](#v1beta1createprojectresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{id}

#### GET
##### Summary

Get Project by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetProjectResponse](#v1beta1getprojectresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Project by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1ProjectRequestBody](#v1beta1projectrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateProjectResponse](#v1beta1updateprojectresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{id}/admins

#### GET
##### Summary

Get all Admins of a Project

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListProjectAdminsResponse](#v1beta1listprojectadminsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Add Admin to Project

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1AddProjectAdminRequestBody](#v1beta1addprojectadminrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AddProjectAdminResponse](#v1beta1addprojectadminresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{id}/admins/{userId}

#### DELETE
##### Summary

Remove Admin from Project

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| userId | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1RemoveProjectAdminResponse](#v1beta1removeprojectadminresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/relations

#### GET
##### Summary

Get all Relations

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListRelationsResponse](#v1beta1listrelationsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Relation

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1RelationRequestBody](#v1beta1relationrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateRelationResponse](#v1beta1createrelationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/relations/{id}

#### GET
##### Summary

Get Relation by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetRelationResponse](#v1beta1getrelationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/resources

#### GET
##### Summary

Get all Resources

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| groupId | query |  | No | string |
| projectId | query |  | No | string |
| organizationId | query |  | No | string |
| namespaceId | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListResourcesResponse](#v1beta1listresourcesresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Resource

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1ResourceRequestBody](#v1beta1resourcerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateResourceResponse](#v1beta1createresourceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/resources/{id}

#### GET
##### Summary

Get Resource by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetResourceResponse](#v1beta1getresourceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Resource by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1ResourceRequestBody](#v1beta1resourcerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateResourceResponse](#v1beta1updateresourceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/roles

#### GET
##### Summary

Get all Roles

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListRolesResponse](#v1beta1listrolesresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Role

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1RoleRequestBody](#v1beta1rolerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateRoleResponse](#v1beta1createroleresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/roles/{id}

#### GET
##### Summary

Get Role by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetRoleResponse](#v1beta1getroleresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Role by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1RoleRequestBody](#v1beta1rolerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateRoleResponse](#v1beta1updateroleresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users

#### GET
##### Summary

Get All Users

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| pageSize | query |  | No | integer |
| pageNum | query |  | No | integer |
| keyword | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListUsersResponse](#v1beta1listusersresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create User

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1UserRequestBody](#v1beta1userrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateUserResponse](#v1beta1createuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/self

#### GET
##### Summary

Get current user

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetCurrentUserResponse](#v1beta1getcurrentuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update current User

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body |  | Yes | [v1beta1UserRequestBody](#v1beta1userrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateCurrentUserResponse](#v1beta1updatecurrentuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}

#### GET
##### Summary

Get a User by id

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetUserResponse](#v1beta1getuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update User by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1UserRequestBody](#v1beta1userrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateUserResponse](#v1beta1updateuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}/groups

#### GET
##### Summary

List Groups of a User

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| id | path |  | Yes | string |
| role | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListUserGroupsResponse](#v1beta1listusergroupsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### Models

#### protobufAny

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| @type | string |  | No |

#### protobufNullValue

`NullValue` is a singleton enumeration to represent the null value for the
`Value` type union.

 The JSON representation for `NullValue` is JSON `null`.

- NULL_VALUE: Null value.

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| protobufNullValue | string | `NullValue` is a singleton enumeration to represent the null value for the `Value` type union.   The JSON representation for `NullValue` is JSON `null`.   - NULL_VALUE: Null value. |  |

#### rpcStatus

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| code | integer |  | No |
| message | string |  | No |
| details | [ [protobufAny](#protobufany) ] |  | No |

#### v1beta1Action

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| namespace | [v1beta1Namespace](#v1beta1namespace) |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| namespaceId | string |  | No |

#### v1beta1ActionRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| namespaceId | string |  | No |

#### v1beta1AddGroupAdminRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| userIds | [ string ] |  | No |

#### v1beta1AddGroupAdminResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1AddGroupUserRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| userIds | [ string ] |  | No |

#### v1beta1AddGroupUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1AddOrganizationAdminRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| userIds | [ string ] |  | No |

#### v1beta1AddOrganizationAdminResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1AddProjectAdminRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| userIds | [ string ] |  | No |

#### v1beta1AddProjectAdminResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1CheckResourcePermissionRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| objectId | string |  | No |
| objectNamespace | string |  | No |
| permission | string |  | No |

#### v1beta1CheckResourcePermissionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| status | boolean |  | No |

#### v1beta1CreateActionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| action | [v1beta1Action](#v1beta1action) |  | No |

#### v1beta1CreateGroupResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| group | [v1beta1Group](#v1beta1group) |  | No |

#### v1beta1CreateMetadataKeyResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| metadatakey | [v1beta1MetadataKey](#v1beta1metadatakey) |  | No |

#### v1beta1CreateNamespaceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| namespace | [v1beta1Namespace](#v1beta1namespace) |  | No |

#### v1beta1CreateOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organization | [v1beta1Organization](#v1beta1organization) |  | No |

#### v1beta1CreatePolicyResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| policies | [ [v1beta1Policy](#v1beta1policy) ] |  | No |

#### v1beta1CreateProjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| project | [v1beta1Project](#v1beta1project) |  | No |

#### v1beta1CreateRelationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| relation | [v1beta1Relation](#v1beta1relation) |  | No |

#### v1beta1CreateResourceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| resource | [v1beta1Resource](#v1beta1resource) |  | No |

#### v1beta1CreateRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| role | [v1beta1Role](#v1beta1role) |  | No |

#### v1beta1CreateUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [v1beta1User](#v1beta1user) |  | No |

#### v1beta1DeleteRelationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |

#### v1beta1GetActionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| action | [v1beta1Action](#v1beta1action) |  | No |

#### v1beta1GetCurrentUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [v1beta1User](#v1beta1user) |  | No |

#### v1beta1GetGroupResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| group | [v1beta1Group](#v1beta1group) |  | No |

#### v1beta1GetNamespaceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| namespace | [v1beta1Namespace](#v1beta1namespace) |  | No |

#### v1beta1GetOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organization | [v1beta1Organization](#v1beta1organization) |  | No |

#### v1beta1GetPolicyResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| policy | [v1beta1Policy](#v1beta1policy) |  | No |

#### v1beta1GetProjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| project | [v1beta1Project](#v1beta1project) |  | No |

#### v1beta1GetRelationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| relation | [v1beta1Relation](#v1beta1relation) |  | No |

#### v1beta1GetResourceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| resource | [v1beta1Resource](#v1beta1resource) |  | No |

#### v1beta1GetRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| role | [v1beta1Role](#v1beta1role) |  | No |

#### v1beta1GetUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [v1beta1User](#v1beta1user) |  | No |

#### v1beta1Group

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| slug | string |  | No |
| orgId | string |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1GroupRelation

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| subjectType | string |  | No |
| role | string |  | No |
| user | [v1beta1User](#v1beta1user) |  | No |
| group | [v1beta1Group](#v1beta1group) |  | No |

#### v1beta1GroupRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| slug | string |  | No |
| metadata | object |  | No |
| orgId | string |  | No |

#### v1beta1ListActionsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| actions | [ [v1beta1Action](#v1beta1action) ] |  | No |

#### v1beta1ListGroupAdminsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1ListGroupRelationsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| relations | [ [v1beta1GroupRelation](#v1beta1grouprelation) ] |  | No |

#### v1beta1ListGroupUsersResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1ListGroupsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| groups | [ [v1beta1Group](#v1beta1group) ] |  | No |

#### v1beta1ListNamespacesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| namespaces | [ [v1beta1Namespace](#v1beta1namespace) ] |  | No |

#### v1beta1ListOrganizationAdminsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1ListOrganizationsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organizations | [ [v1beta1Organization](#v1beta1organization) ] |  | No |

#### v1beta1ListPoliciesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| policies | [ [v1beta1Policy](#v1beta1policy) ] |  | No |

#### v1beta1ListProjectAdminsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1ListProjectsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| projects | [ [v1beta1Project](#v1beta1project) ] |  | No |

#### v1beta1ListRelationsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| relations | [ [v1beta1Relation](#v1beta1relation) ] |  | No |

#### v1beta1ListResourcesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| resources | [ [v1beta1Resource](#v1beta1resource) ] |  | No |

#### v1beta1ListRolesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| roles | [ [v1beta1Role](#v1beta1role) ] |  | No |

#### v1beta1ListUserGroupsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| groups | [ [v1beta1Group](#v1beta1group) ] |  | No |

#### v1beta1ListUsersResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| count | integer |  | No |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1MetadataKey

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| key | string |  | No |
| description | string |  | No |

#### v1beta1MetadataKeyRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| key | string |  | No |
| description | string |  | No |

#### v1beta1Namespace

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1NamespaceRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |

#### v1beta1Organization

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| slug | string |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1OrganizationRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| slug | string |  | No |
| metadata | object |  | No |

#### v1beta1Policy

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| role | [v1beta1Role](#v1beta1role) |  | No |
| action | [v1beta1Action](#v1beta1action) |  | No |
| namespace | [v1beta1Namespace](#v1beta1namespace) |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| namespaceId | string |  | No |
| roleId | string |  | No |
| actionId | string |  | No |

#### v1beta1PolicyRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| roleId | string |  | No |
| actionId | string |  | No |
| namespaceId | string |  | No |

#### v1beta1Project

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| slug | string |  | No |
| orgId | string |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1ProjectRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| slug | string |  | No |
| metadata | object |  | No |
| orgId | string |  | No |

#### v1beta1Relation

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| objectId | string |  | No |
| objectNamespace | string |  | No |
| subject | string |  | No |
| roleName | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1RelationRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| objectId | string |  | No |
| objectNamespace | string |  | No |
| subject | string |  | No |
| roleName | string |  | No |

#### v1beta1RemoveGroupAdminResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |

#### v1beta1RemoveGroupUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |

#### v1beta1RemoveOrganizationAdminResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |

#### v1beta1RemoveProjectAdminResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |

#### v1beta1Resource

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| project | [v1beta1Project](#v1beta1project) |  | No |
| organization | [v1beta1Organization](#v1beta1organization) |  | No |
| namespace | [v1beta1Namespace](#v1beta1namespace) |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| user | [v1beta1User](#v1beta1user) |  | No |
| urn | string |  | No |

#### v1beta1ResourceRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| projectId | string |  | No |
| namespaceId | string |  | No |
| relations | [ [v1beta1Relation](#v1beta1relation) ] |  | No |

#### v1beta1Role

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| types | [ string ] |  | No |
| namespace | [v1beta1Namespace](#v1beta1namespace) |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| namespaceId | string |  | No |

#### v1beta1RoleRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| types | [ string ] |  | No |
| namespaceId | string |  | No |
| metadata | object |  | No |

#### v1beta1UpdateActionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| action | [v1beta1Action](#v1beta1action) |  | No |

#### v1beta1UpdateCurrentUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [v1beta1User](#v1beta1user) |  | No |

#### v1beta1UpdateGroupResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| group | [v1beta1Group](#v1beta1group) |  | No |

#### v1beta1UpdateNamespaceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| namespace | [v1beta1Namespace](#v1beta1namespace) |  | No |

#### v1beta1UpdateOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organization | [v1beta1Organization](#v1beta1organization) |  | No |

#### v1beta1UpdatePolicyResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| policies | [ [v1beta1Policy](#v1beta1policy) ] |  | No |

#### v1beta1UpdateProjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| project | [v1beta1Project](#v1beta1project) |  | No |

#### v1beta1UpdateResourceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| resource | [v1beta1Resource](#v1beta1resource) |  | No |

#### v1beta1UpdateRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| role | [v1beta1Role](#v1beta1role) |  | No |

#### v1beta1UpdateUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [v1beta1User](#v1beta1user) |  | No |

#### v1beta1User

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| slug | string |  | No |
| email | string |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1UserRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| email | string |  | No |
| metadata | object |  | No |
