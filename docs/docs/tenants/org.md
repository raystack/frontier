import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Organization

## Overview 

An Organization in Shield is a top-level resource. Each Project, Group, User, and Audit logs (coming soon) belongs to an Organization. There can be multiple tenants in each Shield deployement and an Organization will usually represent one of your tenant.

An Organization serves as a container for users, groups, projects and resources. Each organization can have its own set of admins and access control policies. Users from one organization are seperated from others.

An Organization slug is must be unique in the entire Shield deployment. No two organization can have same the slug. Whenever a user creates an Organization, one is assigned the role of the Organization Admin by default.

Users are shared between organizations: Any user may belong to multiple organizations and will be able to use the same identity to navigate between organizations.

---
### Permissions and Roles at Organization level

| **Role**                     | **Permissions**                                  | **Description**                                                                                                                     |
| ---------------------------- | ------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------- |
| **app_organization_owner**   | app_organization_administer                      | The owner of an organization has full control over it. They can create, update, delete, and manage all aspects of the organization. |
| **app_organization_manager** | app_organization_update<br/>app_organization_get | A manager of an organization has the ability to update and manage the organization, but they cannot create or delete it.            |
| **app_organization_viewer**  | app_organization_get                             | A viewer of an organization can only view the organization's information. They cannot update or manage the organization in any way. |

Besides this an Organization Owner can add custom permissions and roles at the organization level.

---
### Deactivating an Organization

Disabling an organization suspends its access to IAM services and restricts its ability to perform IAM-related actions. This can be done for various reasons, such as temporary suspension of services, organizational restructuring, or security considerations.

Enabling or disabling an organization helps administrators manage the access and privileges of the organization's members and controls their interaction with IAM services and resources.

---
### Deleting an Organization

:::danger
Caution should be exercised when deleting an organization, as it cannot be undone. Deleting an Organization is fatal as it permanently removes the organization, including all associated groups and projects. While the user accounts themselves are not deleted, they will lose access to all the organization resources.
:::

---

## Managing Organization

An Organization in Shield is a top-level resource. Each Project, Group, User, and Audit logs (coming soon) belongs to an Organization. There can be multiple tenants in each Shield deployement and an Organization will usually represent one of your tenant.

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
    "title": "Raystack",
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

:::tip
Some of these APIs require special privileges to access these endpoints and to authorize these requests, users may need a Client ID/Secret or an Access token to proceed. Read [**Authorization for APIs**](../reference/api-auth.md) to learn more.
:::

### Create an Organization

Any (human) user can create an Organization in Shield. The user must exist in Shield before procceding to create an Organization. This user will be assigned the role of **Organization Admin** once the action is completed successful and will contain all the necessary permissions to manage the org.

**Required Authorization:** User request just needs to be authenticated and no particular roles/permission are required to proceed.

You can create organizations using either the Admin Portal, Shield Command Line Interface or via the HTTP APIs.

1. Using `shield organization create` CLI command
2. Calling to `POST /v1beta1/organizations` API

<Tabs groupId="api">
  <TabItem value="http" label="HTTP">

```bash
$ curl -L -X POST 'http://127.0.0.1:7400/v1beta1/organizations' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1zZWNyZXQ=' \
--data-raw '{
  "name": "raystack",
  "title": "Raystack Foundation",
  "metadata": {"description": "Raystack Foundation"}
}'
```

  </TabItem>
<TabItem value="cli" label="CLI" default>

```bash
$ shield organization create --file=<path to the organization.json file>
```

  </TabItem>
</Tabs>

3. To create an organization via the **Admin Portal**:
   1. Navigate to **Admin Portal > Organizations** from the sidebar
   2. Select **+ New Organization** from top right corner
   3. Enter basic information for your organization, and select **Add Organization**.

---
### Add Users to an Organization

Add users to an organization. A user must exists in Shield before adding it to an org. This request will fail if the user doesn't exists. This API accepts a list of comma seperated (human) user IDs to be added to the organization. In case a user is already a member of the organization, no action will be taken for that particular user.

**Required Authorization** : A user/service account with the role `Organization Admin` or `Organization Manager` is authorized to take this action. Also any user being assigned with custom role with `update` permission at `app/organization` namespace can also perform this action. 

1. Using the **`POST /v1beta1/organizations/:id/users`** API

```bash
$ curl -L -X POST 'http://127.0.0.1:7400/v1beta1/organizations/:id/users' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1zZWNyZXQ=' \
--data-raw '{
  "userIds": [
    "2e73f4a2-3763-4dc6-a00e-7a9aebeaa971"
  ]
}'
```
---
### Invite user to an Organization

Invite user with email to an organization and optionally to the groups within that Org. A user must exists in Shield before sending the invitation otherwise the request will fail. Invitations expire in 7 days.

