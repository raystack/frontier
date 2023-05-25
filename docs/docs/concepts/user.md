# Principal

Principal in Shield are entities that can be authenticated and authorized to access resources and perform actions. A principal will be used in controlling access to resources, as they can be associated with policies in Shield that define what actions the principal can perform and which resources they can access.

### Types of Principals in Shield

- **User**: A User Principal in Shield refers to an individual user who is granted certain privileges and access rights within the system. This type of principal is associated with a specific user account and is used to authenticate and authorize actions performed by the user. User principals are commonly used to manage and control access to resources, enforce security policies, and track user activities within the Shield system.

- **Machine User** (coming up): A Machine User Principal is a type of principal in Shield that represents a non-human entity, such as a software application or a service, that interacts with the system. This principal is specifically designed to authenticate and authorize actions performed by automated processes or scripts, allowing machines to securely access and manipulate resources within the Shield environment. Machine User Principals are commonly used in scenarios where programmatic access is required without human intervention.

- **Group**: A Group in Shield refers to a logical grouping of users or machine users who share common access rights and permissions. This type of principal simplifies the management of access control by assigning privileges and permissions to a group instead of individual users. Group principals enable administrators to efficiently define and enforce security policies across multiple users or machine users, making it easier to grant or revoke access to resources based on membership in a particular group.

- **Super User**: Superusers are platform level admins used to manage a Shield instance. A Shield Super User have the highest level of access and control, allowing them to perform administrative tasks, configure settings, and manage other organization accounts. A superuser email can be added in configurations yaml while setting up the Shield server.

:::info
**Constaints**: In Shield, a (human) user or a machine user can belong to more than one organizations, but a Group will belong to only one organization that created it.
:::

### Authenticating Principals

Shield authenticates a principal using OIDC (OpenID Connect) and the Magic Link approach (coming up). [OIDC](../concepts/glossary.md#oidc) is an authentication protocol that allows applications to verify the identity of users and obtain their basic profile information. The Magic Link approach combines the simplicity of email-based authentication with the security of OIDC to create a seamless and secure authentication experience.

Currently users authenticate with the help of an external IdP, which issues security tokens that are then exchanged with the system.

### Authorizing Principals

Once a Shield principal is authenticated, the next step is authorization. Policies are used to define the permissions and access levels for each principal. These policies are attached to principals (users, or groups) and determine what actions they can perform on which resources.

### Deactivate and Reactivate a User

Disabling a user suspends its access to organization resources and restricts its ability to perform IAM-related actions. This can be done for various reasons, such as employee termination, account compromise, temporary suspension of services, or security considerations. When a user account is diabled it cannot be used to login to Shield. The user account is not deleted, it will lose access to all the organization resources. The user account can be re-enabled at any time.

Enabling or disabling a user helps administrators manage the access and privileges of the users and controls their interaction with IAM services and resources.

### Delete a User

Deleting a user account permanently removes the user from the organization, including all associated groups and projects.

(TODO explain the case if we delete an org admin, what if there is only one org admin and we delete that user, how to add admins again. Should there be a constraint to be added to handle this?)

:::danger
Caution should be exercised when deleting a user, as it cannot be undone. If the user account with same credentials is created again, it's organization, group and project memberships will not be restored.
:::
