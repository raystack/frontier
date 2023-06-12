import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Managing Group

- Create an org group
- List an org groups
- Add or invite users to a group
- View a group members
- Assign roles to group members
- Remove a user from a group
- Enable or disable a group

A group in Shield looks like

<Tabs groupId="model">
  <TabItem value="Model" label="Model" default>

| Field        | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                |
| ------------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **id**       | uuid   | Unique group identifier                                                                                                                                                                                                                                                                                                                                                                                    |
| **name**     | string | The name of the group. The name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores.                                                                                                                                                                                                                                               |
| **title**    | string | The title can contain any UTF-8 character, used to provide a human-readable name for the group. Can also be left empty.                                                                                                                                                                                                                                                                                    |
| **metadata** | object | Metadata object for groups that can hold key value pairs defined in Group Metaschema. The metadata object can be used to store arbitrary information about the group such as labels, descriptions etc. The default Group Metaschema contains labels and descripton fields. Update the Group Metaschema to add more fields. <br/>_Example:{"labels": {"key": "value"}, "description": "Group description"}_ |
| **orgId**    | uuid   | The organization ID to which the group belongs to.                                                                                                                                                                                                                                                                                                                                                         |

</TabItem>
<TabItem value="JSON" label="Sample JSON" default>

```json
{
  "group": {
    "id": "2105beab-5d04-4fc5-b0ec-8d6f60b67ab2",
    "name": "Data Batching",
    "title": "data-batching",
    "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8",
    "metadata": {
      "description": "group for users in data batching domain"
    },
    "createdAt": "2022-12-14T10:22:14.394120Z",
    "updatedAt": "2022-12-14T10:25:34.890645Z"
  }
}
```

</TabItem>
</Tabs>

**Note:** group metadata values are validated using MetaSchemas in Shield [Read More](../reference/metaschemas.md)

### Create an organization group

1. Using `shield group create` CLI command
2. Calling to `POST /v1beta1/organizations/orgId/groups` API

<Tabs groupId="api">
  <TabItem value="http" label="HTTP">
  <CodeBlock className="language-bash">
{` $ curl --location --request POST 'http://localhost:8000/v1beta1/organizations/adf997e8-59d1-4462-a4f2-ab02f60a86e7/groups'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "data-batching",
  "title": "Data Batching",
  "metadata": {
      "description": "group for users in data batching domain"
  }
}'
`}</CodeBlock>
</TabItem>
 <TabItem value="cli" label="CLI">
Todo
 </TabItem>
 </Tabs>

3. To create a group via the Admin Portal:

i. Navigate to **Admin Portal > Groups** from the sidebar

ii. Select **+ New Group** from top right corner

iii. Enter basic information for the group, and select **Add Group**

### List an organization groups

1. Using `shield group list` CLI command
2. Calling to `GET /v1beta1/organizations/orgId/groups` API

### View an organization projects

<Tabs groupId="api">
  <TabItem value="http" label="HTTP">
  <CodeBlock className="language-bash">
  {` curl --location 'http://localhost:8000/v1beta1/organizations/adf997e8-59d1-4462-a4f2-ab02f60a86e7/projects' 
--header 'Accept: application/json' `}
  </CodeBlock>
</TabItem>
 <TabItem value="cli" label="CLI">
Todo
 </TabItem>
 </Tabs>

## API Interface

### Create groups

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request POST 'http://localhost:8000/v1beta1/groups'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "Data Batching",
  "slug": "data-batching",
  "metadata": {
      "description": "group for users in data batching domain"
  },
  "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8"
}'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield group create --file group.yaml --header key:value`
</CodeBlock>

  </TabItem>
</Tabs>

### List groups

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request GET 'http://localhost:8000/v1beta1/organizations/4eb3c3b4-962b-4b45-b55b-4c07d3810ca8/groups'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield group list`
</CodeBlock>

  </TabItem>
</Tabs>

### Get groups

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request GET 'http://localhost:8000/v1beta1/organizations/4eb3c3b4-962b-4b45-b55b-4c07d3810ca8/groups/2105beab-5d04-4fc5-b0ec-8d6f60b67ab2'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield group view 2105beab-5d04-4fc5-b0ec-8d6f60b67ab2 --metadata`
</CodeBlock>

  </TabItem>
</Tabs>

### Update group

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request PUT 'http://localhost:8000/v1beta1/organizations/4eb3c3b4-962b-4b45-b55b-4c07d3810ca8/groups/2105beab-5d04-4fc5-b0ec-8d6f60b67ab2'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
    "name": "Data Batching",
    "slug": "data-batching",
    "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8",
    "metadata": {
        "description": "group for users in data batching domain",
        "org-name": "raystack"
    }
}'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield group edit 457944c2-2a4c-4e6f-b1f7-3e1e109fe94c --file=group.yaml`
</CodeBlock>

  </TabItem>
</Tabs>

### Get all users in a group

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`curl --location --request GET 'http://localhost:8000/v1beta1/organizations/4eb3c3b4-962b-4b45-b55b-4c07d3810ca8/groups/86e2f95d-92c7-4c59-8fed-b7686cccbf4f/relations?subjectType=user&role=manager'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
</Tabs>
