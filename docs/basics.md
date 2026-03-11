# Basics

Let's walk through the basics of Frontier. Any online platform that need to manage users require at least a authentication
system. Once the user is authenticated, users should be persisted in Frontier and should work as an identity server.

## Authentication

Frontier is a headless server so everything is exposed via APIs. Imagining a website that wants to authenticate user, 
we will start with `Email Link` strategy as it's easiest to configure and pretty secure. Set following yaml configuration
in config.yaml where frontier binary is kept.

```yaml

version: 1
log:
  level: debug

app:
  port: 7400
  grpc:
    port: 7401
  host: 127.0.0.1
  # cors_origin is origin value from where we want to allow cors
  cors_origin:
    - http://localhost:8888
  authentication:
    # session managed via cookies
    session:
      # both of them should be 32 chars long
      # hash helps identify if the value is tempered with
      hash_secret_key: "hash-secret-should-be-32-chars--"
      # block helps in encryption
      block_secret_key: "block-secret-should-be-32-chars-"
      # domain used for setting cookies, if not set defaults to request origin host
      domain: "localhost"
    # public facing host used for oidc redirect uri and mail link redirection
    # e.g. http://localhost:7400/v1beta1/auth/callback
    callback_urls:
      - http://localhost:8888/callback
    mail_link:
      subject: "Frontier Login - One time link"
      # body is a go template with `Otp` as a variable
      body: "Click on the following link or copy/paste the url in browser to login.<h3><a href='{{.Link}}' target='_blank'>Login</a></h3>Address: {{.Link}} <br>This link will expire in 10 minutes."
      validity: 10m

  # platform level administration
  admin:
    # email list of users which needs to be converted as superusers
    # if the user is already present in the system, it is promoted to su
    # if not, a new account is created with provided email id and promoted to su
    users:
      - test@example.com
  # smtp configuration for sending emails
  mailer:
    smtp_host: sandbox.smtp.mailtrap.io
    smtp_port: 2525
    smtp_username: h9fdce4ce3ea5c
    smtp_password: 6d71xd9d754505
    smtp_insecure: true
    headers:
      from: test@example.com
db:
  url: postgres://frontier:frontier@localhost:5432/frontier?sslmode=disable

spicedb:
  host: localhost
  port: 50051
  pre_shared_key: frontier

```

### Some key configurations to change are:

- app.callback_urls: Callback urls are public facing endpoint where the user will be redirected to when the user clicks
on the link. This url will contain 2 query params, `state` and `otp` required for authenticating the user. In current design
we assume the frontend using frontier will receive them and send them back to frontier at `/v1beta1/auth/callback`. In
response frontier will set a header `Set-Cookie` and initialize a session. To allow cross domain access of cookie, it is advised
to host frontier at somewhere like auth.example.com and main app at anywhere in same domain like app.example.com or example.com
- app.authentication.session.domain: will be set as `example.com` if this is your hosted domain
- app.authentication.session.hash_secret_key: random 32 char key
- app.authentication.session.block_secret_key: random 32 char key
- app.cors_origin: url of the frontend to allow cross-origin request
- admin.users: list of emails we want to be promoted as instance admins
- app.mailer.*: all these details are used to send the email to user. Get test configurations from `mailtrap.io` for testing.
- db.url: local database instance credentials in following form
  - `postgres://<username>:<password>@<hostname>:<hostport>/<database_name>?sslmode=disable`
- spicedb.host: local instance host address used for authorization
- spicedb.pre_shared_key: used as a password to authenticate frontier with spicedb instance

Once all of them are configured, start a example available at `examples/auth/main.go` directory in repo using

```shell
cd examples/auth/
go run main.go
```

This will start a sample frontend supporting mail and oidc strategies. By default, it looks for froniter at
http://localhost:7400 and starts itself at http://localhost:8888.

Go through the example app and understand a basic authentication flow.

### Multi tenancy

If the online portal is a selling Software As A Service(SaaS), it is very common to handle multiple tenants. Frontier
handles multi-tenancy by creating multiple `organizations`. Each customer/tenant will have it's own organization.
An organization work as a unit to group all the company projects. So Frontier also allows creating `projects` under
the organization hierarchy. Projects are although globally unique always belong to an organization.

When a user is registered, it is not part of any organization. A user can either be invited to an organization using
Frontier `invitations` APIs. Once a multiple users are part of an organization, they can be grouped in multiple
groups as needed using `Group` APIs.

A project work as a workspace where all the external services using frontier as central authority will namespace their
items. Each project has a unique name and uuid, it is advised to use this uuid to partition customer data in all the 
backend services.

### Authorization

Once multiple users starts to onboard the platform, we need to control who has what access. Frontier exposes `Check`
APIs that solves very low level authorization, the API allows to ask if a `Principal` can do a `Action` on a `Resource`.
To simplify the user access management, Frontier provides RBAC system. The flow of giving access to users look something 
like as follows:

1. List all available `permissions` of frontier
2. Create new `permissions` if needed for external service like `compute.instance.get` or `compute.instance.delete`
3. List all available `roles` of frontier
4. Create new `roles` if needed with set of permissions(existing or new)
5. Create a `Policy` that binds the role with `principal`(a user). Frontier allows binding a role at different hierarchy
levels. For example if a policy is assigned at organization level, it allows access across whole organization where as
if the permission is assigned at project level, only that project and all of the resources under the project will be assigned
under the policy to user. User won't be able to access/govern any other project under the same org.
6. Using `Check` API, verify the user has the permission.