# Managing policies

Shield lets you control who \(users\) has what access \(roles\) to which resources by setting IAM policies. IAM policies grant specific role\(s\) to a user giving the user certain permissions.

## Components

Here is a brief description of all the components of a policy. These definitions will help you understand the below example better.

- **Users**: Users of your application.
- **Groups**: Groups are collections of users. Groups can also be used as attributes\(see below\) to map with resources.
- **Actions**: Actions are a representation of your endpoints. For example, action name for `POST /books` could be `book.create`, similarly for `GET /books/:id` could be `book.read`, etc.
- **Roles**: Roles are collections of actions. Example: someone with a `Viewer` role can only perform `book.read`, whereas a `Manager` can perform both `book.read` and `book.create`.
- **Resources**: Resources are the assets you would like to protect via authorization.
- **Attributes**: Resources get mapped with attributes that can be used while creating policies. Policies can also be defined on attributes rather than individual resources to create better access management.

Please check the [Glossary](../reference/glossary.md) section to understand the definitions in detail.

## Example

Let us take an example of a Library Management Application. It has users and groups that manage books and e-books for a library based on their assigned roles.

### Application endpoints to protect

The following are the endpoints of the Library Management App that we would like to protect.

```text
POST /books => Corresponding action = book.create
GET /books/{urn} => Corresponding action = book.read
PUT /books/{urn} => Corresponding action = book.update
DELETE /books/{urn} => Corresponding action = book.delete

POST /e-books => Corresponding action = e-book.create
GET /e-books/{urn} => Corresponding action = e-book.read
PUT /e-books/{urn} => Corresponding action = e-book.update
DELETE /e-books/{urn} => Corresponding action = e-book.delete
```

### Roles and actions

As defined earlier, actions are representations of your endpoints, and roles are just groups of actions. For our library app, we have the following roles:

#### Group Admin

- Description: Users with this role can add users to a group and manage all resources of that group
- Actions: `["*"]`

#### Book Manager

- Description: Users with this role can perform all actions of the `books` endpoints.
- Actions: `["book.read", "book.create", "book.update", "book.delete"]`

#### E-Book Manager

- Description: Users with this role can perform all actions of the `e-books` endpoints.
- Actions: `["e-book.read", "e-book.create", "e-book.update", "e-book.delete"]`

#### Viewer

- Description: Users with this role can only read `books` and `e-books` endpoint
- Actions: `["book.read", "e-book.read"]`

To create a Role, and `Role<->Action` Mapping call the following API:

```text
POST /api/roles
Content-Type: application/json
Accept: application/json

Request Body:

[
    {
        "displayname": "Book Manager",
        "actions": ["book.read", "book.create", "book.update", "book.delete"]
    },
    {
        "displayname": "Group Admin",
        "actions": ["*"]
    }
]
```

### Users and Groups

#### List of users of the Library App:

- Name: Einstein
- Name: Newton
- Name: Feynman
- Name: Darwin
- Name: Ramanujan
- Name: Leibniz

#### List of Groups in the Library App:

- Name: Scientists
- Name: Mathematicians

You can create users and groups using the following APIs:

```text
POST /api/users
Content-Type: application/json
Accept: application/json

Request Body:
{
    "displayname": "Einstein"
}

Response:

{
    "id": "10253480-2ac8-4073-9084-5f059dbf6fae",
    "username": "einstein",
    "displayname": "Einstein"
}


POST /api/groups
Content-Type: application/json
Accept: application/json

{
    "displayname": "Scientists"
}

Response:

{
    "id": "80553880-23c8-4073-9094-7f059avf6ftp",
    "groupname": "scientists",
    "displayname": "Scientists"
}
```

### Resources and Attributes

We have the following books and e-books with certain attributes in our library.

- Note: the value of the group attribute below is the group id of the respective group.
- Group Id for Mathematicians group = 90663880-43c8-5073-1094-7f059avf6ftz
- Group Id for Scientists group = 80553880-23c8-4073-9094-7f059avf6ftp

