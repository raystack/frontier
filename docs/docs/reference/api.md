# Shield General APIs
## Version: 0.2.0

## Authn
Authentication APIs

### /v1beta1/auth

#### GET
##### Summary

Get all authentication strategies

##### Description

Returns a list of identity providers configured on an instance level in Shield. e.g Google, AzureAD, Github etc

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListAuthStrategiesResponse](#v1beta1listauthstrategiesresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/auth/callback

#### GET
##### Summary

Callback from a strategy

##### Description

Callback from a strategy. This is the endpoint where the strategy will redirect the user after successful authentication. This endpoint will be called with the code and state query parameters. The code will be used to get the access token from the strategy and the state will be used to get the return_to url from the session and redirect the user to that url.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| strategyName | query | strategy_name will not be set for oidc but can be utilized for methods like email magic links | No | string |
| state | query | for oidc & magic links | No | string |
| code | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthCallbackResponse](#v1beta1authcallbackresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Callback from a strategy

##### Description

Callback from a strategy. This is the endpoint where the strategy will redirect the user after successful authentication. This endpoint will be called with the code and state query parameters. The code will be used to get the access token from the strategy and the state will be used to get the return_to url from the session and redirect the user to that url.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| strategyName | query | strategy_name will not be set for oidc but can be utilized for methods like email magic links | No | string |
| state | query | for oidc & magic links | No | string |
| code | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthCallbackResponse](#v1beta1authcallbackresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/auth/logout

#### GET
##### Summary

Logout from a strategy

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthLogoutResponse](#v1beta1authlogoutresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Logout from a strategy

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthLogoutResponse](#v1beta1authlogoutresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/auth/register/{strategyName}

#### GET
##### Summary

Authenticate with a strategy

##### Description

Authenticate a user with a strategy. By default, after successful authentication no operation will be performed to apply redirection in case of browsers, provide a url in the authenticate request body that will be used for redirection after authentication. Also set redirect as true for redirecting the user request to the redirect_url after successful authentication.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| strategyName | path | Name of the strategy to use for authentication.<br/> *Example:* `google` | Yes | string |
| redirect | query | by default, location header for redirect if applicable will be skipped unless this is set to true, useful in browser  If set to true, location header will be set for redirect | No | boolean |
| returnTo | query | by default, after successful authentication no operation will be performed to apply redirection in case of browsers, provide a url that will be used for redirection after authentication  URL to redirect after successful authentication.<br/> *Example:*`"https://shield.example.com"` | No | string |
| email | query | email of the user for magic links  Email of the user to authenticate. Used for magic links.<br/> *Example:*`example@acme.org` | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthenticateResponse](#v1beta1authenticateresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Authenticate with a strategy

##### Description

Authenticate a user with a strategy. By default, after successful authentication no operation will be performed to apply redirection in case of browsers, provide a url in the authenticate request body that will be used for redirection after authentication. Also set redirect as true for redirecting the user request to the redirect_url after successful authentication.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| strategyName | path | Name of the strategy to use for authentication.<br/> *Example:* `google` | Yes | string |
| redirect | query | by default, location header for redirect if applicable will be skipped unless this is set to true, useful in browser  If set to true, location header will be set for redirect | No | boolean |
| returnTo | query | by default, after successful authentication no operation will be performed to apply redirection in case of browsers, provide a url that will be used for redirection after authentication  URL to redirect after successful authentication.<br/> *Example:*`"https://shield.example.com"` | No | string |
| email | query | email of the user for magic links  Email of the user to authenticate. Used for magic links.<br/> *Example:*`example@acme.org` | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AuthenticateResponse](#v1beta1authenticateresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Authz
Authorization APIs

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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## MetaSchema
Manage Metadata Schemas which are used to validate metadata object while creating user/org/group/role. Shield automatically generates default metaschemas with `label` and `description` fields. You can also update these metaschemas to add/edit more fields.

### /v1beta1/meta/schemas

#### GET
##### Summary

Get all Metadata Schemas

##### Description

Returns a list of all metaschemas configured on an instance level in Shield. e.g user, project, organization etc

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListMetaSchemasResponse](#v1beta1listmetaschemasresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Metadata Schema

##### Description

Create a new metadata schema. The metaschema **name** must be unique within the entire Shield instance and can contain only alphanumeric characters, dashes and underscores. The metaschema **schema** must be a valid JSON schema.Please refer to <https://json-schema.org/> to know more about json schema. <br/>*Example:* `{name:"user",schema:{"type":"object","properties":{"label":{"type":"object","additionalProperties":{"type":"string"}},"description":{"type":"string"}}}}`

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1MetaSchemaRequestBody](#v1beta1metaschemarequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateMetaSchemaResponse](#v1beta1createmetaschemaresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/meta/schemas/{id}

#### GET
##### Summary

Get MetaSchema by ID

##### Description

Get a metadata schema by ID.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetMetaSchemaResponse](#v1beta1getmetaschemaresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete a MetaSchema permanently

##### Description

Delete a metadata schema permanently. Once deleted the metaschema won't be used to validate the metadata. For example, if a metaschema(with `label` and `description` fields) is used to validate the metadata of a user, then deleting the metaschema will not validate the metadata of the user and metadata field can contain any key-value pair(and say another field called `foo` can be inserted in a user's metadata).

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteMetaSchemaResponse](#v1beta1deletemetaschemaresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update MetaSchema by ID

##### Description

Update a metadata schema. Only `schema` field of a metaschema can be updated. The metaschema `schema` must be a valid JSON schema.Please refer to <https://json-schema.org/> to know more about json schema. <br/>*Example:* `{name:"user",schema:{"type":"object","properties":{"label":{"type":"object","additionalProperties":{"type":"string"}},"description":{"type":"string"}}}}`

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1MetaSchemaRequestBody](#v1beta1metaschemarequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateMetaSchemaResponse](#v1beta1updatemetaschemaresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Namespace

### /v1beta1/namespaces

#### GET
##### Summary

Get all Namespaces

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListNamespacesResponse](#v1beta1listnamespacesresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/namespaces/{id}

#### GET
##### Summary

Get a Namespaces

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetNamespaceResponse](#v1beta1getnamespaceresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Organization

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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}

