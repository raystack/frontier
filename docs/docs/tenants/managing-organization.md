import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Managing Organization

An Organization in Shield is a top-level resource. Each Project, Group, User, and Audit logs (coming soon) belongs to an Organization. There can be multiple tenants in each Shield deployement and an Organization will usually represent one of your tenant.

Read More about an Organization in the [Concepts](../concepts/org.md) section.

<Tabs groupId="model">
  <TabItem value="Model" label="Model" default>

| Field        | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| ------------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --- |
| **id**       | uuid   | Unique Organization identifier                                                                                                                                                                                                                                                                                                                                                                                                                      |
| **name**     | string | The name of the organization. The name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores.<br/> _Example:"shield-org1-acme"_                                                                                                                                                                                                                                               |
| **title**    | string | The title can contain any UTF-8 character, used to provide a human-readable name for the organization. Can also be left empty. <br/>_Example: "Acme Inc"_                                                                                                                                                                                                                                                                                           |     |
| **metadata** | object | Metadata object for organizations that can hold key value pairs defined in Organization Metaschema. The metadata object can be used to store arbitrary information about the organization such as labels, descriptions etc. The default Organization Metaschema contains labels and descripton fields. Update the Organization Metaschema to add more fields.<br/>_Example:{"labels": {"key": "value"}, "description": "Organization description"}_ |

</TabItem>
<TabItem value="JSON" label="Sample JSON" default>

```json
{
  "organization": {
    "id": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8",
    "name": "open-dataops-foundation",
    "title": "Open DataOps Foundation (ODPF)",
    "metadata": {
      "description": "Some Organization details"
    },
    "createdAt": "2023-05-07T14:10:42.755848Z",
    "updatedAt": "2023-05-07T14:10:42.755848Z"
  }
}
```
</TabItem>
</Tabs>

**Note:** Organization metadata values are validated using MetaSchemas in Shield [Read More](../reference/metaschemas.md)

### Create an Organization

You can create organizations using either the Admin Portal, Shield Command Line Interface or via the HTTP APIs.

1. Using `shield organization create` CLI command
2. Calling to `POST /v1beta1/organizations` API

<Tabs groupId="api">
  <TabItem value="http" label="HTTP">

```bash
$ curl -L -X POST 'http://127.0.0.1:7400/v1beta1/organizations' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{
  "name": "odpf",
  "title": "Open DataOps Foundation",
  "metadata": {"description": "Open DataOps Foundation"}
}'
```

  </TabItem>
<TabItem value="cli" label="CLI" default>

```bash
$ shield organization create --file=<path to the organization.json file>
```

  </TabItem>
</Tabs>

3. To create an organization via the Admin Portal:

  i. Navigate to **Admin Portal > Organizations** from the sidebar

  ii. Select **+ New Organization** from top right corner

  iii. Enter basic information for your organization, and select **Add Organization**.

### Add users to an Organization

Add a user to an organization. A user must exists in Shield before adding it to an org. This request will fail if the user doesn't exists.

1. Using the **`POST /v1beta1/organizations/:id/users`** API

```bash
$ curl -L -X POST 'http://127.0.0.1:7400/v1beta1/organizations/:id/users' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{
  "userIds": [
    "2e73f4a2-3763-4dc6-a00e-7a9aebeaa971"
  ]
}'
```

### Invite users to an Organization

Invite users to an organization, if the user doesn't exists, it will be created and notified. Invitations expire in 7 days.

1. Using **`POST /v1beta1/organizations/:orgId/invitations`** API

```bash
$ curl -L -X POST 'http://127.0.0.1:7400/v1beta1/organizations/:orgId/invitations' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{
  "userId": "8e73f4a2-3763-4dc6-a00e-7a9aebeaa971",
  "groupIds": [
    "4e76f4a2-3763-4dc6-b00e-7a9a5beaa923",
    "2e45f4a2-1539-4df6-a70e-7a9aebeaa952"
  ]
}'
```

### List pending invitations

Get all the pending invitations queued for an organization.

1. Using **`GET /v1beta1/organizations/:orgId/invitations`** API

```bash
$ curl -L -X GET 'http://127.0.0.1:7400/v1beta1/organizations/:orgId/invitations' \
-H 'Accept: application/json'
```

### Delete pending invitations

Delete a pending invitation queued for an organization

1. Using **`DELETE /v1beta1/organizations/:orgId/invitations/:id`** API

```bash
$ curl -L -X DELETE 'http://127.0.0.1:7400/v1beta1/organizations/:orgId/invitations/:id' \
-H 'Accept: application/json'
```

### View an organization users

1. Calling to `GET /v1beta1/organizations/{id}/users` API

```bash
curl -L -X GET 'http://127.0.0.1:7400/v1beta1/organizations/:id/users' \
-H 'Accept: application/json'
```

### Create custom permissions for an organization

1. Using the **`POST /v1beta1/permissions`** API. <br/>**Note:** Specify the namespace **`app/organization`** to create a org level custom permission

```bash
$ curl -L -X POST 'http://127.0.0.1:7400/v1beta1/permissions' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{
  "bodies": [
    {
      "name": "customrole",
      "namespace": "app/organization",
      "metadata": {},
      "title": "Custom Role"
    }
  ]
}'
```
### Enable or disable an org

Sets the state of the organization as disabled. The existing users in the org will not be able to access any organization resources when the org is disabled.

1. Using **`POST /v1beta1/organizations/:id/disable`** API

```bash
$ curl --location 'http://localhost:7400/v1beta1/organizations/adf997e8-59d1-4462-a4f2-ab02f60a86e7/disable' 
--header 'Content-Type: application/json' 
--header 'Accept: application/json' 
--data '{}'
```

To Enable the Org again send the request on **`POST /v1beta1/organizations/:id/enable`** API in same way as described above

### Remove an Org user

Removes a user from the organization. The user however will not br deleted from Shield.

1. Calling to **`DELETE {HOST}/v1beta1/organizations/:id/users/:userId`** API

```bash
$ curl -L -X DELETE 'http://127.0.0.1:7400/v1beta1/organizations/:id/users/:userId' \
-H 'Accept: application/json'
```