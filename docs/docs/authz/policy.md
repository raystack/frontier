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
