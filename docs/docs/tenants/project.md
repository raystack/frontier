# Project

Projects in Shield are sub-resources within an organization. They allow for logical grouping of resources and users (including groups and service users). Each project can have its own set of permissions and access controls, enabling fine-grained control over resource allocation and user management.A single organization can contain multiple projects.

Principals(user, groups, service users) can be assigned a pre-defined or a custom role at the project level if multiple resources in a project are to share the same role for a user. A Shield policy can be created for that Project namespace for enabling user to have same role for all the underlying resources. Say a user A has `app_project_viewer` role for both the applications say X and Y in a project.

If fine-grained access is required, a Shield policy can be created to attach a principal to a particular resource along with the user role. This giving flexibility to manage user authorization to organization resources at minute levels. Using two different policies, the same user A can have `app_project_viewer` role for resource X and a `app_project_manager` role for application Y.

### Predefined Permissions and Roles at Project level

| **Role**                | **Permissions**                                                                                               | **Description**                                                                                                                                                                                                                                      |
| ----------------------- | ------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **app_project_owner**   | app_project_administer                                                                                        | A project owner have permissions to manage and modify all aspects of the project, including configuration, settings, and access control. They can add or remove members, grant or revoke permissions, and perform all actions related to the project |
| **app_project_manager** | app_project_update <br/>app_project_get<br/> app_organization_projectcreate <br/>app_organization_projectlist | A project manager can update project information, create new projects within the organization, retrieve project details and list all projects under the organization.                                                                                |
| **app_project_viewer**  | app_project_get                                                                                               | A project viewer has read-only access to the application project. They can view project details, configurations, and information but cannot make any modifications or perform any actions that could alter the project's state.                      |

Besides this we can create custom permissions and roles at the project level.

### Deactivate and Reactivate a Project

Disabling an project suspends its access to granted principals and restricts its ability to perform IAM-related actions. This can be done for various reasons, such as temporary suspension of services for maintainence or updates, or security concerns.

Enabling or disabling a project doesn't removes the policies or users associated with a project resources. However, the project resources wont't be avaliable for access by the users even though they might have relevant permissions to access them. Re-enabling the project restores the expected behaviours.

### Deleting a Project

:::danger
Caution should be exercised when deleting a project, as it cannot be undone. Deleting a Project is fatal as it permanently removes the resources, including all associated users and policies.
:::
