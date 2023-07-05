---
sidebar_label: Example
title: Example of Authorization via Shield
---

## Raystack Store

Let's assume Shield contains a user name **John Doe** who creates an organization called **Raystack Store** with a project called **Financials** inside it.

Using the [Create Permissions API](../apis/admin-service-create-permission) a Shield Admin can create the permissions with the following keys:
1. **storage.file.get :** Allows retrieving files from storage.
2. **storage.file.delete :** Allows deleting files from storage.
3. **storage.file.post :** Allows posting or uploading files to storage.

And roles for storage service can be created using [Create Roles API](../apis/admin-service-create-role.api.mdx) or [Create Org Roles](../apis/shield-service-create-organization-role). For example, the following roles can be defined:

1. **storage file owner**: Grants full ownership and control over the files.
2. **storage file reader**: Allows reading or accessing files.

Using the [Create Resource API](../apis/shield-service-create-project-resource) John Doe can create a file in the **Financials** project with resource name as **sensitive-info.txt** which contains Raystack Store's previous quarter financial details.

Assuming in Shield, there also exists a user named **Jane Doe** who is a member of Raystack Store. In case **John** wants his accountant, **Jane** to get access to this file for verification, he can create a Policy in Shield using [Create Policy API](../apis/shield-service-create-policy.api.mdx) with role of a **`storage file reader`** and providing the resource details **`storage/file:92f69c3a-334b-4f25-90b8-4d4f3be6b825`** in format resource's namespace:uuid.

To check if Jane has access to the **sensitive-info.txt** file, you can utilize the Check API and verify if the user has read access to the file or not. This check will confirm whether Jane has the necessary permissions to view the contents of the file.
