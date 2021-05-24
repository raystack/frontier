# API

## Version: 0.1.0

### Security

**AUTH_KEY**

| apiKey | _API Key_                       |
| ------ | ------------------------------- |
| Name   | X-Goog-Authenticated-User-Email |
| In     | header                          |

### /ping

#### GET

##### Summary

pong the request

##### Responses

| Code | Description | Schema                |
| ---- | ----------- | --------------------- |
| 200  | Successful  | [PingPong](#pingpong) |

### /api/activities

#### GET

##### Summary

get all activities

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [Activities](#activities)                                   |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/groups

#### GET

##### Summary

get list of groups

##### Parameters

| Name      | Located in | Description | Required | Schema |
| --------- | ---------- | ----------- | -------- | ------ |
| user_role | query      | role id     | No       | string |
| group     | query      | group id    | No       | string |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [Groups](#groups)                                           |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

#### POST

##### Summary

create group

##### Parameters

| Name | Located in | Description | Required | Schema                                    |
| ---- | ---------- | ----------- | -------- | ----------------------------------------- |
| body | body       |             | No       | [GroupCreatePayload](#groupcreatepayload) |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [Group](#group)                                             |
| 400  | Bad Request           | [BadRequestResponse](#badrequestresponse)                   |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/profile

#### GET

##### Summary

get current user's profile

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [User](#user)                                               |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

#### PUT

##### Summary

update current user

##### Parameters

| Name | Located in | Description | Required | Schema                                        |
| ---- | ---------- | ----------- | -------- | --------------------------------------------- |
| body | body       |             | No       | [ProfileUpdatePayload](#profileupdatepayload) |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [User](#user)                                               |
| 201  | Created               | [User](#user)                                               |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/roles

#### GET

##### Summary

get roles based on attributes

##### Parameters

| Name       | Located in | Description | Required | Schema     |
| ---------- | ---------- | ----------- | -------- | ---------- |
| attributes | query      |             | No       | [ string ] |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [Roles](#roles)                                             |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

#### POST

##### Summary

create roles along with action mapping

##### Parameters

| Name | Located in | Description | Required | Schema                                  |
| ---- | ---------- | ----------- | -------- | --------------------------------------- |
| body | body       |             | No       | [RoleCreatePayload](#rolecreatepayload) |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [Role](#role)                                               |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/users

#### GET

##### Summary

get list of users

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [Users](#users)                                             |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

#### POST

##### Summary

create user

##### Parameters

| Name | Located in | Description | Required | Schema                                        |
| ---- | ---------- | ----------- | -------- | --------------------------------------------- |
| body | body       |             | No       | [ProfileUpdatePayload](#profileupdatepayload) |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [User](#user)                                               |
| 400  | Bad Request           | [BadRequestResponse](#badrequestresponse)                   |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/groups/{id}

#### GET

##### Summary

get group by id

##### Parameters

| Name      | Located in | Description | Required | Schema |
| --------- | ---------- | ----------- | -------- | ------ |
| id        | path       | group id    | Yes      | string |
| user_role | query      |             | No       | string |
| group     | query      |             | No       | string |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [Group](#group)                                             |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

#### PUT

##### Summary

update group by id

##### Parameters

| Name | Located in | Description | Required | Schema                                    |
| ---- | ---------- | ----------- | -------- | ----------------------------------------- |
| id   | path       | group id    | Yes      | string                                    |
| body | body       |             | No       | [GroupUpdatePayload](#groupupdatepayload) |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [Group](#group)                                             |
| 201  | Created               | [Group](#group)                                             |
| 400  | Bad Request           | [BadRequestResponse](#badrequestresponse)                   |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/users/{userId}

#### GET

##### Summary

get user by id

##### Parameters

| Name   | Located in | Description | Required | Schema |
| ------ | ---------- | ----------- | -------- | ------ |
| userId | path       | user id     | Yes      | string |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [User](#user)                                               |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/groups/{groupId}/users

#### GET

##### Summary

fetch list of users of a group

##### Parameters

| Name    | Located in | Description | Required | Schema |
| ------- | ---------- | ----------- | -------- | ------ |
| groupId | path       | group id    | Yes      | string |
| role    | query      |             | No       | string |
| action  | query      |             | No       | string |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [UsersWithPolicies](#userswithpolicies)                     |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/groups/{id}/activities

#### GET

##### Summary

get activities by group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id   | path       | group id    | Yes      | string |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [Activities](#activities)                                   |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/users/self/groups

#### GET

##### Summary

fetch list of groups of the loggedIn user

##### Parameters

| Name   | Located in | Description | Required | Schema |
| ------ | ---------- | ----------- | -------- | ------ |
| action | query      |             | No       | string |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [GroupPolicies](#grouppolicies)                             |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/users/{userId}/groups

#### GET

##### Summary

fetch list of groups of a user

##### Parameters

| Name   | Located in | Description | Required | Schema |
| ------ | ---------- | ----------- | -------- | ------ |
| userId | path       | user id     | Yes      | string |
| action | query      |             | No       | string |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [GroupPolicies](#grouppolicies)                             |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/groups/{groupId}/users/{userId}

#### GET

##### Summary

fetch user and group mapping

##### Parameters

| Name    | Located in | Description | Required | Schema |
| ------- | ---------- | ----------- | -------- | ------ |
| groupId | path       | group id    | Yes      | string |
| userId  | path       | user id     | Yes      | string |
| role    | query      |             | No       | string |
| action  | query      |             | No       | string |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [UserWithPolicies](#userwithpolicies)                       |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

#### PUT

##### Summary

update group and user mapping

##### Parameters

| Name    | Located in | Description | Required | Schema                                              |
| ------- | ---------- | ----------- | -------- | --------------------------------------------------- |
| groupId | path       | group id    | Yes      | string                                              |
| userId  | path       | user id     | Yes      | string                                              |
| body    | body       |             | No       | [PolciesOperationPayload](#polciesoperationpayload) |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [UserGroupResponse](#usergroupresponse)                     |
| 201  | Created               | [UserGroupResponse](#usergroupresponse)                     |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

#### POST

##### Summary

create group and user mapping

##### Parameters

| Name    | Located in | Description | Required | Schema                                              |
| ------- | ---------- | ----------- | -------- | --------------------------------------------------- |
| groupId | path       | group id    | Yes      | string                                              |
| userId  | path       | user id     | Yes      | string                                              |
| body    | body       |             | No       | [PolciesOperationPayload](#polciesoperationpayload) |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [UserGroupResponse](#usergroupresponse)                     |
| 201  | Created               | [UserGroupResponse](#usergroupresponse)                     |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

#### DELETE

##### Summary

delete group and user mapping

##### Parameters

| Name    | Located in | Description | Required | Schema |
| ------- | ---------- | ----------- | -------- | ------ |
| groupId | path       | group id    | Yes      | string |
| userId  | path       | user id     | Yes      | string |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | boolean                                                     |
| 201  | Created               | [UserGroupResponse](#usergroupresponse)                     |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/roles/{id}

#### PUT

##### Summary

update roles along with action mapping

##### Parameters

| Name | Located in | Description | Required | Schema                                  |
| ---- | ---------- | ----------- | -------- | --------------------------------------- |
| id   | path       |             | Yes      | string                                  |
| body | body       |             | No       | [RoleCreatePayload](#rolecreatepayload) |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [Role](#role)                                               |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/check-access

#### POST

##### Summary

checks whether subject has access to perform action on a given resource/attributes

##### Parameters

| Name | Located in | Description | Required | Schema                |
| ---- | ---------- | ----------- | -------- | --------------------- |
| body | body       |             | No       | [Policies](#policies) |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | [Accesses](#accesses)                                       |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### /api/resources

#### POST

##### Summary

create resource and attributes mapping

##### Parameters

| Name | Located in | Description | Required | Schema                                                                |
| ---- | ---------- | ----------- | -------- | --------------------------------------------------------------------- |
| body | body       |             | No       | [ResourceAttributesMappingPayload](#resourceattributesmappingpayload) |

##### Responses

| Code | Description           | Schema                                                                                |
| ---- | --------------------- | ------------------------------------------------------------------------------------- |
| 200  | Successful            | [ResourceAttributesMappingPayloadResponse](#resourceattributesmappingpayloadresponse) |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)                                         |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                                                 |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse)                           |

### /api/groups/{groupId}/users/self

#### DELETE

##### Summary

delete group and loggedin user mapping

##### Parameters

| Name    | Located in | Description | Required | Schema |
| ------- | ---------- | ----------- | -------- | ------ |
| groupId | path       | group id    | Yes      | string |

##### Responses

| Code | Description           | Schema                                                      |
| ---- | --------------------- | ----------------------------------------------------------- |
| 200  | Successful            | boolean                                                     |
| 201  | Created               | [UserGroupResponse](#usergroupresponse)                     |
| 401  | Unauthorized          | [UnauthorizedResponse](#unauthorizedresponse)               |
| 404  | Not Found             | [NotFoundResponse](#notfoundresponse)                       |
| 500  | Internal Server Error | [InternalServerErrorResponse](#internalservererrorresponse) |

### Models

#### PingPong

| Name       | Type    | Description    | Required |
| ---------- | ------- | -------------- | -------- |
| statusCode | integer | _Enum:_ `200`  | Yes      |
| status     | string  | _Enum:_ `"ok"` | Yes      |
| message    | string  |                | Yes      |

#### diff

| Name | Type   | Description | Required |
| ---- | ------ | ----------- | -------- |
| diff | object |             |          |

#### Activity

| Name      | Type          | Description | Required |
| --------- | ------------- | ----------- | -------- |
| id        | string        |             | No       |
| createdAt | date          |             | No       |
| diff      | [diff](#diff) |             | No       |
| reason    | string        |             | No       |
| user      | string        |             | No       |

#### Activities

| Name       | Type  | Description | Required |
| ---------- | ----- | ----------- | -------- |
| Activities | array |             |          |

#### UnauthorizedResponse

| Name       | Type    | Description   | Required |
| ---------- | ------- | ------------- | -------- |
| error      | string  |               | Yes      |
| message    | string  |               | Yes      |
| statusCode | integer | _Enum:_ `401` | Yes      |

#### NotFoundResponse

| Name       | Type    | Description   | Required |
| ---------- | ------- | ------------- | -------- |
| error      | string  |               | Yes      |
| message    | string  |               | Yes      |
| statusCode | integer | _Enum:_ `404` | Yes      |

#### InternalServerErrorResponse

| Name       | Type    | Description   | Required |
| ---------- | ------- | ------------- | -------- |
| error      | string  |               | Yes      |
| message    | string  |               | Yes      |
| statusCode | integer | _Enum:_ `500` | Yes      |

#### subject

| Name  | Type   | Description | Required |
| ----- | ------ | ----------- | -------- |
| user  | string |             | No       |
| group | string |             | No       |

#### action

| Name   | Type   | Description | Required |
| ------ | ------ | ----------- | -------- |
| action | string |             | No       |
| role   | string |             | No       |

#### Policy

| Name     | Type                | Description | Required |
| -------- | ------------------- | ----------- | -------- |
| subject  | [subject](#subject) |             | Yes      |
| resource | [diff](#diff)       |             | Yes      |
| action   | [action](#action)   |             | Yes      |

#### Policies

| Name     | Type  | Description | Required |
| -------- | ----- | ----------- | -------- |
| Policies | array |             |          |

#### attributes

| Name       | Type  | Description | Required |
| ---------- | ----- | ----------- | -------- |
| attributes | array |             |          |

#### Group

| Name         | Type                      | Description | Required |
| ------------ | ------------------------- | ----------- | -------- |
| id           | string                    |             | Yes      |
| displayname  | string                    |             | No       |
| isMember     | boolean                   |             | No       |
| userPolicies | [Policies](#policies)     |             | No       |
| policies     | [Policies](#policies)     |             | No       |
| memberCount  | integer                   |             | No       |
| attributes   | [attributes](#attributes) |             | No       |

#### Groups

| Name   | Type  | Description | Required |
| ------ | ----- | ----------- | -------- |
| Groups | array |             |          |

#### User

| Name        | Type          | Description | Required |
| ----------- | ------------- | ----------- | -------- |
| id          | string        |             | Yes      |
| username    | string        |             | Yes      |
| displayname | string        |             | Yes      |
| metadata    | [diff](#diff) |             | No       |
| createdAt   | date          |             | Yes      |
| updatedAt   | date          |             | Yes      |

#### Model1

| Name   | Type  | Description | Required |
| ------ | ----- | ----------- | -------- |
| Model1 | array |             |          |

#### Role

| Name        | Type              | Description | Required |
| ----------- | ----------------- | ----------- | -------- |
| id          | string            |             | Yes      |
| displayname | string            |             | Yes      |
| attributes  | [Model1](#model1) |             | No       |
| metadata    | [diff](#diff)     |             | Yes      |
| createdAt   | date              |             | Yes      |
| updatedAt   | date              |             | Yes      |

#### Roles

| Name  | Type  | Description | Required |
| ----- | ----- | ----------- | -------- |
| Roles | array |             |          |

#### Users

| Name  | Type  | Description | Required |
| ----- | ----- | ----------- | -------- |
| Users | array |             |          |

#### Model2

| Name     | Type          | Description | Required |
| -------- | ------------- | ----------- | -------- |
| subject  | [diff](#diff) |             | No       |
| resource | [diff](#diff) |             | No       |
| action   | [diff](#diff) |             | No       |

#### policies

| Name     | Type  | Description | Required |
| -------- | ----- | ----------- | -------- |
| policies | array |             |          |

#### UserWithPolicies

| Name        | Type                  | Description | Required |
| ----------- | --------------------- | ----------- | -------- |
| id          | string                |             | Yes      |
| username    | string                |             | Yes      |
| displayname | string                |             | Yes      |
| metadata    | [diff](#diff)         |             | No       |
| createdAt   | date                  |             | Yes      |
| updatedAt   | date                  |             | Yes      |
| policies    | [policies](#policies) |             | No       |

#### UsersWithPolicies

| Name              | Type  | Description | Required |
| ----------------- | ----- | ----------- | -------- |
| UsersWithPolicies | array |             |          |

#### GroupPolicy

| Name       | Type                      | Description | Required |
| ---------- | ------------------------- | ----------- | -------- |
| policies   | [Policies](#policies)     |             | No       |
| attributes | [attributes](#attributes) |             | No       |

#### GroupPolicies

| Name          | Type  | Description | Required |
| ------------- | ----- | ----------- | -------- |
| GroupPolicies | array |             |          |

#### ProfileUpdatePayload

| Name        | Type          | Description | Required |
| ----------- | ------------- | ----------- | -------- |
| displayname | string        |             | Yes      |
| metadata    | [diff](#diff) |             | No       |

#### Model3

| Name      | Type                | Description                    | Required |
| --------- | ------------------- | ------------------------------ | -------- |
| subject   | [subject](#subject) |                                | No       |
| resource  | [diff](#diff)       |                                | No       |
| action    | [action](#action)   |                                | No       |
| operation | string              | _Enum:_ `"create"`, `"delete"` | Yes      |

#### Model4

| Name   | Type  | Description | Required |
| ------ | ----- | ----------- | -------- |
| Model4 | array |             |          |

#### GroupUpdatePayload

| Name        | Type                      | Description | Required |
| ----------- | ------------------------- | ----------- | -------- |
| displayname | string                    |             | Yes      |
| policies    | [Model4](#model4)         |             | No       |
| attributes  | [attributes](#attributes) |             | No       |
| metadata    | [diff](#diff)             |             | No       |

#### BadRequestResponse

| Name       | Type    | Description   | Required |
| ---------- | ------- | ------------- | -------- |
| error      | string  |               | Yes      |
| message    | string  |               | Yes      |
| statusCode | integer | _Enum:_ `400` | Yes      |

#### ActionOperation

| Name      | Type   | Description                    | Required |
| --------- | ------ | ------------------------------ | -------- |
| operation | string | _Enum:_ `"create"`, `"delete"` | Yes      |
| action    | string |                                | Yes      |

#### actions

| Name    | Type  | Description | Required |
| ------- | ----- | ----------- | -------- |
| actions | array |             |          |

#### RoleCreatePayload

| Name        | Type                | Description | Required |
| ----------- | ------------------- | ----------- | -------- |
| displayname | string              |             | Yes      |
| attributes  | [Model1](#model1)   |             | No       |
| actions     | [actions](#actions) |             | No       |
| metadata    | [diff](#diff)       |             | No       |

#### PolciesOperationPayload

| Name     | Type              | Description | Required |
| -------- | ----------------- | ----------- | -------- |
| policies | [Model4](#model4) |             | No       |

#### Model5

| Name      | Type                | Description                    | Required |
| --------- | ------------------- | ------------------------------ | -------- |
| subject   | [subject](#subject) |                                | No       |
| resource  | [diff](#diff)       |                                | No       |
| action    | [action](#action)   |                                | No       |
| operation | string              | _Enum:_ `"create"`, `"delete"` | Yes      |
| success   | boolean             |                                | Yes      |

#### UserGroupResponse

| Name              | Type  | Description | Required |
| ----------------- | ----- | ----------- | -------- |
| UserGroupResponse | array |             |          |

#### Model6

| Name      | Type                | Description | Required |
| --------- | ------------------- | ----------- | -------- |
| subject   | [subject](#subject) |             | No       |
| resource  | [diff](#diff)       |             | No       |
| action    | [action](#action)   |             | No       |
| hasAccess | boolean             |             | Yes      |

#### Accesses

| Name     | Type  | Description | Required |
| -------- | ----- | ----------- | -------- |
| Accesses | array |             |          |

#### GroupCreatePayload

| Name        | Type                      | Description | Required |
| ----------- | ------------------------- | ----------- | -------- |
| groupname   | string                    |             | No       |
| displayname | string                    |             | Yes      |
| policies    | [Model4](#model4)         |             | No       |
| attributes  | [attributes](#attributes) |             | No       |
| metadata    | [diff](#diff)             |             | No       |

#### Model7

| Name       | Type          | Description                    | Required |
| ---------- | ------------- | ------------------------------ | -------- |
| operation  | string        | _Enum:_ `"create"`, `"delete"` | Yes      |
| resource   | [diff](#diff) |                                | No       |
| attributes | [diff](#diff) |                                | No       |

#### ResourceAttributesMappingPayload

| Name                             | Type  | Description | Required |
| -------------------------------- | ----- | ----------- | -------- |
| ResourceAttributesMappingPayload | array |             |          |

#### Model8

| Name       | Type          | Description                    | Required |
| ---------- | ------------- | ------------------------------ | -------- |
| operation  | string        | _Enum:_ `"create"`, `"delete"` | Yes      |
| resource   | [diff](#diff) |                                | No       |
| attributes | [diff](#diff) |                                | No       |
| success    | boolean       |                                | No       |

#### ResourceAttributesMappingPayloadResponse

| Name                                     | Type  | Description | Required |
| ---------------------------------------- | ----- | ----------- | -------- |
| ResourceAttributesMappingPayloadResponse | array |             |          |
