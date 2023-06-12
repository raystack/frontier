import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Managing Users

- Create a user in Shield
- List a user's organizations
- List a user's groups
- Accept a pending invitation from an org/group
- Decline a pending invitation from an org/group
- Enable or disable a user
- Delete a user from Shield

A project in Shield looks like

<Tabs groupId="model">
  <TabItem value="Model" label="Model" default>

| Field        | Type   | Description                                                                                                                                                                                                                                                                                                                                                                                                                          |
| ------------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **id**       | uuid   | Unique user identifier                                                                                                                                                                                                                                                                                                                                                                                                               |
| **name**     | string | The name of the user. The name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores and must start with a letter. If not provided, Shield automatically generates a name from the user email. <br/> _Example:"john_doe_raystack_io"_                                                                                                                          |
| **email**    | string | The email of the user. The email must be unique within the entire Shield instance.<br/> _Example:"john.doe@raystack.io"_                                                                                                                                                                                                                                                                                                             |
| **metadata** | object | Metadata object for users that can hold key value pairs pre-defined in User Metaschema. The metadata object can be used to store arbitrary information about the user such as label, description etc. By default the user metaschema contains labels and descriptions for the user. Update the same to add more fields to the user metadata object. <br/> _Example:{"label": {"key1": "value1"}, "description": "User Description"}_ |
| **title**    | string | The title can contain any UTF-8 character, used to provide a human-readable name for the user. Can also be left empty. <br/> _Example:"John Doe"_                                                                                                                                                                                                                                                                                    |

</TabItem>
<TabItem value="JSON" label="Sample JSON" default>

```json
{
  "user": {
    "id": "598688c6-8c6d-487f-b324-ef3f4af120bb",
    "name": "john_doe_raystack_io",
    "title": "John Doe",
    "email": "john.doe@raystack.io",
    "metadata": {
      "description": "\"Shield human user\""
    },
    "createdAt": "2022-12-09T10:45:19.134019Z",
    "updatedAt": "2022-12-09T10:45:19.134019Z"
  }
}
```

</TabItem>
</Tabs>

**Note:** user metadata values are validated using MetaSchemas in Shield [Read More](../reference/metaschemas.md)

## API Interface

### Create users

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request POST 'http://localhost:8000/v1beta1/users'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--header 'X-Shield-Email: admin@raystack.io'
--data-raw '{
  "name": "Jonny Doe",
  "email": "jonny.doe@raystack.io",
  "metadata": {
      "role": "user-3"
  }
}'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield user create --file=user.yaml`
</CodeBlock>

  </TabItem>
</Tabs>

### List all users

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`curl --location --request GET 'http://localhost:8000/v1beta1/admin/users'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield user list`
</CodeBlock>

  </TabItem>
</Tabs>

### Get Users

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request GET 'http://localhost:8000/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608'
--header 'Accept: application/json''`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield user view e9fba4af-ab23-4631-abba-597b1c8e6608 --metadata`
</CodeBlock>

  </TabItem>
</Tabs>

### Update Users

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request PUT 'http://localhost:8000/v1beta1/users/e9fba4af-ab23-4631-abba-597b1c8e6608'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "Jonny Doe",
  "email": "john.doe001@raystack.io",
  "metadata": {
      "role" :   "user-3"
  }
}'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield user edit e9fba4af-ab23-4631-abba-597b1c8e6608 --file=user.yaml`
</CodeBlock>

  </TabItem>
</Tabs>
