# Using as a reverse proxy

![](../.gitbook/assets/overview.svg)

Another way to protect your endpoints using Shield is to use it as a reverse proxy by configuring your endpoints, which either forwards or forbids requests based on authorization checks. By using Shield as a reverse proxy you can keep your services free of any `IAM` related logic and move it all to configuration.

Say you want to protect the `PUT /api/books/relativity-the-special-general-theory` endpoint given in the [example](managing_policies.md). To set it up as a reverse proxy you need to do the following:

## Step 1: Create this YAML file for your API

Create an `api.yaml` file with the configuration mentioned below.

```text
---
  -
    method: "PUT"
    path: "/api/books/{urn}"
    proxy:
      uri: "http://www.library.com/books/{urn}"
    permissions:
      -
        action: "book.update"
        attributes:
          -
            urn:
              type: "params"
              key: "urn"
  -
    method: "GET"
    path: "/api/books/{urn}"
    proxy:
      uri: "http://www.library.com/books/{urn}"
    permissions:
      -
        action: "book.read"
        attributes:
          -
            urn:
              type: "params"
              key: "urn"
```

### Explanation

- **method**: The HTTP method of your endpoint
- **path**: The path of the endpoint
- **uri**: The URL of your service to forward the request. It should be added inside the `proxy` key.
- **permissions**: An array of permissions that should resolve to true for the user to be able to access your `uri`.
- **action**: is part of the permission object. Check more about the action here.
- **attributes** are part of the permission object. Attributes are a list of keys that you want to pick from the request object and apply your authorization on. In the above example, we are picking `urn` from `params` of this endpoint `/api/books/{urn}`. You can pick values from `query`, `params`, `headers`, and `payload`

## Step 2: Mount the above proxy file while deploying Shield

You can create a `proxies` folder and add the above `api.yaml` file. Check out our [deployment](deployment.md) guide to mount your `proxies` while deploying Shield

## Hooks

When using Shield as a reverse proxy you might also want to store your resources in your IAM policy while the resource is created or also you might want to update it. You can do this with the following configuration.

```text
---
  -
    method: "PUT"
    path: "/api/books/{urn}"
    proxy:
      uri: "http://www.library.com/books/{urn}"
    permissions:
      -
        action: "book.update"
        attributes:
          -
            urn:
              type: "params"
              key: "urn"
    hooks:
      -
        resources:
          -
            urn:
              type: "params"
              key: "urn"
        attributes:
          -
            group:
              type: "payload"
              key: "group"
          -
            category:
              type: "payload"
              key: "category"
```

Configuring hooks is similar to using the [resources](https://github.com/odpf/shield/tree/e4adf59ae35efc5bd3c615068932e1d780037f13/docs/guides/usage_check_access/README.md#resources-and-attributes) API but here you are able to create the resource and attributes mapping on the fly.
