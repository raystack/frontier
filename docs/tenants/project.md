import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Project

Projects in Frontier are sub-resources within an organization. They allow for logical grouping of resources and users (including groups and service users). Each project can have its own set of permissions and access controls, enabling fine-grained control over resource allocation and user management.A single organization can contain multiple projects.

Principals(user, groups, service users) can be assigned a pre-defined or a custom role at the project level if multiple resources in a project are to share the same role for a user. A Frontier policy can be created for that Project namespace for enabling user to have same role for all the underlying resources. Say a user A has `app_project_viewer` role for both the applications say X and Y in a project.

If fine-grained access is required, a Frontier policy can be created to attach a principal to a particular resource along with the user role. This giving flexibility to manage user authorization to organization resources at minute levels. Using two different policies, the same user A can have `app_project_viewer` role for resource X and a `app_project_manager` role for application Y.

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

# Managing Project

- Create a project in an org
- Create a project resource
- View a project resources
- Create a policy to attach user to a project with pre-defined roles
- View a project users
- List a project admins
- Enable or disable a project

A project in Frontier looks like

<Tabs groupId="model">
  <TabItem value="Model" label="Model" default>

| Field        | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| ------------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **id**       | uuid   | Unique project identifier                                                                                                                                                                                                                                                                                                                                                                                                                       |
| **name**     | string | The name of the project. This name must be unique within the entire Frontier instance. The name can contain only alphanumeric characters, dashes and underscores and must start with a letter. <br/> _Example:"project-alpha"_                                                                                                                                                                                                                  |
| **title**    | string | The title can contain any UTF-8 character, used to provide a human-readable name for the project. Can also be left empty. <br/> _Example:"Project Alpha"_                                                                                                                                                                                                                                                                                       |
| **orgId**    | uuid   | Unique Organization identifier to which the project belongs                                                                                                                                                                                                                                                                                                                                                                                     |
| **metadata** | object | Metadata object for project that can hold key value pairs pre-defined in Project Metaschema. The metadata object can be used to store arbitrary information about the user such as label, description etc. By default the user metaschema contains labels and descriptions for the project. Update the same to add more fields to the user metadata object. <br/> _Example:{"label": {"key1": "value1"}, "description": "Project Description"}_ |

</TabItem>
<TabItem value="JSON" label="Sample JSON" default>

```json
{
  "project": {
    "id": "1b89026b-6713-4327-9d7e-ed03345da288",
    "name": "project-alpha",
    "title": "Project Alpha",
    "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8",
    "metadata": {
      "description": "Project Alpha"
    },
    "createdAt": "2022-12-07T14:31:46.436081Z",
    "updatedAt": "2022-12-07T14:31:46.436081Z"
  }
}
```

</TabItem>
</Tabs>

## API Interface

:::tip
Some of these APIs require special privileges to access these endpoints and to authorize these requests, users may need a Client ID/Secret or an Access token to proceed. Read [**Authorization for APIs**](../reference/api-auth.md) to learn more.
:::

### Create projects

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>

```bash
$ curl --location --request POST 'http://localhost:8000/v1beta1/projects'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "Project Beta",
  "slug": "project-beta",
  "metadata": {
      "description": "Project Beta"
  },
  "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8"
}
```

  </TabItem>
  <TabItem value="CLI" label="CLI" default>

```bash
$ frontier project create --file project.yaml --header key:value
```

  </TabItem>
</Tabs>

### List projects

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>

```bash
$ curl --location --request GET 'http://localhost:8000/v1beta1/admin/projects'
--header 'Accept: application/json'
```

  </TabItem>
  <TabItem value="CLI" label="CLI" default>

```bash
$ frontier project list
```

  </TabItem>
</Tabs>

### Get Projects

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>

```bash
$ curl --location --request GET 'http://localhost:8000/v1beta1/projects/457944c2-2a4c-4e6f-b1f7-3e1e109fe94c'
--header 'Accept: application/json'
```

  </TabItem>
  <TabItem value="CLI" label="CLI" default>

```bash
$ frontier project view 457944c2-2a4c-4e6f-b1f7-3e1e109fe94c --metadata
```

  </TabItem>
</Tabs>

### Update Projects

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>

```bash
$ curl --location --request PUT 'http://localhost:8000/v1beta1/projects/457944c2-2a4c-4e6f-b1f7-3e1e109fe94c'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "Project Beta",
  "slug": "project-beta",
  "metadata": {
      "description": "Project Beta by Raystack"
  },
  "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8"
}'
```

  </TabItem>
  <TabItem value="CLI" label="CLI" default>
  
```bash
$ frontier project edit 457944c2-2a4c-4e6f-b1f7-3e1e109fe94c --file=project.yaml````
```
  </TabItem>
</Tabs>
