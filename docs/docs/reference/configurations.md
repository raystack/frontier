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
  # list of audit events to be ignored
  # e.g. ["app.user.created", "app.permission.checked"]
  ignored_audit_events: []
app:
  port: 8000
  grpc: 
    port: 8001
    # optional tls config
    # tls_cert_file: "temp/server-cert.pem"
    # tls_key_file: "temp/server-key.pem"
    # tls_client_ca_file: "temp/ca-cert.pem"
  # port for application metrics
  metrics_port: 9000
  # enable pprof endpoints for cpu/mem/mutex profiling
  profiler: false
  # WARNING: identity_proxy_header bypass all authorization checks and shouldn't be used in production
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
      # custom claims configuration for the jwt
      claims:
        # if set to true, the jwt will contain the org ids of the user in the claim
        add_org_ids: true
        # if set to true, the jwt will contain the user email in the claim
        add_user_email: true
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
      body: "Please copy/paste the OneTimePassword in login form.<h2>{{.Otp}}</h2>This code will expire in 15 minutes."
      validity: 15m
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
  # webhook configuration for sending events to external services    
  webhook:
    # encryption key used to encrypt the secrets stored in database not to encrypt
    # the webhook payload
    encryption_key: "encryption-key-should-be-32-chars--"
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
  # check_trace enables tracing in check api for spicedb, it adds considerable
  # latency to the check calls and shouldn't be enabled in production
  check_trace: false

billing:
  # stripe key to be used for billing
  # e.g. sk_test_XXXXXXXXXXX
  stripe_key: ""
  # if true, tax will be calculated automatically by stripe
  # before turning it on, make sure you have configured tax rates in stripe
  stripe_auto_tax: false
  # webhook secret to be used for validating stripe webhooks events
  # all the secrets are used to validate the events useful in case of key rotation
  stripe_webhook_secrets: []
  # path to plans spec file that will be used to create plans in billing engine
  # e.g. file:///tmp/plans
  plans_path: ""
  # default currency to be used for billing if not provided by the user
  # e.g. usd, inr, eur
  default_currency: ""
  # billing customer account configuration
  customer:
    # automatically create a default customer account when an org is created
    auto_create_with_org: true
    # name of the plan that should be used subscribed automatically when the org is created
    # it also automatically creates an empty billing account under the org
    default_plan: ""
    # default offline status for the customer account, if true the customer account
    # will not be registered in billing provider
    default_offline: false
    # free credits to be added to the customer account when created as a part of the org
    onboard_credits_with_org: 0
  # plan change configuration applied when a user changes their subscription plan
  plan_change:
    # proration_behavior can be one of "create_prorations", "none", "always_invoice"
    proration_behavior: "create_prorations"
    # immediate_proration_behavior can be one of "create_prorations", "none", "always_invoice"
    # this is applied when the plan is changed immediately instead of waiting for the next billing cycle
    immediate_proration_behavior: "create_prorations"
    # collection_method can be one of "charge_automatically", "send_invoice"
    collection_method: "charge_automatically"
  # product configuration
  product:
    # seat_change_behavior can be one of "exact", "incremental"
    # "exact" will change the seat count to the exact number of users within the organization
    # "incremental" will change the seat count to the number of users within the organization
    # but will not decrease the seat count if reduced
    seat_change_behavior: "exact"
  # refresh interval for billing engine to sync with the billing provider
  # setting it too low can lead to rate limiting by the billing provider
  # setting it too high can lead to stale data in the billing engine
  # e.g. 60s, 2m, 30m
  refresh_interval:
    customer: 1m
    subscription: 1m
    invoice: 5m
    checkout: 1m
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
