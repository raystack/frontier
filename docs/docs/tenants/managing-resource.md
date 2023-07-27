import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Manage Resources

A resource in Frontier looks like

<Tabs groupId="model">
  <TabItem value="Model" label="Model" default>

| Field | Type | Description |
| ----- | ---- | ----------- |

</TabItem>
<TabItem value="JSON" label="Sample JSON" default>

```json
{
  "resource": {
    "id": "5723e961-7259-48b3-b721-292868d652d7",
    "name": "test-random-name",
    "project": {
      "id": "1b89026b-6713-4327-9d7e-ed03345da288",
      "name": "",
      "slug": "",
      "orgId": "",
      "metadata": null,
      "createdAt": null,
      "updatedAt": null
    },
    "organization": {
      "id": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8",
      "name": "",
      "slug": "",
      "metadata": null,
      "createdAt": null,
      "updatedAt": null
    },
    "namespace": {
      "id": "entropy/firehose",
      "name": "",
      "createdAt": null,
      "updatedAt": null
    },
    "createdAt": "2022-12-13T11:59:23.964065Z",
    "updatedAt": "2022-12-13T11:59:23.964065Z",
    "user": {
      "id": "2fd7f306-61db-4198-9623-6f5f1809df11",
      "name": "",
      "slug": "",
      "email": "",
      "metadata": null,
      "createdAt": null,
      "updatedAt": null
    },
    "urn": "r/entropy/firehose/test-random-name"
  }
}
```

</TabItem>
</Tabs>

## API Interface

### Create resources

There are two ways to create a resource in the frontier,

#### API Interface

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request POST 'http://localhost:8000/v1beta1/projects/1b89026b-6713-4327-9d7e-ed03345da288/resources'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--header 'X-Frontier-Email: admin@raystack.org'
--data-raw '{
  "name": "test-resource-beta",
  "projectId": "1b89026b-6713-4327-9d7e-ed03345da288",
  "namespaceId": "entropy/firehose",
  "relations": [
    {
      "subject": "user:john.doe@raystack.org",
      "roleName": "owner"
    }
  ]
}'`}
    </CodeBlock>
  </TabItem>
</Tabs>

#### Proxy Hook

Users can add hooks to rules set to create a resource. The hook will be called after the proxy request is completed.
Hooks can read query, header, params, payload, and response to get the values for Resource.

```yaml
- name: test-res
  path: /test-res
  target: "http://127.0.0.1:3000/"
  methods: ["POST"]
  frontends:
    - name: create test-res
      path: "/test-res"
      method: "POST"
      hooks:
        - name: authz
          config:
            attributes:
              project:
                key: project
                type: json_payload
              organization:
                key: organization
                type: json_payload
              team:
                key: team
                type: json_payload
              resource:
                key: urns.#.id
                type: json_payload
```

### List all resources across projects

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request GET 'http://localhost:8000/v1beta1/admin/resources'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
</Tabs>

### Get resources

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl -curl --location --request GET 'http://localhost:8000/v1beta1/projects/1b89026b-6713-4327-9d7e-ed03345da288/resources/28105b9a-1717-47cf-a5d9-49249b6638df'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
</Tabs>

### Update resource

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request PUT 'http://localhost:8000/v1beta1/projects/1b89026b-6713-4327-9d7e-ed03345da288/resources/a9f784cf-0f29-486f-92d0-51300295f7e8'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "name": "test-resource-beta1",
  "projectId": "1b89026b-6713-4327-9d7e-ed03345da288",
  "namespaceId": "entropy/firehose"
}'`}
    </CodeBlock>
  </TabItem>
</Tabs>
