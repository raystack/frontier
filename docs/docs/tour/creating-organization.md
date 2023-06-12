# Creating an organization

Before creating a new organization, let's create an organization admin user. Note, the metadata in the user body is validated using the default MetaSchemas defined in Shield. These metadata schema validations can always be changed or disabled. For more details read MetaSchema Guides.

## User creation in Shield

```sh
curl --location --request POST 'http://localhost:8000/v1beta1/users'
--header 'Content-Type: application/json'
--header 'X-Shield-Email: admin@raystack.io'
--data-raw '{
    "name": "Shield Org Admin",
    "email": "admin@raystack.io",
    "metadata": {
        "label": {
            "foo":"bar"
        },
        "description":"some user details"
    }
}'
```

Expected response for the user created is of type.

```json
200
{
    "user": {
        "id": "2fd7f306-61db-4198-9623-6f5f1809df11",
        "name": "Shield Org Admin",
        "slug": "admin_raystack_io",
        "email": "admin@raystack.io",
        "metadata": {
            "label": {
                "key":"value"
            },
            "description":"some user details"
        },
        "createdAt": "2022-12-07T13:35:19.005545Z",
        "updatedAt": "2022-12-07T13:35:19.005545Z"
    }
}
```

From now onwards, we can use the above user to perform all the admin operations. Let's begin with organization creation.

## Organization creation in Shield

```sh
curl --location --request POST 'http://localhost:8000/v1beta1/organizations'
--header 'Content-Type: application/json'
--header 'X-Shield-Email: admin@raystack.io'
--data-raw '{
    "name": "Raystack",
    "slug": "raystack",
    "metadata": {
        "description": "Open DataOps Foundation"
    }
}'
```

```json
200
{
    "organization": {
        "id": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8",
        "name": "Raystack",
        "slug": "raystack",
        "metadata": {
            "description": "Open DataOps Foundation"
        },
        "createdAt": "2022-12-07T14:10:42.755848Z",
        "updatedAt": "2022-12-07T14:10:42.755848Z"
    }
}
```

Now, let's have a look at relations table where an `organization:owner` relationship is created.

```sh
                  id                  | subject_namespace_id |              subject_id              | object_namespace_id |              object_id               |      role_id       |          created_at           |          updated_at           | deleted_at
--------------------------------------+----------------------+--------------------------------------+---------------------+--------------------------------------+--------------------+-------------------------------+-------------------------------+------------
 460c44a6-f074-4abe-8f8e-949e7a3f5ec2 | user                 | 2fd7f306-61db-4198-9623-6f5f1809df11 | organization        | 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 | organization:owner | 2022-12-07 14:10:42.881572+00 | 2022-12-07 14:10:42.881572+00 |
(1 row)
```
