---
title: Service User
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Service User Authentication

Service User authentication is used to authenticate a service to another service where a human is not actively
involved in the authentication process. For example, an external service is authenticating to a backend service.
Before the authentication is started for a service user, a service user should exist in an organization. A service
user can have multiple credentials. For example, a service user can have a client id/secret and a public/private key
pair and both of them will be valid for the service user authentication.

## Client ID/Secret

Once a service user is created, a client id and secret can be generated for the service user using following API:
<Tabs groupId="api">
<TabItem value="HTTP" label="HTTP" default>
<CodeBlock className="language-bash">
{`$ curl --location --request POST 'http://localhost:7400/v1beta1/serviceusers/{id}/secrets'
--header 'Content-Type: application/json'
--header 'Accept: application/json' --data-raw '{"title": "command line tool"}'`}
</CodeBlock>
</TabItem>
</Tabs>

The response of this request will contain an `id` and `secret` field. The secret field is never persisted in the database
and is only returned once. If you lose the secret, you will have to generate a new one.
To authenticate a service user, you need to pass the client id and secret in the **Authorization ** header first to
request an access token as follows:

<Tabs groupId="api">
    <TabItem value="HTTP" label="HTTP" default>
    <CodeBlock className="language-bash">
    {`$ curl --location 'http://localhost:7400/v1beta1/auth/token'
    --header 'Accept: application/json'
    --header 'Authorization: Basic {base64(id:secret)}'
    --data-raw 'grant_type=client_credentials'`}
    </CodeBlock>
    </TabItem>
</Tabs>

The response of this request will contain a `access_token` that can be used to authenticate the service user in all
subsequent requests. The token can be used in the **Authorization** header as follows:

<Tabs groupId="api">
<TabItem value="HTTP" label="HTTP" default>
<CodeBlock className="language-bash">
{`$ curl --location 'http://localhost:7400/v1beta1/users/self'
--header 'Accept: application/json'
--header 'Authorization: Bearer <access_token>'`}
</CodeBlock>
</TabItem>
</Tabs>

By default, the access token is valid for 1 hour.

## Private JWT Grant

Once a service user is created, a public/private key pair can be generated for the service user using following API:
<Tabs groupId="api">
<TabItem value="HTTP" label="HTTP" default>
<CodeBlock className="language-bash">
{`$ curl --location --request POST 'http://localhost:7400/v1beta1/serviceusers/{id}/keys'
--header 'Content-Type: application/json'
--header 'Accept: application/json' --data-raw '{"title": "command line tool"}'`}
</CodeBlock>
</TabItem>
</Tabs>

The response of this request will be as follows:

```protobuf
message KeyCredential {
  string type = 1;
  string kid = 2;
  string principal_id = 3;

  // RSA private key as string
  string private_key = 4;
}
```

The private key is never persisted in the database and is only returned once. If you lose the private key,
you will have to generate a new one. The public key can be retrieved using the following API:
<Tabs groupId="api">
<TabItem value="HTTP" label="HTTP" default>
<CodeBlock className="language-bash">
{`$ curl --location --request GET 'http://localhost:7400/v1beta1/serviceusers/{id}/keys/{kid}'
    --header 'Accept: application/json'`}
</CodeBlock>
</TabItem>
</Tabs>

