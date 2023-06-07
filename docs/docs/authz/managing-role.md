# Managing Role

| Field           | Type     | Description                                                                                                                                                                                                                                                                                         |
| --------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **id**          | uuid     | Unique Role identifier                                                                                                                                                                                                                                                                              |
| **name**        | string   | The name of the role. The name must be unique within the entire Shield instance. The name can contain only alphanumeric characters, dashes and underscores.<br/> _Example:"app_organization_owner"_                                                                                                 |
| **title**       | string   | The title can contain any UTF-8 character, used to provide a human-readable name for the organization. Can also be left empty. <br/>_Example: "Organization Owner"_                                                                                                                                 |
| **permissions** | string[] | List of permission slugs to be assigned to the role <br/> _Example: ["app_organization_administer"]_                                                                                                                                                                                                |
| **metadata**    | object   | Metadata object for organizations that can hold key value pairs defined in Role Metaschema. The default Role Metaschema contains labels and descripton fields. Update the Organization Metaschema to add more fields.<br/>_Example:{"labels": {"key": "value"}, "description": "Role description"}_ |

### List Organization Roles

1. Using Shield CLI
2. Using Shield APIs
3. Admin Poral