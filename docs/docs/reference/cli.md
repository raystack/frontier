# CLI

## `frontier auth`

Auth configs that need to be used with frontier

## `frontier completion [bash|zsh|fish|powershell]`

Generate shell completion scripts

## `frontier config <command>`

Manage client configurations

### `frontier config init`

Initialize a new client configuration

### `frontier config list`

List client configuration settings

## `frontier environment`

List of supported environment variables

## `frontier group`

Manage groups

### `frontier group create [flags]`

Upsert a group

```
-f, --file string     Path to the group body file
-H, --header string   Header <key>:<value>
```

### `frontier group edit [flags]`

Edit a group

```
-f, --file string   Path to the group body file
```

### `frontier group list`

List all groups

### `frontier group view [flags]`

View a group

```
-m, --metadata   Set this flag to see metadata
```

## `frontier namespace`

Manage namespaces

### `frontier namespace list`

List all namespaces

### `frontier namespace view`

View a namespace

## `frontier organization`

Manage organizations

### `frontier organization admlist`

list admins of an organization

### `frontier organization create [flags]`

Upsert an organization

```
-f, --file string     Path to the organization body file
-H, --header string   Header <key>:<value>
```

### `frontier organization edit [flags]`

Edit an organization

```
-f, --file string   Path to the organization body file
```

### `frontier organization list`

List all organizations

### `frontier organization view [flags]`

View an organization

```
-m, --metadata   Set this flag to see metadata
```

## `frontier permission`

Manage permissions

### `frontier permission create [flags]`

Upsert a permission

```
-f, --file string     Path to the permission body file
-H, --header string   Header <key>:<value>
```

### `frontier permission edit [flags]`

Edit a permission

```
-f, --file string   Path to the permission body file
```

### `frontier permission list`

List all permissions

### `frontier permission view`

View a permission

## `frontier policy`

Manage policies

### `frontier policy create [flags]`

Upsert a policy

```
-f, --file string     Path to the policy body file
-H, --header string   Header <key>:<value>
```

### `frontier policy edit [flags]`

Edit a policy

```
-f, --file string   Path to the policy body file
```

### `frontier policy view`

View a policy

## `frontier project`

Manage projects

### `frontier project create [flags]`

Upsert a project

```
-f, --file string     Path to the project body file
-H, --header string   Header <key>:<value>
```

### `frontier project edit [flags]`

Edit a project

```
-f, --file string   Path to the project body file
```

### `frontier project list`

List all projects

### `frontier project view [flags]`

View a project

```
-m, --metadata   Set this flag to see metadata
```

## `frontier role`

Manage roles

### `frontier role create [flags]`

Upsert a role

```
-f, --file string     Path to the role body file
-H, --header string   Header <key>:<value>
```

### `frontier role edit [flags]`

Edit a role

```
-f, --file string   Path to the role body file
```

### `frontier role list`

List all roles

### `frontier role view [flags]`

View a role

```
-m, --metadata   Set this flag to see metadata
```

## `frontier seed [flags]`

Seed the database with initial data

```
-H, --header string   Header <key>
```

## `frontier server`

Server management

### `frontier server init [flags]`

Initialize server

```
-o, --output string      Output config file path (default "./config.yaml")
-r, --resources string   URL path of resources. Full path prefixed with scheme where resources config yaml files are kept
                         e.g.:
                         local storage file "file:///tmp/resources_config"
                         GCS Bucket "gs://frontier-bucket-example"
                         (default: file://{pwd}/resources_config)

-u, --rule string        URL path of rules. Full path prefixed with scheme where ruleset yaml files are kept
                         e.g.:
                         local storage file "file:///tmp/rules"
                         GCS Bucket "gs://frontier-bucket-example"
                         (default: file://{pwd}/rules)

```

### `frontier server keygen [flags]`

Generate 2 rsa keys as jwks for auth token generation

```
-k, --keys int   num of keys to generate (default 2)
```

### `frontier server migrate [flags]`

Run DB Schema Migrations

```
-c, --config string   config file path
```

### `frontier server migrate-rollback [flags]`

Run DB Schema Migrations Rollback to last state

```
-c, --config string   config file path
```

### `frontier server start [flags]`

Start server and proxy default on port 8080

```
-c, --config string   config file path
```

## `frontier user`

Manage users

### `frontier user create [flags]`

Upsert an user

```
-f, --file string     Path to the user body file
-H, --header string   Header <key>:<value>
```

### `frontier user edit [flags]`

Edit an user

```
-f, --file string   Path to the user body file
```

### `frontier user list`

List all users

### `frontier user view [flags]`

View an user

```
-m, --metadata   Set this flag to see metadata
```

## `frontier version`

Print version information
