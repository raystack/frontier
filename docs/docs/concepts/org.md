# Organization

An Organization in Shield is a top-level resource. Each Project, Group, User, and Audit logs (coming soon) belongs to an Organization. There can be multiple tenants in each Shield deployement and an Organization will usually represent one of your tenant.

An Organization serves as a container for users, groups, projects and resources. Each organization can have its own set of admins and access control policies. Users from one organization are seperated from others.

An Organization slug is must be unique in the entire Shield deployment. No two organization can have same the slug. Whenever a user creates an Organization, one is assigned the role of the Organization Admin by default.

Users are shared between organizations: Any user may belong to multiple organizations and will be able to use the same identity to navigate between organizations.
### Predefined Permissions and Roles at Organization level

| **Role**                     | **Permissions**                                  | **Description**                                                                                                                     |
| ---------------------------- | ------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------- |
| **app_organization_owner**   | app_organization_administer                      | The owner of an organization has full control over it. They can create, update, delete, and manage all aspects of the organization. |
| **app_organization_manager** | app_organization_update<br/>app_organization_get | A manager of an organization has the ability to update and manage the organization, but they cannot create or delete it.            |
| **app_organization_viewer**  | app_organization_get                             | A viewer of an organization can only view the organization's information. They cannot update or manage the organization in any way. |

Besides this an Organization Owner can add custom permissions and roles at the organization level.

### Deactivate and Reactivate an Organization

Disabling an organization suspends its access to IAM services and restricts its ability to perform IAM-related actions. This can be done for various reasons, such as temporary suspension of services, organizational restructuring, or security considerations.

Enabling or disabling an organization helps administrators manage the access and privileges of the organization's members and controls their interaction with IAM services and resources.

:::danger
Caution should be exercised when deleting an organization, as it cannot be undone. Deleting an Organization is fatal as it permanently removes the organization, including all associated groups and projects. While the user accounts themselves are not deleted, they will lose access to all the organization resources.
:::
