import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Permissions

Permissions determine what operations are allowed on a resource. In Shield, permissions are represented in the form of `service.resource.verb`, for example, `potato.cart.list`.

Permissions often correspond one-to-one with API methods. That is, each service has an associated set of permissions for each API method that it exposes. The caller of that method needs those permissions to call that method. For example, if you want to create a new project you must have the projecr create permission.

You don't grant permissions to users directly. Instead, you identify roles that contain the appropriate permissions, and then grant those roles to the user.

Shield comes with predefined permissions which allow users to perform common operations on IAM resources itself. For example managing organisation, assigning roles to users.

Permissions in Shield are attached to a logical partition called Namespace.

#### Namespace

A Namespace in Shield is a logical container or partitioning mechanism that helps organize and manage resources and access control policies. It is used to group related entities and define the scope of authorization within a system. With the namespaces we compartmentalize and control access to different parts of the application or system, allowing more granular authorization and improved security.

Each namespace can have its own set of permissions, roles, and access control policies, enabling fine-grained control over who can access what resources within the defined scope. For example, the namespace `app/organization` in Shield contains permissions related to managing organizations. These permissions could include `create` `update` `delete` `get` and so on.

### Predefined Org Permissions

Includes a list of Shield's predefined permissions at the organization level.

:::info
Shield allows inheritance of permissions for a hierarchical structure, where higher-level permissions grant access to lower-level entities. In this case, granting permissions at the organization level automatically extends those permissions to the projects, resources, and groups within that org.
:::

| **Permission Name**                       | **Permission Title**                 | **Description**                                                                  |
| ----------------------------------------- | ------------------------------------ | -------------------------------------------------------------------------------- |
| **_`app.organization.administer`_**       | **`Organization Administer`**        | Grants administrative privileges for managing the organization.                  |
| **_`app.organization.delete`_**           | **`Organization Delete`**            | Allows deleting the organization.                                                |
| **_`app.organization.update`_**           | **`Organization Update`**            | Allows updating or modifying the organization's information.                     |
| **_`app.organization.get`_**              | **`Organization Get`**               | Allows retrieving or accessing a specific organization.                          |
| **_`app.organization.rolemanage`_**       | **`Organization Role Manage`**       | Enables managing or controlling roles within the organization.                   |
| **_`app.organization.policymanage`_**     | **`Organization Policy Manage`**     | Enables managing or controlling access control policies within the organization. |
| **_`app.organization.projectlist`_**      | **`Organization Project List`**      | Allows listing or retrieving a list of projects within the organization.         |
| **_`app.organization.grouplist`_**        | **`Organization Group List`**        | Allows listing or retrieving a list of groups within the organization.           |
| **_`app.organization.invitationlist`_**   | **`Organization Invitation List`**   | Allows listing or retrieving a list of user invitations in the organization.     |
| **_`app.organization.projectcreate`_**    | **`Organization Project Create`**    | Allows creating new projects within the organization.                            |
| **_`app.organization.groupcreate`_**      | **`Organization Group Create`**      | Allows creating new groups within the organization.                              |
| **_`app.organization.invitationcreate`_** | **`Organization Invitation Create`** | Allows creating new invitations or access requests within the organization.      |

### Predefined Project Permissions

Includes a list of Shield's predefined permissions at the project level.

| **Permission Name**              | **Permission Title**        | **Description**                                                             |
| -------------------------------- | --------------------------- | --------------------------------------------------------------------------- |
| **_`app.project.administer`_**   | **`Project Administer`**    | Grants administrative privileges for managing the project.                  |
| **_`app.project.delete`_**       | **`Project Delete`**        | Allows deleting the project.                                                |
| **_`app.project.update`_**       | **`Project Update`**        | Allows updating or modifying the project's information.                     |
| **_`app.project.get`_**          | **`Project Get`**           | Allows retrieving or accessing a specific project.                          |
| **_`app.project.policymanage`_** | **`Project Policy Manage`** | Enables managing or controlling access control policies within the project. |
| **_`app.project.resourcelist`_** | **`Project Resource List`** | Allows listing or retrieving a list of resources within the project.        |

### Predefined Group Permissions

Contains a list of Shield's predefined permissions at the group level.

| **Permission Name**          | **Permission Title**   | **Description**                                          |
| ---------------------------- | ---------------------- | -------------------------------------------------------- |
| **_`app.group.administer`_** | **`Group Administer`** | Grants administrative privileges for managing the group. |
| **_`app.group.delete`_**     | **`Group Delete`**     | Allows deleting the group.                               |
| **_`app.group.update`_**     | **`Group Update`**     | Allows updating or modifying the group's information.    |
| **_`app.group.get`_**        | **`Group Get`**        | Allows retrieving or accessing a specific group.         |

:::note
Permissions in Shield follow a hierarchical structure, where higher-level permissions include the capabilities of lower-level permissions. For example, a higher-level permission like **adminiter** includes all the capabilities of a lower-level permission like **get**.
:::

---

### Custom Permissions

Shield allow the its platform administrators (Superusers) to create permissions to meet specific authorization requirements beyond the predefined set of permissions. Custom permissions allow for more granular control over access to resources and actions within an application or system. They enable organizations to define and enforce fine-grained access policies tailored to their unique needs.

When creating custom permissions, administrators typically define the name and scope (namespace) of the permission. The name should be descriptive and indicative of the action or resource it governs. The scope determines where the permission applies, such as the platform(making it available across all the organizations in Shield), organization, project or group.

:::note
The permissions can either be created dynamically when the Shield is running, or can either be provided in the server configurations.
:::

For example, let's say you have an e-commerce application where users can manage their shopping carts. To control access to cart-related actions, you can create custom permissions in the **potato/cart** namespace.

Sample custom permission requirements:

| **Permission Name** | **Permission Title** | **Namespace** | **Description**                                            |
| ------------------- | -------------------- | ------------- | ---------------------------------------------------------- |
| **delete**          | Cart Delete          | potato/cart   | Grants the ability to delete items from the shopping cart. |
| **update**          | Cart Update          | potato/cart   | Allows updating the contents of the shopping cart.         |
| **get**             | Cart Get             | potato/cart   | Enables retrieving the details of the shopping cart.       |

This is how the [resource config file](https://github.com/raystack/shield/blob/7cae6bf86e99fe96650c6dcd4e8207cb916b8184/test/e2e/smoke/testdata/resource/potato.yaml#L4) will look like

:::info
While creating permissions, it is important to note that the namespace is appended to the permission name to generate a permission slug. For instance, the **delete** permission name will be appended with namespace **potato/cart**, to generate the final permission slug as **potato_cart_delete** in Shield. This naming convention is used to ensure uniqueness and organization of permissions within Shield. While creating a role, we pass the list of these permission slugs being binded to that role.
:::

## Managing Permission

:::tip
Some of these APIs require special privileges to access these endpoints and to authorize these requests, users may need a Client ID/Secret or an Access token to proceed. Read [**Authorization for APIs**](../reference/api-auth.md) to learn more.
:::
