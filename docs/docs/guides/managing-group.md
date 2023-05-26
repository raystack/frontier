import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Managing Group

- Add or invite users to a group (coming up!!)
- View a group members
- View a user's groups
- Assign roles to group members
- Remove a user from a group
- Enable or disable a group
- Add a group to another group (coming up!!)

A group in Shield looks like

```json
{
  "group": {
    "id": "2105beab-5d04-4fc5-b0ec-8d6f60b67ab2",
    "name": "Data Batching",
    "slug": "data-batching",
    "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8",
    "metadata": {
      "description": "group for users in data batching domain",
      "org-name": "odpf"
    },
    "createdAt": "2022-12-14T10:22:14.394120Z",
    "updatedAt": "2022-12-14T10:25:34.890645Z"
  }
}
```

**Note:** group metadata values are validated using MetaSchemas in Shield [Read More](./managing-metaschemas.md)

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
        "org-name": "odpf"
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