# Configuration

Shield binary contains both the CLI client and the server. Each has it's own configuration in order to run. Server configuration contains information such as database credentials, spicedb connection, proxy, log severity, etc. while CLI client configuration only has configuration about which server to connect.

## Server Setup

There are several approaches to setup Shield Server

1. [Using the CLI](#using-the-cli)
1. [Using the Docker](#using-the-docker)
1. [Using the Helm Chart](#using-the-helm-chart)

#### General pre-requisites

- PostgreSQL (version 13 or above)
- [SpiceDB](https://authzed.com/docs/spicedb/installing)

## Using the CLI

### Using config file

Create a config file with the following command

```bash
$ shield server init
```

Alternatively you can [use `--config` flag](#using---config-flag) to customize to config file location.You can also [use environment variables](#using-environment-variable) to provide the server configuration.

Setup up the Postgres database, and SpiceDB instance and provide the details as shown in the example below.

> If you're new to YAML and want to learn more, see [Learn YAML in Y minutes.](https://learnxinyminutes.com/docs/yaml/)

Following is a sample server configuration yaml:

<details>
<summary> config.yaml </summary>

```yaml
version: 1

log:
  level: debug

app:
  port: 8000
  grpc:
    port: 8001
  metrics_port: 9000
  identity_proxy_header: X-Shield-Email
  resources_config_path: file:///tmp/resources_config\
  resources_config_path_secret: env://TEST_RESOURCE_CONFIG_SECRET
  disable_orgs_listing: false
  disable_users_listing: false
  cors_origin: http://localhost:3000
  authentication:
    session:
      hash_secret_key: "hash-secret-should-be-32-chars--"
      block_secret_key: "block-secret-should-be-32-chars-"
    token:
      rsa_path: ./temp/rsa
      iss: "http://localhost.shield"
    oidc_callback_host: http://localhost:8000/v1beta1/auth/callback
    oidc_config:
      google:
        client_id: "xxxxx.apps.googleusercontent.com"
        client_secret: "xxxxx"
        issuer_url: "https://accounts.google.com"
  admin:
    users: []

db:
  driver: postgres
  url: postgres://shield:@localhost:5432/shield?sslmode=disable
  max_query_timeout: 500ms

spicedb:
  host: localhost
  pre_shared_key: randomkey
  port: 50051
  fully_consistent: false

proxy:
  services:
    - name: test
      host: 0.0.0.0
      port: 5556
      ruleset: file:///tmp/rules
      ruleset_secret: env://TEST_RULESET_SECRET
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
APP_METRICS_PORT=9000
APP_IDENTITY_PROXY_HEADER=X-Shield-Email
DB_DRIVER=postgres
DB_URL=postgres://shield:@localhost:5432/shield?sslmode=disable
DB_MAX_QUERY_TIMEOUT=500ms
SPICEDB_HOST=spicedb.localhost
SPICEDB_PRE_SHARED_KEY=randomkey
SPICEDB_PORT=50051
SPICEDB_FULLY_CONSISTENT=false
PROXY_SERVICES_0_NAME=test
PROXY_SERVICES_0_HOST=0.0.0.0
PROXY_SERVICES_0_PORT=5556
PROXY_SERVICES_0_RULESET=file:///tmp/rules
PROXY_SERVICES_0_RULESET_SECRET=env://TEST_RULESET_SECRET
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
$ shield server migrate
```

To run the Shield server use command:

```bash
$ shield server start
```

#### Using `--config` flag

```bash
$ shield server migrate --config=<path-to-file>
```

```bash
$ shield server start --config=<path-to-file>
```

## Using the Docker

To run the Shield server using Docker, you need to have Docker installed on your system. You can find the installation instructions [here](https://docs.docker.com/get-docker/).

You can choose to set the configuration using environment variables or a config file. The environment variables will override the config file.

If you use Docker to build shield, then configuring networking requires extra steps. Following is one of doing it by running postgres and spicedb inside with `docker-compose` first.

Go to the root of this project and run `docker-compose`.

```bash
$ docker-compose up
```

Once postgres and spicedb has been ready, we can run Shield by passing in the config of postgres and elasticsearch defined in `docker-compose.yaml` file.

### Using config file

Alternatively you can use the `shield.yaml` config file defined [above](#using-config-file) and run the following command.

```bash
$ docker run -d \
    --restart=always \
    -p 7400:7400 \
    -v $(pwd)/shield.yaml:/shield.yaml \
    --name shield-server \
    odpf/shield:<version> \
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
    --name shield-server \
    odpf/shield:<version> \
    server start
```

## Using the Helm chart

### Pre-requisites for Helm chart

Shield can be installed in Kubernetes using the Helm chart from https://github.com/odpf/charts.

Ensure that the following requirements are met:

- Kubernetes 1.14+
- Helm version 3.x is [installed](https://helm.sh/docs/intro/install/)

### Add ODPF Helm repository

Add ODPF chart repository to Helm:

```bash
helm repo add odpf https://odpf.github.io/charts/
```

You can update the chart repository by running:

```bash
helm repo update
```

### Setup helm values

The following table lists the configurable parameters of the Shield chart and their default values.

See full helm values guide [here](https://github.com/odpf/charts/tree/main/stable/shield#values) and [values.yaml](https://github.com/odpf/charts/blob/main/stable/shield/values.yaml) file

Install it with the helm command line:

```bash
helm install my-release -f values.yaml odpf/shield
```

## Client Initialisation

Add a client configurations file with the following command:

```bash
shield config init
```

Open the config file and edit the gRPC host for Shield CLI

```yml title="shield.yaml"
client:
  host: localhost:8081
```

List the client configurations with the following command:

```bash
shield config list
```

#### Required Header/Metadata in API

In the current version, all HTTP & gRPC APIs in Shield requires an identity header/metadata in the request. The header key is configurable but the default name is `X-Shield-Email`.

If everything goes well, you should see something like this:

```bash
2023-05-17T00:02:54.324+0530    info    shield starting {"version": "v0.5.1"}
2023-05-17T00:02:54.331+0530    debug   resource config cache refreshed {"resource_config_count": 0}
2023-05-17T00:02:54.333+0530    info    Connected to spiceDB: localhost:50051
2023-05-17T00:02:54.339+0530    info    metaschemas loaded      {"count": 4}
```