```text
Resource => {"urn": "relativity-the-special-general-theory"}
Attributes => {"group": "80553880-23c8-4073-9094-7f059avf6ftp", "category": "physics"}
Type: Book

Resource => {"urn": "on-the-origin-of-species"}
Attributes => {"group": "80553880-23c8-4073-9094-7f059avf6ftp", "category": "biology"}
Type: E-Book


Resource => {"urn": "calculus-made-easy"}
Attributes => {"group": "90663880-43c8-5073-1094-7f059avf6ftz", "category": "calculus"}
Type: E-Book

Resource => {"urn": "trignometry-for-dummies"}
Attributes => {"group": "90663880-43c8-5073-1094-7f059avf6ftz, "category": "trignometry"}
Type: E-Book
```

You can call the following API to map Resources with Attributes:

```text
POST /api/resources
Content-Type: application/json
Accept: application/json

Request Body

[
    {
        "resource": { "urn": "relativity-the-special-general-theory" },
        "attributes": {"group": "80553880-23c8-4073-9094-7f059avf6ftp", "category": "physics"}
    },

    {
        "resource": { "urn": "calculus-made-easy" },
        "attributes": {"group": "90663880-43c8-5073-1094-7f059avf6ftz", "category": "calculus"}
    }
]
```

### Creating Policies

With users, groups, roles, actions, resources, and attributes setup, we can create policies to restrict/give access to our users.

Let's say we want the following roles for users in the Scientists Group:

- Newton =&gt; `Group Admin` role for the Scientists Group
- Einstein =&gt; `Book Manager` role, but only for books tagged with `{"category": "physics"}`
- Feynman =&gt; `E-book Manager` role, but only for e-books tagged with `{"category": "physics"}`
- Charles Darwin =&gt; `Book Manager` and `E-book Manager` role but only for e-books and books tagged with `{"category": "biology"}`

We can call the following API to setup the above Policies:

#### Einstein =&gt; `Book Manager` role within the Scientists Group, but only for books tagged with `{"category": "physics"}`

```text
POST /api/groups/{groupId}/users/{userId}
groupId (Scientists) = 80553880-23c8-4073-9094-7f059avf6ftp
userId (Einstein) = 10253480-2ac8-4073-9084-5f059dbf6fae
roleId (Book Manager) = 13f87b13-c4c8-4be1-b6d6-a9359bcd05fc

Request Body

{
    operation: "create",
    "subject": {
        "user": "10253480-2ac8-4073-9084-5f059dbf6fae"
    },
    "resource": {
        "group": "80553880-23c8-4073-9094-7f059avf6ftp",
        "category": "physics"
    },
    "action": {
        "role": "13f87b13-c4c8-4be1-b6d6-a9359bcd05fc"
    }
}
```

#### Newton =&gt; `Group Admin` role for the Scientists Group

```text
POST /api/groups/{groupId}/users/{userId}
groupId (Scientists) = 80553880-23c8-4073-9094-7f059avf6ftp
userId (Newton) = 3d3b2dad-e1c9-4c05-a9f2-dfb474c4ff35
roleId (Group Admin) = 096495d1-3e5b-40a0-9e9a-e69d4c390266

Request Body

{
    operation: "create",
    "subject": {
        "user": "3d3b2dad-e1c9-4c05-a9f2-dfb474c4ff35"
    },
    "resource": {
        "group": "80553880-23c8-4073-9094-7f059avf6ftp",
    },
    "action": {
        "role": "096495d1-3e5b-40a0-9e9a-e69d4c390266"
    }
}
```

### Protecting Endpoints

Once the policies have been added, you can protect your endpoints in two ways:

- [Using Shield as an external authorization service](./using_auth_server.md)
- [Using Shield as a reverse proxy](./usage_reverse_proxy.md)

### Access Management for Shield's APIs

You need correct permissions to access every API in Shield, even the native ones such as creating roles, groups, etc. You can find all the native APIs of Shield along with their permissions [here](https://odpf.gitbook.io/shield/reference/api)

#### Make yourself a Super Admin

One way to get around this issue is to give yourself all access using the following DB entry in `casbin_rule` table:

| v0                       | v1           | v2               |
| ------------------------ | ------------ | ---------------- |
| {"user": "your_user_id"} | {"\*": "\*"} | {"action": "\*"} |

#### Create specific Roles

Another way to solve this issue is to create roles for each of the endpoints/actions and assign them to appropriate users. However, you may still need a super admin user to create these roles.
