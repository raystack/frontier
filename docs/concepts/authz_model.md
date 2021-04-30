# Auth model

Shield uses [casbin](https://casbin.org/docs/en/rbac) for its RBAC policy management, along with a custom JSON-based role manager.

## Model

You can define your policy model in Casbin. In Shield, we have used the RBAC model. RBAC model contains the following sections:

* request\_definition
* policy\_definition
* role\_definition
* policy\_effect
* matchers

Check out Shield's model definition [here](https://github.com/odpf/shield/blob/main/lib/casbin/model.conf)

## request\_definition and policy\_definition

`request_definition` is the definition for the access request, and `policy_definition` defines the meaning of the policy.

Shield's `request_definition` and `policy_definition` use three elements while defining policy.

```text
[request_definition]
r = subject, resource, action

[policy_definition]
p = subject, resource, action
```

* **Subject**: This element contains the users/groups of your app.
* **Resource**: This element contains resources/attributes that need to be protected
* **Action**: This element contains actions/roles that will get assigned to the user/group on the particular resource/attribute.

While authorizing HTTP requests Shield picks three elements: the user who is making the request, the resource, and the action\(read, write, etc\) that the user is performing on the request.

## role\_definition

`role_definition` is the definition for the RBAC role inheritance relations. Casbin supports multiple instances of RBAC systems, e.g., users can have groups and their inheritance relations, and resources can have their groupings and their inheritance relations too.

```text
[role_definition]
g = _, _, _
g2 = _, _, _
g3 = _, _, _
```

### Subject role\_definition

g is the role\_definition for subjects, and it contains user and group mapping. Sample representation in the database:

```text
{"user": "10253480-2ac8-4073-9084-5f059dbf6fae"} | {"group": "80553880-23c8-4073-9094-7f059avf6ftp"} | subject
```

### Resource role\_definition

g2 is the role\_definition for resources, and it contains resource and attributes mapping. Sample representation in the database:

```text
{"urn": "relativity-the-special-general-theory"} | {"group": "80553880-23c8-4073-9094-7f059avf6ftp", "category": "physics"} | resource
```

### Action role\_definition

g3 is the role\_definition for actions, and it contains actions and roles mapping. Sample representation in the database:

```text
{"action": "book.read"} | {"role": "13f87b13-c4c8-4be1-b6d6-a9359bcd05fc"} | action
```

## policy\_effect

\[policy\_effect\] is the definition for the policy effect. It defines whether the access request should be approved if multiple policy rules match the request. Shield evaluates a request to authorized if even one matching policy is found:

```text
e = some(where (p.eft == allow))
```

## matchers

The matchers define how the policy rules are evaluated against the request. You can check Shield matcher [here](https://github.com/odpf/shield/blob/main/lib/casbin/model.conf#L16)

## JSON-based role manager

Shield uses a JSON-based role manager as it uses JSON schema in its policies. Moreover, Shield also only loads a subset of policies while evaluating a request i.e only the policies related to the subject making the request are loaded in memory while authorizing a request.

### Example

Here's an example that shows how the JSON-based role manager works: Let's say you have the following policy:

```text
policy:
{"user": "10253480-2ac8-4073-9084-5f059dbf6fae"} | {"category": "physics"} | {"role": "13f87b13-c4c8-4be1-b6d6-a9359bcd05fc"}

resource definition:
{"urn": "relativity-the-special-general-theory"} | {"group": "80553880-23c8-4073-9094-7f059avf6ftp", "category": "physics"} | resource

action definition:
{"action": "book.read"} | {"role": "13f87b13-c4c8-4be1-b6d6-a9359bcd05fc"} | action
```

So, when a request with the following values is made:

```text
user = "10253480-2ac8-4073-9084-5f059dbf6fae"
resource = "relativity-the-special-general-theory"
action = "book.read"
```

Shield will evaluate the above request to true because `{"user": "10253480-2ac8-4073-9084-5f059dbf6fae"}` has `{"role": "13f87b13-c4c8-4be1-b6d6-a9359bcd05fc"}`, which is mapped with `book.read` action, on `{"category": "physics"}`, which is the attribute of `relativity-the-special-general-theory`

