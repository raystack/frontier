# What is in Shield?

Now, with the initial setup done, let's go throuh the pre-polulated data in Shield's databases.


## Namespace Table

```sh
SELECT * FROM namespaces;
```

```sh
        id        |       name       |          created_at           |          updated_at           | deleted_at | backend | resource_type 
------------------+------------------+-------------------------------+-------------------------------+------------+---------+---------------
 organization     | organization     | 2022-12-06 15:47:57.898578+00 | 2022-12-06 16:00:11.687748+00 |            |         | 
 entropy/firehose | entropy/firehose | 2022-12-06 15:47:58.93294+00  | 2022-12-06 16:00:11.898179+00 |            | entropy | firehose
 project          | project          | 2022-12-06 15:47:58.25897+00  | 2022-12-06 16:00:12.218625+00 |            |         | 
 group            | group            | 2022-12-06 15:47:58.603929+00 | 2022-12-06 16:00:12.42399+00  |            |         | 
 user             | user             | 2022-12-06 15:47:57.853201+00 | 2022-12-06 16:00:12.618185+00 |            |         | 
(5 rows)
```

## Roles Table

```sh
SELECT * FROM roles;
```

```sh
              id               |     name     |        types         | metadata |          created_at           |          updated_at           | deleted_at |   namespace_id   
-------------------------------+--------------+----------------------+----------+-------------------------------+-------------------------------+------------+------------------
 organization:sink_editor      | sink_editor  | {user,group}         | null     | 2022-12-06 15:47:58.125789+00 | 2022-12-06 15:47:58.125789+00 |            | organization
 organization:owner            | owner        | {user,group}         | null     | 2022-12-06 15:47:57.932181+00 | 2022-12-06 15:47:57.932181+00 |            | organization
 organization:editor           | editor       | {user,group}         | null     | 2022-12-06 15:47:58.015648+00 | 2022-12-06 15:47:58.015648+00 |            | organization
 organization:viewer           | viewer       | {user,group}         | null     | 2022-12-06 15:47:58.082954+00 | 2022-12-06 15:47:58.082954+00 |            | organization
 entropy/firehose:viewer       | viewer       | {user,group}         | null     | 2022-12-06 15:47:58.960705+00 | 2022-12-06 15:47:58.960705+00 |            | entropy/firehose
 entropy/firehose:sink_editor  | sink_editor  | {user,group}         | null     | 2022-12-06 15:47:59.027849+00 | 2022-12-06 15:47:59.027849+00 |            | entropy/firehose
 entropy/firehose:owner        | owner        | {user,group}         | null     | 2022-12-06 15:47:59.091254+00 | 2022-12-06 15:47:59.091254+00 |            | entropy/firehose
 entropy/firehose:editor       | editor       | {user,group}         | null     | 2022-12-06 15:47:59.150816+00 | 2022-12-06 15:47:59.150816+00 |            | entropy/firehose
 entropy/firehose:organization | organization | {InheritedNamespace} | null     | 2022-12-06 15:47:59.18893+00  | 2022-12-06 15:47:59.18893+00  |            | entropy/firehose
 entropy/firehose:project      | project      | {InheritedNamespace} | null     | 2022-12-06 15:47:59.250437+00 | 2022-12-06 15:47:59.250437+00 |            | entropy/firehose
 project:owner                 | owner        | {group,user}         | null     | 2022-12-06 15:47:58.267137+00 | 2022-12-06 15:47:58.267137+00 |            | project
 project:editor                | editor       | {user,group}         | null     | 2022-12-06 15:47:58.324931+00 | 2022-12-06 15:47:58.324931+00 |            | project
 project:viewer                | viewer       | {user,group}         | null     | 2022-12-06 15:47:58.387529+00 | 2022-12-06 15:47:58.387529+00 |            | project
 project:organization          | organization | {InheritedNamespace} | null     | 2022-12-06 15:47:58.447548+00 | 2022-12-06 15:47:58.447548+00 |            | project
 group:member                  | member       | {user}               | null     | 2022-12-06 15:47:58.612887+00 | 2022-12-06 15:47:58.612887+00 |            | group
 group:manager                 | manager      | {user}               | null     | 2022-12-06 15:47:58.654405+00 | 2022-12-06 15:47:58.654405+00 |            | group
 group:organization            | organization | {InheritedNamespace} | null     | 2022-12-06 15:47:58.711528+00 | 2022-12-06 15:47:58.711528+00 |            | group
(17 rows)
```

## Actions Table

```sh
SELECT * FROM actions;
```

