# Setup Server

Shield binary contains both the CLI client and the server itself. Each has it's own configuration in order to run. Server configuration contains information such as database credentials, log severity, etc. while CLI client configuration only has configuration about which server to connect.

## Server

### Pre-Requisites

Dependencies:

- PostgreSQL
- [SpiceDB]((https://github.com/authzed/spicedb))

You need to prepare and run above dependencies first before running Shield. Shield also has a `docker-compose.yaml` file in its repo that has all required dependencies. If you are interested to use it, you just need to git clone the repo and run `docker-compose up` in the root project.

### Initialization

This steps assumes all dependencies already up and running. Create a server config `config.yaml` file (shield server init) in the root folder of shield project or use `--config` flag to customize to a certain config file location or you can also use environment variables to provide the server config.

Setup up a database in postgres and provide the details in the DB field as given in the example below. For the purpose of this tutorial, we'll assume that the username is shield, password is 12345, database name is shield, host and port are localhost and 5432. Also, setup a SpiceDB database on localhost port 50051 and pre_shared_key value shield.

Note, that you can also user the `docker-compose` file to avoid all the hassle.

```
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
  port: 8080
  host: 127.0.0.1
  identity_proxy_header: X-Shield-Email
  resources_config_path: file:///path/to/shield/resources_config
db:
  driver: postgres
  url: postgres://shield:12345@localhost:5432/shield?sslmode=disable
  max_idle_conns: 10
  max_open_conns: 10
  conn_max_life_time: 10ms
  max_query_timeout: 100ms
spicedb:
  host: localhost
  port: 50051
  pre_shared_key: shield

```

You need to define the policies in a YAML file and pass it's directory path to `resources_config_path`. The rules for each path shall be defined in another YAML file and pass it's path to `ruleset`.

Next, let's look at a example policy configuration for a backend `entropy` with `firehose` and `dagger` resource types. Also, we have defined roles for `organizations` and `project` to demonstrate shield's flexibility to define policy for different category of namespaces. 

```
entropy:
  type: resource_group
  resource_types:
    - name: firehose
      roles:
        - name: viewer
          principals:
            - user
            - group
        - name: sink_editor
          principals:
            - user
            - group
      permissions:
        - name: view
          roles:
            - owner
            - organization/owner
            - viewer
        - name: sink_edit
          roles:
            - owner
            - sink_editor
            - organization/sink_editor

organization:
  type: system
  roles:
    - name: sink_editor
      principals:
        - user
        - group
    - name: database_editor
      principals:
        - group

project:
  type: system
  roles:
    - name: viewer
      principals:
        - user
        - group
    - name: owner
      principals:
        - group

```

Finally, we'll have a look at an example rule configuration.

```
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
                      subject_principal: group
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