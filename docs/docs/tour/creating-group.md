# Creating a group in organization

In this, we will be using the organization id of the organization we created. Groups in shield belong to one organization.

```sh
curl --location --request POST 'http://localhost:8000/admin/v1beta1/groups'
--header 'Content-Type: application/json'
--data-raw '{
    "name": "Data Streaming",
    "slug": "data-streaming",
    "metadata": {
        "description": "group for users in data streaming domain"
    },
    "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8"
}'
```

```json
200
{
    "group": {
        "id": "86e2f95d-92c7-4c59-8fed-b7686cccbf4f",
        "name": "Data Streaming",
        "slug": "data-streaming",
        "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8",
        "metadata": {
            "description": "group for users in data streaming domain"
        },
        "createdAt": "2022-12-07T17:03:59.456847Z",
        "updatedAt": "2022-12-07T17:03:59.456847Z"
    }
}
```

### Relations Table

It got an entry for the role `group:organization` for the organization `4eb3c3b4-962b-4b45-b55b-4c07d3810ca8`.

```sh
                  id                  | subject_namespace_id |              subject_id              | object_namespace_id |              object_id               |        role_id         |          created_at           |          updated_at           | deleted_at 
--------------------------------------+----------------------+--------------------------------------+---------------------+--------------------------------------+------------------------+-------------------------------+-------------------------------+------------
 460c44a6-f074-4abe-8f8e-949e7a3f5ec2 | user                 | 2fd7f306-61db-4198-9623-6f5f1809df11 | organization        | 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 | organization:owner     | 2022-12-07 14:10:42.881572+00 | 2022-12-07 14:10:42.881572+00 | 
 10797ec9-6744-4064-8408-c0919e71fbca | organization         | 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 | project             | 1b89026b-6713-4327-9d7e-ed03345da288 | project:organization   | 2022-12-07 14:31:46.517828+00 | 2022-12-07 14:31:46.517828+00 | 
 29b82d6e-b6fd-4009-9727-1e619c802e23 | organization         | 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 | group               | 86e2f95d-92c7-4c59-8fed-b7686cccbf4f | group:organization     | 2022-12-07 17:03:59.537254+00 | 2022-12-07 17:03:59.537254+00 |
(3 rows)
```
