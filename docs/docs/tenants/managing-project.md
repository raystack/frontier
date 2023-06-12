import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Managing Project

- Create a project in an org
- Create a project resource
- View a project resources
- Create a policy to attach user to a project with pre-defined roles
- View a project users
- List a project admins
- Enable or disable a project

A project in Shield looks like

<Tabs groupId="model">
  <TabItem value="Model" label="Model" default>

| Field        | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| ------------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **id**       | uuid   | Unique project identifier                                                                                                                                                                                                                                                                                                                                                                                                                       |
| **name**     | string | The name of the project. This name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores and must start with a letter. <br/> _Example:"project-alpha"_                                                                                                                                                                                                                    |
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

### Create projects

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request POST 'http://localhost:8000/v1beta1/projects'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "Project Beta",
  "slug": "project-beta",
  "metadata": {
      "description": "Project Beta"
  },
  "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8"
}'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield project create --file project.yaml --header key:value`
</CodeBlock>

  </TabItem>
</Tabs>

### List projects

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request GET 'http://localhost:8000/v1beta1/admin/projects'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield project list`
</CodeBlock>

  </TabItem>
</Tabs>

### Get Projects

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request GET 'http://localhost:8000/v1beta1/projects/457944c2-2a4c-4e6f-b1f7-3e1e109fe94c'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield project view 457944c2-2a4c-4e6f-b1f7-3e1e109fe94c --metadata`
</CodeBlock>

  </TabItem>
</Tabs>

### Update Projects

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request PUT 'http://localhost:8000/v1beta1/projects/457944c2-2a4c-4e6f-b1f7-3e1e109fe94c'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "Project Beta",
  "slug": "project-beta",
  "metadata": {
      "description": "Project Beta by Raystack"
  },
  "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8"
}'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield project edit 457944c2-2a4c-4e6f-b1f7-3e1e109fe94c --file=project.yaml`
</CodeBlock>

  </TabItem>
</Tabs>
