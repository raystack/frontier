import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

# Platform Preferences

Platform wide settings can be configured using `Preferences` in Frontier. These preferences are stored in the database and are applied to all the users of the platform. Since these settings are crucial to the platform, only the platform admin can set these settings. In case the preferences are not set, the default values are used.

The settings can be set using the preferences CLI or the API as described below.

Note: The Frontier backend predefines some settings which are used by the platform. These settings are defined in the default traits of Frontier. The platform admin can override these settings by setting the same setting name using the preferences CLI or the API.

For example, the **disable_orgs_on_create** setting is defined in the default traits of Frontier. This setting is used by the platform to make the default state of new organizations in Frontier to be disabled. The platform admin can override this setting by setting the same setting name using the preferences CLI or the API.

## Usage

### Create and Update Preferences

<Tabs groupId="cli">
<TabItem value="API" label="API">

Use this [Create Platform Preference API](https://raystack-frontier.vercel.app/apis/admin-service-create-preferences) to create platform wide preferences.

```bash
$ curl -L -X POST 'http://127.0.0.1:7400/v1beta1/preferences' \
-H 'Content-Type: application/json' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdFVzZXJuYW1lOnRlc3RQYXNzd29yZA==' \
--data-raw '{
  "preferences": [
    {
      "name": "disable_orgs_on_create",
      "value": "true"
    }
  ]
}'
```

</TabItem>
  <TabItem value="CLI" label="CLI" default>

```bash
## Syntax
$ frontier preferences set -n <name> -v <value>
$ frontier preferences set --name <name> --value <value>

## Example
$ frontier preferences set -n disable_orgs_on_create -v true -H X-Frontier-Email:user@raystack.org
```

</TabItem>
</Tabs>

Note: The setting name is case sensitive.

### List Preferences

Once the preferences are set, they can be listed using the following command:

<Tabs groupId="cli">
<TabItem value="API" label="API">

```bash
$ curl -L -X GET 'http://127.0.0.1:7400/v1beta1/preferences' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdFVzZXJuYW1lOnRlc3RQYXNzd29yZA=='
```

</TabItem>

<TabItem value="CLI" label="CLI" default>

```bash
$ frontier preferences list
```

</TabItem>
</Tabs>

### Describe Preferences

The Preference traits are defined in the Frontier backend to be used by the platform. Only these traits can be used to set the platform wide preferences. The platform admin can view the default traits defined in Frontier using the following command:

<Tabs groupId="cli">
<TabItem value="API" label="API">

Use this [API](https://raystack-frontier.vercel.app/apis/frontier-service-describe-preferences) to list the default traits defined in Frontier, the traits with `app/platform` namespace denotes the platform wide traits.

```bash
$ curl -L -X GET 'http://127.0.0.1:7400/v1beta1/preferences/traits' \
-H 'Accept: application/json' \
-H 'Authorization: Basic dGVzdFVzZXI6dGVzdFBhc3M='
```

</TabItem>
  <TabItem value="CLI" label="CLI" default>

```bash
 $  frontier preferences get
```

</TabItem>
</Tabs>

This command will list all the default traits defined in Frontier. The traits with `app/platform` namespace denotes the platform wide traits. Similarly, the traits with `app/user`, `app/organization` and `app/group` namespace denotes the user, organization and group wide traits respectively.

## Platform Traits

All the trait description, default value and usage can be found with the above command. Just a brief description of the platform wide traits is given below. The input hints show the kind of values that can be provided for the trait separated by a comma. The default value is shown in bold.

### Usage and Description of Frontier Traits

| Name                         | Namespace    | Description                                                                                                                                                                       | Input Hints      |
| ---------------------------- | ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------- |
| disable_orgs_on_create       | app/platform | If selected the new orgs created by members will be disabled by default.This can be used to prevent users from accessing the org until they contact the admin and get it enabled. | "**true**,false" |
| disable_orgs_listing         | app/platform | If selected will disallow non-admin APIs to list all organizations on the platform.                                                                                               | "true,**false**" |
| disable_users_listing        | app/platform | If selected will will disallow non-admin APIs to list all users on the platform                                                                                                   | "true,**false**" |
| invite_with_roles            | app/platform | Allow inviting new members with set of role ids. When the invitation is accepted, the user will be added to the org with the roles specified.                                     | "**true**,false" |
| invite_mail_template_subject | app/platform | The subject of the invite mail sent to new members                                                                                                                                |                  |
| invite_mail_template_body    | app/platform | The body of the invite mail sent to new members                                                                                                                                   |                  |
---