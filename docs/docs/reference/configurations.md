# Server Configurations

<details>
<summary> Sample Config </summary>

```yaml title=config.yaml
version: 1

# logging configuration
log:
  # debug, info, warning, error, fatal - default 'info'
  level: debug
  #  none(default), stdout, db
  audit_events: none

app:
  port: 8000
  grpc:
    port: 8001
    # optional tls configuration for grpc server
    tls_cert_file: "temp/server-cert.pem"
    tls_key_file: "temp/server-key.pem"
    tls_client_ca_file: "temp/ca-cert.pem"
  metrics_port: 9000
  identity_proxy_header: X-Frontier-Email
  # full path prefixed with scheme where resources config yaml files are kept
  # e.g.:
  # local storage file "file:///tmp/resources_config"
  # GCS Bucket "gs://frontier/resources_config"
  resources_config_path: file:///tmp/resources_config\
  # secret required to access resources config
  # e.g.:
  # system environment variable "env://TEST_RULESET_SECRET"
  # local file "file:///opt/auth.json"
  # secret string "val://user:password"
  # optional
  resources_config_path_secret: env://TEST_RESOURCE_CONFIG_SECRET
  # disable_orgs_on_create if set to true will set the org status to disabled on creation. This can be used to
  # prevent users from accessing the org until they contact the admin and get it enabled. Default is false
  disable_orgs_on_create: false
  # disable_orgs_listing if set to true will disallow non-admin APIs to list all organizations
  disable_orgs_listing: false
  # disable_orgs_listing if set to true will disallow non-admin APIs to list all users
  disable_users_listing: false
  # configs for user invitation to join an organization
  invite:
    # with_roles if set to true will allow people in org with the permission to send invitation to users
    # with set of role ids. When the invitation is accepted, the user will be added to the org with the roles specified
    # This can be a security risk if the user who is inviting is not careful about the roles he is adding
    # and cause permission escalation
    # Note: this is dangerous and should be used with caution
    with_roles: false
    # invite email template (if not specified, default template will be used)
    mail_template:
      subject: "You have been invited to join an organization"
      body: "<div>Hi {{.UserID}},</div><br><p>You have been invited to join an organization: {{.Organization}}. Login to your account to accept the invitation.</p><br><div>Thanks,<br>Team Frontier</div>"
  # cross-origin resource sharing configuration
  cors:
    # allowed_origins is origin value from where we want to allow cors
    allowed_origins:
      - "https://example.com" # use "*" to allow all origins
    allowed_methods:
      - POST
      - GET
      - PUT
      - PATCH
      - DELETE
    allowed_headers:
      - Authorization
    exposed_headers:
      - Content-Type
  # configuration to allow authentication in frontier
  authentication:
    # to use frontier as session store
    session:
      # both of them should be 32 chars long
      # hash helps identify if the value is tempered with
      hash_secret_key: "hash-secret-should-be-32-chars--"
      # block helps in encryption
      block_secret_key: "block-secret-should-be-32-chars-"
      # domain used for setting cookies, if not set defaults to request origin host
      domain: ""
      # same site policy for cookies
      # can be one of: "", "lax"(default value), "strict", "none"
      same_site: "lax"
      # secure flag for cookies
      secure: false
      # validity of the session
      validity: "720h"
    # once authenticated, server responds with a jwt with user context
    # this jwt works as a bearer access token for all APIs
    token:
      # generate key file via "./frontier server keygen"
      # if not specified, access tokens will be disabled
      # example: /opt/rsa
      rsa_path: ""
      # if rsa_path is not specified, rsa_base64 can be used to provide the rsa key in base64 encoded format
      rsa_base64: ""
      # issuer claim to be added to the jwt
      iss: "http://localhost.frontier"
      # validity of the token
      validity: "1h"
    # Public facing host used for oidc redirect uri and mail link redirection
    # after user credentials are verified.
    # If frontier is exposed behind a proxy, this should set as proxy endpoint
    # e.g. http://localhost:7400/v1beta1/auth/callback
    # Only the first host is used for callback by default, if multiple hosts are provided
    # they can be used to override the callback host for specific strategies using query param
    callback_urls: ["http://localhost:8000/v1beta1/auth/callback"]
    # by default, after successful authentication(flow completes) no operation will be performed,
    # to apply redirection in case of browsers, provide a list of urls one of which will be used
    # after authentication where users will be redirected to.
    # this is optional
    authorized_redirect_urls: []
    # oidc auth server configs
    oidc_config:
      google:
        client_id: "xxxxx.apps.googleusercontent.com"
        client_secret: "xxxxx"
        issuer_url: "https://accounts.google.com"
        # validity of the verification duration
        validity: "10m"
    mail_otp:
      subject: "Frontier - Login Link"
      # body is a go template with `Otp` as a variable
      body: "Please copy/paste the OneTimePassword in login form.<h2>{{.Otp}}</h2>This code will expire in 10 minutes."
      validity: "1h"
    mail_link:
      subject: "Frontier Login - One time link"
      # body is a go template with `Otp` as a variable
      body: "Click on the following link or copy/paste the url in browser to login.<br><h2><a href='{{.Link}}' target='_blank'>Login</a></h2><br>Address: {{.Link}} <br>This link will expire in 15 minutes."
      validity: 15m
  # platform level administration
  admin:
    # Email list of users which needs to be converted as superusers
    # if the user is already present in the system, it is promoted to su
    # if not, a new account is created with provided email id and promoted to su.
    # UUIDs/slugs of existing users can also be provided instead of email ids
    # but in that case a new user will not be created.
    users: []
  # smtp configuration for sending emails
  mailer:
    smtp_host: smtp.example.com
    smtp_port: 587
    smtp_username: "username"
    smtp_password: "password"
    smtp_insecure: true
    headers:
      from: "username@acme.org"
db:
  driver: postgres
  url: postgres://frontier:@localhost:5432/frontier?sslmode=disable
  max_query_timeout: 500ms

spicedb:
  host: spicedb.localhost
  pre_shared_key: randomkey
  port: 50051
  # fully_consistent ensures APIs although slower than usual will result in responses always most consistent
  # suggested to keep it false for performance
  fully_consistent: false
```

