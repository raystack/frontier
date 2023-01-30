import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Managing Project

A project in Shield looks like

```json
{
    "projects": [
        {
            "id": "1b89026b-6713-4327-9d7e-ed03345da288",
            "name": "Project Alpha",
            "slug": "project-alpha",
            "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8",
            "metadata": {
                "description": "Project Alpha"
            },
            "createdAt": "2022-12-07T14:31:46.436081Z",
            "updatedAt": "2022-12-07T14:31:46.436081Z"
        }
    ]
}
```

## API Interface

### Create projects

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request POST 'http://localhost:8000/admin/v1beta1/projects'
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
    {`$ curl --location --request GET 'http://localhost:8000/admin/v1beta1/projects'
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
    {`$ curl --location --request GET 'http://localhost:8000/admin/v1beta1/projects/457944c2-2a4c-4e6f-b1f7-3e1e109fe94c'
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
    {`$ curl --location --request PUT 'http://localhost:8000/admin/v1beta1/projects/457944c2-2a4c-4e6f-b1f7-3e1e109fe94c'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "Project Beta",
  "slug": "project-beta",
  "metadata": {
      "description": "Project Beta by ODPF"
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