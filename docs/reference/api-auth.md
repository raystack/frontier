import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Authorization for APIs

There are 3 ways to authenticate in Frontier via APIs
1. API Token
2. Client Credentials
3. JWK Token

## API Token

API Token is a long-lived token that can be used to authenticate requests to Frontier APIs.
1. Create a service account inside an organization using [Create Service User](../apis/frontier-service-create-service-user.api.mdx) API
2. Create the token using [Create Service User Token](../apis/frontier-service-create-service-user-token.api.mdx) API

:::caution
The client secret you have created is never persisted in the database and will only be displayed once. If you happen to lose the secret, you will need to generate a new one. It is crucial to save the client secret before closing the response.
:::

Use the generated value in header for subsequent requests.

```
Authorization: "Basic " + base64( client_id + ":" + client_token )
```

## Client ID and Secret

When using client id and secret on token or introspection endpoints, provide an Authorization header with a Basic auth value in the following form:

```
Authorization: "Basic " + base64( client_id + ":" + client_secret )
```

Using this header on AuthToken(/v1beta1/auth/token) endpoint will provide a short-lived access token which should be used in the Authorization header for further requests.

```
Authorization: "Bearer " + <access Token>
```

#### Creating a client id and secret

1. Create a service account inside an organization using [Create Service User](../apis/frontier-service-create-service-user.api.mdx) API
2. Create the secret using [Create Service User Secret](../apis/frontier-service-create-service-user-credential.api.mdx) API

:::caution
The client secret you have created is never persisted in the database and will only be displayed once. If you happen to lose the secret, you will need to generate a new one. It is crucial to save the client secret before closing the response.
:::

---

## JWT Token

Alternatively, a Bearer token can also be used to verify user's identity.

```
Authorization: "Bearer " + <JWT Token>
```

#### Getting the Access token issued by Frontier after user login

Access token by default is returned as part of the response header "x-user-token" after successful login with either an Email OTP or Social login. This can be requested again by sending a request to the Frontier server with the cookies containing session details on endpoint /v1beta1/users/self.

One can use this token as the Bearer token in Authorization headers.

#### Creating JWT token from Private Keys for a service user

1. Create a service account inside an organization using [Create Service User](../apis/frontier-service-create-service-user.api.mdx) API
2. Create using [Create Service User Keys](../apis/frontier-service-create-service-user-jwk.api.mdx) API

:::caution
The private key you created never persists in Frontier and is only returned once. If you lose the private key, you will have to generate a new one. Public keys for a service user can be retrieved using [**this**](../apis/frontier-service-get-service-user-jwk.api.mdx) API
:::

3. Refer [frontier-go](https://github.com/raystack/frontier-go/blob/01b6fc925b355e69d79fcde66e1f6bb5bfd475ab/pkg/serviceuser.go) to see a Golang implementation to get a JWT token from private key using various libraries. This JWT token can be used in headers for user verification.

4. Alternatively, Frontier also exposes a [Create Access Token](../apis/frontier-service-auth-token.api.mdx) API from Client ID and Secret. Use the access token returned from the API response in the headers for authentication as discussed above.

---

## X-Frontier-Email

:::danger Warning
Currently Frontier CLI and APIs also allow an identity header like `X-Frontier-Email` which can be configured via the server configurations file. This will be deprecated in the upcoming versions and should not be used in deployment.
:::
