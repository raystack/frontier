# Checking permissions in SpiceDB

In this part of the tour, we'll learn how can we use Frontier's permission checking system. Through this we are going to verify all the relations we have created till now.

We can either use the check API or `zed` tool. In this tour we will be using the `zed` tool.

## 1 Check the owner of the organization

```sh
zed permission check organization:4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 owner user:2fd7f306-61db-4198-9623-6f5f1809df11
```

```sh
true
```

## 2 Check the organization of the project

```sh
frontier % zed permission check project:1b89026b-6713-4327-9d7e-ed03345da288 organization organization:4eb3c3b4-962b-4b45-b55b-4c07d3810ca8
```

```sh
true
```

## 3 Check the organization of the group

```sh
zed permission check group:86e2f95d-92c7-4c59-8fed-b7686cccbf4f organization organization:4eb3c3b4-962b-4b45-b55b-4c07d3810ca8
```

```sh
true
```

## 4 Check the owner group of the resource

```sh
zed permission check entropy/firehose:28105b9a-1717-47cf-a5d9-49249b6638df owner group:86e2f95d-92c7-4c59-8fed-b7686cccbf4f
```

```sh
false
```

Note, this returns a false, even though we created a relation, where this group was an owner of this resource.
Let's revisit the authz schema for `entropy/firehose` where we have

```json
definition entropy/firehose {
    ...
    relation owner: user | group#membership
    ...
}
```

We gave the owner right not to the group, but the entities holding `membership` permisssion in the `group` schema, namely `member + manager` which are of type `user`.

```json
definition group {
    ...
	relation member: user
	relation manager: user
	permission membership = member + manager
    ...
}
```

So, all the members of the group have the permission on this resource, which we have validated below.

```sh
zed permission check entropy/firehose:28105b9a-1717-47cf-a5d9-49249b6638df owner user:598688c6-8c6d-487f-b324-ef3f4af120bb
```

```sh
true
```
