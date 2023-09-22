# Configuration

Frontier binary contains both the CLI client and the server. Each has it's own configuration in order to run. Server configuration contains information such as database credentials, spicedb connection, log severity, etc. while CLI client configuration only has configuration about which server to connect.

## Server Setup

There are several approaches to setup Frontier Server

1. [Using the CLI](#using-the-cli)
2. [Using the Docker](#using-the-docker)
3. [Using the Helm Chart](#using-the-helm-chart)

#### General pre-requisites

- PostgreSQL (version 13 or above)
- [SpiceDB](https://authzed.com/docs/spicedb/installing)

## Using the CLI

### Using config file

Create a config file with the following command

```bash
$ frontier server init
```

Alternatively you can [use `--config` flag](#using---config-flag) to customize to config file location.You can also [use environment variables](#using-environment-variable) to provide the server configuration.

Setup up the Postgres database, and SpiceDB instance and provide the details as shown in the example below.

> If you're new to YAML and want to learn more, see [Learn YAML in Y minutes.](https://learnxinyminutes.com/docs/yaml/)

Following is a sample server configuration yaml:

<details>
<summary> config.yaml </summary>

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
    # optional tls config
    # tls_cert_file: "temp/server-cert.pem"
    # tls_key_file: "temp/server-key.pem"
    # tls_client_ca_file: "temp/ca-cert.pem"
  metrics_port: 9000
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

See [configuration reference](./reference/configurations.md) for more details.

### Using environment variables

All the server configurations can be passed as environment variables using underscore \_ as the delimiter between nested keys.

<details>
<summary> .env </summary>

```bash
LOG_LEVEL=debug
APP_PORT=8000
APP_GRPC_PORT=8001
DB_DRIVER=postgres
DB_URL=postgres://frontier:@localhost:5432/frontier?sslmode=disable
DB_MAX_QUERY_TIMEOUT=500ms
SPICEDB_HOST=spicedb.localhost
SPICEDB_PRE_SHARED_KEY=randomkey
SPICEDB_PORT=50051
SPICEDB_FULLY_CONSISTENT=false
```

</details>

Set the env variable using export

```bash
$ export DB_PORT = 5432
```

### Starting the server

Database migration is required during the first server initialization. In addition, re-running the migration command might be needed in a new release to apply the new schema changes (if any). It's safer to always re-run the migration script before deploying/starting a new release.

To initialize the database schema, Run Migrations with the following command:

```bash
$ frontier server migrate
```

To run the Frontier server use command:

```bash
$ frontier server start
```

#### Using `--config` flag

```bash
$ frontier server migrate --config=<path-to-file>
```

```bash
$ frontier server start --config=<path-to-file>
```

## Using the Docker

To run the Frontier server using Docker, you need to have Docker installed on your system. You can find the installation instructions [here](https://docs.docker.com/get-docker/).

You can choose to set the configuration using environment variables or a config file. The environment variables will override the config file.

If you use Docker to build frontier, then configuring networking requires extra steps. Following is one of doing it by running postgres and spicedb inside with `docker-compose` first.

Go to the root of this project and run `docker-compose`.

```bash
$ docker-compose up
```

Once postgres and spicedb has been ready, we can run Frontier by passing in the config of postgres and elasticsearch defined in `docker-compose.yaml` file.

### Using config file

Alternatively you can use the `frontier.yaml` config file defined [above](#using-config-file) and run the following command.

```bash
$ docker run -d \
    --restart=always \
    -p 7400:7400 \
    -v $(pwd)/frontier.yaml:/frontier.yaml \
    --name frontier-server \
    raystack/frontier:<version> \
    server start -c /config.yaml
```

### Using environment variables

All the configs can be passed as environment variables using underscore `_` as the delimiter between nested keys. See the example as discussed [above](#using-environment-variable)

Run the following command to start the server

```bash
$ docker run -d \
    --restart=always \
    -p 7400:7400 \
    --env-file .env \
    --name frontier-server \
    raystack/frontier:<version> \
    server start
```

## Using the Helm chart

### Pre-requisites for Helm chart

Frontier can be installed in Kubernetes using the Helm chart from https://github.com/raystack/charts.

Ensure that the following requirements are met:

- Kubernetes 1.14+
- Helm version 3.x is [installed](https://helm.sh/docs/intro/install/)

### Add Raystack Helm repository

Add Raystack chart repository to Helm:

```bash
helm repo add raystack https://raystack.github.io/charts/
```

You can update the chart repository by running:

```bash
helm repo update
```

### Setup helm values

The following table lists the configurable parameters of the Frontier chart and their default values.

See full helm values guide [here](https://github.com/raystack/charts/tree/main/stable/frontier#values) and [values.yaml](https://github.com/raystack/charts/blob/main/stable/frontier/values.yaml) file

Install it with the helm command line:

```bash
helm install my-release -f values.yaml raystack/frontier
```

## Client Initialisation

Add a client configurations file with the following command:

```bash
frontier config init
```

Open the config file and edit the gRPC host for Frontier CLI

```yml title="frontier.yaml"
client:
  host: localhost:8081
```

List the client configurations with the following command:

```bash
frontier config list
```

#### Required Header/Metadata in API

In the current version, all HTTP & gRPC APIs in Frontier requires an identity header/metadata in the request. The header key is configurable but the default name is `X-Frontier-Email`.

If everything goes well, you should see something like this:

```bash
2023-05-17T00:02:54.324+0530    info    frontier starting {"version": "v0.5.1"}
2023-05-17T00:02:54.331+0530    debug   resource config cache refreshed {"resource_config_count": 0}
2023-05-17T00:02:54.333+0530    info    Connected to spiceDB: localhost:50051
2023-05-17T00:02:54.339+0530    info    metaschemas loaded      {"count": 4}
```
