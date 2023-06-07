# Policy

Policies define the permissions and access control rules that govern user access to resources within the Shield platform. By configuring policies, you can control and manage the level of access granted to users, groups, or service accounts.

A policy in Shield consists of a set of permissions associated with a role. The role determines the actions that a user can perform on specific resources. 

## Policy Usage

Shield internally creates a SpiceDB relation between the Subject(Principal) and Object(Resource, Org, Project, Group) whenever a policy is created in Shield. And when a user tries to access the resource, the Shield looks up SpiceDB to check if a given relation has a specified permission. Thus, whenever a request is made for accessing a resource in Shield, SpiceDB evaluates the defined policies and permissions to determine if the request is authorized. It checks the relationships, roles, and permissions associated with the requested resource and the user making the request. If the conditions specified in the policies are satisfied, access is granted; otherwise, access is denied. 

The check API in Shield works with policy for authorization by using the following steps:

1. The user sends a request to the check API with the following information: <br/>i. The resource they want to check <br/>ii. The action they want to perform on the resource

2. The check API queries the SpiceDB engine to determine if the user is authorized to perform the requested action on the resource.

3. The check API returns a response to the user indicating whether or not they are authorized to perform the requested action on the resource.