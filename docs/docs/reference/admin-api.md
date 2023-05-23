# Shield Administration API
## Version: 0.2.0

## Group

### /v1beta1/admin/groups

#### GET
##### Summary

Get all groups

##### Description

Lists all the groups from all the organizations in a Shield instance. It can be filtered by organization and state.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| userId | query |  | No | string |
| orgId | query | The organization id to filter by. | No | string |
| state | query | The state to filter by. It can be enabled or disabled. | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListGroupsResponse](#v1beta1listgroupsresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Organization

### /v1beta1/admin/organizations

#### GET
##### Summary

Get all organization

##### Description

Lists all the organizations in a Shield instance. It can be filtered by user and state.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| userId | query | The user id to filter by. | No | string |
| state | query | The state to filter by. It can be enabled or disabled. | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListAllOrganizationsResponse](#v1beta1listallorganizationsresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Project

### /v1beta1/admin/projects

#### GET
##### Summary

Get all project

##### Description

Lists all the projects from all the organizations in a Shield instance. It can be filtered by organization and state.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | query | The organization id to filter by. | No | string |
| state | query | The state to filter by. It can be enabled or disabled. | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListProjectsResponse](#v1beta1listprojectsresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Relation

### /v1beta1/admin/relations

#### GET
##### Summary

Get all relations

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListRelationsResponse](#v1beta1listrelationsresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Resource

### /v1beta1/admin/resources

#### GET
##### Summary

Get all resources

##### Description

Lists all the resources from all the organizations in a Shield instance. It can be filtered by user, project, organization and namespace.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| userId | query | The user id to filter by. | No | string |
| projectId | query | The project id to filter by. | No | string |
| organizationId | query | The organization id to filter by. | No | string |
| namespace | query | The namespace to filter by. | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListResourcesResponse](#v1beta1listresourcesresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## User

### /v1beta1/admin/users

#### GET
##### Summary

Get all users

##### Description

Lists all the users from all the organizations in a Shield instance. It can be filtered by keyword, organization, group and state.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| pageSize | query | The maximum number of users to return per page. The default is 50. | No | integer |
| pageNum | query | The page number to return. The default is 1. | No | integer |
| keyword | query | The keyword to search for. It can be a user's name, email,metadata or id. | No | string |
| orgId | query | The organization id to filter by. | No | string |
| groupId | query | The group id to filter by. | No | string |
| state | query | The state to filter by. It can be enabled or disabled. | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListAllUsersResponse](#v1beta1listallusersresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Permission

### /v1beta1/permissions

#### POST
##### Summary

Create permission

##### Description

Creates a permission. It can be used to grant permissions to all the resources in a Shield instance.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1CreatePermissionRequest](#v1beta1createpermissionrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreatePermissionResponse](#v1beta1createpermissionresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/permissions/{id}

#### DELETE
##### Summary

Delete permission by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeletePermissionResponse](#v1beta1deletepermissionresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update permission by ID

##### Description

Updates a permission by ID. It can be used to grant permissions to all the resources in a Shield instance.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1PermissionRequestBody](#v1beta1permissionrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdatePermissionResponse](#v1beta1updatepermissionresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Policy

### /v1beta1/policies

#### GET
##### Summary

Get all policies

##### Description

Lists all the policies from all the organizations in a Shield instance. It can be filtered by organization, project, user, role and group.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | query | The organization id to filter by. | No | string |
| projectId | query | The project id to filter by. | No | string |
| userId | query | The user id to filter by. | No | string |
| roleId | query | The role id to filter by. | No | string |
| groupId | query | The group id to filter by. | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListPoliciesResponse](#v1beta1listpoliciesresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Role

### /v1beta1/roles

#### POST
##### Summary

Create platform wide role

##### Description

Creates a platform wide role. It can be used to grant permissions to all the resources in a Shield instance.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1RoleRequestBody](#v1beta1rolerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateRoleResponse](#v1beta1createroleresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/roles/{id}

#### DELETE
##### Summary

Delete a role permanently forever and all of its relations

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path | The role id to delete. | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteRoleResponse](#v1beta1deleteroleresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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

#### v1beta1CreatePermissionRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| bodies | [ [v1beta1PermissionRequestBody](#v1beta1permissionrequestbody) ] |  | No |

#### v1beta1CreatePermissionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| permissions | [ [v1beta1Permission](#v1beta1permission) ] |  | No |

#### v1beta1CreateRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| role | [v1beta1Role](#v1beta1role) |  | No |

#### v1beta1DeletePermissionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeletePermissionResponse | object |  |  |

#### v1beta1DeleteRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteRoleResponse | object |  |  |

#### v1beta1Group

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| title | string |  | No |
| orgId | string |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1ListAllOrganizationsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organizations | [ [v1beta1Organization](#v1beta1organization) ] |  | No |

#### v1beta1ListAllUsersResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| count | integer |  | No |
| users | [ [v1beta1User](#v1beta1user) ] |  | No |

#### v1beta1ListGroupsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| groups | [ [v1beta1Group](#v1beta1group) ] |  | No |

#### v1beta1ListPoliciesResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| policies | [ [v1beta1Policy](#v1beta1policy) ] |  | No |

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

#### v1beta1Organization

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| title | string |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1Permission

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| title | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| namespace | string |  | No |
| metadata | object |  | No |

#### v1beta1PermissionRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string | The name of the permission. It should be unique across a Shield instance and can contain only alphanumeric characters. | Yes |
| namespace | string | The namespace of the permission.The namespace should be in service/resource format.<br/>*Example:*`app/guardian` | Yes |
| metadata | object | The metadata object for permissions that can hold key value pairs. | No |
| title | string | The title can contain any UTF-8 character, used to provide a human-readable name for the permissions. Can also be left empty. | No |

#### v1beta1Policy

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| title | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| roleId | string |  | No |
| resource | string |  | No |
| principal | string |  | No |
| metadata | object |  | No |

#### v1beta1Project

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| title | string |  | No |
| orgId | string |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1Relation

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| subjectSubRelation | string |  | No |
| relation | string |  | No |
| object | string |  | No |
| subject | string |  | No |

#### v1beta1Resource

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
| urn | string |  | No |
| projectId | string |  | No |
| namespace | string |  | No |
| userId | string |  | No |
| metadata | object |  | No |

#### v1beta1Role

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| permissions | [ string ] |  | No |
| title | string |  | No |
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
| title | string |  | No |

#### v1beta1UpdatePermissionResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| permission | [v1beta1Permission](#v1beta1permission) |  | No |

#### v1beta1User

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| title | string |  | No |
| email | string |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |
