# Schema

![](../.gitbook/assets/schema.png)

&nbsp;
&nbsp;

### users

| column name | column type | foreign key                                   | notes                                      |
| ----------- | ----------- | --------------------------------------------- | ------------------------------------------ |
| id          | uuid        | one to many map with v0 column of casbin_rule | unique                                     |
| username    | varchar     |                                               | unique                                     |
| displayname | varchar     |                                               |                                            |
| metadata    | jsonb       |                                               | can contain any extra metadata information |
| created_at  | string      |                                               | data/time when the record was created      |
| updated_at  | string      |                                               | data/time when the record was last updated |

&nbsp;
&nbsp;

### groups

| column name | column type | foreign key                                   | notes                                      |
| ----------- | ----------- | --------------------------------------------- | ------------------------------------------ |
| id          | uuid        | one to many map with v1 column of casbin_rule | unique                                     |
| groupname   | varchar     |                                               | unique                                     |
| displayname | varchar     |                                               |                                            |
| metadata    | jsonb       |                                               | can contain any extra metadata information |
| created_at  | string      |                                               | data/time when the record was created      |
| updated_at  | string      |                                               | data/time when the record was last updated |

&nbsp;
&nbsp;

### roles

| column name | column type | foreign key                                   | notes                                      |
| ----------- | ----------- | --------------------------------------------- | ------------------------------------------ |
| id          | uuid        | one to many map with v1 column of casbin_rule | unique                                     |
| displayname | varchar     |                                               |                                            |
| metadata    | jsonb       |                                               | can contain any extra metadata information |
| created_at  | string      |                                               | data/time when the record was created      |
| updated_at  | string      |                                               | data/time when the record was last updated |

&nbsp;
&nbsp;

### casbin_rule

| column name | column type | foreign key                 | notes |
| ----------- | ----------- | --------------------------- | ----- |
| v0          | string      | maps with user id           |       |
| v1          | string      | maps with role id, group id |       |
| v2          | string      |                             |       |
| v3          | string      |                             |       |
