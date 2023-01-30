import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Adding Metadata Keys

A metadata-key in Shield looks like

```json
{
    "metadatakey": {
        "key": "manager",
        "description": "manager of this user"
    }
}
```

## API Interface

### Create metadata keys

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request POST 'http://localhost:8000/admin/v1beta1/metadatakey'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--data-raw '{
  "key": "manager",
  "description": "manager of this user"
}'`}
    </CodeBlock>
  </TabItem>
</Tabs>