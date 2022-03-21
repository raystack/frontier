# Check Permission

There are two ways to check permission in the shield,

1. REST/gRPC API
2. Proxy Middleware

## REST/gRPC API

```sh
POST v1beta1/check/{resourceId} HTTP/1.1
Host: localhost:8000
Content-Type: application/json

{
  "actionId": "read",
  "namespaceId": "test-namespace"
}
```

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
