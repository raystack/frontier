import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Managing Organization

A organization in Shield looks like

```json
{
    "organizations": [
        {
            "id": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8",
            "name": "ODPF",
            "slug": "odpf",
            "metadata": {
                "description": "Open DataOps Foundation"
            },
            "createdAt": "2022-12-07T14:10:42.755848Z",
            "updatedAt": "2022-12-07T14:10:42.755848Z"
        }
    ]
}
```

## API Interface

### Create Organizations

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request POST 'http://localhost:8000/v1beta1/organizations'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "ODPF",
  "slug": "odpf",
  "metadata": {
      "description": "Open DataOps Foundation"
  }
}'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield organization create --file org.yaml --header key:value`
</CodeBlock>

  </TabItem>
</Tabs>

### List All Organizations

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request GET 'http://localhost:8000/v1beta1/admin/organizations'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield organization list`
</CodeBlock>

  </TabItem>
</Tabs>

### Get Organization

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request GET 'http://localhost:8000/v1beta1/organizations/4eb3c3b4-962b-4b45-b55b-4c07d3810ca8'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield organization view 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 --metadata`
</CodeBlock>

  </TabItem>
</Tabs>

### Update Organization

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request PUT 'http://localhost:8000/v1beta1/organizations/4eb3c3b4-962b-4b45-b55b-4c07d3810ca8'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "ODPF",
  "slug": "odpf",
  "metadata": {
      "description": "Open DataOps Foundation",
      "url": "github.com/odpf"
  }
} '`}
    </CodeBlock>
  </TabItem>
  <TabItem value="CLI" label="CLI" default>
<CodeBlock>

`$ shield organization edit 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 --file=org.yaml`
</CodeBlock>

  </TabItem>
</Tabs>