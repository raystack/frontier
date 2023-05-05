import Tabs from "@theme/Tabs";
import TabItem from "@theme/TabItem";

# Managing MetaSchemas

MetaSchemas in Shield are default JSON-schemas designed to validate metadata that is included in the body of a resource. These schemas provide a standard way of describing the expected structure and content of metadata, which can be used to ensure consistency and accuracy of metadata across different resources.

## Why MetaSchemas?

Metadata is an essential component of many resources, including user profiles, organization descriptions, group memberships, and project details in Shield. However, the structure and content of metadata can vary widely between different resources, making it challenging to validate and compare metadata across resources.

Metaschemas address this challenge by providing a standardized way of describing the expected structure and content of metadata. With the help of this, Shield users can ensure that metadata is consistently structured and accurately represents the information it is intended to convey.

## How Metaschemas Work!!

Metaschemas are based on the JSON schema format, typically including properties that describe the expected structure and content of metadata, such as data types, formats, required fields, and allowed values. Additionally it include properties that provide context about the metadata, such as a description of the metadata, a version number, and authorship information.

For the ease of users of Shield, we populate the Shield database with default MetaSchemas for Users, Group, Organisation and Roles in Shield during the database migrations.

One can easily updated these Schemas using the Shield MetaSchema APIs.

## Example MetaSchema

A sample user MetaSchema is given below, and the second tab shows an example user metadata which is valid according to the given example schema.

<Tabs groupId="cli" >
<TabItem value="User MetaSchema" label="User MetaSchema">

```json
{
  "title": "user metadata",
  "description": "JSON-schema for validating user metadata",
  "type": "object",
  "properties": {
    "label": {
      "title": "Label",
      "description": "Additional context about the user",
      "type": "object",
      "properties": {
        "role": {
          "type": "string",
          "title": "Role",
          "description": "User's designation in the org"
        }
      },
    },
    "description": {
      "title": "Description",
      "description": "Some additional information for the user",
      "type": "string"
    }
  },
  "additionalProperties": false
}
```

</TabItem>
<TabItem value="Valid User" label="Valid User">

```json
{
  "name": "John Doe",
  "email": "john.doe@odpf.io",
  "metadata": {
    "label": {
      "role": "team leader"
    },
    "description": "ODPF user description"
  }
}
```

</TabItem>
</Tabs>

**Note:** The default user, organization, group and project MetaSchemas are in this [repository](https://github.com/odpf/shield/tree/feat/json-schema-validation/internal/store/postgres/metaschemas)

## Disabling MetaSchemas

In a Shield instance if one wants to diable the MetaSchema validation in either of users, group, organization or roles metadata. It is recommended to updated the MetaSchema's **`additionalProperties`** field value to **`true`**. Shield provides APIs to manipulate these schemas at the endpoint **`/v1beta1/meta/schemas`**. See the [API Reference](../reference/api.md#default-8) for more details.