**Required Authorization** : A user/service account with the role `Organization Admin` is authorized to take this action. Also any user being assigned with custom role with `invitationcreate` permission at `app/organization` namespace can also perform this action. 

1. Using **`POST /v1beta1/organizations/:orgId/invitations`** API

```bash
$ curl -L -X POST 'http://127.0.0.1:7400/v1beta1/organizations/:orgId/invitations' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1zZWNyZXQ=' \
--data-raw '{
  "userId": "john.doe@raystack.org",
  "groupIds": [
    "4e76f4a2-3763-4dc6-b00e-7a9a5beaa923",
    "2e45f4a2-1539-4df6-a70e-7a9aebeaa952"
  ]
}'
```
---
### List pending invitations

This API can be used to list all the pending invitations queued for an organization which the user haven't take any action on. Once the user accepts an invitation or rejects it, the invitation is deleted from the Shield database. 

**Required Authorization** : A user/service account with the role `Organization Admin` is authorized to take this action. Also any user being assigned with custom role with `invitationlist` or `invitationcreate` permission at the `app/organization` namespace can also perform this action. 

1. Using **`GET /v1beta1/organizations/:orgId/invitations`** API

```bash
$ curl -L -X GET 'http://127.0.0.1:7400/v1beta1/organizations/:orgId/invitations' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1zZWNyZXQ=' 
```
---
### Delete pending invitations

Delete a pending invitation queued for an organization.

**Required Authorization** : Any user/service account with the role `Organization Admin` is authorized to take this action. Also the user for whom the invite was created can delete the invite. Any user being assigned with custom role with `invitationcreate` permission at the `app/organization` namespace for that particular organization can also perform this action. 

1. Using **`DELETE /v1beta1/organizations/:orgId/invitations/:id`** API

```bash
$ curl -L -X DELETE 'http://127.0.0.1:7400/v1beta1/organizations/:orgId/invitations/:id' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1zZWNyZXQ=' 
```
---
### View an organization users

This API only returns a list of human users in an Organization. For listing service users of the Org refer this [API](../apis/shield-service-list-service-users.api.mdx)

**Required Authorization** : Any user/service user with `get` permission at Organization level can perform this action. Meaning all the Organization users/service users with role `Organization Admin` , `Organization Manager` or an `Organization Viewer` can list the org users.

1. Calling to `GET /v1beta1/organizations/{id}/users` API

```bash
curl -L -X GET 'http://127.0.0.1:7400/v1beta1/organizations/:id/users' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1zZWNyZXQ=' 
```
---
### Create custom role for an Organization

**Required Authorization** : Any user/service user with the role `Organization Admin` is authorized to take this action. Also any user being assigned a custom role with `rolemanage` permission at the `app/organization` namespace can also perform this action. 

1. Using the **`POST /v1beta1/permissions`** API. <br/>**Note:** Specify the namespace **`app/organization`** to create a org level custom permission

```bash
$ curl -L -X POST 'http://127.0.0.1:7400/v1beta1/organizations/92f69c3a-334b-4f25-90b8-4d4f3be6b825/roles' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1zZWNyZXQ=' \
--data-raw '{
  "name": "app_organization_member",
  "permissions": [
    "app_organization_get",
    "app_organization_update",
    "app_organization_invitationcreate"
  ],
  "metadata": {"decription": "Custom Role with managing invitations permissions"},
  "title": "Organization Member"
}'
```
---
### Enable or disable an Organization

Sets the state of the organization as disabled. The existing users in the org will not be able to access any organization resources when the org is disabled.

**Required Authorization** : Any user/service account with the role `Organization Admin` is authorized to take this action. Also any user being assigned with custom role with `delete` permission at the `app/organization` namespace can also perform this action. 

1. Using **`POST /v1beta1/organizations/:id/disable`** API

```bash
$ curl --location 'http://localhost:7400/v1beta1/organizations/adf997e8-59d1-4462-a4f2-ab02f60a86e7/disable' \
--header 'Content-Type: application/json' \
--header 'Accept: application/json' \
--header 'Authorization: Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1zZWNyZXQ=' \
--data '{}'
```

To Enable the Org again send the request on **`POST /v1beta1/organizations/:id/enable`** API in same way as described above

---
### Remove Organization User

Removes a user from the organization. The user will be removed from all the groups inside the Organization ans will lose all the access to underlying projects and resources, the user will not be deleted from the Shield instance.

**Required Authorization** : A user/service account with the role `Organization Admin` or `Organization Manager` is authorized to take this action. Also any user being assigned with custom role with `update` permission at `app/organization` namespace can also perform this action. 

1. Calling to **`DELETE {HOST}/v1beta1/organizations/:id/users/:userId`** API

```bash
$ curl -L -X DELETE 'http://127.0.0.1:7400/v1beta1/organizations/:id/users/:userId' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1zZWNyZXQ=' 
```
