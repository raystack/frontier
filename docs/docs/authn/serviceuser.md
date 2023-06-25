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

The response of this request will contain an id and secret field. The secret field is never persisted in the database
and is only returned once. If you lose the secret, you will have to generate a new one.
To authenticate a service user, you need to pass the client id and secret in the **Authorization ** header as follows:

<Tabs groupId="api">
    <TabItem value="HTTP" label="HTTP" default>
    <CodeBlock className="language-bash">
    {`$ curl --location 'http://localhost:7400/v1beta1/users/self'
    --header 'Content-Type: application/json'
    --header 'Accept: application/json'
    --header 'Authorization: Basic {base64(id:secret)}'`}
    </CodeBlock>
    </TabItem>
</Tabs>

## Public/Private Key Pair

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
    --header 'Content-Type: application/json'
    --header 'Accept: application/json'`}
    </CodeBlock>
    </TabItem>
</Tabs>

To authenticate a request, you need to generate a JWT token using the private key and pass it in the **Authorization**
header as follows:
<Tabs groupId="api">
<TabItem value="HTTP" label="HTTP" default>
<CodeBlock className="language-bash">
{`$ curl --location 'http://localhost:7400/v1beta1/users/self'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--header 'Authorization: Bearer <jwt token>'`}
</CodeBlock>
</TabItem>
</Tabs>