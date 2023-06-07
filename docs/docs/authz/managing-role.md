import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Managing Role

<Tabs groupId="model">
  <TabItem value="Model" label="Model" default>

| Field           | Type     | Description                                                                                                                                                                                                                                                                                         |
| --------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **id**          | uuid     | Unique Role identifier                                                                                                                                                                                                                                                                              |
| **name**        | string   | The name of the role. The name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores.<br/> _Example:"app_organization_owner"_                                                                                                 |
| **title**       | string   | The title can contain any UTF-8 character, used to provide a human-readable name for the organization. Can also be left empty. <br/>_Example: "Organization Owner"_                                                                                                                                 |
| **permissions** | string[] | List of permission slugs to be assigned to the role <br/> _Example: ["app_organization_administer"]_                                                                                                                                                                                                |
| **metadata**    | object   | Metadata object for organizations that can hold key value pairs defined in Role Metaschema. The default Role Metaschema contains labels and descripton fields. Update the Organization Metaschema to add more fields.<br/>_Example:{"labels": {"key": "value"}, "description": "Role description"}_ |
| **orgId**       | string   | Unique Organization identifier to which the role is attached                                                                                                                                                                                                                                        |
| **state**       | string   | Represents the status of the role indicating whether it can be used in a policy or not. One of **`enabled`** or **`disabled`**                                                                                                                                                              |

</TabItem>
<TabItem value="JSON" label="Sample JSON" default>

```json
{
  "role": {
    "id": "8e73f4a2-3763-4dc6-a00e-7a9aebeaa971",
    "name": "app_organization_owner",
    "permissions": ["app_organization_administer"],
    "title": "Organization Owner",
    "metadata": {
        {"labels": {"key": "value"}, "description": "Role description"}
    },
    "createdAt": "2023-06-07T05:39:56.941Z",
    "updatedAt": "2023-06-07T05:39:56.941Z",
    "orgId": "4d726cf5-52f6-46f1-9c87-1a79f29e3abf",
    "state": "enabled"
  }
}
```

</TabItem>
</Tabs>

### List Organization Roles

To get a list of custom roles created under an organization with their associated permissions. Note the default roles avaliable across the Shield platform and inherited in every organizations wont be displayed with this 

1. Using **`GET /v1beta1/organizations/:orgId/roles`** API. One can also filter with an optional parameter for role status

```bash
$ curl -L -X GET 'http://127.0.0.1:7400/v1beta1/organizations/4d726cf5-52f6-46f1-9c87-1a79f29e3abf/roles?state=enabled' \
-H 'Accept: application/json'
```
---

### List Default Roles (Across Platform)

To get the list of all the predefined roles in Shield use the **`GET /v1beta1/roles`** API. You can additionally filter the results with an optional parameter for role status.
Note: these default roles are avaliable across all the organizations in Shield (platform wide).

```bash
$ curl -L -X GET 'http://127.0.0.1:7400/v1beta1/roles?state=enabled' \
-H 'Accept: application/json'
```
---

### Create organization role

Create a custom role under an organization. This custom role will only be available for assignment to the principals within the organization.

Using the **`POST /v1beta1/organizations/:orgId/roles`** API we can create a custom org role like this:

```bash
$ curl -L -X POST 'http://127.0.0.1:7400/v1beta1/organizations/4d726cf5-52f6-46f1-9c87-1a79f29e3abf/roles' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{
  "name": "manager",
  "permissions": [
    "potato_cart_update",
    "potato_cart_get"
  ],
  "metadata": {},
  "title": "Cart Manager"
}'
```
---
### Update organization role

Update a custom role under an organization using **`PUT /v1beta1/organizations/:orgId/roles/:id`** API. Lets update the permissions of the Cart Manager role we created in the previous example.

```bash
curl -L -X PUT 'http://127.0.0.1:7400/v1beta1/organizations/4d726cf5-52f6-46f1-9c87-1a79f29e3abf/roles/c2d85306-96f4-4895-98b4-c3e5c2f3084d' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
--data-raw '{
  "name": "manager",
  "permissions": [
    "potato_cart_update",
    "potato_cart_get",
    "potato_cart_delete"
  ],
  "metadata": {},
  "title": "Cart Manager"
}'
```
---

### Delete Organization Role

Delete a custom role in an organization will also delete all the policies assocaited with the role being deleted. One can do this using **`DELETE /v1beta1/organizations/:orgId/roles/:id`** API.

```bash
curl -L -X DELETE 'http://127.0.0.1:7400/v1beta1/organizations/4d726cf5-52f6-46f1-9c87-1a79f29e3abf/roles/c2d85306-96f4-4895-98b4-c3e5c2f3084d' \
-H 'Accept: application/json'
```