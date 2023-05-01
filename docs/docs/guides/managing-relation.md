import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Managing Relations

A relation in Shield looks like

```json
{
    "relations": [
        {
            "id": "08effbce-42cb-4b7e-a808-ad17cd3445df",
            "objectId": "a9f784cf-0f29-486f-92d0-51300295f7e8",
            "objectNamespace": "entropy/firehose",
            "subject": "user:598688c6-8c6d-487f-b324-ef3f4af120bb",
            "roleName": "entropy/firehose:owner",
            "createdAt": null,
            "updatedAt": null
        }
    ]
}
```

## API Interface

### Create Relations

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request POST 'http://localhost:8000/v1beta1/relations'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "objectId": "a9f784cf-0f29-486f-92d0-51300295f7e8",
  "objectNamespace": "entropy/firehose",
  "subject": "user:doe.john@odpf.io",
  "roleName": "owner"
}'`}
    </CodeBlock>
  </TabItem>
</Tabs>

### List Relations

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request GET 'http://localhost:8000/v1beta1/admin/relations'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
</Tabs>

### Get Relations

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request GET 'http://localhost:8000/v1beta1/relations/f959a605-8755-4ee4-b898-a1e26f596c4d'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
</Tabs>

### Delete relation

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request DELETE 'http://localhost:8000/v1beta1/
    object/a9f784cf-0f29-486f-92d0-51300295f7e8/
    subject/448d52d4-48cb-495e-8ec5-8afc55c624ca/
    role/owner'
--header 'Accept: application/json'`}
    </CodeBlock>
  </TabItem>
</Tabs>