```sh
             id             |    name    |          created_at           |          updated_at           | deleted_at |   namespace_id   
----------------------------+------------+-------------------------------+-------------------------------+------------+------------------
 edit.organization          | edit       | 2022-12-06 15:47:58.182569+00 | 2022-12-06 15:47:58.182569+00 |            | organization
 view.organization          | view       | 2022-12-06 15:47:58.225102+00 | 2022-12-06 15:47:58.225102+00 |            | organization
 view.entropy/firehose      | view       | 2022-12-06 15:47:59.316085+00 | 2022-12-06 15:47:59.316085+00 |            | entropy/firehose
 sink_edit.entropy/firehose | sink_edit  | 2022-12-06 15:47:59.347891+00 | 2022-12-06 15:47:59.347891+00 |            | entropy/firehose
 edit.entropy/firehose      | edit       | 2022-12-06 15:47:59.377112+00 | 2022-12-06 15:47:59.377112+00 |            | entropy/firehose
 delete.entropy/firehose    | delete     | 2022-12-06 15:47:59.407132+00 | 2022-12-06 15:47:59.407132+00 |            | entropy/firehose
 delete.project             | delete     | 2022-12-06 15:47:58.571933+00 | 2022-12-06 15:47:58.571933+00 |            | project
 edit.project               | edit       | 2022-12-06 15:47:58.507712+00 | 2022-12-06 15:47:58.507712+00 |            | project
 view.project               | view       | 2022-12-06 15:47:58.539204+00 | 2022-12-06 15:47:58.539204+00 |            | project
 edit.group                 | edit       | 2022-12-06 15:47:58.758155+00 | 2022-12-06 15:47:58.758155+00 |            | group
 view.group                 | view       | 2022-12-06 15:47:58.788941+00 | 2022-12-06 15:47:58.788941+00 |            | group
 delete.group               | delete     | 2022-12-06 15:47:58.850114+00 | 2022-12-06 15:47:58.850114+00 |            | group
 membership.group           | membership | 2022-12-06 15:47:58.90078+00  | 2022-12-06 15:47:58.90078+00  |            | group
(13 rows)
```

## Policies Table

```sh
SELECT * FROM policies;
```

