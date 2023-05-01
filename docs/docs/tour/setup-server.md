# Setup Server

Shield binary contains both the CLI client and the server itself. Each has it's own configuration in order to run. Server configuration contains information such as database credentials, log severity, etc. while CLI client configuration only has configuration about which server to connect.

## Server

### Pre-Requisites

Dependencies:

- PostgreSQL
- [SpiceDB](https://github.com/authzed/spicedb)

You need to prepare and run above dependencies first before running Shield. Shield also has a `docker-compose.yaml` file in its repo that has all required dependencies. If you are interested to use it, you just need to git clone the repo and run `docker-compose up` in the root project.

### Initialization

This steps assumes all dependencies already up and running. Create a server config `config.yaml` file (shield server init) in the root folder of shield project or use `--config` flag to customize to a certain config file location or you can also use environment variables to provide the server config.

Setup a database in postgres and provide the details in the DB field as given in the example below. For the purpose of this tutorial, we'll assume that the username is shield, password is 12345, database name is shield, host and port are localhost and 5432. Also, setup a SpiceDB database on localhost port 50051 and pre_shared_key value shield.

```yaml title=config.yaml
version: 1
proxy:
  services:
    - port: 5556
      host: 0.0.0.0
      name: base
      ruleset: file:///path/to/shield/rules
log:
  level: info
  format: json
new_relic:
  app_name: ""
  license: ""
  enabled: false
app:
  port: 8000
  grpc:
    port: 8001
  metrics_port: 9000
  host: 127.0.0.1
  identity_proxy_header: X-Shield-Email
  resources_config_path: file:///path/to/shield/resources_config
  resources_config_path_secret: env://TEST_RESOURCE_CONFIG_SECRET
  authentication:
    session:
      hash_secret_key: "hash-secret-should-be-32-chars--"
      block_secret_key: "block-secret-should-be-32-chars-"
    token:
      rsa_path: ./temp/rsa
      iss: "http://localhost.shield"
    oidc_callback_host: http://localhost:8888/callback
    oidc_config:
      google:
        client_id: xxxxx.apps.googleusercontent.com
        client_secret: xxxxx
        issuer_url: "https://accounts.google.com"
db:
  driver: postgres
  url: postgres://username:password@localhost:5432/databaseName?sslmode=disable
  max_idle_conns: 10
  max_open_conns: 10
  conn_max_life_time: 10ms
  max_query_timeout: 100ms
spicedb:
  host: localhost
  port: 50051
  pre_shared_key: random_key
```

You need to define the policies in a YAML file and pass it's directory path to `resources_config_path`. The rules for each path shall be defined in another YAML file and pass it's path to `ruleset`.

Next, let's look at a example policy configuration for a backend `entropy` with `firehose` and `dagger` resource types. Also, we have defined roles for `organizations` and `project` to demonstrate shield's flexibility to define policy for different category of namespaces.

```yaml title=ruleset.yaml
entropy:
  type: resource_group
  resource_types:
    - name: firehose
      roles:
        - name: viewer
          principals:
            - shield/user
            - shield/group
        - name: sink_editor
          principals:
            - shield/user
            - shield/group
      permissions:
        - name: view
          roles:
            - owner
            - organization:owner
            - viewer
        - name: sink_edit
          roles:
            - owner
            - sink_editor

shield/organization:
  type: system
  roles:
    - name: owner
      principals:
        - shield/group
    - name: sink_editor
      principals:
        - shield/user
        - shield/group
    - name: database_editor
      principals:
        - shield/group

shield/project:
  type: system
  roles:
    - name: viewer
      principals:
        - shield/user
        - shield/group
    - name: owner
      principals:
        - shield/group
```

Finally, we'll have a look at an example rule configuration.

```yaml title="rules.yaml"
rules:
  - backends:
      - name: entropy
        target: "http://entropy.io"
        prefix: "/api"
        frontends:
          - name: list_firehoses
            path: "/api/firehoses"
            method: "GET"
          - name: list_firehoses
            path: "/api/firehoses/{firehose_id}"
            method: "GET"
          - name: create_firehose
            path: "/api/firehoses"
            method: "POST"
            hooks:
              - name: authz
                config:
                  action: authz_action
                  attributes:
                    resource:
                      key: firehose.name
                      type: json_payload
                    project:
                      key: X-Shield-Project
                      type: header
                      source: request
                    organization:
                      key: X-Shield-Org
                      type: header
                      source: request
                    resource_type:
                      value: "firehose"
                      type: constant
                    group_attribute:
                      key: X-Shield-Group
                      type: header
                      source: request
                  relations:
                    - role: owner
                      subject_principal: shield/group
                      subject_id_attribute: group_attribute
          - name: update_firehose_status
            path: "/firehoses/{resource}/{action}"
            method: "PUT"
            middlewares:
              - name: authz
                config:
                  actions:
                    - odin_firehose_update
                    - odin_firehose_all_actions
                  attributes:
                    resource_type:
                      value: "firehose"
                      type: constant
                    project:
                      key: X-Shield-Project
                      type: header
                      source: request
          - name: update_firehose
            path: "/api/firehoses/{resource}"
            method: "PUT"
            middlewares:
              - name: authz
                config:
                  actions:
                    - odin_firehose_update
                    - odin_firehose_all_actions
                  attributes:
                    resource_type:
                      value: "firehose"
                      type: constant
                    project:
                      key: X-Shield-Project
                      type: header
                      source: request
          - name: firehose_history
            path: "/api/history/firehose/{firehose_name}"
            method: "GET"
          - name: dagger_history
            path: "/api/history/dagger/{dagger_name}"
            method: "GET"
          - name: update_firehose_alerts
            path: "/api/alerts/firehose/{resource}"
            method: "PUT"
            middlewares:
              - name: authz
                config:
                  actions:
                    - odin_firehose_update
                    - odin_firehose_all_actions
                  attributes:
                    project:
                      key: X-Shield-Project
                      type: header
                      source: request
                    resource_type:
                      value: "firehose"
                      type: constant
```

### Migrating the server

Database migration is required during the first server initialization. In addition, re-running the migration command might be needed in a new release to apply the new schema changes (if any). It's safer to always re-run the migration script before deploying/starting a new release.

To initialize the database schema, Run Migrations with the following command:

```sh
$ shield server migrate
```

Using `--config` flag

```sh
$ shield server migrate --config=<path-to-file>
```

If migration command throws the following error, you need to create the databases first.

```sh
pq: database "shield" does not exist
```

```sh
pq: database "spicedb" does not exist
```

Let's verify the migration by looking at the tables created in both the databases.

We will use `SQL Shell (psql)` to connect to each instance.

![Shield connection config](./shield-connection-config.png)

![SpiceDB connection config](./spicedb-connection-config.png)

Let's display all the tables in both the postgres databases

```sh
$ \dt
```

![Shield tables](./shield-tables.png)

![SpiceDB tables](./spicedb-tables.png)

### Starting the server

Now, it's time to start the server, but before that, let's open a browser tab and open `localhost:8080` to have a look at the permission schema on SpiceDB.

![SpiceDB permission schema before](./permission-schema-before.png)

To run the Shield server use command:

```sh
$ shield server start
```

```sh
$ shield server start --config=<path-to-file>
```

You will see the similar logs as following, based on your configs

![Server start cmd output](./server-start-cmd-output.png)

You can ping the server to verify

```sh
curl --location --request GET 'http://localhost:8000/ping'
```

```sh
pong
```

Let's verify in the browser, if the SpiceDB permission schema is updated

![SpiceDB permission schema after](./permission-schema-after.png)
