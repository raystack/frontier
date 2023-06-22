import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Managing Users

A project in Shield looks like

<Tabs groupId="model">
  <TabItem value="Model" label="Model" default>

| Field        | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                          |
| ------------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **id**       | uuid   | Unique user identifier                                                                                                                                                                                                                                                                                                                                                                                                               |
| **name**     | string | The name of the user. The name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores and must start with a letter. If not provided, Shield automatically generates a name from the user email. <br/> _Example:"john_doe_raystack_io"_                                                                                                                          |
| **email**    | string | The email of the user. The email must be unique within the entire Shield instance.<br/> _Example:"john.doe@raystack.org"_                                                                                                                                                                                                                                                                                                            |
| **metadata** | object | Metadata object for users that can hold key value pairs pre-defined in User Metaschema. The metadata object can be used to store arbitrary information about the user such as label, description etc. By default the user metaschema contains labels and descriptions for the user. Update the same to add more fields to the user metadata object. <br/> _Example:{"label": {"key1": "value1"}, "description": "User Description"}_ |
| **title**    | string | The title can contain any UTF-8 character, used to provide a human-readable name for the user. Can also be left empty. <br/> _Example:"John Doe"_                                                                                                                                                                                                                                                                                    |

</TabItem>
<TabItem value="JSON" label="Sample JSON" default>

```json
{
  "user": {
    "id": "598688c6-8c6d-487f-b324-ef3f4af120bb",
    "name": "john_doe_raystack_io",
    "title": "John Doe",
    "email": "john.doe@raystack.org",
    "metadata": {
      "description": "\"Shield human user\""
    },
    "createdAt": "2022-12-09T10:45:19.134019Z",
    "updatedAt": "2022-12-09T10:45:19.134019Z"
  }
}
```

</TabItem>
</Tabs>

**Note:** The metadata values are validated while creating and updating User using MetaSchemas in Shield [Read More](../reference/metaschemas.md)

---

### Create users

Create a user in Shield with the given details. A user is not attached to an organization or a group by default, and can be invited to the org/group. The name of the user must be unique within the entire Shield instance. If a user name is not provided, Shield automatically generates a name from the user email.

1. Using **`POST /v1beta1/users`** API
2. Using  **`shield user create`** CLI command

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>

```bash    
curl -L -X POST 'http://127.0.0.1:7400/v1beta1/users' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{
  "name": "john_doe_raystack_org",
  "email": "john.doe@raystack.org",
  "metadata": {
    "description": "Some User Description"
  },
  "title": "John Doe"
}'
```
</TabItem>
<TabItem value="CLI" label="CLI" default>

```bash
$ shield user create --file=user.yaml
```
</TabItem>
</Tabs>

---

### List all users

Lists all the users from all the organizations in a Shield instance. The results can be filtered by keyword, organization, group and state.

1. Using **`GET /v1beta1/users`** API
2. Using **`shield user list`** CLI command (will list all users, no flags for filtering currently)

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>

```bash
$ curl -L -X GET 'http://127.0.0.1:7400/v1beta1/admin/users?pageSize=10&pageNum=2&keyword=John&orgId=4d726cf5-52f6-46f1-9c87-1a79f29e3abf&groupId=c2d85306-96f4-4895-98b4-c3e5c2f3084d&state=enabled' \
-H 'Accept: application/json'
```
  </TabItem>
  <TabItem value="CLI" label="CLI" default>

```bash
$ shield user list 
```
  </TabItem>
</Tabs>

---

### Get User

Query a user in Shield by its unique id. In Shield a user can belong to more than one organizations, and result will get us a user in the entire Shield instance and not specific to an organization. 

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>

```bash
curl -L -X GET 'http://127.0.0.1:7400/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608%20' \
-H 'Accept: application/json'
```
  </TabItem>
<TabItem value="CLI" label="CLI" default>

```bash
$ shield user view e9fba4af-ab23-4631-abba-597b1c8e6608
```
  </TabItem>
</Tabs>

---

### List User Organizations

To get all the organizations a particular user belongs to utilize the **`GET /v1beta1/users/:id/organizations`** API

```bash
$ curl -L -X GET 'http://127.0.0.1:7400/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608/organizations' \
-H 'Accept: application/json'
```

---

### Accept and Decline an Organization Invitation

In Shield a user can be invited to an Organization and it's underlying Groups. A User can Accept or Descline the invitation with the following APIs.

To Accept the invitation use **`POST /v1beta1/organizations/:orgId/invitations/:id/accept`** API

```bash
curl -L -X POST 'http://127.0.0.1:7400/v1beta1/organizations/4d726cf5-52f6-46f1-9c87-1a79f29e3abf/invitations/e9fba4af-ab23-4631-abba-597b1c8e6608/accept' \
-H 'Accept: application/json'
```

To Decline the invitation use **`DELETE /v1beta1/organizations/:orgId/invitations/:id`** API. Here id is the unique invitation identifier.

```bash
curl -L -X DELETE 'http://127.0.0.1:7400/v1beta1/organizations/4d726cf5-52f6-46f1-9c87-1a79f29e3abf/invitations/8e73f4a2-3763-4dc6-a00e-7a9aebeaa971' \
-H 'Accept: application/json'
```

(TODO) list user invitations API is currently missing 

---
### Update User

To Update a user title, metadata, name use **`PUT /v1beta1/users/:id`** API. <br/>**Note:** it is not allowed to update a user email in Shield.

Shield also provides **`shield user edit`** CLI command for this operation.

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>

```bash
curl -L -X PUT 'http://127.0.0.1:7400/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{
  "name": "john_doe_raystack_org",
  "email": "john.doe@raystack.org",
  "metadata": {
    "description": "Update User Description"
  },
  "title": "Johnny Doe"
}'
```
  </TabItem>
<TabItem value="CLI" label="CLI" default>

```bash
$ shield user edit e9fba4af-ab23-4631-abba-597b1c8e6608 --file=user.yaml
```
  </TabItem>
</Tabs>

---

### Disable and Enable a User

This operation temporarily remove the user from being invited to Organizations(and Groups), also disabling it's access to the Organization resources. The user won't be authenticated in the first place viaShield. But its organization/group memberships, policies remain intact. Once the user is enabled again, all it's relations are restored back.

To disable a user use **`POST /v1beta1/users/:id/disable`** API

```bash
curl -L -X POST 'http://127.0.0.1:7400/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608/disable' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{}'
```

To enable it back  use **`POST /v1beta1/users/:id/enable`** API

```bash
curl -L -X POST 'http://127.0.0.1:7400/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608/enable' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{}'
```

---

### Delete User

Deleting a Shield user deletes it from the Shield instance and all it's relations to an Organization or Group. All the policies created for that user for access control and invitations for that user too is deleted. 

Use **`DELETE /v1beta1/users/:id`** API for this operation.

```bash
curl -L -X DELETE 'http://127.0.0.1:7400/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608' \
-H 'Accept: application/json'
```