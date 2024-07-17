# Policy

Policies define the permissions and access control rules that govern user access to resources within the Frontier platform. By configuring policies, you can control and manage the level of access granted to users, groups, or service accounts.

A policy in Frontier consists of a set of permissions associated with a role. The role determines the actions that a user can perform on specific resources.

## Policy Usage

Frontier internally creates a SpiceDB relation between the Subject(Principal) and Object(Resource, Org, Project, Group) whenever a policy is created in Frontier. And when a user tries to access the resource, the Frontier looks up SpiceDB to check if a given relation has a specified permission. Thus, whenever a request is made for accessing a resource in Frontier, SpiceDB evaluates the defined policies and permissions to determine if the request is authorized. It checks the relationships, roles, and permissions associated with the requested resource and the user making the request. If the conditions specified in the policies are satisfied, access is granted; otherwise, access is denied.

### Check API

To check if the current logged in user has the permission to access a particular resource, Frontier provides the [Check](../apis/frontier-service-check-resource-permission.api.mdx) API. The check API in Frontier works with policy for authorization by using the following steps:

1. The user sends a request to the check API with the following information: <br/>i. The resource they want to check <br/>ii. The action they want to perform on the resource

2. The check API queries the SpiceDB engine to determine if the user is authorized to perform the requested action on the resource.

3. The check API returns a response to the user indicating whether or not they are authorized to perform the requested action on the resource.

**Getting current user's credentials:** The principal crendentials required to authorize the user is obtained from the current logged in user's **session** (if available). Otherwise, the **client ID and secret** or **access token** from the request headers can also be used to validate the user request to access the resource.

#### Sample Query

<Tabs groupId="api">
  <TabItem value="HTTP" label="HTTP" default>

```bash
curl -L -X POST 'http://127.0.0.1:7400/v1beta1/check' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1zZWNyZXQ=' \
--data-raw '{
  "permission": "get",
  "resource": "compute/instance:92f69c3a-334b-4f25-90b8-4d4f3be6b825"
}'
```

</TabItem>
</Tabs>

:::info
The resource field should contain the resource namespace and resource id/name in this format **namespace:uuid** or **namespace:name**.<br/><br/>Instead of passing complete namespaces for organization and projects, one can also use the aliases of **org**, **organization** or **project** instead of **app/organization** and **app/project**.

Similarly, instead of passing the unique UUID for the organization or project, one can pass the unique organization or project **name** used while creating these resources.

In case the object is not an organization or project, it can be a project resource and one can use format of **namespace:uuid** or **namespace:urn** in the resource field above.
:::

## Internals of Policy and Permission

Frontier uses the [SpiceDB](https://authzed.com/docs) permission system to manage and enforce access control policies. 
SpiceDB is an open-source permission system inspired by Google Zanzibar that is capable of answering questions like: 

```
does <User> have <permission> on <resource>?
```

Permissions are defined as relationships between users and resources. A user can be a user, a group, or any other entity 
that needs to be granted permissions. A resource can be a file, a database, or any other object that needs to be protected.
There are well known pre-defined namespaces reserved for frontier resources like `app/user`, `app/organization`, 
`app/project`, `app/role`, `app/rolebinding` etc prefixed with `app`.
Even though SpiceDB is used internally, the user does not need to interact with SpiceDB directly. Frontier provides a 
Hierarchical Role Based Access Control built on top of relationships and permissions defined in SpiceDB. Authzed has a blog
that talks about how RBAC can be modeled on top of SpiceDB. You can read more about it [here](https://authzed.com/blog/google-cloud-iam-modeling).

It works by defining Role and RoleBinding objects. A Role is a collection of relations that can be bound to a object namespace.
A RoleBinding binds a Role to a principal namespace(user/group/service account). The permissions are defined as relationships
between the principal and the object.

We maintain spiceDB schema in [base_schema.zed](https://github.com/raystack/frontier/blob/4434204940f3a599b7e0ef399679f381db46c09a/internal/bootstrap/schema/base_schema.zed)
file in Frontier. The schema is then used to build the base set of permissions on predefined namespaces. Some permissions are
hardcoded in the schema, but Frontier allows custom permission set on Project resources. Custom permissions are dynamically 
generated at frontier boot-up by merging them in the SpiceDB schema. To create custom permissions on a project resource
a [rule set](https://github.com/raystack/frontier/blob/ccf00da8a636f363546b1dbe6c7933323c369dca/internal/bootstrap/testdata/compute_service.yml) 
can be configured in the [config.yaml](https://github.com/raystack/frontier/blob/7fca56ef567791bf4ea1280aa3f9b4ae997b1bf5/config/sample.config.yaml#L29)
file. Same is also exposed in the [Frontier API](../apis/admin-service-create-permission.api.mdx) to create 
custom permissions on a project resource. 

Frontier models roles as collection of permissions which can be created/updated/deleted dynamically at run time.

To give principal access to a resource, a RoleBinding object is created that binds the principal to a role. The role contains
the permissions that the principal is granted on the resource. When a permission check is performed, SpiceDB traverses the permissions
graph to determine whether the principal has relations that grant the required permission to access the resource.

For example: To give a user access to a project, frontier create a policy that does following:
- A `app/role` object is created that creates a relation between a permission and `app/user`
- A `app/rolebinding` object is created that creates a relation between `role` and the created `app/role` object
- Another relation on same `app/rolebinding` object is created between `app/user` and `bearer` of role

When the user tries to access the project, the SpiceDB engine checks if the user has the permission by traversing the 
graph, checking if the role binding contains requested bearer and role->permission relation.
