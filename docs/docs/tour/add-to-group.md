# Adding to a group

In this part we'll learn to add `members` and `managers` to a group. For this, we'll be using relations API. Also, we have added two new users to shield `john.doe@odpf.io` and  `doe.john@odpf.io`.

## Add a Member to a Group

```sh
curl --location --request POST 'http://localhost:8000/admin/v1beta1/relations' \
--header 'Content-Type: application/json' \
--data-raw '{
"objectId": "86e2f95d-92c7-4c59-8fed-b7686cccbf4f",
  "objectNamespace": "group",
  "subject": "user:doe.john@odpf.io",
  "roleName": "member"
}'
```

```json
{
    "relation": {
        "id": "7cd5d527-6304-4dc7-9e35-4b1a7d3988a0",
        "objectId": "86e2f95d-92c7-4c59-8fed-b7686cccbf4f",
        "objectNamespace": "group",
        "subject": "user:448d52d4-48cb-495e-8ec5-8afc55c624ca",
        "roleName": "group:member",
        "createdAt": null,
        "updatedAt": null
    }
}
```

## Add a Manager to a Group

```sh
curl --location --request POST 'http://localhost:8000/admin/v1beta1/relations' \
--header 'Content-Type: application/json' \
--data-raw '{
"objectId": "86e2f95d-92c7-4c59-8fed-b7686cccbf4f",
  "objectNamespace": "group",
  "subject": "user:doe.john@odpf.io",
  "roleName": "manager"
}'
```

```json
200
{
    "relation": {
        "id": "d8c5d2ca-73db-4185-bed8-c802c212a287",
        "objectId": "86e2f95d-92c7-4c59-8fed-b7686cccbf4f",
        "objectNamespace": "group",
        "subject": "user:448d52d4-48cb-495e-8ec5-8afc55c624ca",
        "roleName": "group:manager",
        "createdAt": null,
        "updatedAt": null
    }
}
```