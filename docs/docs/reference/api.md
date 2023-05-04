# Shield Administration API
## Version: 0.2.0

## default

### /v1beta1/admin/groups

#### GET
##### Summary

Get all groups

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| userId | query |  | No | string |
| orgId | query |  | No | string |
| state | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListGroupsResponse](#v1beta1listgroupsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{body.orgId}/groups

#### POST
##### Summary

Create Group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body.orgId | path |  | Yes | string |
| body | body |  | Yes | { **"name"**: string, **"slug"**: string, **"metadata"**: object } |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateGroupResponse](#v1beta1creategroupresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{body.orgId}/groups/{id}

#### PUT
##### Summary

Update Group by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body.orgId | path |  | Yes | string |
| id | path |  | Yes | string |
| body | body |  | Yes | { **"name"**: string, **"slug"**: string, **"metadata"**: object } |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateGroupResponse](#v1beta1updategroupresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/groups

#### GET
##### Summary

Create Group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| userId | query |  | No | string |
| state | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationGroupsResponse](#v1beta1listorganizationgroupsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/groups/{id}

#### GET
##### Summary

Get Group by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetGroupResponse](#v1beta1getgroupresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete a Group permanently forever and all of its relations

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteGroupResponse](#v1beta1deletegroupresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/groups/{id}/disable

#### POST
##### Summary

Disable a Group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DisableGroupResponse](#v1beta1disablegroupresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/groups/{id}/enable

#### POST
##### Summary

Enable a Group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1EnableGroupResponse](#v1beta1enablegroupresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/groups/{id}/users

#### GET
##### Summary

Get all users for a group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListGroupUsersResponse](#v1beta1listgroupusersresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## default

### /v1beta1/admin/organizations

#### GET
##### Summary

Get all organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| userId | query |  | No | string |
| state | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListAllOrganizationsResponse](#v1beta1listallorganizationsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations

#### GET
##### Summary

Get all organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| userId | query |  | No | string |
| state | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationsResponse](#v1beta1listorganizationsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1OrganizationRequestBody](#v1beta1organizationrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateOrganizationResponse](#v1beta1createorganizationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}

#### GET
##### Summary

Get Organization by ID or slug

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetOrganizationResponse](#v1beta1getorganizationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete an organization permanently forever and all of its relations

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteOrganizationResponse](#v1beta1deleteorganizationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Organization by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
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
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationAdminsResponse](#v1beta1listorganizationadminsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}/disable

#### POST
##### Summary

Disable an organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DisableOrganizationResponse](#v1beta1disableorganizationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}/enable

#### POST
##### Summary

Enable an Organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1EnableOrganizationResponse](#v1beta1enableorganizationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}/projects

#### GET
##### Summary

Get all projects that belong to an organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| state | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationProjectsResponse](#v1beta1listorganizationprojectsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}/users

#### GET
##### Summary

List Organization users

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| permissionFilter | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationUsersResponse](#v1beta1listorganizationusersresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## default

### /v1beta1/admin/projects

#### GET
##### Summary

Get all project

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | query |  | No | string |
| state | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListProjectsResponse](#v1beta1listprojectsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects

#### POST
##### Summary

Create Project

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
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
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetProjectResponse](#v1beta1getprojectresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete a Project permanently forever and all of its relations

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteProjectResponse](#v1beta1deleteprojectresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Project by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
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
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListProjectAdminsResponse](#v1beta1listprojectadminsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{id}/disable

#### POST
##### Summary

Disable a Project

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DisableProjectResponse](#v1beta1disableprojectresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{id}/enable

#### POST
##### Summary

Enable a Project

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1EnableProjectResponse](#v1beta1enableprojectresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{id}/users

#### GET
##### Summary

Get all users of a Project

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| permissionFilter | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListProjectUsersResponse](#v1beta1listprojectusersresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## default

### /v1beta1/admin/relations

#### GET
##### Summary

Get all relations

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListRelationsResponse](#v1beta1listrelationsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/object/{objectNamespace}/{objectId}/subject/{subjectNamespace}/{subjectId}/relation/{relation}

#### DELETE
##### Summary

Remove a subject having a relation from an object

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| objectNamespace | path |  | Yes | string |
| objectId | path |  | Yes | string |
| subjectNamespace | path |  | Yes | string |
| subjectId | path |  | Yes | string |
| relation | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteRelationResponse](#v1beta1deleterelationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/relations

#### POST
##### Summary

Create Relation

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
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
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetRelationResponse](#v1beta1getrelationresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## default

### /v1beta1/admin/resources

#### GET
##### Summary

Get all resources

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| groupId | query |  | No | string |
| projectId | query |  | No | string |
| organizationId | query |  | No | string |
| namespaceId | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListResourcesResponse](#v1beta1listresourcesresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{body.projectId}/resources

#### POST
##### Summary

Create Resource

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body.projectId | path |  | Yes | string |
| body | body |  | Yes | { **"name"**: string, **"namespaceId"**: string, **"relations"**: [ [v1beta1Relation](#v1beta1relation) ], **"userId"**: string, **"metadata"**: object } |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateResourceResponse](#v1beta1createresourceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{body.projectId}/resources/{id}

#### PUT
##### Summary

Update Resource by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body.projectId | path |  | Yes | string |
| id | path |  | Yes | string |
| body | body |  | Yes | { **"name"**: string, **"namespaceId"**: string, **"relations"**: [ [v1beta1Relation](#v1beta1relation) ], **"userId"**: string, **"metadata"**: object } |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateResourceResponse](#v1beta1updateresourceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{projectId}/resources

#### GET
##### Summary

Get all resources

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| projectId | path |  | Yes | string |
| namespaceId | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListProjectResourcesResponse](#v1beta1listprojectresourcesresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{projectId}/resources/{id}

#### GET
##### Summary

Get Resource by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| projectId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetResourceResponse](#v1beta1getresourceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete a resource permanently forever

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| projectId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteResourceResponse](#v1beta1deleteresourceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## default

### /v1beta1/admin/users

#### GET
##### Summary

Get all users

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| pageSize | query |  | No | integer |
| pageNum | query |  | No | integer |
| keyword | query |  | No | string |
| orgId | query |  | No | string |
| groupId | query |  | No | string |
| state | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListAllUsersResponse](#v1beta1listallusersresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users

#### GET
##### Summary

Get all public users

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| pageSize | query |  | No | integer |
| pageNum | query |  | No | integer |
| keyword | query |  | No | string |
| orgId | query |  | No | string |
| groupId | query |  | No | string |
| state | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListUsersResponse](#v1beta1listusersresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
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
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1UserRequestBody](#v1beta1userrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateCurrentUserResponse](#v1beta1updatecurrentuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/self/groups

#### GET
##### Summary

List groups of a User

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListCurrentUserGroupsResponse](#v1beta1listcurrentusergroupsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/self/organizations

#### GET
##### Summary

Get all organizations a user belongs to

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetOrganizationsByCurrentUserResponse](#v1beta1getorganizationsbycurrentuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}

#### GET
##### Summary

Get a user by id

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetUserResponse](#v1beta1getuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete an user permanently forever and all of its relations

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteUserResponse](#v1beta1deleteuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update User by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1UserRequestBody](#v1beta1userrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateUserResponse](#v1beta1updateuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}/disable

#### POST
##### Summary

Disable a user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DisableUserResponse](#v1beta1disableuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}/enable

#### POST
##### Summary

Enable a user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1EnableUserResponse](#v1beta1enableuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}/groups

#### GET
##### Summary

List Groups of a User

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| role | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListUserGroupsResponse](#v1beta1listusergroupsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}/organizations

#### GET
##### Summary

Get all organizations a user belongs to

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetOrganizationsByUserResponse](#v1beta1getorganizationsbyuserresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## ShieldService

### /v1beta1/auth

#### GET
##### Summary

Authn

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListAuthStrategiesResponse](#v1beta1listauthstrategiesresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/auth/callback

#### GET
##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| strategyName | query | strategy_name will not be set for oidc but can be utilized for methods like email magic links | No | string |
| state | query | for oidc | No | string |
| code | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthCallbackResponse](#v1beta1authcallbackresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| strategyName | query | strategy_name will not be set for oidc but can be utilized for methods like email magic links | No | string |
| state | query | for oidc | No | string |
| code | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthCallbackResponse](#v1beta1authcallbackresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/auth/logout

#### GET
##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthLogoutResponse](#v1beta1authlogoutresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthLogoutResponse](#v1beta1authlogoutresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/auth/register/{strategyName}

#### GET
##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| strategyName | path |  | Yes | string |
| redirect | query | by default, location header for redirect if applicable will be skipped unless this is set to true, useful in browser | No | boolean |
| returnTo | query | be default, after successful authentication no operation will be performed to apply redirection in case of browsers, provide a url that will be used for redirection after authentication | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthenticateResponse](#v1beta1authenticateresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| strategyName | path |  | Yes | string |
| redirect | query | by default, location header for redirect if applicable will be skipped unless this is set to true, useful in browser | No | boolean |
| returnTo | query | be default, after successful authentication no operation will be performed to apply redirection in case of browsers, provide a url that will be used for redirection after authentication | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthenticateResponse](#v1beta1authenticateresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## default

### /v1beta1/check

#### POST
##### Summary

check permission on a resource of an user

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1CheckResourcePermissionRequest](#v1beta1checkresourcepermissionrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CheckResourcePermissionResponse](#v1beta1checkresourcepermissionresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## default

### /v1beta1/meta/schemas

#### GET
##### Summary

Get all Metadata Schemas

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListMetaSchemasResponse](#v1beta1listmetaschemasresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Metadata Schema

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1MetaSchemaRequestBody](#v1beta1metaschemarequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateMetaSchemaResponse](#v1beta1createmetaschemaresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/meta/schemas/{id}

#### GET
##### Summary

Get MetaSchema by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetMetaSchemaResponse](#v1beta1getmetaschemaresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete a MetaSchema permanently

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteMetaSchemaResponse](#v1beta1deletemetaschemaresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update MetaSchema by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1MetaSchemaRequestBody](#v1beta1metaschemarequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateMetaSchemaResponse](#v1beta1updatemetaschemaresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## default

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
| ---- | ---------- | ----------- | -------- | ------ |
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
| ---- | ---------- | ----------- | -------- | ------ |
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
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1NamespaceRequestBody](#v1beta1namespacerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateNamespaceResponse](#v1beta1updatenamespaceresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## default

### /v1beta1/organizations/{body.orgId}/roles

#### POST
##### Summary

Create Role

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body.orgId | path |  | Yes | string |
| body | body |  | Yes | { **"id"**: string, **"name"**: string, **"permissions"**: [ string ], **"metadata"**: object } |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateOrganizationRoleResponse](#v1beta1createorganizationroleresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{body.orgId}/roles/{id}

#### PUT
##### Summary

Update Role by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body.orgId | path |  | Yes | string |
| id | path |  | Yes | string |
| body | body |  | Yes | { **"id"**: string, **"name"**: string, **"permissions"**: [ string ], **"metadata"**: object } |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateOrganizationRoleResponse](#v1beta1updateorganizationroleresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/roles

#### GET
##### Summary

Get custom roles under an org

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationRolesResponse](#v1beta1listorganizationrolesresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/roles/{id}

#### GET
##### Summary

Get Role by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetOrganizationRoleResponse](#v1beta1getorganizationroleresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete a role permanently forever and all of its relations

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteOrganizationRoleResponse](#v1beta1deleteorganizationroleresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/roles

#### GET
##### Summary

List all pre-defined roles

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListRolesResponse](#v1beta1listrolesresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create platform wide role

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1RoleRequestBody](#v1beta1rolerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateRoleResponse](#v1beta1createroleresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/roles/{id}

#### DELETE
##### Summary

Delete a role permanently forever and all of its relations

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteRoleResponse](#v1beta1deleteroleresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## default

### /v1beta1/permissions

#### GET
##### Summary

Get all Permissions

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListPermissionsResponse](#v1beta1listpermissionsresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Permission

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1PermissionRequestBody](#v1beta1permissionrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreatePermissionResponse](#v1beta1createpermissionresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/permissions/{id}

#### GET
##### Summary

Get Permission by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetPermissionResponse](#v1beta1getpermissionresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete Permission by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeletePermissionResponse](#v1beta1deletepermissionresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Permission by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1PermissionRequestBody](#v1beta1permissionrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdatePermissionResponse](#v1beta1updatepermissionresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## default

### /v1beta1/policies

#### GET
##### Summary

Get all policies

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | query |  | No | string |
| projectId | query |  | No | string |
| userId | query |  | No | string |
| roleId | query |  | No | string |

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
| ---- | ---------- | ----------- | -------- | ------ |
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
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetPolicyResponse](#v1beta1getpolicyresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete a policy permanently forever and all of its relations

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeletePolicyResponse](#v1beta1deletepolicyresponse) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Policy by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1PolicyRequestBody](#v1beta1policyrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdatePolicyResponse](#v1beta1updatepolicyresponse) |
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

#### v1beta1AuthCallbackResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1AuthCallbackResponse | object |  |  |

#### v1beta1AuthLogoutResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1AuthLogoutResponse | object |  |  |

#### v1beta1AuthStrategy

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| params | object |  | No |

#### v1beta1AuthenticateResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| endpoint | string |  | No |

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

#### v1beta1CreateGroupResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| group | [v1beta1Group](#v1beta1group) |  | No |

#### v1beta1CreateMetaSchemaResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| metaschema | [v1beta1MetaSchema](#v1beta1metaschema) |  | No |

#### v1beta1CreateNamespaceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| namespace | [v1beta1Namespace](#v1beta1namespace) |  | No |

#### v1beta1CreateOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organization | [v1beta1Organization](#v1beta1organization) |  | No |

#### v1beta1CreateOrganizationRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| role | [v1beta1Role](#v1beta1role) |  | No |

#### v1beta1CreatePermissionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| permission | [v1beta1Permission](#v1beta1permission) |  | No |

#### v1beta1CreatePolicyResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| policy | [v1beta1Policy](#v1beta1policy) |  | No |

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

#### v1beta1DeleteGroupResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteGroupResponse | object |  |  |

#### v1beta1DeleteMetaSchemaResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteMetaSchemaResponse | object |  |  |

#### v1beta1DeleteOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteOrganizationResponse | object |  |  |

#### v1beta1DeleteOrganizationRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteOrganizationRoleResponse | object |  |  |

#### v1beta1DeletePermissionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeletePermissionResponse | object |  |  |

#### v1beta1DeletePolicyResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeletePolicyResponse | object |  |  |

#### v1beta1DeleteProjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteProjectResponse | object |  |  |

#### v1beta1DeleteRelationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| message | string |  | No |

#### v1beta1DeleteResourceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteResourceResponse | object |  |  |

#### v1beta1DeleteRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteRoleResponse | object |  |  |

#### v1beta1DeleteUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteUserResponse | object |  |  |

#### v1beta1DisableGroupResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DisableGroupResponse | object |  |  |

#### v1beta1DisableOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DisableOrganizationResponse | object |  |  |

#### v1beta1DisableProjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DisableProjectResponse | object |  |  |

#### v1beta1DisableUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DisableUserResponse | object |  |  |

#### v1beta1EnableGroupResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1EnableGroupResponse | object |  |  |

#### v1beta1EnableOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1EnableOrganizationResponse | object |  |  |

#### v1beta1EnableProjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1EnableProjectResponse | object |  |  |

#### v1beta1EnableUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1EnableUserResponse | object |  |  |

#### v1beta1GetCurrentUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [v1beta1User](#v1beta1user) |  | No |

#### v1beta1GetGroupResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| group | [v1beta1Group](#v1beta1group) |  | No |

#### v1beta1GetMetaSchemaResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| metaschema | [v1beta1MetaSchema](#v1beta1metaschema) |  | No |

#### v1beta1GetNamespaceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| namespace | [v1beta1Namespace](#v1beta1namespace) |  | No |

#### v1beta1GetOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organization | [v1beta1Organization](#v1beta1organization) |  | No |

#### v1beta1GetOrganizationRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| role | [v1beta1Role](#v1beta1role) |  | No |

#### v1beta1GetOrganizationsByCurrentUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organizations | [ [v1beta1Organization](#v1beta1organization) ] |  | No |

#### v1beta1GetOrganizationsByUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organizations | [ [v1beta1Organization](#v1beta1organization) ] |  | No |

#### v1beta1GetPermissionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| permission | [v1beta1Permission](#v1beta1permission) |  | No |

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

#### v1beta1GroupRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| slug | string |  | No |
| metadata | object |  | No |
| orgId | string |  | No |

#### v1beta1ListAllOrganizationsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organizations | [ [v1beta1Organization](#v1beta1organization) ] |  | No |

#### v1beta1ListAllUsersResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| count | integer |  | No |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1ListAuthStrategiesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| strategies | [ [v1beta1AuthStrategy](#v1beta1authstrategy) ] |  | No |

#### v1beta1ListCurrentUserGroupsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| groups | [ [v1beta1Group](#v1beta1group) ] |  | No |

#### v1beta1ListGroupUsersResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1ListGroupsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| groups | [ [v1beta1Group](#v1beta1group) ] |  | No |

#### v1beta1ListMetaSchemasResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| metaschemas | [ [v1beta1MetaSchema](#v1beta1metaschema) ] |  | No |

#### v1beta1ListNamespacesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| namespaces | [ [v1beta1Namespace](#v1beta1namespace) ] |  | No |

#### v1beta1ListOrganizationAdminsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1ListOrganizationGroupsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| groups | [ [v1beta1Group](#v1beta1group) ] |  | No |

#### v1beta1ListOrganizationProjectsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| projects | [ [v1beta1Project](#v1beta1project) ] |  | No |

#### v1beta1ListOrganizationRolesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| roles | [ [v1beta1Role](#v1beta1role) ] |  | No |

#### v1beta1ListOrganizationUsersResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1ListOrganizationsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organizations | [ [v1beta1Organization](#v1beta1organization) ] |  | No |

#### v1beta1ListPermissionsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| permissions | [ [v1beta1Permission](#v1beta1permission) ] |  | No |

#### v1beta1ListPoliciesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| policies | [ [v1beta1Policy](#v1beta1policy) ] |  | No |

#### v1beta1ListProjectAdminsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1ListProjectResourcesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| resources | [ [v1beta1Resource](#v1beta1resource) ] |  | No |

#### v1beta1ListProjectUsersResponse

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

#### v1beta1MetaSchema

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| schema | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1MetaSchemaRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| schema | string |  | No |

#### v1beta1Namespace

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| metadata | object |  | No |
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

#### v1beta1Permission

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| namespaceId | string |  | No |
| metadata | object |  | No |
| slug | string |  | No |

#### v1beta1PermissionRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| namespaceId | string |  | No |
| metadata | object |  | No |
| slug | string |  | No |

#### v1beta1Policy

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| namespaceId | string |  | No |
| roleId | string |  | No |
| resourceId | string |  | No |
| userId | string |  | No |
| metadata | object |  | No |

#### v1beta1PolicyRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| roleId | string |  | No |
| resourceId | string |  | No |
| namespaceId | string |  | No |
| userId | string |  | No |
| metadata | object |  | No |

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
| subjectId | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| subjectNamespace | string |  | No |
| subjectSubRelationName | string |  | No |
| relationName | string |  | No |

#### v1beta1RelationRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| objectId | string |  | No |
| objectNamespace | string |  | No |
| subjectId | string |  | No |
| relationName | string |  | No |
| subjectNamespace | string |  | No |
| subjectSubRelation | string |  | No |

#### v1beta1Resource

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| urn | string |  | No |
| projectId | string |  | No |
| namespaceId | string |  | No |
| userId | string |  | No |
| metadata | object |  | No |

#### v1beta1ResourceRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| projectId | string |  | No |
| namespaceId | string |  | No |
| relations | [ [v1beta1Relation](#v1beta1relation) ] |  | No |
| userId | string |  | No |
| metadata | object |  | No |

#### v1beta1Role

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| permissions | [ string ] |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| orgId | string |  | No |
| state | string |  | No |

#### v1beta1RoleRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| permissions | [ string ] |  | No |
| orgId | string |  | No |
| metadata | object |  | No |

#### v1beta1UpdateCurrentUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [v1beta1User](#v1beta1user) |  | No |

#### v1beta1UpdateGroupResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| group | [v1beta1Group](#v1beta1group) |  | No |

#### v1beta1UpdateMetaSchemaResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| metaschema | [v1beta1MetaSchema](#v1beta1metaschema) |  | No |

#### v1beta1UpdateNamespaceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| namespace | [v1beta1Namespace](#v1beta1namespace) |  | No |

#### v1beta1UpdateOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organization | [v1beta1Organization](#v1beta1organization) |  | No |

#### v1beta1UpdateOrganizationRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| role | [v1beta1Role](#v1beta1role) |  | No |

#### v1beta1UpdatePermissionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| permission | [v1beta1Permission](#v1beta1permission) |  | No |

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
| slug | string |  | No |
