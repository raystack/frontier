# API

## Version: 1.0.0

### Security

**AUTH\_KEY**

| apiKey | _API Key_ |
| :--- | :--- |
| Name | X-Goog-Authenticated-User-Email |
| In | header |

### /ping

#### GET

**Summary**

pong the request

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |


**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [PingPong](api.md#pingpong) |

### /api/activities

#### GET

**Summary**

get all activities

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |


**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [Activities](api.md#activities) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/groups

#### GET

**Summary**

get list of groups

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| user\_role | query | role id | No | string |
| group | query | group id | No | string |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [GroupsPolicies](api.md#groupspolicies) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

#### POST

**Summary**

create group

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| body | body |  | No | [GroupCreatePayload](api.md#groupcreatepayload) |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [GroupPolicy](api.md#grouppolicy) |
| 400 | Bad Request | [BadRequestResponse](api.md#badrequestresponse) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/profile

#### GET

**Summary**

get current user's profile

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |


**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [User](api.md#user) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

#### PUT

**Summary**

update current user

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| body | body |  | No | [ProfileUpdatePayload](api.md#profileupdatepayload) |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [User](api.md#user) |
| 201 | Created | [User](api.md#user) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/roles

#### GET

**Summary**

get roles based on attributes

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| attributes | query |  | No | \[ string \] |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [Roles](api.md#roles) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

#### POST

**Summary**

create roles along with action mapping

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| body | body |  | No | [RoleCreatePayload](api.md#rolecreatepayload) |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [Role](api.md#role) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/users

#### GET

**Summary**

get list of users

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| action | query |  | No | string |
| role | query |  | No | string |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [Users](api.md#users) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

#### POST

**Summary**

create user

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| body | body |  | No | [ProfileUpdatePayload](api.md#profileupdatepayload) |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [User](api.md#user) |
| 400 | Bad Request | [BadRequestResponse](api.md#badrequestresponse) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/groups/{id}

#### GET

**Summary**

get group by id

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| id | path | group id | Yes | string |
| user\_role | query |  | No | string |
| group | query |  | No | string |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [GroupPolicy](api.md#grouppolicy) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

#### PUT

**Summary**

update group by id

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| id | path | group id | Yes | string |
| body | body |  | No | [GroupUpdatePayload](api.md#groupupdatepayload) |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [GroupPolicy](api.md#grouppolicy) |
| 201 | Created | [GroupPolicy](api.md#grouppolicy) |
| 400 | Bad Request | [BadRequestResponse](api.md#badrequestresponse) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/users/{userId}

#### GET

**Summary**

get user by id

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| userId | path | user id | Yes | string |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [User](api.md#user) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/groups/{groupId}/users

#### GET

**Summary**

fetch list of users of a group

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| groupId | path | group id | Yes | string |
| role | query |  | No | string |
| action | query |  | No | string |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [GroupsPolicies](api.md#groupspolicies) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/groups/{id}/activities

#### GET

**Summary**

get activities by group

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| id | path | group id | Yes | string |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [Activities](api.md#activities) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/users/self/groups

#### GET

**Summary**

fetch list of groups of the loggedIn user

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| action | query |  | No | string |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [GroupPolicies](api.md#grouppolicies) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/users/{userId}/groups

#### GET

**Summary**

fetch list of groups of a user

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| userId | path | user id | Yes | string |
| action | query |  | No | string |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [GroupPolicies](api.md#grouppolicies) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/groups/{groupId}/users/{userId}

#### GET

**Summary**

fetch user and group mapping

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| groupId | path | group id | Yes | string |
| userId | path | user id | Yes | string |
| role | query |  | No | string |
| action | query |  | No | string |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [UserWithPolicies](api.md#userwithpolicies) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

#### POST

**Summary**

create group and user mapping

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| groupId | path | group id | Yes | string |
| userId | path | user id | Yes | string |
| body | body |  | No | [PolciesOperationPayload](api.md#polciesoperationpayload) |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [UserGroupResponse](api.md#usergroupresponse) |
| 201 | Created | [UserGroupResponse](api.md#usergroupresponse) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

#### PUT

**Summary**

update group and user mapping

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| groupId | path | group id | Yes | string |
| userId | path | user id | Yes | string |
| body | body |  | No | [PolciesOperationPayload](api.md#polciesoperationpayload) |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [UserGroupResponse](api.md#usergroupresponse) |
| 201 | Created | [UserGroupResponse](api.md#usergroupresponse) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

#### DELETE

**Summary**

delete group and user mapping

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| groupId | path | group id | Yes | string |
| userId | path | user id | Yes | string |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | boolean |
| 201 | Created | [UserGroupResponse](api.md#usergroupresponse) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/roles/{id}

#### PUT

**Summary**

update roles along with action mapping

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| id | path |  | Yes | string |
| body | body |  | No | [RoleCreatePayload](api.md#rolecreatepayload) |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [Role](api.md#role) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/check-access

#### POST

**Summary**

checks whether subject has access to perform action on a given resource/attributes

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| body | body |  | No | [Policies](api.md#policies) |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [Accesses](api.md#accesses) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/resources

#### POST

**Summary**

create resource and attributes mapping

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| body | body |  | No | [ResourceAttributesMappingPayload](api.md#resourceattributesmappingpayload) |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | [ResourceAttributesMappingPayloadResponse](api.md#resourceattributesmappingpayloadresponse) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### /api/groups/{groupId}/users/self

#### DELETE

**Summary**

delete group and loggedin user mapping

**Parameters**

| Name | Located in | Description | Required | Schema |
| :--- | :--- | :--- | :--- | :--- |
| groupId | path | group id | Yes | string |

**Responses**

| Code | Description | Schema |
| :--- | :--- | :--- |
| 200 | Successful | boolean |
| 201 | Created | [UserGroupResponse](api.md#usergroupresponse) |
| 401 | Unauthorized | [UnauthorizedResponse](api.md#unauthorizedresponse) |
| 404 | Not Found | [NotFoundResponse](api.md#notfoundresponse) |
| 500 | Internal Server Error | [InternalServerErrorResponse](api.md#internalservererrorresponse) |

### Models

#### PingPong

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| statusCode | integer |  | Yes |
| status | string | _Enum:_ `"ok"` | Yes |
| message | string |  | Yes |

#### diff

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| diff | object |  |  |

#### Activity

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| id | string |  | No |
| createdAt | date |  | No |
| diff | [diff](api.md#diff) |  | No |
| reason | string |  | No |
| user | string |  | No |

#### Activities

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| Activities | array |  |  |

#### UnauthorizedResponse

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| error | string |  | Yes |
| message | string |  | Yes |
| statusCode | integer |  | Yes |

#### NotFoundResponse

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| error | string |  | Yes |
| message | string |  | Yes |
| statusCode | integer |  | Yes |

#### InternalServerErrorResponse

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| error | string |  | Yes |
| message | string |  | Yes |
| statusCode | integer |  | Yes |

#### subject

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| user | string |  | No |
| group | string |  | No |

#### action

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| action | string |  | No |
| role | string |  | No |

#### Policy

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| subject | [subject](api.md#subject) |  | Yes |
| resource | [diff](api.md#diff) |  | Yes |
| action | [action](api.md#action) |  | Yes |

#### Policies

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| Policies | array |  |  |

#### attributes

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| attributes | array |  |  |

#### GroupPolicy

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| id | string |  | Yes |
| displayname | string |  | No |
| isMember | boolean |  | No |
| userPolicies | [Policies](api.md#policies) |  | No |
| policies | [Policies](api.md#policies) |  | No |
| memberCount | integer |  | No |
| attributes | [attributes](api.md#attributes) |  | No |

#### GroupsPolicies

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| GroupsPolicies | array |  |  |

#### User

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| id | string |  | Yes |
| username | string |  | Yes |
| displayname | string |  | Yes |
| metadata | [diff](api.md#diff) |  | No |
| createdAt | date |  | Yes |
| updatedAt | date |  | Yes |

#### Model1

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| Model1 | array |  |  |

#### Role

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| id | string |  | Yes |
| displayname | string |  | Yes |
| attributes | [Model1](api.md#model1) |  | No |
| metadata | [diff](api.md#diff) |  | Yes |
| createdAt | date |  | Yes |
| updatedAt | date |  | Yes |

#### Roles

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| Roles | array |  |  |

#### Users

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| Users | array |  |  |

#### Model2

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| policies | [Policies](api.md#policies) |  | No |
| attributes | [attributes](api.md#attributes) |  | No |

#### GroupPolicies

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| GroupPolicies | array |  |  |

#### Model3

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| subject | [diff](api.md#diff) |  | No |
| resource | [diff](api.md#diff) |  | No |
| action | [diff](api.md#diff) |  | No |

#### policies

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| policies | array |  |  |

#### UserWithPolicies

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| id | string |  | Yes |
| username | string |  | Yes |
| displayname | string |  | Yes |
| metadata | [diff](api.md#diff) |  | No |
| createdAt | date |  | Yes |
| updatedAt | date |  | Yes |
| policies | [policies](api.md#policies) |  | No |

#### ProfileUpdatePayload

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| displayname | string |  | Yes |
| metadata | [diff](api.md#diff) |  | No |

#### Model4

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| subject | [subject](api.md#subject) |  | No |
| resource | [diff](api.md#diff) |  | No |
| action | [action](api.md#action) |  | No |
| operation | string | _Enum:_ `"create"`, `"delete"` | Yes |

#### Model5

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| Model5 | array |  |  |

#### GroupUpdatePayload

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| displayname | string |  | Yes |
| policies | [Model5](api.md#model5) |  | No |
| attributes | [attributes](api.md#attributes) |  | No |
| metadata | [diff](api.md#diff) |  | No |

#### BadRequestResponse

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| error | string |  | Yes |
| message | string |  | Yes |
| statusCode | integer |  | Yes |

#### ActionOperation

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| operation | string | _Enum:_ `"create"`, `"delete"` | Yes |
| action | string |  | Yes |

#### actions

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| actions | array |  |  |

#### RoleCreatePayload

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| displayname | string |  | Yes |
| attributes | [Model1](api.md#model1) |  | No |
| actions | [actions](api.md#actions) |  | No |
| metadata | [diff](api.md#diff) |  | No |

#### PolciesOperationPayload

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| policies | [Model5](api.md#model5) |  | No |

#### Model6

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| subject | [subject](api.md#subject) |  | No |
| resource | [diff](api.md#diff) |  | No |
| action | [action](api.md#action) |  | No |
| operation | string | _Enum:_ `"create"`, `"delete"` | Yes |
| success | boolean |  | Yes |

#### UserGroupResponse

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| UserGroupResponse | array |  |  |

#### Model7

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| subject | [subject](api.md#subject) |  | No |
| resource | [diff](api.md#diff) |  | No |
| action | [action](api.md#action) |  | No |
| hasAccess | boolean |  | Yes |

#### Accesses

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| Accesses | array |  |  |

#### GroupCreatePayload

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| groupname | string |  | No |
| displayname | string |  | Yes |
| policies | [Model5](api.md#model5) |  | No |
| attributes | [attributes](api.md#attributes) |  | No |
| metadata | [diff](api.md#diff) |  | No |

#### Model8

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| operation | string | _Enum:_ `"create"`, `"delete"` | Yes |
| resource | [diff](api.md#diff) |  | No |
| attributes | [diff](api.md#diff) |  | No |

#### ResourceAttributesMappingPayload

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| ResourceAttributesMappingPayload | array |  |  |

#### Model9

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| operation | string | _Enum:_ `"create"`, `"delete"` | Yes |
| resource | [diff](api.md#diff) |  | No |
| attributes | [diff](api.md#diff) |  | No |
| success | boolean |  | No |

#### ResourceAttributesMappingPayloadResponse

| Name | Type | Description | Required |
| :--- | :--- | :--- | :--- |
| ResourceAttributesMappingPayloadResponse | array |  |  |