```sh
                  id                  |           role_id            |   namespace_id   |         action_id          |          created_at           |          updated_at           | deleted_at 
--------------------------------------+------------------------------+------------------+----------------------------+-------------------------------+-------------------------------+------------
 b7612685-6aca-4f9a-bd95-db23b1711e65 | organization:owner           | organization     | view.organization          | 2022-12-06 15:48:03.095428+00 | 2022-12-06 15:48:03.095428+00 | 
 76823b4c-ad47-45dc-b36e-08c89c945151 | organization:editor          | organization     | view.organization          | 2022-12-06 15:48:03.170675+00 | 2022-12-06 15:48:03.170675+00 | 
 9464634e-84c7-4e53-81f1-6bec929fb0ee | organization:viewer          | organization     | view.organization          | 2022-12-06 15:48:03.244282+00 | 2022-12-06 15:48:03.244282+00 | 
 4d652f7e-5a18-4c5c-91b7-fb715679f0c8 | organization:owner           | organization     | edit.organization          | 2022-12-06 15:48:02.938871+00 | 2022-12-06 15:48:02.938871+00 | 
 13de9e96-1131-4bf1-b208-bd2c3b7fbc3e | organization:editor          | organization     | edit.organization          | 2022-12-06 15:48:03.019583+00 | 2022-12-06 15:48:03.019583+00 | 
 f4d3ee6d-7430-4db2-8020-84718d974b5a | entropy/firehose:owner       | entropy/firehose | view.entropy/firehose      | 2022-12-06 15:48:01.349746+00 | 2022-12-06 15:48:01.349746+00 | 
 18d62e6c-9b08-4c09-afe2-3b33221f4bb7 | entropy/firehose:editor      | entropy/firehose | view.entropy/firehose      | 2022-12-06 15:48:01.426043+00 | 2022-12-06 15:48:01.426043+00 | 
 32f70f7c-0d42-45a8-b89d-397dc5c6f071 | entropy/firehose:viewer      | entropy/firehose | view.entropy/firehose      | 2022-12-06 15:48:01.503689+00 | 2022-12-06 15:48:01.503689+00 | 
 4886ee85-2139-45db-a65f-f0479ef410af | organization:owner           | entropy/firehose | view.entropy/firehose      | 2022-12-06 15:48:01.57801+00  | 2022-12-06 15:48:01.57801+00  | 
 fb4ad469-d4fe-4a7b-9ccb-c2ff210b94ff | organization:editor          | entropy/firehose | view.entropy/firehose      | 2022-12-06 15:48:01.653929+00 | 2022-12-06 15:48:01.653929+00 | 
 f8213e9c-38a5-4b0b-81ed-262d9eafaeec | organization:viewer          | entropy/firehose | view.entropy/firehose      | 2022-12-06 15:48:01.731061+00 | 2022-12-06 15:48:01.731061+00 | 
 0c0f432b-7bfb-4ff6-a123-19894b69d1ce | project:owner                | entropy/firehose | view.entropy/firehose      | 2022-12-06 15:48:01.803733+00 | 2022-12-06 15:48:01.803733+00 | 
 1e33316e-fef1-4b67-9fe4-9ae827778e42 | project:editor               | entropy/firehose | view.entropy/firehose      | 2022-12-06 15:48:01.881091+00 | 2022-12-06 15:48:01.881091+00 | 
 1f5bcce5-ac2f-49dc-8e6e-58a01756ab51 | project:viewer               | entropy/firehose | view.entropy/firehose      | 2022-12-06 15:48:01.956207+00 | 2022-12-06 15:48:01.956207+00 | 
 6bcbef6c-395b-4075-8c1a-59c745e09f3a | entropy/firehose:owner       | entropy/firehose | sink_edit.entropy/firehose | 2022-12-06 15:48:02.032317+00 | 2022-12-06 15:48:02.032317+00 | 
 904f4d75-8ba4-41c4-be6f-e462151ad1dd | entropy/firehose:sink_editor | entropy/firehose | sink_edit.entropy/firehose | 2022-12-06 15:48:02.109671+00 | 2022-12-06 15:48:02.109671+00 | 
 4bc60718-1f9e-4592-a962-4d316f54f840 | organization:sink_editor     | entropy/firehose | sink_edit.entropy/firehose | 2022-12-06 15:48:02.179724+00 | 2022-12-06 15:48:02.179724+00 | 
 e50cde2a-5e09-4a2e-857e-57dac3bf2a41 | entropy/firehose:owner       | entropy/firehose | edit.entropy/firehose      | 2022-12-06 15:48:02.256486+00 | 2022-12-06 15:48:02.256486+00 | 
 28aa10f8-359a-4474-a481-b6955ef6402e | entropy/firehose:editor      | entropy/firehose | edit.entropy/firehose      | 2022-12-06 15:48:02.330007+00 | 2022-12-06 15:48:02.330007+00 | 
 be59cd7e-c6ef-4594-9e62-e14bf8bb3b3e | organization:owner           | entropy/firehose | edit.entropy/firehose      | 2022-12-06 15:48:02.404029+00 | 2022-12-06 15:48:02.404029+00 | 
 edbeea61-8cf5-4f67-bcaf-847ba30cfd8c | organization:editor          | entropy/firehose | edit.entropy/firehose      | 2022-12-06 15:48:02.483747+00 | 2022-12-06 15:48:02.483747+00 | 
 29f6f1c1-e45c-4eb3-8cc4-614b1faa4a62 | project:owner                | entropy/firehose | edit.entropy/firehose      | 2022-12-06 15:48:02.554789+00 | 2022-12-06 15:48:02.554789+00 | 
 d6252544-0a4d-4c52-a6c0-dfd2276842d8 | project:editor               | entropy/firehose | edit.entropy/firehose      | 2022-12-06 15:48:02.627653+00 | 2022-12-06 15:48:02.627653+00 | 
 a83c7eae-67da-4938-87c0-371f6734ec8d | entropy/firehose:owner       | entropy/firehose | delete.entropy/firehose    | 2022-12-06 15:48:02.704619+00 | 2022-12-06 15:48:02.704619+00 | 
 fc06cbc5-2eee-4475-9a9e-f213e2a32eb3 | organization:owner           | entropy/firehose | delete.entropy/firehose    | 2022-12-06 15:48:02.780577+00 | 2022-12-06 15:48:02.780577+00 | 
 54480a13-7855-4931-9114-ab122ffe6e87 | project:owner                | entropy/firehose | delete.entropy/firehose    | 2022-12-06 15:48:02.852104+00 | 2022-12-06 15:48:02.852104+00 | 
 eeff8fa6-ce5f-4381-bbd1-551bdf2b2250 | project:owner                | project          | edit.project               | 2022-12-06 15:48:00.076655+00 | 2022-12-06 15:48:00.076655+00 | 
 a373d50e-6322-4e95-ac98-220ef8f2ffd5 | project:editor               | project          | edit.project               | 2022-12-06 15:48:00.151133+00 | 2022-12-06 15:48:00.151133+00 | 
 df6414f0-0eca-4ddf-bdba-56dc9413af27 | organization:owner           | project          | edit.project               | 2022-12-06 15:48:00.225033+00 | 2022-12-06 15:48:00.225033+00 | 
 9ab6f0f2-924e-4e14-9754-8bd7c362509e | organization:editor          | project          | edit.project               | 2022-12-06 15:48:00.302639+00 | 2022-12-06 15:48:00.302639+00 | 
 16aa9afa-dd29-453a-8b4c-f0dbe3b148bd | project:owner                | project          | view.project               | 2022-12-06 15:47:59.443527+00 | 2022-12-06 15:47:59.443527+00 | 
 2dd3f679-d326-4871-b9b6-4eacee560c96 | project:editor               | project          | view.project               | 2022-12-06 15:47:59.534718+00 | 2022-12-06 15:47:59.534718+00 | 
 f1ebf420-fb68-44a3-b7c0-04d5f6b6a4fa | project:viewer               | project          | view.project               | 2022-12-06 15:47:59.620557+00 | 2022-12-06 15:47:59.620557+00 | 
 012bbbd1-1a22-4200-8f90-8c388e06cddc | organization:owner           | project          | view.project               | 2022-12-06 15:47:59.696618+00 | 2022-12-06 15:47:59.696618+00 | 
 3105fbd3-284d-4fec-bb6e-12728e12bd25 | organization:editor          | project          | view.project               | 2022-12-06 15:47:59.770896+00 | 2022-12-06 15:47:59.770896+00 | 
 4a4a4a73-25a9-4a03-9afc-87aaf9bd33a1 | organization:viewer          | project          | view.project               | 2022-12-06 15:47:59.846239+00 | 2022-12-06 15:47:59.846239+00 | 
 d00de717-9fec-4798-b15e-123fbefc7510 | project:owner                | project          | delete.project             | 2022-12-06 15:47:59.925547+00 | 2022-12-06 15:47:59.925547+00 | 
 084c879f-ed1e-4536-8f33-9b1066ddabd9 | organization:owner           | project          | delete.project             | 2022-12-06 15:47:59.996819+00 | 2022-12-06 15:47:59.996819+00 | 
 05d76946-928e-47a4-9b75-6b0173b64a98 | group:manager                | group            | delete.group               | 2022-12-06 15:48:01.048715+00 | 2022-12-06 15:48:01.048715+00 | 
 2046b3db-5228-4e4b-9b90-114409418304 | organization:owner           | group            | delete.group               | 2022-12-06 15:48:01.123014+00 | 2022-12-06 15:48:01.123014+00 | 
 b0745565-d1d4-4de1-85c5-f00e9a92b47e | group:member                 | group            | membership.group           | 2022-12-06 15:48:01.196163+00 | 2022-12-06 15:48:01.196163+00 | 
 bd37f1dc-6872-4538-a941-7432aef26ade | group:manager                | group            | membership.group           | 2022-12-06 15:48:01.270635+00 | 2022-12-06 15:48:01.270635+00 | 
 b1e4041e-5fdc-4d61-9eeb-a03491467a29 | group:manager                | group            | edit.group                 | 2022-12-06 15:48:00.385612+00 | 2022-12-06 15:48:00.385612+00 | 
 31b889fb-802b-4e46-8d9a-8894906784a8 | organization:owner           | group            | edit.group                 | 2022-12-06 15:48:00.45974+00  | 2022-12-06 15:48:00.45974+00  | 
 ceb8fd6a-199c-4442-aaf3-2be68a17bb57 | organization:editor          | group            | edit.group                 | 2022-12-06 15:48:00.533017+00 | 2022-12-06 15:48:00.533017+00 | 
 322a85e1-09a1-451b-bf6e-66a575fdd94a | group:manager                | group            | view.group                 | 2022-12-06 15:48:00.625583+00 | 2022-12-06 15:48:00.625583+00 | 
 0a039ee7-debe-4c08-9373-95724531b5fa | group:member                 | group            | view.group                 | 2022-12-06 15:48:00.733941+00 | 2022-12-06 15:48:00.733941+00 | 
 f11b898b-2ab5-44f4-8d4b-ef19722c35d3 | organization:owner           | group            | view.group                 | 2022-12-06 15:48:00.814642+00 | 2022-12-06 15:48:00.814642+00 | 
 786dbe5f-7bc7-4e5b-b52a-a69780325b2f | organization:editor          | group            | view.group                 | 2022-12-06 15:48:00.893097+00 | 2022-12-06 15:48:00.893097+00 | 
 33b7cae9-55f9-444f-b90d-539d7470fe00 | organization:viewer          | group            | view.group                 | 2022-12-06 15:48:00.971109+00 | 2022-12-06 15:48:00.971109+00 | 
(50 rows)
```