#### GET
##### Summary

Get organization by ID or slug

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetOrganizationResponse](#v1beta1getorganizationresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update organization by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1OrganizationRequestBody](#v1beta1organizationrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateOrganizationResponse](#v1beta1updateorganizationresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}/admins

#### GET
##### Summary

Get all admins of an organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationAdminsResponse](#v1beta1listorganizationadminsresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}/users

#### GET
##### Summary

List organization users

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| permissionFilter | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationUsersResponse](#v1beta1listorganizationusersresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Add a user to an organization, request fails if the user doesn't exists

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | { **"userIds"**: [ string ] } |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AddOrganizationUsersResponse](#v1beta1addorganizationusersresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{id}/users/{userId}

#### DELETE
##### Summary

Remove a user to an organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| userId | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1RemoveOrganizationUserResponse](#v1beta1removeorganizationuserresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/invitations

#### GET
##### Summary

List pending invitations queued for an organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| userId | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationInvitationsResponse](#v1beta1listorganizationinvitationsresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Invite users to an organization, if the user doesn't exists, it will be created and notified. Invitations expire in 7 days

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| body | body |  | Yes | { **"userId"**: string, **"groupIds"**: [ string ] } |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateOrganizationInvitationResponse](#v1beta1createorganizationinvitationresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/invitations/{id}

#### GET
##### Summary

Get pending invitation queued for an organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetOrganizationInvitationResponse](#v1beta1getorganizationinvitationresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete pending invitation queued for an organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteOrganizationInvitationResponse](#v1beta1deleteorganizationinvitationresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/invitations/{id}/accept

#### POST
##### Summary

Accept pending invitation queued for an organization

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AcceptOrganizationInvitationResponse](#v1beta1acceptorganizationinvitationresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Group
Groups in Shield are used to manage users and their access to resources. Each group has a unique name and id that can be used to grant access to resources. When a user is added to a group, they inherit the access permissions that have been granted to the group. This allows you to manage access to resources at scale, without having to grant permissions to individual users.

### /v1beta1/organizations/{orgId}/groups

#### GET
##### Summary

List Organization Groups

##### Description

Get all groups that belong to an organization. It can be filtered by keyword, organization, group and state. Additionally you can include page number and page size for pagination.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| userId | query |  | No | string |
| state | query | The state to filter by. It can be enabled or disabled. | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationGroupsResponse](#v1beta1listorganizationgroupsresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Group

##### Description

Create a new group in an organization which serves as a container for users. The group can be assigned roles and permissions and can be used to manage access to resources. Also a group can also be assigned to other groups.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path | The organization ID to which the group belongs to. | Yes | string |
| body | body |  | Yes | [v1beta1GroupRequestBody](#v1beta1grouprequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateGroupResponse](#v1beta1creategroupresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/groups/{id}

#### GET
##### Summary

Get group by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetGroupResponse](#v1beta1getgroupresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete a group permanently and all of its relations

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteGroupResponse](#v1beta1deletegroupresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update group by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path | The organization ID to which the group belongs to. | Yes | string |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1GroupRequestBody](#v1beta1grouprequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateGroupResponse](#v1beta1updategroupresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/groups/{id}/disable

#### POST
##### Summary

Disable a group

##### Description

Sets the state of the group as disabled. The group will not be available for access control and the existing users in the group will not be able to access any resources that are assigned to the group. No other users can be added to the group while it is disabled.

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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/groups/{id}/enable

#### POST
##### Summary

Enable a Group

##### Description

Sets the state of the group as enabled. The `enabled` flag is used to determine if the group can be used for access control.

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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Add users to a group, existing users will be ignored

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |
| body | body |  | Yes | { **"userIds"**: [ string ] } |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1AddGroupUsersResponse](#v1beta1addgroupusersresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/organizations/{orgId}/groups/{id}/users/{userId}

#### DELETE
##### Summary

Remove user from a group

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |
| userId | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1RemoveGroupUserResponse](#v1beta1removegroupuserresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Role

### /v1beta1/organizations/{orgId}/roles

#### GET
##### Summary

Get custom roles under an org

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| state | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListOrganizationRolesResponse](#v1beta1listorganizationrolesresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Role

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path | The organization ID to which the role belongs to. | Yes | string |
| body | body |  | Yes | [v1beta1RoleRequestBody](#v1beta1rolerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateOrganizationRoleResponse](#v1beta1createorganizationroleresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Role by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| orgId | path |  | Yes | string |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1RoleRequestBody](#v1beta1rolerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateOrganizationRoleResponse](#v1beta1updateorganizationroleresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/roles

#### GET
##### Summary

List all pre-defined roles

##### Description

Returns a collection of Shield predefined roles with their associated permissions

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| state | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListRolesResponse](#v1beta1listrolesresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Permission

### /v1beta1/permissions

#### GET
##### Summary

Get all Permissions

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListPermissionsResponse](#v1beta1listpermissionsresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/permissions/{id}

#### GET
##### Summary

Get permission by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetPermissionResponse](#v1beta1getpermissionresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Policy

### /v1beta1/policies

#### POST
##### Summary

Create policy

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1PolicyRequestBody](#v1beta1policyrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreatePolicyResponse](#v1beta1createpolicyresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/policies/{id}

#### GET
##### Summary

Get policy by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetPolicyResponse](#v1beta1getpolicyresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update policy by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1PolicyRequestBody](#v1beta1policyrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdatePolicyResponse](#v1beta1updatepolicyresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Project

### /v1beta1/projects

#### POST
##### Summary

Create oroject

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1ProjectRequestBody](#v1beta1projectrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateProjectResponse](#v1beta1createprojectresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{id}

#### GET
##### Summary

Get oroject by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetProjectResponse](#v1beta1getprojectresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete a oroject permanently forever and all of its relations

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteProjectResponse](#v1beta1deleteprojectresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update oroject by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1ProjectRequestBody](#v1beta1projectrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateProjectResponse](#v1beta1updateprojectresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{id}/admins

#### GET
##### Summary

Get all Admins of a oroject

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListProjectAdminsResponse](#v1beta1listprojectadminsresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{id}/disable

#### POST
##### Summary

Disable a oroject

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DisableProjectResponse](#v1beta1disableprojectresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{id}/enable

#### POST
##### Summary

Enable a oroject

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1EnableProjectResponse](#v1beta1enableprojectresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/projects/{id}/users

#### GET
##### Summary

Get all users of a oroject

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| permissionFilter | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListProjectUsersResponse](#v1beta1listprojectusersresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Resource

### /v1beta1/projects/{projectId}/resources

#### GET
##### Summary

Get all resources

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| projectId | path |  | Yes | string |
| namespace | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListProjectResourcesResponse](#v1beta1listprojectresourcesresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create Resource

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| projectId | path |  | Yes | string |
| body | body |  | Yes | [v1beta1ResourceRequestBody](#v1beta1resourcerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateProjectResourceResponse](#v1beta1createprojectresourceresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 200 | A successful response. | [v1beta1GetProjectResourceResponse](#v1beta1getprojectresourceresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 200 | A successful response. | [v1beta1DeleteProjectResourceResponse](#v1beta1deleteprojectresourceresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### PUT
##### Summary

Update Resource by ID

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| projectId | path |  | Yes | string |
| id | path |  | Yes | string |
| body | body |  | Yes | [v1beta1ResourceRequestBody](#v1beta1resourcerequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1UpdateProjectResourceResponse](#v1beta1updateprojectresourceresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## Relation

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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/relations/{relation}/object/{object}/subject/{subject}

#### DELETE
##### Summary

Remove a subject having a relation from an object

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| relation | path |  | Yes | string |
| object | path | objectnamespace:uuid | Yes | string |
| subject | path | subjectnamespace:uuid | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteRelationResponse](#v1beta1deleterelationresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

## User

### /v1beta1/users

#### GET
##### Summary

Get all public users

##### Description

Returns the users from all the organizations in a Shield instance. It can be filtered by keyword, organization, group and state. Additionally you can include page number and page size for pagination.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| pageSize | query | The maximum number of users to return per page. The default is 50. | No | integer |
| pageNum | query | The page number to return. The default is 1. | No | integer |
| keyword | query | The keyword to search for in name or email. | No | string |
| orgId | query | The organization ID to filter users by. | No | string |
| groupId | query | The group id to filter by. | No | string |
| state | query | The state to filter by. It can be enabled or disabled. | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListUsersResponse](#v1beta1listusersresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### POST
##### Summary

Create user

##### Description

Create a user with the given details. A user is not attached to an organization or a group by default,and can be invited to the org/group. The name of the user must be unique within the entire Shield instance. If a user name is not provided, Shield automatically generates a name from the user email. The user metadata is validated against the user metaschema. By default the user metaschema contains `labels` and `descriptions` for the user. The `title` field can be optionally added for a user-friendly name. <br/><br/>*Example:*`{"email":"john.doe@odpf.io","title":"John Doe",metadata:{"label": {"key1": "value1"}, "description": "User Description"}}`

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| body | body |  | Yes | [v1beta1UserRequestBody](#v1beta1userrequestbody) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1CreateUserResponse](#v1beta1createuserresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/self

#### GET
##### Summary

Get current user

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetCurrentUserResponse](#v1beta1getcurrentuserresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/self/groups

#### GET
##### Summary

List groups of a User

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListCurrentUserGroupsResponse](#v1beta1listcurrentusergroupsresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/self/organizations

#### GET
##### Summary

Get My Organizations

##### Description

Get all organizations the current user belongs to

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetOrganizationsByCurrentUserResponse](#v1beta1getorganizationsbycurrentuserresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}

#### GET
##### Summary

Get a user by id

##### Description

Get a user by id searched over all organizations in Shield.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetUserResponse](#v1beta1getuserresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

#### DELETE
##### Summary

Delete user

##### Description

Delete an user permanently forever and all of its relations (organizations, groups, etc)

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DeleteUserResponse](#v1beta1deleteuserresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
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
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}/disable

#### POST
##### Summary

Disable user

##### Description

Sets the state of the user as diabled.The user's membership to groups and organizations will still exist along with all it's roles for access control, but the user will not be able to log in and access the Shield instance.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1DisableUserResponse](#v1beta1disableuserresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}/enable

#### POST
##### Summary

Enable user

##### Description

Sets the state of the user as enabled. The user will be able to log in and access the Shield instance.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| body | body |  | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1EnableUserResponse](#v1beta1enableuserresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}/groups

#### GET
##### Summary

List Groups of a User

##### Description

Lists all the groups a user belongs to across all organization in Shield.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |
| role | query |  | No | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1ListUserGroupsResponse](#v1beta1listusergroupsresponse) |
| 400 | Bad Request - The request was malformed or contained invalid parameters. | [rpcStatus](#rpcstatus) |
| 401 | Unauthorized - Authentication is required | [rpcStatus](#rpcstatus) |
| 403 | Forbidden - User does not have permission to access the resource | [rpcStatus](#rpcstatus) |
| 404 | Not Found - The requested resource was not found | [rpcStatus](#rpcstatus) |
| 500 | Internal Server Error. Returned when theres is something wrong with Shield server. | [rpcStatus](#rpcstatus) |
| default | An unexpected error response. | [rpcStatus](#rpcstatus) |

### /v1beta1/users/{id}/organizations

#### GET
##### Summary

Get Organizations by User

##### Description

Get all the organizations a user belongs to.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ------ |
| id | path |  | Yes | string |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | A successful response. | [v1beta1GetOrganizationsByUserResponse](#v1beta1getorganizationsbyuserresponse) |
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

#### v1beta1AcceptOrganizationInvitationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1AcceptOrganizationInvitationResponse | object |  |  |

#### v1beta1AddGroupUsersResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1AddGroupUsersResponse | object |  |  |

#### v1beta1AddOrganizationUsersResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1AddOrganizationUsersResponse | object |  |  |

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
| state | string |  | No |

#### v1beta1CheckResourcePermissionRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| objectId | string |  | Yes |
| objectNamespace | string |  | Yes |
| permission | string |  | Yes |

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

#### v1beta1CreateOrganizationInvitationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| invitation | [v1beta1Invitation](#v1beta1invitation) |  | No |

#### v1beta1CreateOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organization | [v1beta1Organization](#v1beta1organization) |  | No |

#### v1beta1CreateOrganizationRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| role | [v1beta1Role](#v1beta1role) |  | No |

#### v1beta1CreatePolicyResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| policy | [v1beta1Policy](#v1beta1policy) |  | No |

#### v1beta1CreateProjectResourceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| resource | [v1beta1Resource](#v1beta1resource) |  | No |

#### v1beta1CreateProjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| project | [v1beta1Project](#v1beta1project) |  | No |

#### v1beta1CreateRelationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| relation | [v1beta1Relation](#v1beta1relation) |  | No |

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

#### v1beta1DeleteOrganizationInvitationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteOrganizationInvitationResponse | object |  |  |

#### v1beta1DeleteOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteOrganizationResponse | object |  |  |

#### v1beta1DeleteOrganizationRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteOrganizationRoleResponse | object |  |  |

#### v1beta1DeletePolicyResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeletePolicyResponse | object |  |  |

#### v1beta1DeleteProjectResourceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteProjectResourceResponse | object |  |  |

#### v1beta1DeleteProjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteProjectResponse | object |  |  |

#### v1beta1DeleteRelationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1DeleteRelationResponse | object |  |  |

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

#### v1beta1GetOrganizationInvitationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| invitation | [v1beta1Invitation](#v1beta1invitation) |  | No |

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

#### v1beta1GetProjectResourceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| resource | [v1beta1Resource](#v1beta1resource) |  | No |

#### v1beta1GetProjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| project | [v1beta1Project](#v1beta1project) |  | No |

#### v1beta1GetRelationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| relation | [v1beta1Relation](#v1beta1relation) |  | No |

#### v1beta1GetUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [v1beta1User](#v1beta1user) |  | No |

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

#### v1beta1GroupRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string | The name of the group. The name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores. | Yes |
| title | string | The title can contain any UTF-8 character, used to provide a human-readable name for the group. Can also be left empty. | No |
| metadata | object | Metadata object for groups that can hold key value pairs defined in Group Metaschema. The metadata object can be used to store arbitrary information about the group such as labels, descriptions etc. The default Group Metaschema contains labels and descripton fields. Update the Group Metaschema to add more fields.<br/>*Example:*`{"labels": {"key": "value"}, "description": "Group description"}` | No |
| orgId | string | The organization ID to which the group belongs to. | No |

#### v1beta1Invitation

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| userId | string |  | No |
| orgId | string |  | No |
| groupIds | [ string ] |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| expiresAt | dateTime |  | No |

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

#### v1beta1ListOrganizationInvitationsResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| invitations | [ [v1beta1Invitation](#v1beta1invitation) ] |  | No |

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
| name | string | The name of the metaschema. The name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores. | Yes |
| schema | string | The schema of the metaschema. The schema must be a valid JSON schema.Please refer to <https://json-schema.org/> to know more about json schema. | Yes |

#### v1beta1Namespace

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1Organization

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| id | string |  | No |
| name | string |  | No |
| title | string |  | No |
| metadata | object |  | No |
| createdAt | dateTime |  | No |
| updatedAt | dateTime |  | No |

#### v1beta1OrganizationRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string | The name of the organization. The name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores.<br/>*Example:*`"shield-org1-acme"` | Yes |
| title | string | The title can contain any UTF-8 character, used to provide a human-readable name for the organization. Can also be left empty.<br/> *Example*: `"Acme Inc"` | No |
| metadata | object | Metadata object for organizations that can hold key value pairs defined in Organization Metaschema. The metadata object can be used to store arbitrary information about the organization such as labels, descriptions etc. The default Organization Metaschema contains labels and descripton fields. Update the Organization Metaschema to add more fields. <br/>*Example*:`{"labels": {"key": "value"}, "description": "Organization description"}` | No |

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

#### v1beta1PolicyRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| roleId | string |  | No |
| title | string |  | No |
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

#### v1beta1ProjectRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
| title | string |  | No |
| metadata | object |  | No |
| orgId | string |  | No |

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

#### v1beta1RelationRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| object | string |  | No |
| subject | string |  | No |
| relation | string |  | No |
| subjectSubRelation | string |  | No |

#### v1beta1RemoveGroupUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1RemoveGroupUserResponse | object |  |  |

#### v1beta1RemoveOrganizationUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| v1beta1RemoveOrganizationUserResponse | object |  |  |

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

#### v1beta1ResourceRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string |  | No |
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

#### v1beta1UpdateOrganizationResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| organization | [v1beta1Organization](#v1beta1organization) |  | No |

#### v1beta1UpdateOrganizationRoleResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| role | [v1beta1Role](#v1beta1role) |  | No |

#### v1beta1UpdatePolicyResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| policies | [ [v1beta1Policy](#v1beta1policy) ] |  | No |

#### v1beta1UpdateProjectResourceResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| resource | [v1beta1Resource](#v1beta1resource) |  | No |

#### v1beta1UpdateProjectResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| project | [v1beta1Project](#v1beta1project) |  | No |

#### v1beta1UpdateUserResponse

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| user | [v1beta1User](#v1beta1user) |  | No |

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

#### v1beta1UserRequestBody

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| name | string | The name of the user. The name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores and must start with a letter. If not provided, Shield automatically generates a name from the user email.  | No |
| email | string | The email of the user. The email must be unique within the entire Shield instance.<br/>*Example:*`"john.doe@odpf.io"` | Yes |
| metadata | object | Metadata object for users that can hold key value pairs pre-defined in User Metaschema. The metadata object can be used to store arbitrary information about the user such as label, description etc. By default the user metaschema contains labels and descriptions for the user. Update the same to add more fields to the user metadata object. <br/>*Example:*`{"label": {"key1": "value1"}, "description": "User Description"}` | No |
| title | string | The title can contain any UTF-8 character, used to provide a human-readable name for the user. Can also be left empty. <br/>*Example:*`"John Doe"` | No |
