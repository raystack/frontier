# Managing Policies

Shield lets you control who \(users\) has what access \(roles\) to which resources by setting IAM policies. IAM policies grant specific role\(s\) to a user giving the user certain permissions.

## Components of a policy

Here is a brief description of all the components of a policy. These definitions will help you understand the below example better.

* **Users**: Users of your application.
* **Groups**: Groups are just collections of users. Groups can also be used as attributes\(see below\) to map with resources.
* **Actions**: Actions are representation of your endpoints. For example: action name for `POST /books` would be `book.create`, similarly for `GET /books/:id` would be `book.read`, etc.
* **Roles**: Roles are nothing but collections of actions. Example: someone with a `Viewer` role can only perform `book.read`, whereas a `Manager` can perform both `book.read` and `book.create`.
* **Resources**: Resources are the assets you would like to protect via authorization.
* **Attributes**: Resources get mapped with attributes that can be turn used while creating policies. Policies can get defined on attributes rather than individual resources to create better access management.

Please check the [Glossary](../reference/glossary.md) section to understand the definitions in detail.

## Example

Let us take an example of a Library Management Application. It has users and groups that manage books and e-books for a library based on the roles assigned to them.

### Application endpoints to protect

The following are the endpoints of the Library Management App that we would like to protect

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

* Description: Users with this role can add users to a group and manage all resources all resources of that group
* Actions: `["*"]`

#### Book Manager

* Description: Users with this role can perform all actions of the `books` endpoints.
* Actions: `["book.read", "book.create", "book.update", "book.delete"]`

#### E-Book Manager

* Description: Users with this role can perform all actions of the `e-books` endpoints.
* Actions: `["e-book.read", "e-book.create", "e-book.update", "e-book.delete"]`

#### Viewer

* Description: Users with this role can only read `books` and `e-books` endpoint
* Actions: `["book.read", "e-book.read"]`

To create a Role, and Role&lt;-&gt;Action Mapping call the following API:

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

* Name: Einstein
* Name: Newton
* Name: Feynman
* Name: Darwin
* Name: Ramanujan
* Name: Leibniz

#### List of Groups in the Library App:

* Name: Scientists
* Name: Mathematicians

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

We have the following books and e-books with certain attributes in our library

* Note: the value of the group attribute below is the group id of the respective group.
* Group Id for Mathematicians group = 90663880-43c8-5073-1094-7f059avf6ftz
* Group Id for Scientists group = 80553880-23c8-4073-9094-7f059avf6ftp

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

* Newton =&gt; `Group Admin` role for the Scientists Group
* Einstein =&gt; `Book Manager` role, but only for books tagged with `{"category": "physics"}`
* Feynman =&gt; `E-book Manager` role, but only for e-books tagged with `{"category": "physics"}`
* Charles Darwin =&gt; `Book Manager` and `E-book Manager` role but only for e-books and books tagged with `{"category": "biology"}`

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

### Checking Access

Now that we have setup our policies, we can check whether a user has the necessary permissions to perform actions on a resource.

#### Scenario 1

Let's say Einstein tries edit to edit the book `relativity-the-special-general-theory` by calling the following endpoint:

```text
PUT /api/books/relativity-the-special-general-theory
```

Then within your service you can call the `check-access` endpoint of Shield to check whether Einstein has access.

```text
POST /api/check-access


Request Body:

[
    {
        "suject": {
            "user": "10253480-2ac8-4073-9084-5f059dbf6fae"
        },
        "resource": {
            "urn": "relativity-the-special-general-theory"
        },
        "action": {
            "action: "book.update"
        }
    }
]

Response:

[
    {
        "suject": {
            "user": "10253480-2ac8-4073-9084-5f059dbf6fae"
        },
        "resource": {
            "urn": "relativity-the-special-general-theory"
        },
        "action": {
            "action: "book.update"
        },
        "hasAccess": true
    }
]
```

The above response will return `hasAccess:true` since `Einstein` is permitted to `book.update` for the Book `relativity-the-special-general-theory` as he was assigned `Book Manager` role for `{"group": "80553880-23c8-4073-9094-7f059avf6ftp", "category": "physics"}`

Based on this APIs response, you can decide within your application to either forbid or allow the user.

