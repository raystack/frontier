# Introduction

Shield is a cloud native role-based authorization aware reverse-proxy service. With Shield, you can assign roles to users or groups of users to configure policies that determine whether a particular user has the ability to perform a certain action on a given resource.

## Key Features

Here's a list of all the key features of Shield

- **Policy Management**: Policies help you assign various roles to users/groups that determine their access to various resources
- **Group Management**: Group is nothing but another word for team. Shield provides APIs to add/remove users to/from a group, fetch list of users in a group along with their roles in the group, and fetch list of groups a user is part of.
- **Activity Logs**: Shield has APIs that store and retrieve all the access related logs. You can see who added/removed a user to/from group in these logs.
- **Reverse Proxy**: In addition to configuring access management, you can also use Shield as a reverse proxy to directly protect your endpoints by configuring appropriate permissions for them.
- **Google IAP**: Shield also utilizes Google IAP as an authentication mechanism. So if your services are behind a Google IAP, Shield will seemlessly integrate with it.
- **Runtime**: Shield can run inside containers or VMs in a fully managed runtime environment like Kubernetes. Shield also depends on a Postgres server to store data.

## How can I get started?

- [Guides](#) provide guidance on how to use Shield and configure it to your needs
- [Concepts](#) descibe the primary concepts and architecture behind Shield
- [Reference](#) contains the list of all the APIs that Shield exposes
- [Contributing](#) contains resources for anyone who wants to contribute to Shield
