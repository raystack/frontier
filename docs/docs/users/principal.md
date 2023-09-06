import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Principal

Principal in Frontier are entities that can be authenticated and authorized to access resources and perform actions. A principal will be used in controlling access to resources, as they can be associated with policies in Frontier that define what actions the principal can perform and which resources they can access.

### Types of Principals in Frontier

- **User**: A User Principal in Frontier refers to an individual user who is granted certain privileges and access rights within the system. This type of principal is associated with a specific user account and is used to authenticate and authorize actions performed by the user. User principals are commonly used to manage and control access to resources, enforce security policies, and track user activities within the Frontier system.

- **Service User**: A Service User Principal is a type of principal in Frontier that represents a non-human entity, such as a software application or a service, that interacts with the system. This principal is specifically designed to authenticate and authorize actions performed by automated processes or scripts, allowing machines to securely access and manipulate resources within the Frontier environment. Machine User Principals are commonly used in scenarios where programmatic access is required without human intervention.

- **Group**: A Group in Frontier refers to a logical grouping of users or machine users who share common access rights and permissions. This type of principal simplifies the management of access control by assigning privileges and permissions to a group instead of individual users. Group principals enable administrators to efficiently define and enforce security policies across multiple users or machine users, making it easier to grant or revoke access to resources based on membership in a particular group.

- **Super User**: Superusers are platform level admins used to manage a Frontier instance. A Frontier Super User have the highest level of access and control, allowing them to perform administrative tasks, configure settings, and manage other organization accounts. A superuser email can be added in configurations yaml while setting up the Frontier server.

:::info
**Constraints**: In Frontier, a (human) user can belong to more than one organization, but a Group will belong to only one organization that created it.
:::

### Authenticating Principals

Frontier authenticates a principal using OIDC (OpenID Connect) and the Magic Link approach. [OIDC](../concepts/glossary.md#oidc) is an authentication protocol that allows applications to verify the identity of users and obtain their basic profile information. The Magic Link approach combines the simplicity of email-based authentication with the security of OIDC to create a seamless and secure authentication experience.

Currently, users authenticate with the help of an external IdP, which issues security tokens that are then exchanged with the system.

### Authorizing Principals

Once a Frontier principal is authenticated, the next step is authorization. Policies are used to define the permissions and access levels for each principal. These policies are attached to principals (users, or groups) and determine what actions they can perform on which resources.

### Deactivate and Reactivate a User

Disabling a user suspends its access to organization resources and restricts its ability to perform IAM-related actions. This can be done for various reasons, such as employee termination, account compromise, temporary suspension of services, or security considerations. When a user account is diabled it cannot be used to login to Frontier. The user account is not deleted, it will lose access to all the organization resources. The user account can be re-enabled at any time.

Enabling or disabling a user helps administrators manage the access and privileges of the users and controls their interaction with IAM services and resources.

### Delete a User

Deleting a user account permanently removes the user from the organization, including all associated groups and projects.

:::note
There must always be a minimum of one Admin in an Organization, so if the user being deleted is the sole Organization Admin, then this action should be halted until another user is promoted to the role of the Organization Admin.
:::
:::danger
Caution should be exercised when deleting a user, as it cannot be undone. If the user account with same credentials is created again, it's organization, group and project memberships will not be restored.
:::

## Managing Users

A project in Frontier looks like

<Tabs groupId="model">
  <TabItem value="Model" label="Model" default>

| Field        | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                          |
| ------------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **id**       | uuid   | Unique user identifier                                                                                                                                                                                                                                                                                                                                                                                                               |
| **name**     | string | The name of the user. The name must be unique within the entire Frontier instance. The name can contain only alphanumeric characters, dashes and underscores and must start with a letter. If not provided, Frontier automatically generates a name from the user email. <br/> _Example:"john_doe_raystack_io"_                                                                                                                      |
| **email**    | string | The email of the user. The email must be unique within the entire Frontier instance.<br/> _Example:"john.doe@raystack.org"_                                                                                                                                                                                                                                                                                                          |
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
      "description": "\"Frontier human user\""
    },
    "createdAt": "2022-12-09T10:45:19.134019Z",
    "updatedAt": "2022-12-09T10:45:19.134019Z"
  }
}
```

</TabItem>
</Tabs>

**Note:** The metadata values are validated while creating and updating User using MetaSchemas in Frontier [Read More](../reference/metaschemas.md)

---

:::tip
Some of these APIs require special privileges to access these endpoints and to authorize these requests, users may need a Client ID/Secret or an Access token to proceed. Read [**Authorization for APIs**](../reference/api-auth.md) to learn more.
:::

### Create users

Create a user in Frontier with the given details. A user is not attached to an organization or a group by default, and can be invited to the org/group. The name of the user must be unique within the entire Frontier instance. If a user name is not provided, Frontier automatically generates a name from the user email.

1. Using **`POST /v1beta1/users`** API
2. Using **`frontier user create`** CLI command

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
$ frontier user create --file=user.yaml
```

