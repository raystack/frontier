# Architecture

Frontier is a cloud-native role-based authentication & authorization server that helps you secure microservices of given resources. It uses [SpiceDB](https://github.com/authzed/spicedb) authorization engine, which is an open source fine-grained permissions database inspired by [Google Zanzibar](https://authzed.com/blog/what-is-zanzibar/).

We can configure role assignments to certain user or group on this resource as well during the resource creation.

We will talk more with example about the rule configuration in detail, in the guides.
Frontier exposes both HTTP and gRPC APIs to manage data. It also proxy APIs to other services. Frontier talks to SpiceDB instance to check for authorization.

## Tools and Technologies

Frontier is developed with

- Golang - Programming language
- Docker - container engine to start postgres and cortex to aid development
- Postgres - a relational database
- SpiceDB - SpiceDB is an open source database system for managing security-critical application permissions.

## Components

### API Server

Frontier server exposes both HTTP and gRPC APIs (via GRPC gateway) to manage users, groups, policies, etc. It also runs a proxy server on different port.

### PostgresDB

There are 2 PostgresDB instances. One instance is required for Frontier to store all the business logic like user detail, team detail, User's role in the team, etc.

Another DB instance is for SpiceDB to store all the data needed for authorization.

### SpiceDB

Frontier push all the policies and relationships data to SpiceDB. All this data is needed to make the authorization decision. Frontier connects to SpiceDB instance via gRPC.

## Overall System Architecture - Frontier as an Authorization Service

Frontier when used as authentication service support variety of strategies. Once a user is authenticated, user identity
is stored in db. To manage browser session, it uses Cookies and for API calls, it uses JWT tokens. JWT tokens are created
and signed by Frontier using configured RSA keys and all backend services are expected to verify the token using public keys
when the token is received. Frontier exposes an API to get the public keys.

Frontier when used as an authorization service delegates all of its check to SpiceDB using the `check` API. 
A basic check is composed of 3 items, i.e. `can a USER do an ACTION on this RESOURCE`.

![Overall System Architecture Authorization](./frontier-authorization-architecture.png)

The API gives a boolean response. You can refer this [guide](../authz/permission.md#managing-permission) for usage information.

