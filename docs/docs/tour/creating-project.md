# Creating a project in organization

In this, we will be using the organization id of the organization we just created. Projects in frontier belong to one organization.

```sh
curl --location --request POST 'http://localhost:8000/v1beta1/projects'
--header 'Content-Type: application/json'
--data-raw '{
    "name": "Project Alpha",
    "slug": "project-alpha",
    "metadata": {
        "description": "Project Alpha"
    },
    "orgId": "4eb3c3b4-962b-4b45-b55b-4c07d3810ca8"
}'
```

```json
200
{
    "project": {
        "id": "1b89026b-6713-4327-9d7e-ed03345da288",
        "name": "Project Alpha",
        "slug": "project-alpha",
        "orgId": "",
        "metadata": {
            "description": "Project Alpha"
        },
        "createdAt": "2022-12-07T14:31:46.436081Z",
        "updatedAt": "2022-12-07T14:31:46.436081Z"
    }
}
```

### Relations Table

It got an entry for the role `project:organization` for the organization `4eb3c3b4-962b-4b45-b55b-4c07d3810ca8`.

```sh
                  id                  | subject_namespace_id |              subject_id              | object_namespace_id |              object_id               |        role_id         |          created_at           |          updated_at           | deleted_at
--------------------------------------+----------------------+--------------------------------------+---------------------+--------------------------------------+------------------------+-------------------------------+-------------------------------+------------
 460c44a6-f074-4abe-8f8e-949e7a3f5ec2 | user                 | 2fd7f306-61db-4198-9623-6f5f1809df11 | organization        | 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 | organization:owner     | 2022-12-07 14:10:42.881572+00 | 2022-12-07 14:10:42.881572+00 |
 10797ec9-6744-4064-8408-c0919e71fbca | organization         | 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 | project             | 1b89026b-6713-4327-9d7e-ed03345da288 | project:organization   | 2022-12-07 14:31:46.517828+00 | 2022-12-07 14:31:46.517828+00 |
(2 rows)
```
