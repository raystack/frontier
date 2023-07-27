# Glossary

Terminology and concepts used in Frontier documentation.

### Access Token

An access token is a credential which serves as proof of authentication and authorisation to access protected resources in a system or an API. After successful authorisation, an access token is included in the headers or parameters of API requests to validate the user's identity and permissions. The access token contains information such as the client's identity and the permissions granted to it. The server validates the access token to determine if the client is authorised to access the requested resource.

### Backend

External Service which wants to use Frontier for authz. It can verify access via Frontier Proxy & API.

### Client ID

The Client ID is a public identifier assigned to a client application by the authorization server. It is typically known to the server and can be shared openly. It helps identify the client application when making requests to the authorization server.

### Client Secret

The Client Secret is a confidential string or password associated with the client application. It is meant to be kept confidential and only known to the client application and the OAuth identity provider. The Client Secret adds an extra layer of security by ensuring that only authorized client applications can access protected resources.

### Identity Providers (IdPs)

Frontier supports integration with external identity providers through OpenID Connect (OIDC). This enables users to authenticate using their existing credentials from external systems, such as Google accounts, Microsoft Azure AD, or Github. The authentication process is handled by the respective IdP, while Frontier handles the subsequent authorization based on the authenticated identity.

### Issuer URL

The Issuer URL in OIDC (OpenID Connect) is the URL that uniquely identifies the identity provider (IdP) or an authorization server. It serves as a reference point for the OIDC client to discover and retrieve necessary information about the IdP's configuration and supported endpoints. To find the Issuer URL for an OIDC provider, you typically need to refer to the provider's documentation or specifications.

### JWT Token

A JSON Web Token(JWT) is a specific type of access token that is encoded as a JSON object and conforms to the [JWT standard](https://datatracker.ietf.org/doc/html/rfc7519). It is a self-contained token that includes the user's identity and any additional claims or information as a JSON payload. The JWT is digitally signed by the issuer (usually an authorization server) to ensure its integrity and authenticity. JWTs are commonly used in modern authentication and authorization protocols, such as OAuth 2.0 and OpenID Connect.

### Namespace

Type of objects over which we want authorization. They are of two types:

1. System Namespace: Objects like Organization, Project & Team, over which we need authorization to actions such as adding user to team, adding user as owner of project.

2. Resource Namespace: Resources Types over which we need authorization. For example, we need edit & view permissions over GCE Instances.

### OIDC

OpenID Connect (or OIDC) is an identity layer built on top of the OAuth 2.0 protocol that enables clients (applications) to verify the identity of end-users based on authentication performed by an authorization server.

### Organization

An Organization is a top-level resource in Frontier. Each Project, Role, Group, Policy and Principal belongs to an organization. Each organization will represent a single tenant in your SaaS application.

### Permission

Ability to carry out any action, in Frontier or configured Backends.

### Policy

It defines the specific actions that are allowed or denied on resources. Aliter, Policies are used to enforce authorization and access control by binding one or more principals to individual roles. Policies can be attached to users, groups, or roles to grant or restrict access to resources.

### Principal

Principal refers to an entity that can be authenticated and authorized to access resources. It can represent a user, an application, or even another system. Principals are typically identified by unique identifiers, such as usernames, email addresses, or service account IDs.
In Frontier Principals are of following type:

1. User: A person or service account who can be a Principal. It is identified by an Email ID.
2. Group: Collection of Users.
3. All Registered Users: Collection of users who have registered in Frontier. Any user who registers in Frontier becomes part of this Principal.

### Project

By which we can group Resources, of various different Resource Types, who have common environment.

### Redirect URI

The Redirect URI is registered with the OIDC provider and serves as a callback URL, indicating the location which the OIDC provider redirects an authorized user with an authorization code or an ID token after authentication is complete.

### Resource

Entity which needs authorization to be accessed. For example, a GCE instance is a resource over which we need permission such as edit & view.

### Resource Type

Classification that contains Resource instances. For example, GCE can be a resource type for GCE instances.

### RBAC

Role-based access control (or RBAC) is a methodology for restricting app resources to only authorized users. RBAC defines primitives such as roles, resources, users, and policies to configure access control within an app

### Role

A Role is a collection of permissions or privileges that can be assigned to principals. Permissions determine what actions are allowed on a resource. Whenver we attach a role, a principal is granted all the permissions that the role contains.

### SpiceDB

[SpiceDB](https://github.com/authzed/spicedb) is a Zanzibar-inspired open source database system for managing security-critical application permissions.

### SSO

A Single Sign-On (SSO) mechanism allows users to authenticate once with the authorization server and then access multiple applications or services without needing to reauthenticate for each one. OIDC supports SSO by providing a standardized way to authenticate and share user identity across applications.
