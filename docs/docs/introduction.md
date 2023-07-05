---
id: introduction
slug: /
---

# Introduction

Welcome to the introductory guide to Shield! We cover what Shield is, what problems it can solve, how it works, and how you can get started using it. If you are familiar with the basics of Shield, the guides provides a more detailed reference of available features.

## What is Shield?

Shield by Raystack is a role-based cloud-native user management system and authorization server for your applications and API endpoints. It is API-driven and provides CLI, HTTP/GRPC APIs and an Admin Portal. Shield is designed to be easy to use Identity and Access Management tool which handles user authentication by providing a Single-Sign-On(SSO) with any provider which adheres to the OIDC Protocol. It checks the client's credentials and permissions against the defined access policies before allowing or denying access to the backend services.

![Shield flow diagram](./shield-flow-diagram.png)

## How does shield work?

Here are the steps to work with Shield.

1. Configure shield: Shield has various tuning parameters that can be configured to suit the needs of the organization. 
This includes the configuration of the database, OIDC provider, email provider, etc.
2. Configure policies: This step involves definition of resource types that will exist in the connected backend. 
A resource is a protected backend service for example an order management service or user picture library. User can 
also configure the roles and permissions to be assigned to the users.
3. Connecting frontend: Shield provides a set of APIs that can be used by the frontend to authenticate, authorize and
manage users. The frontend can be a web application, mobile application or any other application that can make HTTP
requests.

## Key Features

- **Single-Sign-On(SSO)**: Shield implements the OpenID Connect(OIDC) Protocol that extends the OAuth 2.0 framework to provide authentication and identity information, allowing users to authenticate once and access multiple applications seamlessly. It also enables applications to obtain user identity information for authorization purposes. A Single login and logout for all the underlying application and resources ensure a seamless user experience.

- **Multiple sources of user identities**: Shield allows seamless integration with any identity provider that supports OIDC, enabling a wide range of options for user authentication. This flexibility empowers organizations to leverage their preferred identity providers, such as Google, Microsoft Azure AD, Github, LinkedIn or others, to authenticate and manage user identities within their organisation's ecosystem.

- **Role-Based-Access-Control(RBAC)**: Shield follows the RBAC model, which means you can assign roles to users, groups, or service accounts. This simplifies access management by providing predefined roles with specific permissions, reducing the need for manual permission assignments.

- **Resource Management**: Shield provides API to create and manage organizations/projects. Admins can create projects and groups within organizations. In addition, group admins can manage groups, add-remove members to the groups, and assign roles to these principals. Using Shield, admins can define policies which bind roles to the underlying resources.

## Using Shield

You can manage organizations, projects, group, users and resources in any of the following ways:

### Shield Command Line Interface

You can use the Shield command line interface to issue commands and to perform the entire Shield features. Using the command line can be faster and more convenient than using API. For more information on using the Shield CLI, see the CLI Reference page.

### HTTPS API

You can manage relation creation, checking authorization on a resource and much more by using the Shield HTTPS API, which lets you issue HTTPS requests directly to the service. For more information, see the API Reference page.

### Admin Portal

Besides HTTP APIs and CLI tool, Shield provides an out-of-the-box UI for admins to configure SSO for the clients and manage roles, users, groups and organisations in one place.

## Where to go from here

See the [installation](./installation) page to install the Shield CLI. Next, we recommend completing the guides. The tour provides an overview of most of the existing functionality of Shield and takes approximately 30 minutes to complete.

After completing the tour, check out the remainder of the documentation in the reference and concepts sections for your specific areas of interest. We've aimed to provide as much documentation as we can for the various components of Shield to give you a full understanding of Shield's surface area.

Finally, follow the project on [GitHub](https://github.com/raystack/shield), and contact us if you'd like to get involved.
