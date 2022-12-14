# Creating an organization

Before creating a new organization, let's create an organization admin user.

## User creation in Shield

```sh
curl --location --request POST 'http://localhost:8000/admin/v1beta1/users'
--header 'Content-Type: application/json'
--header 'X-Shield-Email: admin@odpf.io'
--data-raw '{
    "name": "Shield Org Admin",
    "email": "admin@odpf.io",
    "metadata": {
        "role": "organization admin"
    }
}'
```

Note that this will return an error response

```json
500
{
    "code": 13,
    "message": "internal server error",
    "details": []
}
```

This is because metadata key `role` is not defined in `metadata_keys` table. So, let's first create it.

```sh
curl --location --request POST 'http://localhost:8000/admin/v1beta1/metadatakey'
--header 'Content-Type: application/json'
--data-raw '{
    "key": "role",
    "description": "role of user in organization"
}'
```

```json
200
{
    "metadatakey": {
        "key": "role",
        "description": "role of user in organization"
    }
}
```

Now, we can retry the above user creation request and it should be successful.

```json
200
{
    "user": {
        "id": "2fd7f306-61db-4198-9623-6f5f1809df11",
        "name": "Shield Org Admin",
        "slug": "",
        "email": "admin@odpf.io",
        "metadata": {
            "role": "organization admin"
        },
        "createdAt": "2022-12-07T13:35:19.005545Z",
        "updatedAt": "2022-12-07T13:35:19.005545Z"
    }
}
```

From now onwards, we can use the above user to perform all the admin operations. Let's begin with organization creation.

## Organization creation in Shield

```sh
curl --location --request POST 'http://localhost:8000/admin/v1beta1/organizations'
--header 'Content-Type: application/json'
--header 'X-Shield-Email: admin@odpf.io'
--data-raw '{
    "name": "ODPF",
    "slug": "odpf",
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
        "name": "ODPF",
        "slug": "odpf",
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