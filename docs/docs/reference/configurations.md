# Configurations

## Server Configurations

<details>
<summary> Sample Config </summary>

```yaml title=config.yaml
app:
  port: 8000
  grpc:
    port: 8001
  metrics_port: 9000
  host: 127.0.0.1
  identity_proxy_header: X-Shield-Email
  # full path prefixed with scheme where resources config yaml files are kept
  # e.g.:
  # local storage file "file:///tmp/resources_config"
  # GCS Bucket "gs://shield/resources_config"
  resources_config_path: file:///tmp/resources_config\
  # secret required to access resources config
  # e.g.:
  # system environment variable "env://TEST_RULESET_SECRET"
  # local file "file:///opt/auth.json"
  # secret string "val://user:password"
  # optional
  resources_config_path_secret: env://TEST_RESOURCE_CONFIG_SECRET
  # disable_orgs_listing if set to true will disallow non-admin APIs to list all organizations
  disable_orgs_listing: false
  # disable_orgs_listing if set to true will disallow non-admin APIs to list all users
  disable_users_listing: false
  # cors_origin is origin value from where we want to allow cors
  cors_origin: http://localhost:3000
  # configuration to allow authentication in shield
  authentication:
    # to use shield as session store
    session:
      # both of them should be 32 chars long
      # hash helps identify if the value is tempered with
      hash_secret_key: "hash-secret-should-be-32-chars--"
      # block helps in encryption
      block_secret_key: "block-secret-should-be-32-chars-"
    # once authenticated, server responds with a jwt with user context
    token:
      # generate key file via "./shield server keygen"
      rsa_path: ./temp/rsa
      iss: "http://localhost.shield"
    # external host used for oidc redirect uri, e.g. http://localhost:8000/v1beta1/auth/callback
    oidc_callback_host: http://localhost:8000/v1beta1/auth/callback
    # oidc auth server configs
    oidc_config:
      google:
        client_id: "xxxxx.apps.googleusercontent.com"
        client_secret: "xxxxx"
        issuer_url: "https://accounts.google.com"
  # platform level administration
  admin:
    # Email list of users which needs to be converted as superusers
    # if the user is already present in the system, it is promoted to su
    # if not, a new account is created with provided email id and promoted to su.
    # UUIDs/slugs of existing users can also be provided instead of email ids
    # but in that case a new user will not be created.
    users: []

db:
  driver: postgres
  url: postgres://username:password@localhost:5432/databaseName?sslmode=disable
  max_idle_conns: 10
  max_open_conns: 10
  conn_max_life_time: 10ms
  max_query_timeout: 100ms

spicedb:
  host: localhost
  pre_shared_key: random_key
  port: 50051
  # fully_consistent ensures APIs although slower than usual will result in responses always most consistent
  # suggested to keep it false for performance
  fully_consistent: false

# proxy configuration
proxy:
  services:
    - name: test
      host: 0.0.0.0
      # port where the proxy will be listening on for requests
      port: 5556
      # full path prefixed with scheme where ruleset yaml files are kept
      # e.g.:
      # local storage file "file:///tmp/rules"
      # GCS Bucket "gs://shield/rules"
      ruleset: file:///tmp/rules
      # secret required to access ruleset
      # e.g.:
      # system environment variable "env://TEST_RULESET_SECRET"
      # local file "file:///opt/auth.json"
      # secret string "val://user:password"
      # optional
      ruleset_secret: env://TEST_RULESET_SECRET
```

</details>

This page contains reference for all the application configurations for Shield.

### Version

| **Field**   | **Type** | **Description**                                  | **Required** |
| ----------- | -------- | ------------------------------------------------ | ------------ |
| **version** | `int`    | Version number of the Shield configuration file. | No           |

### Loggin Configuration

| **Field**     | **Type** | **Description**                                                                            | **Required** |
| ------------- | -------- | ------------------------------------------------------------------------------------------ | ------------ |
| **log.level** | `string` | Logging level for Shield. Possible values **`debug`, `info`, `warning`, `error`, `fatal`** | No           |

### App Configuration

| **Field**                            | **Description**                                                                  | **Example** | **Required** |
| ------------------------------------ | -------------------------------------------------------------------------------- | ----------- | ------------ |
| **app.port**                         | Port number for HTTP communication.                                              | 8000        | Yes          |
| **app.grpc_port**                    | Port number for gRPC communication.                                              | 8001        | Yes          |
| **app.metrics_port**                 | Port number for metrics reporting.                                               | 9000        | Yes          |
| **app.host**                         | Host address for the Shield application.                                         | 127.0.0.1   | Yes          |
| **app.identity_proxy_header**        | Header key used for identity proxy.                                              |             |              |
| **app.resources_config_path**        | Full path prefixed with the scheme where resources config YAML files are stored.<br/>Either new resources can be added dynamically via the apis, or can be passed in this YAML file |             | No           |
| **app.resources_config_path_secret** | Secret required to access resources config.                                      |             | No           |
| **app.disable_orgs_listing**         | If set to true, disallows non-admin APIs to list all organizations.              |             | No           |
| **app.disable_users_listing**        | If set to true, disallows non-admin APIs to list all users.                      |             | No           |
| **app.cors_origin**                  | Origin value from where CORS is allowed.                                         |             | Yes(for Admin UI)|

### Authentication Configurations

Configuration to allow authentication in Shield.

| **Field**                                               | **Description**                                    | **Required** | **Example**                                   |
| ------------------------------------------------------- | -------------------------------------------------- | ------------ | --------------------------------------------- |
| **app.authentication.session.hash_secret_key**          | Secret key for session hashing.                    | Yes          | "hash-secret-should-be-32-chars--"            |
| **app.authentication.session.block_secret_key**         | Secret key for session encryption.                 | Yes          | "block-secret-should-be-32-chars-"            |
| **app.authentication.token.rsa_path**                   | Path to the RSA key file for token authentication. | Yes          | "./temp/rsa"                                  |
| **app.authentication.token.iss**                        | Issuer URL for token authentication.               | Yes          | "http://localhost.shield"                     |
| **app.authentication.oidc_callback_host**               | External host used for OIDC redirect URI.          | Yes          | "http://localhost:8000/v1beta1/auth/callback" |
| **app.authentication.oidc_config.google.client_id**     | Google client ID for OIDC authentication.          | No           | "xxxxx.apps.googleusercontent.com"            |
| **app.authentication.oidc_config.google.client_secret** | Google client secret for OIDC authentication.      | No           | "xxxxx"                                       |
| **app.authentication.oidc_config.google.issuer_url**    | Google issuer URL for OIDC authentication.         | No           | "https://accounts.google.com"                 |

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
