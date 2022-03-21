# Manage Resources

There are two ways to create a resource in the shield,

1. REST/gRPC API
2. Proxy Hooks

## REST/gRPC API

```sh
POST /admin/v1beta1/resources HTTP/1.1
Host: localhost:8000
Content-Type: application/json

{
    "name": "proident Lorem minim occaecat aute",
    "groupId": "ea nostrud consectetur laboris",
    "projectId": "sed occaecat dolor",
    "organizationId": "labore quis consectetur cillu",
    "namespaceId": "fugiat exercitation ut cupid",
    "userId": "anim reprehenderit sed ea sint"
}
```

## Proxy Hook

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
