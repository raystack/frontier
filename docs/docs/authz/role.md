import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Roles

A Role is a collection of permissions in Shield. Roles are typically associated with one or more policies, which specify the permissions granted to the users. When a user is assigned a role, they inherit the permissions defined within that role. This simplifies access management by allowing administrators to assign roles to users rather than individually assigning permissions.

Roles in Shield is used to implement the [Role based acces control (RBAC)](../concepts/glossary.md#rbac)

### Predefined Roles

| **Role Name**                | **Permissions**                                                                                            | **Description**                                                                                                 |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| **app_organization_owner**   | app_organization_administer                                                                                | Grants administrative privileges for managing the organization and all the projects, groups and users under it. |
| **app_organization_manager** | app_organization_update<br/>app_organization_get                                                           | Allows updating and retrieving organization information including its resources.                                |
| **app_organization_viewer**  | app_organization_get                                                                                       | Allows retrieving or viewing a specific organization and its underlying resources.                              |
| **app_project_owner**        | app_project_administer                                                                                     | Grants administrative privileges for managing the project.                                                      |
| **app_project_manager**      | app_project_update<br/>app_project_get<br/>app_organization_projectcreate<br/>app_organization_projectlist | Allows updating, retrieving project information, creating and listing organization projects                     |
| **app_project_viewer**       | app_project_get                                                                                            | Allows retrieving or accessing a specific project.                                                              |
| **app_group_owner**          | app_group_administer                                                                                       | Grants administrative privileges for managing the group.                                                        |

Besides this a Shield Superuser can add custom roles at a particular namespace if required.

### Deactivate and Reactivate a role

In Shield, a role can be disabled when it needs to be temporarily or permanently deactivated. Disabling a role restricts users assigned to that role from exercising the permissions associated with it. A role can be disabled due to policy or regulatory compliance violations, temporary suspensions during investigations or disciplinary actions, or security concerns regarding potential compromise or misuse. If there are security concerns surrounding a role, such as suspected compromise or misuse, disabling the role immediately revokes the associated permissions, mitigating potential risks and safeguarding the system and its resources.

### Deleting a role

:::danger
Caution should be exercised when deleting a role, as it cannot be undone. Deleting a Role is fatal as it permanently removes the role and the all associated policies created for granting a role to a principal. Users will not be able to perform any actions or access any resources that are associated with the deleted role.
:::

### Custom Roles and How to use the roles for access control

In continuation of example from custom permissions in the previous page, let's see how the bird's eye view of how we will create custom roles and attach it to a principal in Shield.

To start using Shield Roles, follow these steps:

1. **Define Roles**: Identify the different roles you need in your system. Think about the specific sets of permissions each role should have.<br/><br/> For example, you might have roles like **CartAdmin**, **CartManager** and **CartUser** each with different levels of access to our e-commerce shopping carts.

2. **Create Custom Permissions**: If the predefined permissions in Shield do not meet your requirements, you can create [custom permissions](permission.md#custom-permissions). Custom permissions allow you to define granular access control tailored to your application's needs.

3. **Assign Permissions to Roles**: Once you have defined the roles and permissions, assign the appropriate permissions to each role. Consider the specific actions and resources each role should have access to. <br/><br/>For example, **CartManager** will have **potato_cart_update** and **potato_cart_get** permissions from the previous example for managing our e-commerce shopping cart.

4. **Assign Roles to Users or Groups via Policy**: Finally, assign roles to individual users or groups. This determines the access rights and privileges for each user or group in your org. Users inherit the permissions associated with the roles they are assigned. We will read more on this in the next pages.

:::note
Shield superusers can create a custom role platform wide which mean this role will be visible in all the organizations in the Shield instance.
Shield also provides flexibility to Org Admins to create a custom role specific to the organization.
:::

---

## Managing Roles

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
| **state**       | string   | Represents the status of the role indicating whether it can be used in a policy or not. One of **`enabled`** or **`disabled`**                                                                                                                                                                      |

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
