import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';
import CodeBlock from '@theme/CodeBlock';

# Checking Pemrissions

There are two ways to check a user permission on a resource in shield,
## API Interface

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>
        <CodeBlock className="language-bash">
    {`$ curl --location --request POST 'http://localhost:8000/admin/v1beta1/check'
--header 'Content-Type: application/json'
--header 'Accept: application/json'
--header 'X-Shield-Email: doe.john@odpf.io'
--data-raw '{
  "objectId": "test-resource-beta1",
  "objectNamespace": "entropy/firehose",
  "permission": "owner"
}'`}
    </CodeBlock>
  </TabItem>
</Tabs>

## Proxy Middleware

Users can add middleware in the rules set to check permission. Middlewares will be called before the proxy call and will not call the services if authorization fails.

The shield will read the action from the config, resource id from the path params, and UserId of the current user.

```yaml
- name: test-res
  path: /test-res
  target: "http://127.0.0.1:3000/"
  methods: ["PUT"]
  frontends:
    - name: test_api
      path: "/test-res/{resource_id}"
      method: "PUT"
      middlewares:
        - name: authz
          config:
            actions:
              - test-res_all_actions
              - test-res_cancel
            attributes:
              project:
                key: X-Shield-Project
                type: header
                source: request
```