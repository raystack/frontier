# CLI

## `shield auth`

Auth configs that need to be used with shield

## `shield completion [bash|zsh|fish|powershell]`

Generate shell completion scripts

## `shield config <command>`

Manage client configurations

### `shield config init`

Initialize a new client configuration

### `shield config list`

List client configuration settings

## `shield environment`

List of supported environment variables

## `shield group`

Manage groups

### `shield group create [flags]`

Upsert a group

```
-f, --file string     Path to the group body file
-H, --header string   Header <key>:<value>
````

### `shield group edit [flags]`

Edit a group

```
-f, --file string   Path to the group body file
````

### `shield group list`

List all groups

### `shield group view [flags]`

View a group

```
-m, --metadata   Set this flag to see metadata
````

## `shield namespace`

Manage namespaces

### `shield namespace list`

List all namespaces

### `shield namespace view`

View a namespace

## `shield organization`

Manage organizations

### `shield organization admlist`

list admins of an organization

### `shield organization create [flags]`

Upsert an organization

```
-f, --file string     Path to the organization body file
-H, --header string   Header <key>:<value>
````

### `shield organization edit [flags]`

Edit an organization

```
-f, --file string   Path to the organization body file
````

### `shield organization list`

List all organizations

### `shield organization view [flags]`

View an organization

```
-m, --metadata   Set this flag to see metadata
````

## `shield permission`

Manage permissions

### `shield permission create [flags]`

Upsert a permission

```
-f, --file string     Path to the permission body file
-H, --header string   Header <key>:<value>
````

### `shield permission edit [flags]`

Edit a permission

```
-f, --file string   Path to the permission body file
````

### `shield permission list`

List all permissions

### `shield permission view`

View a permission

## `shield policy`

Manage policies

### `shield policy create [flags]`

Upsert a policy

```
-f, --file string     Path to the policy body file
-H, --header string   Header <key>:<value>
````

### `shield policy edit [flags]`

Edit a policy

```
-f, --file string   Path to the policy body file
````

### `shield policy view`

View a policy

## `shield project`

Manage projects

### `shield project create [flags]`

Upsert a project

```
-f, --file string     Path to the project body file
-H, --header string   Header <key>:<value>
````

### `shield project edit [flags]`

Edit a project

```
-f, --file string   Path to the project body file
````

### `shield project list`

List all projects

### `shield project view [flags]`

View a project

```
-m, --metadata   Set this flag to see metadata
````

## `shield role`

Manage roles

### `shield role create [flags]`

Upsert a role

```
-f, --file string     Path to the role body file
-H, --header string   Header <key>:<value>
````

### `shield role edit [flags]`

Edit a role

```
-f, --file string   Path to the role body file
````

### `shield role list`

List all roles

### `shield role view [flags]`

View a role

```
-m, --metadata   Set this flag to see metadata
````

## `shield server`

Server management

### `shield server init [flags]`

Initialize server

```
-o, --output string      Output config file path (default "./config.yaml")
-r, --resources string   URL path of resources. Full path prefixed with scheme where resources config yaml files are kept
                         e.g.:
                         local storage file "file:///tmp/resources_config"
                         GCS Bucket "gs://shield-bucket-example"
                         (default: file://{pwd}/resources_config)
                         
-u, --rule string        URL path of rules. Full path prefixed with scheme where ruleset yaml files are kept
                         e.g.:
                         local storage file "file:///tmp/rules"
                         GCS Bucket "gs://shield-bucket-example"
                         (default: file://{pwd}/rules)
                         
````

### `shield server keygen [flags]`

Generate 2 rsa keys as jwks for auth token generation

```
-k, --keys int   num of keys to generate (default 2)
````

### `shield server migrate [flags]`

Run DB Schema Migrations

```
-c, --config string   config file path
````

### `shield server migrate-rollback [flags]`

Run DB Schema Migrations Rollback to last state

```
-c, --config string   config file path
````

### `shield server start [flags]`

Start server and proxy default on port 8080

```
-c, --config string   config file path
````

## `shield user`

Manage users

### `shield user create [flags]`

Upsert an user

```
-f, --file string     Path to the user body file
-H, --header string   Header <key>:<value>
````

### `shield user edit [flags]`

Edit an user

```
-f, --file string   Path to the user body file
````

### `shield user list`

List all users

### `shield user view [flags]`

View an user

```
-m, --metadata   Set this flag to see metadata
````

## `shield version`

Print version information

