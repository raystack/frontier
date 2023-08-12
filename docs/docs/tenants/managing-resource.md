import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Manage Resources

A resource is a logical entity that represents any user-defined entity in the system. A resource always belongs to 
a `project` and is identified by a unique identifier called `urn` or via it's `id`. For example, in a system that
manages databases, a resource can be a database instance. For a database instance, it's namespace can be `db/instance`
and when the actual db service creates a database instance, it can create a resource in the frontier to manage it authorization.
When working with custom resources, ensure that the namespace has enough permissions to `create`, `update` and `delete` resources.
Without these permissions, the frontier will not be able to manage the resource, specially when the resource is deleted.

A resource in Frontier looks like

<Tabs groupId="model">
<TabItem value="JSON" label="Sample JSON" default>

```json
{
  "id": "39aee58b-ea9a-474d-ad99-3f5e0e53d588",
  "name": "instance-1",
  "created_at": "2023-08-10T11:58:03.607320Z",
  "updated_at": "2023-08-10T11:58:03.607320Z",
  "urn": "frn:proj1:compute/instance:instance-1",
  "project_id": "4bcb528f-b397-47f6-8e1f-397d7ae88b32",
  "namespace": "compute/instance",
  "principal": "app/user:f4641672-cfdc-493f-95f0-c440515ad032",
  "metadata": null
}
```

</TabItem>
</Tabs>

## API Interface

### Create resources

To create a resource in the frontier:

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

### List all resources across projects

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request GET 'http://localhost:8000/v1beta1/projects/:project_id/resources'
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

### Delete resources

Ensure `delete` permission is created for the resource and provided to caller.

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl -curl --location --request DELETE 'http://localhost:8000/v1beta1/projects/1b89026b-6713-4327-9d7e-ed03345da288/resources/28105b9a-1717-47cf-a5d9-49249b6638df'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
</Tabs>