</TabItem>
</Tabs>

---

### List all users

Lists all the users from all the organizations in a Frontier instance. The results can be filtered by keyword, organization, group and state.

1. Using **`GET /v1beta1/users`** API
2. Using **`frontier user list`** CLI command (will list all users, no flags for filtering currently)

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>

```bash
$ curl -L -X GET 'http://127.0.0.1:7400/v1beta1/admin/users?pageSize=10&pageNum=2&keyword=John&orgId=4d726cf5-52f6-46f1-9c87-1a79f29e3abf&groupId=c2d85306-96f4-4895-98b4-c3e5c2f3084d&state=enabled' \
-H 'Accept: application/json'
```

  </TabItem>
  <TabItem value="CLI" label="CLI" default>

```bash
$ frontier user list
```

  </TabItem>
</Tabs>

---

### Get User

Query a user in Frontier by its unique id. In Frontier a user can belong to more than one organizations, and result will get us a user in the entire Frontier instance and not specific to an organization.

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>

```bash
curl -L -X GET 'http://127.0.0.1:7400/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608%20' \
-H 'Accept: application/json'
```

  </TabItem>
<TabItem value="CLI" label="CLI" default>

```bash
$ frontier user view e9fba4af-ab23-4631-abba-597b1c8e6608
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

### Create Organization Invitation

To invite a user to an organization, use the **`POST /v1beta1/organizations/:orgId/invitations`** API. The user will receive an email with a link to accept the invitation.
The link will do nothing but redirect the user to the Frontier login page, and after successful login, the user will be redirected to the Frontier UI to confirm the acceptance.

When an invitation is created, it can have role ids and group ids attached to it. The user will be assigned to the roles and groups when the invitation is accepted.
Adding role to the invitation can be dangerous causing policy escalation and should be used with caution. To avoid misuse of this feature by default the role ids are not added to the invitation.
To enable it, set `invite.with_roles` flag to `true` under `app` config in `frontier.yaml` file.


### Accept and Decline an Organization Invitation

In Frontier a user can be invited to an Organization and it's underlying Groups. A User can Accept or Descline the invitation with the following APIs.

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



---

### List User Invitations

- [GET /v1beta1/users/:id/invitations API](../apis/frontier-service-list-user-invitations.api.mdx)

---

### Update User

To Update a user title, metadata, name use **`PUT /v1beta1/users/:id`** API. <br/>**Note:** it is not allowed to update a user email in Frontier.

Frontier also provides **`frontier user edit`** CLI command for this operation.

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
$ frontier user edit e9fba4af-ab23-4631-abba-597b1c8e6608 --file=user.yaml
```

  </TabItem>
</Tabs>

---

### Disable and Enable a User

This operation temporarily remove the user from being invited to Organizations(and Groups), also disabling it's access to the Organization resources. The user won't be authenticated in the first place viaFrontier. But its organization/group memberships, policies remain intact. Once the user is enabled again, all it's relations are restored back.

To disable a user use **`POST /v1beta1/users/:id/disable`** API

```bash
curl -L -X POST 'http://127.0.0.1:7400/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608/disable' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{}'
```

To enable it back use **`POST /v1beta1/users/:id/enable`** API

```bash
curl -L -X POST 'http://127.0.0.1:7400/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608/enable' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{}'
```

---

### Delete User

Deleting a Frontier user deletes it from the Frontier instance and all it's relations to an Organization or Group. All the policies created for that user for access control and invitations for that user too is deleted.

Use **`DELETE /v1beta1/users/:id`** API for this operation.

```bash
curl -L -X DELETE 'http://127.0.0.1:7400/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608' \
-H 'Accept: application/json'
```
