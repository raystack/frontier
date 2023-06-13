# Shield as a proxy

Untill now, we were using Shield's admin APIs. Those were responsible for managing Shield's entities. Next, we are use Shield as a proxy.

We had attached the backend service `entropy` to Shield earlier, and now we are going to create a `firehose` resource in it.
Before, going ahead have a look at the configuration file below. Detailed explaining on this configuration file would be in resources/concepts.

```sh
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
```

Let's make the following request.

```sh
curl --location --request POST 'http://localhost:5556/api/firehoses'
--header 'Content-Type: application/json'
--header 'X-Shield-Email: admin@raystack.org'
--header 'X-Shield-Org: 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8'
--header 'X-Shield-Project: 1b89026b-6713-4327-9d7e-ed03345da288'
--header 'X-Shield-Group: 86e2f95d-92c7-4c59-8fed-b7686cccbf4f'
--data-raw '{
    "created_by": "Shield Org Admin",
    "configuration": {
        "SOURCE_KAFKA_CONSUMER_CONFIG_AUTO_COMMIT_ENABLE": false,
        "SOURCE_KAFKA_CONSUMER_CONFIG_FETCH_MIN_BYTES": "1",
        "SOURCE_KAFKA_CONSUMER_CONFIG_MANUAL_COMMIT_MIN_INTERVAL_MS": "-1",
        "SOURCE_KAFKA_CONSUMER_CONFIG_AUTO_OFFSET_RESET": "latest",
        "SINK_TYPE": "log",
        "FILTER_ENGINE": "no_op",
        "RETRY_MAX_ATTEMPTS": "2147483647",
        "LOG_LEVEL": "INFO",
        "INPUT_SCHEMA_PROTO_CLASS": "xxxxx",
        "SOURCE_KAFKA_TOPIC": "delete-me-abcdef",
        "SCHEMA_REGISTRY_STENCIL_URLS": "xxxxx",
        "SOURCE_KAFKA_CONSUMER_CONFIG_MAX_POLL_RECORDS": "1000"
    },
    "replicas": 2,
    "title": "test-firehose-creation-xxxxx",
    "group_id": "5ea18244-8e7a-xxxx-xxxx-ddf4b3fe3698",
    "team": "data_engineering",
    "cluster": "g-xxxxx",
    "stream_name": "g-xxxxx",
    "description": "Creating this firehose for testing purpose.",
    "projectID": "g-xxxxx",
    "orgID": "26ab9a89-de8d-xxxx-xxxx-5ba3f84be7b2",
    "entity": "xxxxx"
}'
```

Now this request will produce a series of events.

- It will hit Shield(proxy) at `/api/firehoses` path, since there are no middleware the request shall be forwarded to the backend.
  We expect that a resource will be created in `entropy` and we'll get a response.
- Now, hooks will be engaged. We only have a single `authz` hook, which creates a resource inside Shield. It will use resource name, org, project and type from either of request, response or as a constant, to create a resource.
- By deafult, no relation is created for this resource, but we can confire this. Here, we have configured to add the group with `owner` role.

We'll get a firehose object sent by `entropy` as a response, though we don't have interest in that.

```json
201
{
    "firehose": {
        "replicas": 2,
        "created_by": "Shield Org Admin",
        "title": "test-firehose-creation-xxxxx",
        "group_id": "5ea18244-8e7a-xxxx-xxxx-ddf4b3fe3698",
        "team": "data_engineering",
        "stream_name": "g-xxxxx",
        "description": "Creating this firehose for testing purpose.",
        "projectID": "g-xxxxx",
        "entity": "xxxxx",
        "environment": "integration",
        "name": "g-xxxxx-firehose-creation-xxxxx-firehose",
        "configuration": {
            "xxxxx": "xxxxx"
        },
        "state": "running",
        "stop_date": null,
        "status": {
            "xxxxx": "xxxxx"
        },
        "pods": ["xxxxx"]
    }
}
```

What we have interest in is going to the resource and relations tables and checking for the entries.

### Resource Table

It got an entry for the resource we just created.

```sh
                                       urn                                       |                             name                             |              project_id              |                org_id                |   namespace_id   |          created_at           |          updated_at           | deleted_at |               user_id                |                  id
---------------------------------------------------------------------------------+--------------------------------------------------------------+--------------------------------------+--------------------------------------+------------------+-------------------------------+-------------------------------+------------+--------------------------------------+--------------------------------------
 r/entropy/firehose/g-xxxxx-firehose-creation-xxxxx-firehose | g-xxxxx-firehose-creation-xxxxx-firehose | 1b89026b-6713-4327-9d7e-ed03345da288 | 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 | entropy/firehose | 2022-12-08 13:25:37.335962+00 | 2022-12-08 13:25:37.335962+00 |            | 2fd7f306-61db-4198-9623-6f5f1809df11 | 28105b9a-1717-47cf-a5d9-49249b6638df
(1 row)
```

### Relations Table

It got an entry for the role `entropy/firehose:owner` for the group `86e2f95d-92c7-4c59-8fed-b7686cccbf4f`.

```sh
                  id                  | subject_namespace_id |              subject_id              | object_namespace_id |              object_id               |        role_id         |          created_at           |          updated_at           | deleted_at
--------------------------------------+----------------------+--------------------------------------+---------------------+--------------------------------------+------------------------+-------------------------------+-------------------------------+------------
 460c44a6-f074-4abe-8f8e-949e7a3f5ec2 | user                 | 2fd7f306-61db-4198-9623-6f5f1809df11 | organization        | 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 | organization:owner     | 2022-12-07 14:10:42.881572+00 | 2022-12-07 14:10:42.881572+00 |
 10797ec9-6744-4064-8408-c0919e71fbca | organization         | 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 | project             | 1b89026b-6713-4327-9d7e-ed03345da288 | project:organization   | 2022-12-07 14:31:46.517828+00 | 2022-12-07 14:31:46.517828+00 |
 29b82d6e-b6fd-4009-9727-1e619c802e23 | organization         | 4eb3c3b4-962b-4b45-b55b-4c07d3810ca8 | group               | 86e2f95d-92c7-4c59-8fed-b7686cccbf4f | group:organization     | 2022-12-07 17:03:59.537254+00 | 2022-12-07 17:03:59.537254+00 |
 0cec1f0a-68ef-4a70-aabd-f3dd1e0eacac | group                | 86e2f95d-92c7-4c59-8fed-b7686cccbf4f | entropy/firehose    | 28105b9a-1717-47cf-a5d9-49249b6638df | entropy/firehose:owner | 2022-12-08 13:25:37.550927+00 | 2022-12-08 13:25:37.550927+00 |
(4 rows)
```