To authenticate a request, you need to generate a JWT using the private key. There are various libraries that can
use a PEM file and generate a JWT. The `private_key` is in PEM format of the KeyCredential message. One example
of generating a JWT using the private key can be found in the
[frontier-go](https://github.com/raystack/frontier-go/blob/01b6fc925b355e69d79fcde66e1f6bb5bfd475ab/pkg/serviceuser.go) SDK.

```go
package pkg

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"time"
)

type ServiceUserTokenGenerator func() ([]byte, error)

func GetServiceUserTokenGenerator(credential *frontierv1beta1.KeyCredential) (ServiceUserTokenGenerator, error) {
	// generate a token out of key
	rsaKey, err := jwk.ParseKey([]byte(credential.GetPrivateKey()), jwk.WithPEM(true))
	if err != nil {
		return nil, err
	}
	if err = rsaKey.Set(jwk.KeyIDKey, credential.GetKid()); err != nil {
		return nil, err
	}
	return func() ([]byte, error) {
		return utils.BuildToken(rsaKey, "//frontier-go-sdk", credential.GetPrincipalId(), time.Hour*12, nil)
	}, nil
}
```

To identify your key, it is necessary that you provide a JWT with a kid header claim representing your key id from the
`KeyCredential`:

```json
{
  "alg": "RS256",
  "kid": "c029a17d-0bad-472c-b335-ed58ba370d84"
}
```

The generated JWT needs to be exchanged to get an access token. It can be requested in the **Authorization** header as follows:

<Tabs groupId="api">
<TabItem value="HTTP" label="HTTP" default>
<CodeBlock className="language-bash">
{`$ curl --location 'http://localhost:7400/v1beta1/auth/token'
--header 'Accept: application/json'
--header 'Authorization: Bearer <jwt token>'
--data-raw 'grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer'`}
</CodeBlock>
</TabItem>
</Tabs>

Response of this request will contain a `access_token` that can be used to authenticate the service user.

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjE2ODAyMzgyNDQyOTQiLCJ0eXAiOiJKV1QifQ.eyJleHAiOjE2OTA4NzI0NjcsImdlbiI6InN5c3RlbSIsImlhdCI6MTY4ODI4MDQ2NywiaXNzIjoiaHR0cDovL2xvY2FsaG9zdC5zaGllbGQiLCJqdGkiOiI0ZjJiNWNlMS00MGFjLTQ4ZWMtOTM0OC0xN2RhODM1NjZmNTYiLCJraWQiOiIxNjgwMjM4MjQ0Mjk0IiwibmJmIjoxNjg4MjgwNDY3LCJvcmdzIjoiIiwic3ViIjoiMTEyODc5NmUtNmM2ZS00ZTM5LTljMjgtOWM5ZWI0NjEwMjc2In0.rDkU6WjrqlLuyQv4Vvyk-iP55C-CodnGIhk2rvR8MasV2byffdu6tRs0koTOv_SCn78bXfDxiW9vilqXeSNWBULFKixUO6095ON2ZNuQrZSVFWD9xqDrNj6wxTNRiR8g6nKJOOqFogQV7qI92-JfBguIZGPhrZbgKHYbseN2FL3ZHs1Zyi_NYh5FaMS9bIEuwGil4B_yMas10dstCVw4aSzFqsXWjPBFMSqRvRcQpOlGXo0TZWtkndiakQ3Ox2PLDRnrdlAzpTlB8kkZ5uwEjNSFgjk_fccSosNtUeuSLJ-uiT52SoujAq-yft2iOL-_tJudpS3Dsm-SODmBg1HSBw",
  "token_type": "Bearer"
}
```

The token can be used in the **Authorization** header same as the client id/secret token as follows:

<Tabs groupId="api">
<TabItem value="HTTP" label="HTTP" default>
<CodeBlock className="language-bash">
{`$ curl --location 'http://localhost:7400/v1beta1/users/self'
--header 'Accept: application/json'
--header 'Authorization: Bearer <access_token>'`}
</CodeBlock>
</TabItem>
</Tabs>

## Access token

As we already mentioned, the access token is a JWT that is generated by the frontier server. Whoever has the access token
can access the resources of the service user. Access token is used to avoid supporting multiple authentication mechanisms
in various part of the infrastructure to verify the request. It can be passed along the request in services behind the frontier server
and all of them only need to validate it using frontier server's public key. Access token endpoint is available
at `/v1beta1/auth/token`.

Token comes with few custom claims that can be used to identify the request scope. The claims are:
- org_id: IDs of the organization principal belongs to
- project_id: ID of the project the request is made to. This is useful when the user has access to multiple projects.
Ideally the frontend should be able to show a list of projects to user, and it can select one out of many and pass it 
along the request in the header "X-Project". If the user has access to this project, it will be added as a claim.
