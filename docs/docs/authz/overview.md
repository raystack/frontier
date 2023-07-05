import Mermaid from '@theme/Mermaid';

# Overview

Authorization is the process of determining whether a user is allowed to perform an action. In the context of Shield, authorization is the process of determining whether a user is allowed to perform an action on a resource. This is done after the system has already confirmed that user has proven their identity (authentication).

Shield authorization is based on the Role Based Access Control (RBAC) model. In RBAC, access is granted to users based on their roles. A role is a collection of permissions that can be assigned to a user. Permissions determine what actions are allowed on a resource. When a role is assigned to a user, the user is granted all the permissions that the role contains.

In the Shield, several key components and concepts come into play for authorization of user request to access resources. Let's first look into how a normal user request flow with Shield will look like.

<Mermaid chart = {`sequenceDiagram
    participant User
    participant App
    participant Shield 
    participant SpiceDB
    %
    User->>App: Request Resource
    App->>Shield : Authenticate User
    Shield->>App: User Authenticated
    %
    Note over Shield,User: Assuming user is authenticated <br/> Now check if user is authorized to access the resource
    alt Case A: User Unauthorized
        App->>Shield : Check User is Authorized
        Shield ->>SpiceDB: Check API
        SpiceDB->>App: Invalid Permission
        App-->>User: Return Permission Denied
    else Case B: User Authorized
        App->>Shield : Check User is Authorized
        Shield->>SpiceDB: Check API
        SpiceDB->>App: Valid Permission
        App-->>User: Return Resource 
    end`}
/>

<br/><br/>

[SpiceDB](https://authzed.com/docs) is Authzed's open source [Google Zanzibar](https://research.google/pubs/pub48190/) inspired permission system which is capable of answering the question like

```
does <User> have <permission> on <resource>?
```

Permissions are defined as relationships between Users and resources. A User can be a user, a group, or any other entity that needs to be granted permissions. A resource can be a file, a database, or any other object that needs to be protected.

The SpiceDB permissions system works by first creating a permissions schema that defines the relationships between Users and resources. The relationships between **resources** and **Users** used for permissions checks are stored within SpiceDB's data store. The schema is then used to create a graph of permissions, where each node in the graph represents a User or resource, and each edge in the graph represents a permission.

When a permission check is performed, SpiceDB traverses the permissions graph to determine whether the User has the required permission to access the resource. 