</details>

This page contains reference for all the application configurations for Frontier.

### Version

| **Field**   | **Type** | **Description**                                    | **Required** |
| ----------- | -------- | -------------------------------------------------- | ------------ |
| **version** | `int`    | Version number of the Frontier configuration file. | No           |

### Logging Configuration

| **Field**            | **Type** | **Description**                                                                              | **Required** |
| -------------------- | -------- | -------------------------------------------------------------------------------------------- | ------------ |
| **log.level**        | `string` | Logging level for Frontier. Possible values **`debug`, `info`, `warning`, `error`, `fatal`** | No           |
| **log.audit_events** | `string` | Audit level for Frontier. Possible values **`none`, `stdout`, `db`**                         | No           |

### App Configuration

| **Field**                            | **Description**                                                                                                                                                                     | **Example** | **Required**      |
| ------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- | ----------------- |
| **app.port**                         | Port number for HTTP communication.                                                                                                                                                 | 8000        | Yes               |
| **app.grpc_port**                    | Port number for gRPC communication.                                                                                                                                                 | 8001        | Yes               |
| **app.metrics_port**                 | Port number for metrics reporting.                                                                                                                                                  | 9000        | Yes               |
| **app.host**                         | Host address for the Frontier application.                                                                                                                                          | 127.0.0.1   | Yes               |
| **app.identity_proxy_header**        | Header key used for identity proxy.                                                                                                                                                 |             |                   |
| **app.resources_config_path**        | Full path prefixed with the scheme where resources config YAML files are stored.<br/>Either new resources can be added dynamically via the apis, or can be passed in this YAML file |             | No                |
| **app.resources_config_path_secret** | Secret required to access resources config.                                                                                                                                         |             | No                |
| **app.disable_orgs_listing**         | If set to true, disallows non-admin APIs to list all organizations.                                                                                                                 |             | No                |
| **app.disable_users_listing**        | If set to true, disallows non-admin APIs to list all users.                                                                                                                         |             | No                |
| **app.cors_origin**                  | Origin value from where CORS is allowed.                                                                                                                                            |             | Yes(for Admin UI) |

### Authentication Configurations

Configuration to allow authentication in Frontier.

| **Field**                                          | **Description**                                     | **Required** | **Example**                                       |
| -------------------------------------------------- |-----------------------------------------------------| ------------ |---------------------------------------------------|
| **app.authentication.session.hash_secret_key**     | Secret key for session hashing.                     | Yes          | "hash-secret-should-be-32-chars--"                |
| **app.authentication.session.block_secret_key**    | Secret key for session encryption.                  | Yes          | "block-secret-should-be-32-chars-"                |
| **app.authentication.token.rsa_path**              | Path to the RSA key file for token authentication.  | Yes          | "./temp/rsa"                                      |
| **app.authentication.token.iss**                   | Issuer URL for token authentication.                | Yes          | "http://localhost.frontier"                       |
| **app.authentication.callback_urls**               | External host used for OIDC/Mail link redirect URI. | Yes          | "['http://localhost:8000/v1beta1/auth/callback']" |
| **app.authentication.oidc_config.google.client_id** | Google client ID for OIDC authentication.           | No           | "xxxxx.apps.googleusercontent.com"                |
| **app.authentication.oidc_config.google.client_secret** | Google client secret for OIDC authentication.       | No           | "xxxxx"                                           |
| **app.authentication.oidc_config.google.issuer_url** | Google issuer URL for OIDC authentication.          | No           | "https://accounts.google.com"                     |

### Admin Configurations

| **Field**           | **Description**                                                                                                              | **Example** | **Required** |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ----------- | ------------ |
| **app.admin.users** | Email list of users to be converted as superusers. <br/> If the user is already present, they will be promoted to superuser. |             | Optional     |

### Database Configurations

| **Field**                 | **Description**                               | **Example**                                                                | **Required** |
| ------------------------- | --------------------------------------------- | -------------------------------------------------------------------------- | ------------ |
| **db.driver**             | Database driver. Currently supports postgres. | `postgres`                                                                 | Yes          |
| **db.url**                | Database connection URL.                      | `postgres://username:password@localhost:5432/databaseName?sslmode=disable` | Yes          |
| **db.max_idle_conns**     | Maximum number of idle database connections.  | `10`                                                                       | No           |
| **db.max_open_conns**     | Maximum number of open database connections.  | `10`                                                                       | No           |
| **db.conn_max_life_time** | Maximum connection lifetime.                  | `10ms`                                                                     | No           |
| **db.max_query_timeout**  | Maximum query execution timeout.              | `500ms`                                                                    | No           |

### SpiceDB Configurations

| **Field**                    | **Type**  | **Description**                                                     | **Example** | **Required** |
| ---------------------------- | --------- | ------------------------------------------------------------------- | ----------- | ------------ |
| **spicedb.host**             | `string`  | Hostname or IP address of the SpiceDB service                       | localhost   | Yes          |
| **spicedb.pre_shared_key**   | `string`  | Random key for authentication and secure communication with SpiceDB | random_key  | Yes          |
| **spicedb.port**             | `uint`    | Port number on which the SpiceDB service is listening               | 50051       | Yes          |
| **spicedb.fully_consistent** | `boolean` | Enable consistent API responses (slower but most consistent)        | false       | No           |
