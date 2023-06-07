# Permission

Each Shield deployment is populated with predefined permissions which allow users to perform common operations such as listing, getting, creating, updating, and deleting entities at different levels of the platform's hierarchy. They also include more specialized permissions like role management, policy management, and acceptance of invitations.

Each predefined permission has a specific purpose and grants a defined level of access to a particular resource or service. By assigning these permissions to users, groups, or service accounts, administrators can effectively control and manage access to an organization resources. Each permission is attached to a logical partition in Shield called Namespace.

#### Namespace

A Namespace in Shield is a logical container or partitioning mechanism that helps organize and manage resources and access control policies. It is used to group related entities and define the scope of authorization within a system. With the namespaces we compartmentalize and control access to different parts of the application or system, allowing more granular authorization and improved security.

Each namespace can have its own set of permissions, roles, and access control policies, enabling fine-grained control over who can access what resources within the defined scope. For example, the namespace `app/organization` in Shield contains permissions related to managing organizations. These permissions could include `create` `update` `delete` `get` and so on.

### Predefined Org Permissions

Includes a list of Shield's predefined permissions at the organization level.

:::info
Shield allows inheritance of permissions for a hierarchical structure, where higher-level permissions grant access to lower-level entities. In this case, granting permissions at the organization level automatically extends those permissions to the projects, resources, and groups within that org.
:::

| **Permission Title**                 | **Description**                                                                  | **Permission Name**      | **Namespace**      |
| ------------------------------------ | -------------------------------------------------------------------------------- | ------------------------ | ------------------ |
| **`Organization Administer`**        | Grants administrative privileges for managing the organization.                  | **_`administer`_**       | _app/organization_ |
| **`Organization Delete`**            | Allows deleting the organization.                                                | **_`delete`_**           | _app/organization_ |
| **`Organization Update`**            | Allows updating or modifying the organization's information.                     | **_`update`_**           | _app/organization_ |
| **`Organization Get`**               | Allows retrieving or accessing a specific organization.                          | **_`get`_**              | _app/organization_ |
| **`Organization Role Manage`**       | Enables managing or controlling roles within the organization.                   | **_`rolemanage`_**       | _app/organization_ |
| **`Organization Policy Manage`**     | Enables managing or controlling access control policies within the organization. | **_`policymanage`_**     | _app/organization_ |
| **`Organization Project List`**      | Allows listing or retrieving a list of projects within the organization.         | **_`projectlist`_**      | _app/organization_ |
| **`Organization Group List`**        | Allows listing or retrieving a list of groups within the organization.           | **_`grouplist`_**        | _app/organization_ |
| **`Organization Invitation List`**   | Allows listing or retrieving a list of user invitations in the organization.     | **_`invitationlist`_**   | _app/organization_ |
| **`Organization Project Create`**    | Allows creating new projects within the organization.                            | **_`projectcreate`_**    | _app/organization_ |
| **`Organization Group Create`**      | Allows creating new groups within the organization.                              | **_`groupcreate`_**      | _app/organization_ |
| **`Organization Invitation Create`** | Allows creating new invitations or access requests within the organization.      | **_`invitationcreate`_** | _app/organization_ |

### Predefined Project Permissions

Includes a list of Shield's predefined permissions at the project level.

| **Permission Title**        | **Description**                                                             | **Permission Name**  | **Namespace** |
| --------------------------- | --------------------------------------------------------------------------- | -------------------- | ------------- |
| **`Project Administer`**    | Grants administrative privileges for managing the project.                  | **_`administer`_**   | _app/project_ |
| **`Project Delete`**        | Allows deleting the project.                                                | **_`delete`_**       | _app/project_ |
| **`Project Update`**        | Allows updating or modifying the project's information.                     | **_`update`_**       | _app/project_ |
| **`Project Get`**           | Allows retrieving or accessing a specific project.                          | **_`get`_**          | _app/project_ |
| **`Project Policy Manage`** | Enables managing or controlling access control policies within the project. | **_`policymanage`_** | _app/project_ |
| **`Project Resource List`** | Allows listing or retrieving a list of resources within the project.        | **_`resourcelist`_** | _app/project_ |

### Predefined Group Permissions

Contains a list of Shield's predefined permissions at the group level.

| **Permission Title**   | **Description**                                          | **Namespace**      | **Permission Name** |
| ---------------------- | -------------------------------------------------------- | ------------------ | ------------------- |
| **`Group Administer`** | Grants administrative privileges for managing the group. | **_`administer`_** | _app/group_         |
| **`Group Delete`**     | Allows deleting the group.                               | **_`delete`_**     | _app/group_         |
| **`Group Update`**     | Allows updating or modifying the group's information.    | **_`update`_**     | _app/group_         |
| **`Group Get`**        | Allows retrieving or accessing a specific group.         | **_`get`_**        | _app/group_         |

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

This is how the [resource config file](https://github.com/odpf/shield/blob/7cae6bf86e99fe96650c6dcd4e8207cb916b8184/test/e2e/smoke/testdata/resource/potato.yaml#L4) will look like

:::info
While creating permissions, it is important to note that the namespace is appended to the permission name to generate a permission slug. For instance, the **delete** permission name will be appended with namespace **potato/cart**, to generate the final permission slug as **potato_cart_delete** in Shield. This naming convention is used to ensure uniqueness and organization of permissions within Shield. While creating a role, we pass the list of these permission slugs being binded to that role. 
:::