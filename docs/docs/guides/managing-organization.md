import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Managing Organization

An Organization in Shield is a top-level resource. Each Project, Group, User, and Audit logs (coming soon) belongs to an Organization. There can be multiple tenants in each Shield deployement and an Organization will usually represent one of your tenant.

Read More about an Organization in the [Concepts](../concepts/org.md) section.

<Tabs groupId="model">
  <TabItem value="Model" label="Model" default>

| Field        | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| ------------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --- |
| **id**       | uuid   | Unique Organization identifier                                                                                                                                                                                                                                                                                                                                                                                                                      |
| **name**     | string | The name of the organization. The name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores.<br/> _Example:"shield-org1-acme"_                                                                                                                                                                                                                                               |
| **title**    | string | The title can contain any UTF-8 character, used to provide a human-readable name for the organization. Can also be left empty. <br/>_Example: "Acme Inc"_                                                                                                                                                                                                                                                                                           |     |
| **metadata** | object | Metadata object for organizations that can hold key value pairs defined in Organization Metaschema. The metadata object can be used to store arbitrary information about the organization such as labels, descriptions etc. The default Organization Metaschema contains labels and descripton fields. Update the Organization Metaschema to add more fields.<br/>_Example:{"labels": {"key": "value"}, "description": "Organization description"}_ |

</TabItem>
<TabItem value="JSON" label="Sample JSON" default>

```json
{
  "organization": {
    "id": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8",
    "name": "open-dataops-foundation",
    "title": "Open DataOps Foundation (ODPF)",
    "metadata": {
      "description": "Some Organization details"
    },
    "createdAt": "2023-05-07T14:10:42.755848Z",
    "updatedAt": "2023-05-07T14:10:42.755848Z"
  }
}
```
</TabItem>
</Tabs>

**Note:** Organization metadata values are validated using MetaSchemas in Shield [Read More](./managing-metaschemas.md)

### Create an Organization

You can create organizations using either the Admin Portal, Shield Command Line Interface or via the HTTP APIs.

1. Using `shield organization create` CLI command
2. Calling to `POST /v1beta1/organizations` API

<Tabs groupId="api">
  <TabItem value="http" label="HTTP">
        <CodeBlock className="language-bash">
{`$ curl --location --request POST 'http://localhost:8000/v1beta1/organizations'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "odpf",
  "title": "ODPF",
  "metadata": {
      "description": "Open DataOps Foundation"
  }
}'`}
    </CodeBlock>
  </TabItem>
<TabItem value="cli" label="CLI" default>

```bash
$ shield organization create --file=<path to the organization.json file>
```

  </TabItem>
</Tabs>

3. To create an organization via the Admin Portal:

  i. Navigate to **Admin Portal > Organizations** from the sidebar

  ii. Select **+ New Organization** from top right corner

  iii. Enter basic information for your organization, and select **Add Organization**.

### Add or invite users to an organization

### List pending invitations queued for an org

### Delete pending invitations queued for an org

### View an organization principals

1. CLI command missing (Todo)
2. Calling to `GET /v1beta1/organizations/{id}/users` API

<Tabs groupId="api">
  <TabItem value="http" label="HTTP">
  <CodeBlock className="language-bash">
{`$ curl --location 'http://localhost:8000/v1beta1/organizations/adf997e8-59d1-4462-a4f2-ab02f60a86e7/users' 
--header 'Accept: application/json'
`}
</CodeBlock>
</TabItem>
 <TabItem value="cli" label="CLI">
Todo
 </TabItem>
</Tabs>

### Create custom roles and permissions for an org

<Tabs groupId="api">
  <TabItem value="http" label="HTTP">
  <CodeBlock className="language-bash">
  {` $curl
  `}</CodeBlock>
</TabItem>
 <TabItem value="cli" label="CLI">
Todo
 </TabItem>
 </Tabs>

### Assign roles to organization principals

<Tabs groupId="api">
  <TabItem value="http" label="HTTP">
  <CodeBlock className="language-bash">
  {` $curl
  `}</CodeBlock>
</TabItem>
 <TabItem value="cli" label="CLI">
Todo
 </TabItem>
 </Tabs>

### Enable or disable an org

<Tabs groupId="api">
  <TabItem value="http" label="HTTP">
  <CodeBlock className="language-bash">
  {` curl --location 'http://localhost:7400/v1beta1/organizations/adf997e8-59d1-4462-a4f2-ab02f60a86e7/enable' 
--header 'Content-Type: application/json' 
--header 'Accept: application/json' 
--data '{}'`}
</CodeBlock>
</TabItem>
 <TabItem value="cli" label="CLI">
Todo
 </TabItem>
 </Tabs>

### Remove a user from an org
<Tabs groupId="api">
  <TabItem value="http" label="HTTP">
  <CodeBlock className="language-bash">
  {` $curl
  `}</CodeBlock>
</TabItem>
 <TabItem value="cli" label="CLI">
Todo
 </TabItem>
 </Tabs>


### Remove user from an organization 