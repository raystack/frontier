# Membership Package — Task Breakdown

Parent issue: [#1478](https://github.com/raystack/frontier/issues/1478)
Prior work: [#1459](https://github.com/raystack/frontier/issues/1459) (SetOrganizationMemberRole), [#1461](https://github.com/raystack/frontier/issues/1461) (SetProjectMemberRole + RemoveProjectMember)

---

## Executive Summary

### What we are doing

Introducing a `core/membership` package that becomes the **single owner** of all member access management across organizations, projects, and groups. Today, membership logic (adding members, changing roles, removing members, listing members) is scattered across 5 packages (`organization`, `project`, `group`, `user`, `serviceuser`) with each independently calling `policyService` and `relationService`. The membership package centralizes this into one place with atomic guarantees.

### Why

Frontier manages access through two mechanisms: **policies** (role bindings that create 3 SpiceDB relations internally) and **direct relations** (explicit SpiceDB relations like `org#owner@user`). Today, the code that manages these is scattered across `organization`, `project`, `group`, `user`, and `serviceuser` packages — each with its own way of creating policies, relations, or both. Some create both, some create only one, some forget to clean up on role changes. This scattered surface area is the root cause of leaky relations: there are too many code paths that can create or modify access, and no single place that enforces consistency. The most visible bug: demoting an org owner to viewer replaces the policy but the code path never touches the `org#owner@user` relation, so the user retains owner access.

### What we achieve

- **Bug fix:** Every mutation (add/set/remove) updates both policy and relation through a single code path — no more stale permissions from scattered call sites
- **Single source of truth:** All reads (list members, list resources for a user) go through policies, not SpiceDB relations that can be stale
- **Simpler codebase:** Organization, project, group, user, and serviceuser packages become pure CRUD (create, read, update, delete entities). All access management logic moves into a single `membership` package. One package for entities, one package for access — clear separation.
- **Code consolidation:** ~30 service functions across 5 packages collapse into 10 public membership functions + shared private core
- **Consistent API:** New RPCs for org and group follow the same `(resource_id, principal_id, principal_type, role_id)` pattern that project already has
- **All principal types:** Org and group RPCs gain support for serviceusers and groups as principals (today they only support users)
- **Explicit roles:** `AddOrganizationMember` and `SetGroupMemberRole` require an explicit role (today `AddOrganizationUsers` hardcodes viewer, `AddGroupUsers` hardcodes member)
- **Better audit logging:** All access mutations flow through one package, giving fine-grained control over what gets logged. Today audit records are inconsistently created — some code paths log, some don't. With membership owning all access changes, every add/set/remove can produce a consistent audit trail with uniform event types and metadata

### High-level approach

```
Before:
  handler → orgService.AddMember()     → policyService.Create() + relationService.Create()  [not atomic]
  handler → orgService.SetMemberRole() → policyService.Delete/Create()                      [no relation update]
  handler → orgService.RemoveUsers()   → policyService.Delete() + relationService.Delete()   [not atomic]

After:
  handler → membershipService.AddOrganizationMember()      → replacePolicy() + replaceRelation()  [atomic]
  handler → membershipService.SetOrganizationMemberRole()  → replacePolicy() + replaceRelation()  [atomic]
  handler → membershipService.RemoveOrganizationMember()   → removeAllPolicies() + removeRelations() [atomic]
```

The membership package has **public functions per resource type** (org/project/group) for type-safe, validation-specific entry points, and **private shared core** (`replacePolicy`, `replaceRelation`, `removeAllPolicies`, `removeRelations`) that handles the atomic policy+relation operations.

### What gets added

| Addition | Count |
|---|---|
| New package: `core/membership/` | 1 |
| New public functions (SetRole, Remove, List for org/project/group) | ~10 |
| New proto RPCs (`AddOrganizationMember`, `RemoveOrganizationMember`, `SetGroupMemberRole`, `RemoveGroupMember`) | 4 |

### What gets deleted

| Deletion | Count |
|---|---|
| Service functions removed from `core/organization` | 9 (`AddMember`, `AddUsers`, `SetMemberRole`, `RemoveUsers`, `ListByUser`, + 4 helpers) |
| Service functions removed from `core/project` | 9 (`SetMemberRole`, `RemoveMember`, `ListByUser`, `ListUsers`, `ListServiceUsers`, `ListGroups`, + 3 helpers) |
| Service functions removed from `core/group` | 7 (`AddMember`, `addOwner`, `AddUsers`, `addMemberPolicy`, `RemoveUsers`, `removeUsers`, `ListByUser`) |
| Service functions removed from `core/user` | 3 (`ListByOrg`, `ListByGroup`, `getUserIDsFromPolicies`) |
| Cascade logic removed from `core/deleter` | 1 (`RemoveUsersFromOrg`) |
| Old proto RPCs deprecated then deleted | 4 (`AddOrganizationUsers`, `RemoveOrganizationUser`, `AddGroupUsers`, `RemoveGroupUser`) |
| **Total functions removed** | **~29** |

### Task overview

| # | Task | New proto? | Priority |
|---|---|---|---|
| 1 | `AddOrganizationMember` — create package + org add/set + serviceuser migration | Yes | High |
| 2 | `RemoveOrganizationMember` — org remove + cascade | Yes | High |
| 3 | Project member mutations — rewire existing RPCs | No | High |
| 4 | `SetGroupMemberRole` — group add/set | Yes | Last |
| 5 | `RemoveGroupMember` — group remove | Yes | Last |
| 6 | `ListPrincipalsByResource` — list members of a resource | No | Medium |
| 7 | `ListResourcesByPrincipal` — list resources for a user | No | Medium |

### What changes for the SDK

- **No more policy/relation RPCs for member management:** The SDK stops using `CreatePolicy`, `DeletePolicy`, `ListPolicies`, `CreateRelation`, `DeleteRelation` for adding/removing/changing members. These were low-level building blocks that leaked internal implementation details (policy IDs, relation triples, principal string formatting) to the client. The SDK now works entirely with role-based RPCs: `AddOrganizationMember`, `SetOrganizationMemberRole`, `RemoveOrganizationMember`, `SetProjectMemberRole`, `RemoveProjectMember`, and the new group equivalents.
- **Explicit roles everywhere:** `AddOrganizationMember` requires a `role_id` — no more hidden defaults. The SDK fetches available roles via `ListRoles(scope)` and passes the chosen role UUID. Same pattern already working for project (migrated in #1461).
- **All principal types from one RPC:** The SDK can add users, service users, and groups to any resource using the same RPC with `principal_type`. Today org and group RPCs only accept `user_ids[]`.
- **Simpler error handling:** One RPC call instead of multi-step `list → delete → create` sequences. No more `Promise.allSettled` workarounds in the React SDK for partial failures during role changes.

### What changes for developer experience

- **One package to understand:** A new developer working on member management only needs to read `core/membership/`. Today they'd have to trace through `organization`, `project`, `group`, `user`, `serviceuser`, `deleter`, and handler-level policy calls to understand the full picture.
- **One place to debug:** If a user has unexpected access, you check their policies via the membership package. No need to cross-reference direct SpiceDB relations against policies to find which path is granting access.
- **Hard to get wrong:** Adding member management to a new feature (e.g., supporting service accounts in groups, or adding role-based access to billing accounts) means implementing a few public functions in the membership package following the established pattern. No risk of forgetting the relation, the audit log, or the min-owner check — the pattern is right there.
- **Entity services stay simple:** `core/organization/`, `core/project/`, `core/group/` become straightforward CRUD services. No interleaved access management logic to reason about when modifying entity lifecycle code.

### Vendor portability (SpiceDB → OpenFGA or similar)

Today, SpiceDB specifics are spread across every service that calls `relationService.Create`, `relationService.LookupResources`, `policyService.Create` (which internally creates SpiceDB relations). Swapping SpiceDB for another authorization engine (OpenFGA, Ory Keto, custom) would require touching every call site — org, project, group, user, serviceuser, deleter services plus 6+ handlers.

After this work:
- **Mutations:** All SpiceDB writes for membership go through the membership package's private core (`replacePolicy`, `replaceRelation`). Changing the authorization backend means changing these ~4 private functions. Zero changes to org/project/group services or handlers.
- **Reads:** All membership reads go through `ListPrincipalsByResource` and `ListResourcesByPrincipal`. Swapping the query backend is localized to these 2 functions.
- **Authz checks:** `CheckPermission`/`BatchCheckPermission` are already behind the `relation.AuthzRepository` interface — unrelated to this work, but also a single-point swap.
- **Net effect:** Membership package becomes a clean abstraction boundary. The rest of the codebase talks in terms of "add member with role" / "remove member" / "list members" — it doesn't know or care whether the backing store is SpiceDB, OpenFGA, or a Postgres table.

### What does NOT change

- **Authorization checks:** `CheckPermission` and `BatchCheckPermission` continue to work exactly as they do today — they read from SpiceDB's permission graph, which is still populated by policies (via role bindings) and any remaining direct relations. No change to how the SDK or handlers verify access before an action.
- **SpiceDB schema:** No schema changes in this phase. The `owner`, `member`, `granted`, `role`, `bearer` relations all stay as-is. Schema cleanup (removing direct relations) may happen later depending on how much benefit we see — it's not committed.
- **Role definitions:** `core/role/` service stays independent — creating roles, managing permissions on roles, role-permission relations in SpiceDB are untouched.
- **Hierarchy relations:** `project→org`, `group→org`, `serviceuser→org` (identity link), `invitation→org` — these are entity links, not membership. They stay as direct relation calls in their respective services.
- **Platform relations:** `platform→admin`, `platform→member` for superuser access — stays in user service.
- **Billing, preferences, invitations:** Completely unrelated, no changes.
- **Existing List RPCs:** `ListOrganizationUsers`, `ListProjectUsers`, `ListGroupUsers` etc. — same proto, same response shape. The handlers rewire to call the membership package instead of calling policy/user/project services directly (Tasks 6 and 7).

### Phase 2 (future, if needed)

Once all call sites go through the membership package, removing direct SpiceDB relations becomes a one-file change inside the package — delete the `replaceRelation`/`removeRelations` calls, update SpiceDB schema, run migration to clean stale relations. No external callers change. Whether we do this depends on the benefits observed after Phase 1 stabilizes.

---

## Part 1: Mutations

### Task 1: `AddOrganizationMember` — create membership package + proto + impl + migrate

**Context:** This is the first task — it creates `core/membership/` from scratch. The private shared core (`replacePolicy`, `replaceRelation`, etc.) gets extracted here as it emerges from implementing the first public function. The existing `AddOrganizationUsers` RPC hides the role (hardcodes viewer) and only supports users. The new RPC makes role explicit and supports all principal types.

**Proto (raystack/proton):**

Add new RPC to `raystack/frontier/v1beta1/frontier.proto`:

```protobuf
rpc AddOrganizationMember(AddOrganizationMemberRequest) returns (AddOrganizationMemberResponse);

message AddOrganizationMemberRequest {
  string org_id = 1;
  string principal_id = 2;
  string principal_type = 3;  // app/user, app/serviceuser
  string role_id = 4;
}
message AddOrganizationMemberResponse {}
```

Do NOT modify or delete `AddOrganizationUsers` — it stays for now.

**Package creation (`core/membership/`):**

- `service.go` — Service struct with dependency interfaces:
  - `PolicyService` (Create, List, Delete)
  - `RelationService` (Create, Delete)
  - `RoleService` (Get)
  - `OrgRepository` (GetByID — for validation)
  - `UserService` (GetByID — for audit)
  - `AuditRecordRepository` (Create)
- Private shared core (extracted during implementation, not designed upfront):
  - `replacePolicy(ctx, resourceID, resourceType, principalID, principalType, roleID)` — delete existing policies for principal+resource, create new one
  - `replaceRelation(ctx, resourceID, resourceType, principalID, principalType, relationName)` — delete existing relations, create new one matching role
  - Helper to map roleID to relation name (e.g. `app_organization_owner` → `owner`, `app_organization_viewer` → `member`)
- Wire into bootstrap/DI (`internal/bootstrap/`)

**Public functions to implement:**

`AddOrganizationMember(ctx, orgID, principalID, principalType, roleID) error`:
- Validate role is org-scoped (one of `app_organization_viewer`, `app_organization_manager`, `app_organization_owner`, or custom org role)
- List existing policies for principal+org → if any exist → `ErrAlreadyMember`
- `replacePolicy()` — create policy
- `replaceRelation()` — create matching relation (e.g. `org#member@user` for viewer/manager, `org#owner@user` for owner)
- Create audit record (org member added event)

`SetOrganizationMemberRole(ctx, orgID, principalID, principalType, roleID) error`:
- This is the existing `org.SetMemberRole` logic moved here with enhancements
- Validate role is org-scoped
- List existing policies for principal+org → if none → `ErrNotMember` (must already be a member)
- Validate min owner constraint — if principal is last owner and new role is not owner → `ErrLastOwner`
- `replacePolicy()` — delete old policies + create new one
- `replaceRelation()` — delete old relation + create new one matching role ← **THIS FIXES THE LEAK BUG** (today `org.SetMemberRole` never touches relations)
- Create audit record

**Difference between Add and Set:**
- `Add` → rejects if already a member (`ErrAlreadyMember`)
- `Set` → rejects if NOT a member (`ErrNotMember`), validates min owner constraint

**Call chain — AddOrganizationMember (new RPC):**

```
Before (AddOrganizationUsers):
  handler → orgService.AddUsers(orgID, userIDs[])
    → loop: orgService.AddMember(orgID, "member", principal)
      → policyService.Create(role=viewer)       // hardcoded default
      → relationService.Create(org#member@user)  // separate, not atomic
      → audit

After (AddOrganizationMember):
  handler → membershipService.AddOrganizationMember(orgID, principalID, principalType, roleID)
    → validateOrgRole(roleID)
    → listExistingPolicies(orgID, principalID, principalType)
    → if len > 0 → ErrAlreadyMember
    → s.replacePolicy(orgID, "app/organization", principalID, principalType, roleID)
    → s.replaceRelation(orgID, "app/organization", principalID, principalType, derivedRelation)
    → audit
```

**Call chain — SetOrganizationMemberRole (existing RPC, rewire):**

```
Before:
  handler → orgService.SetMemberRole(orgID, userID, roleID)
    → validateSetMemberRoleRequest()
    → getUserOrgPolicies()
    → if len == 0 → ErrNotMember
    → validateMinOwnerConstraint()
    → replaceUserOrgPolicies()       // policy only, NO relation update ← BUG

After:
  handler → membershipService.SetOrganizationMemberRole(orgID, userID, "app/user", roleID)
    → validateOrgRole(roleID)
    → listExistingPolicies(orgID, principalID, principalType)
    → if len == 0 → ErrNotMember
    → validateMinOwnerConstraint()
    → s.replacePolicy(...)           // delete old + create new
    → s.replaceRelation(...)         // delete old + create new ← FIXES BUG
    → audit
```

**ServiceUser migration (also in this task):**

```
Before (serviceuser.Create):
  → repo.Create(su)
  → relationService.Create(org#member@su)    // no policy!
  → relationService.Create(su#org@org)       // identity link
  → relationService.Create(su#user@creator)  // identity link

After:
  → repo.Create(su)
  → membershipService.AddOrganizationMember(orgID, suID, "app/serviceuser", defaultOrgRole)
  → relationService.Create(su#org@org)       // identity link stays
  → relationService.Create(su#user@creator)  // identity link stays
```

ServiceUser service keeps `relationService` for identity links only.

**Handler wiring:**
- New `AddOrganizationMember` RPC handler → calls `membershipService.AddOrganizationMember`
- Existing `SetOrganizationMemberRole` RPC handler (from #1459) → rewire from `orgService.SetMemberRole` to `membershipService.SetOrganizationMemberRole`
- Existing `AddOrganizationUsers` RPC handler → rewire from `orgService.AddUsers` to loop `membershipService.AddOrganizationMember` with default viewer role (backward compat until deleted)

**Delete from `core/organization/service.go`:**
- `AddMember()` (line ~207) — replaced by `membership.AddOrganizationMember`
- `AddUsers()` (line ~348) — replaced by handler looping `membership.AddOrganizationMember`
- `SetMemberRole()` (line ~361) — replaced by `membership.SetOrganizationMemberRole`
- `validateSetMemberRoleRequest()` — moves into membership
- `getUserOrgPolicies()` — moves into membership
- `validateMinOwnerConstraint()` — moves into membership
- `replaceUserOrgPolicies()` — absorbed by `replacePolicy` private core
- Remove `PolicyService` and `RelationService` interfaces from org service if no other org functions use them (check `Create` org — it creates platform relation + owner policy; these may need to stay or also move)

**Explicit relations that must be managed in every org mutation:**

After upserting/deleting the policy, each mutation must also handle these explicit direct relations. If not cleaned up, the old relation continues granting the old permissions — this is the core bug today (e.g., org creator demoted from owner to viewer still retains owner access via stale `org#owner@user` relation).

| Role | Explicit relation to create | On role change / add, must first delete |
|---|---|---|
| `app_organization_owner` | `org#owner@principal` | Any existing `org#owner` AND `org#member` for this principal |
| `app_organization_manager` | `org#member@principal` | Any existing `org#owner` AND `org#member` for this principal |
| `app_organization_viewer` | `org#member@principal` | Any existing `org#owner` AND `org#member` for this principal |

On remove: delete ALL explicit relations (`org#owner` and `org#member`) for the principal.

ServiceUser special case: `serviceuser.Create()` today creates `org#member@serviceuser` with NO policy. After migration, `AddOrganizationMember` creates both the policy and the relation.

**Schema constants to be aware of:**
- `schema.RoleOrganizationOwner` = `app_organization_owner`
- `schema.RoleOrganizationManager` = `app_organization_manager`
- `schema.RoleOrganizationViewer` = `app_organization_viewer`
- `schema.OwnerRelationName` = `owner`
- `schema.MemberRelationName` = `member`
- `schema.OrganizationNamespace` = `app/organization`
- Org constants: `AdminRole = schema.RoleOrganizationOwner`, `MemberRole = schema.RoleOrganizationViewer`

**Tests:**
- `AddOrganizationMember`: happy path, already-member error, invalid role error
- `SetOrganizationMemberRole`: happy path, not-member error, last-owner error, relation cleanup
- Verify the bug fix: after `SetOrganizationMemberRole` from owner→viewer, the old `org#owner@user` relation must be gone
- ServiceUser: after creation, serviceuser appears in policy-based listing (has policy now)

---

### Task 2: `RemoveOrganizationMember` — proto + impl + migrate

**Context:** Current `RemoveOrganizationUser` RPC goes through `deleterService.RemoveUsersFromOrg` which does a cascade: removes user from all org projects, all org groups, then removes org-level policies + org relation. This cascade logic either moves into membership or deleter calls membership at each level. The new RPC supports all principal types, not just users.

**Proto (raystack/proton):**

```protobuf
rpc RemoveOrganizationMember(RemoveOrganizationMemberRequest) returns (RemoveOrganizationMemberResponse);

message RemoveOrganizationMemberRequest {
  string org_id = 1;
  string principal_id = 2;
  string principal_type = 3;
}
message RemoveOrganizationMemberResponse {}
```

Do NOT modify or delete `RemoveOrganizationUser` — it stays for now.

**Public function — `RemoveOrganizationMember(ctx, orgID, principalID, principalType) error`:**
- List existing policies for principal+org → if none → `ErrNotMember`
- Validate min owner constraint — can't remove the last owner
- **Cascade:** list org's projects → call `s.removeAllPolicies` for each project where principal has policies
- **Cascade:** list org's groups → call `s.removeAllPolicies` + `s.removeRelations` for each group where principal has policies
- `removeAllPolicies(orgID, "app/organization", principalID, principalType)` — delete org-level policies
- `removeRelations(orgID, "app/organization", principalID, principalType)` — delete org-level relations
- Create audit record (org member removed event)

**Private shared core (new in this task):**
- `removeAllPolicies(ctx, resourceID, resourceType, principalID, principalType)` — find and delete all policies for principal+resource
- `removeRelations(ctx, resourceID, resourceType, principalID, principalType)` — delete all relations for principal+resource

**Decision to make during implementation — cascade approach:**
- Option A: `membership.RemoveOrganizationMember` does cascade internally (calls its own `RemoveProjectMember`/`RemoveGroupMember`)
- Option B: keep cascade in deleter, deleter calls membership at each level
- Option A is cleaner if Tasks 3-5 are done. Option B works regardless of task order. Pick based on what's implemented when you get here.

**Call chain:**

```
Before (RemoveOrganizationUser):
  handler → deleterService.RemoveUsersFromOrg(orgID, [userID])
    → list ALL user policies across all resources
    → for each policy:
        project-level? delete policy if belongs to org's projects
        group-level? groupService.RemoveUsers() if belongs to org's groups
        resource-level? delete policy if belongs to org's project resources
    → orgService.RemoveUsers(orgID, [userID])
      → policyService.List(orgID, userID) + Delete per policy
      → relationService.Delete(org#*@user)
      → audit

After (RemoveOrganizationMember):
  handler → membershipService.RemoveOrganizationMember(orgID, principalID, principalType)
    → validate min owner constraint
    → cascade: for each org project → s.removeAllPolicies(projectID, ...)
    → cascade: for each org group → s.removeAllPolicies + s.removeRelations(groupID, ...)
    → s.removeAllPolicies(orgID, "app/organization", principalID, principalType)
    → s.removeRelations(orgID, "app/organization", principalID, principalType)
    → audit
```

**Handler wiring:**
- New `RemoveOrganizationMember` RPC handler → calls `membershipService.RemoveOrganizationMember`
- Existing `RemoveOrganizationUser` RPC handler → rewire from `deleterService.RemoveUsersFromOrg` to `membershipService.RemoveOrganizationMember` (backward compat until deleted)

**Delete from `core/organization/service.go`:**
- `RemoveUsers()` (line ~508) — replaced by membership

**Delete/refactor from `core/deleter/service.go`:**
- `RemoveUsersFromOrg()` (line ~254) — cascade logic absorbed by membership. The `DeleteUser()` method in deleter calls `RemoveUsersFromOrg` — update it to call membership instead.

**Explicit relations to clean up on remove:**

Removing an org member must delete ALL explicit relations for that principal across the cascade:
- Org level: delete `org#owner@principal` and `org#member@principal`
- For each org project: no explicit relations (project is clean) — only policies
- For each org group: delete `group#owner@principal` and `group#member@principal`

If these are not cleaned up, the principal retains access via stale relations even after all policies are deleted.

**Edge cases to handle:**
- Last owner removal → reject
- Principal has policies at org + project + group level → all must be cleaned
- Resource-level policies under org projects (see deleter line ~308) → also clean up
- Partial failure handling: if removing from one project fails, should we abort or continue? Currently deleter uses `errors.Join` and continues.

**Tests:**
- Happy path, last-owner error, not-member error
- Cascade: removing user from org also removes their project and group memberships within that org
- Verify both policies AND relations are cleaned up at every level

---

### Task 3: Project member mutations — impl + migrate (no new proto, RPCs from #1461)

**Context:** Project RPCs (`SetProjectMemberRole`, `RemoveProjectMember`) already exist from #1461 and are already clean — policy only, no direct relations. This task moves the logic from `core/project/service.go` into `core/membership/` for consistency. The private shared core from Task 1 gets reused here. The key project-specific logic is: validate org membership of principal before allowing project access.

**Public function — `SetProjectMemberRole(ctx, projectID, principalID, principalType, roleID) error`:**
- This is the existing `project.SetMemberRole` moved here
- Validate role is project-scoped (`app_project_owner`, `app_project_manager`, `app_project_viewer`, or custom)
- Resolve project → get parent org ID (needs `ProjectRepository.GetByID` as dependency)
- Validate org membership: principal must have at least one policy on the parent org. If not → `ErrNotOrgMember` (this is `validatePrincipal` logic from project service today)
- Validate principal_type is one of `app/user`, `app/serviceuser`, `app/group`
- `replacePolicy()` — reuse private core from Task 1
- No relation ops — project stays clean
- Note: this is an **upsert** — works for both adding a new member and changing an existing member's role (unlike org which has separate Add and Set)

**Public function — `RemoveProjectMember(ctx, projectID, principalID, principalType) error`:**
- This is the existing `project.RemoveMember` moved here
- Validate principal_type
- `removeAllPolicies(projectID, "app/project", principalID, principalType)` — reuse private core
- If nothing removed → `ErrNotMember`

**Call chain — SetProjectMemberRole:**

```
Before:
  handler → projectService.SetMemberRole(projectID, principalID, principalType, roleID)
    → projectService.Get(projectID)                           // fetch project → get orgID
    → validatePrincipal(orgID, principalID, principalType)    // must be org member
    → validateProjectRole(roleID)                             // must be project-scoped
    → policyService.List(projectID, principalID) → Delete old → Create new

After:
  handler → membershipService.SetProjectMemberRole(projectID, principalID, principalType, roleID)
    → validateProjectRole(roleID)
    → resolve project → get orgID
    → validateOrgMembership(orgID, principalID, principalType)
    → s.replacePolicy(projectID, "app/project", principalID, principalType, roleID)
    // no relations — project stays clean
```

**Call chain — RemoveProjectMember:**

```
Before:
  handler → projectService.RemoveMember(projectID, principalID, principalType)
    → projectService.Get(projectID)
    → validate principalType
    → policyService.List → if len == 0 → ErrNotMember
    → policyService.Delete per policy

After:
  handler → membershipService.RemoveProjectMember(projectID, principalID, principalType)
    → validate principalType
    → s.removeAllPolicies(projectID, "app/project", principalID, principalType)
    → if nothing removed → ErrNotMember
```

**Note on org membership validation:**
`SetProjectMemberRole` must verify the principal is an org member. This means membership package needs a way to check "does this principal have any policy on this org?" — which is `listExistingPolicies(orgID, principalID)`. This is the same helper used in Task 1 for `SetOrganizationMemberRole`'s `ErrNotMember` check.

**Handler wiring:**
- `SetProjectMemberRole` RPC handler → rewire from `projectService.SetMemberRole` to `membershipService.SetProjectMemberRole`
- `RemoveProjectMember` RPC handler → rewire from `projectService.RemoveMember` to `membershipService.RemoveProjectMember`

**Delete from `core/project/service.go`:**
- `SetMemberRole()` (line ~365) — moved to membership
- `RemoveMember()` (line ~406) — moved to membership
- `validatePrincipal()` — moves into membership as `validateOrgMembership()`
- `validateProjectRole()` — moves into membership
- Remove `RoleService` interface from project service if unused after deletions
- `PolicyService` and `RelationService` may still be needed for listing functions (Part 2) — keep if so

**Explicit relations: none.**

Project has no explicit direct relations for membership — only policies. No relation cleanup needed on add, role change, or remove. This is the clean model.

**Schema constants:**
- `schema.RoleProjectOwner` = `app_project_owner`
- `schema.RoleProjectManager` = `app_project_manager`
- `schema.RoleProjectViewer` = `app_project_viewer`
- `schema.ProjectNamespace` = `app/project`

**Tests:**
- `SetProjectMemberRole`: happy path (new member), happy path (role change), not-org-member error, invalid role error, invalid principal_type error
- `RemoveProjectMember`: happy path, not-member error

---

### Task 4: `SetGroupMemberRole` — proto + impl + migrate (last priority)

**Context:** Current `AddGroupUsers` RPC has no role param (hardcodes `group_member`), only supports users, and creates policy + relation non-atomically. The group also has a separate `addOwner` internal method with hardcoded `group_owner` role. The new RPC unifies both with an explicit role and all principal types.

**Proto (raystack/proton):**

```protobuf
rpc SetGroupMemberRole(SetGroupMemberRoleRequest) returns (SetGroupMemberRoleResponse);

message SetGroupMemberRoleRequest {
  string group_id = 1;
  string org_id = 2;
  string principal_id = 3;
  string principal_type = 4;
  string role_id = 5;
}
message SetGroupMemberRoleResponse {}
```

Do NOT modify or delete `AddGroupUsers` — it stays for now.

**Public function — `SetGroupMemberRole(ctx, groupID, principalID, principalType, roleID) error`:**
- Validate role is group-scoped (`app_group_owner`, `app_group_member`)
- If principal has existing group policies and is being changed from owner → check min owner constraint
- `replacePolicy()` — reuse private core
- `replaceRelation()` — reuse private core. Relation name depends on role:
  - `app_group_owner` → `group#owner@principal`
  - `app_group_member` → `group#member@principal`
- Note: group `member` relation stays per #1478 (needed for SpiceDB to resolve group-level policies). `owner` relation is being removed in Phase 2 but created in Phase 1 for backward compat.

**Call chain:**

```
Before (AddGroupUsers):
  handler → groupService.AddUsers(groupID, userIDs[])
    → loop: groupService.AddMember(groupID, principal)
      → addMemberPolicy(): policyService.Create(role=group_member)  // hardcoded
      → relationService.Create(group#member@user)                   // separate

Before (group.addOwner — called during group creation):
  → policyService.Create(role=group_owner)
  → relationService.Create(group#owner@user)

After (SetGroupMemberRole):
  handler → membershipService.SetGroupMemberRole(groupID, principalID, principalType, roleID)
    → validateGroupRole(roleID)
    → listExistingPolicies(groupID, principalID, principalType)
    → if changing from owner → validateMinOwnerConstraint()
    → s.replacePolicy(groupID, "app/group", principalID, principalType, roleID)
    → s.replaceRelation(groupID, "app/group", principalID, principalType, derivedRelation)
```

**Note on group creation:** When a group is created, the creator is added as owner via `group.addOwner()`. After migration, group creation should call `membershipService.SetGroupMemberRole(groupID, creatorID, "app/user", groupOwnerRole)`.

**Handler wiring:**
- New `SetGroupMemberRole` RPC handler → calls `membershipService.SetGroupMemberRole`
- Existing `AddGroupUsers` RPC handler → rewire from `groupService.AddUsers` to loop `membershipService.SetGroupMemberRole` with default `group_member` role (backward compat until deleted)
- Group `Create` handler: wherever `addOwner` was called, call membership instead

**Delete from `core/group/service.go`:**
- `AddMember()` (line ~169) — replaced by membership
- `addOwner()` (line ~194) — replaced by membership
- `AddUsers()` (line ~311) — replaced by handler looping membership
- `addMemberPolicy()` — absorbed into membership
- Remove `PolicyService` and `RelationService` interfaces from group service if no other group functions use them

**Explicit relations that must be managed in every group mutation:**

Same pattern as org — after upserting the policy, the matching explicit relation must be created and any old one deleted. If not cleaned up, role changes have no effect.

| Role | Explicit relation to create | On role change / add, must first delete |
|---|---|---|
| `app_group_owner` | `group#owner@principal` | Any existing `group#owner` AND `group#member` for this principal |
| `app_group_member` | `group#member@principal` | Any existing `group#owner` AND `group#member` for this principal |

Note: `group#member` relation is one that stays long-term per #1478 — needed for SpiceDB to resolve group-level policies. `group#owner` relation may be removed in Phase 2.

**Schema constants:**
- `schema.GroupOwnerRole` = `app_group_owner`
- `schema.GroupMemberRole` = `app_group_member`
- `schema.GroupNamespace` = `app/group`

**Tests:**
- Add member with explicit role
- Change role (member→owner, owner→member) — verify old relation is deleted, new one created
- Min owner constraint
- Group creation still assigns owner through membership

---

### Task 5: `RemoveGroupMember` — proto + impl + migrate (last priority)

**Context:** Current `RemoveGroupUser` only supports users. The new RPC supports all principal types and ensures atomic cleanup of both policies and relations.

**Proto (raystack/proton):**

```protobuf
rpc RemoveGroupMember(RemoveGroupMemberRequest) returns (RemoveGroupMemberResponse);

message RemoveGroupMemberRequest {
  string group_id = 1;
  string org_id = 2;
  string principal_id = 3;
  string principal_type = 4;
}
message RemoveGroupMemberResponse {}
```

Do NOT modify or delete `RemoveGroupUser` — it stays for now.

**Public function — `RemoveGroupMember(ctx, groupID, principalID, principalType) error`:**
- List existing policies → if none → `ErrNotMember`
- Validate min owner constraint (can't remove last owner)
- `removeAllPolicies(groupID, "app/group", principalID, principalType)` — reuse private core
- `removeRelations(groupID, "app/group", principalID, principalType)` — reuse private core
- Audit record

**Note:** Min owner check currently happens in the handler (`group.go` line ~424) by calling `userService.ListByGroup(groupID, AdminRole)`. After migration, this moves into the membership function. Membership can check this via `listExistingPolicies` filtered by owner role.

**Call chain:**

```
Before (RemoveGroupUser):
  handler → check min owner count via userService.ListByGroup(groupID, AdminRole)
  handler → groupService.RemoveUsers(groupID, [userID])
    → policyService.List(groupID, userID) + Delete per policy
    → relationService.Delete(group#*@user)
    → audit

After (RemoveGroupMember):
  handler → membershipService.RemoveGroupMember(groupID, principalID, principalType)
    → validate min owner constraint
    → s.removeAllPolicies(groupID, "app/group", principalID, principalType)
    → s.removeRelations(groupID, "app/group", principalID, principalType)
    → audit
```

**Handler wiring:**
- New `RemoveGroupMember` RPC handler → calls `membershipService.RemoveGroupMember`
- Existing `RemoveGroupUser` RPC handler → rewire to call `membershipService.RemoveGroupMember` (backward compat until deleted)
- Remove min owner check from old handler (now in membership)

**Explicit relations to clean up on remove:**

Delete ALL explicit relations for the principal on this group: both `group#owner@principal` and `group#member@principal`. If not cleaned up, the principal retains group access via stale relations even after all policies are deleted.

**Delete from `core/group/service.go`:**
- `RemoveUsers()` (line ~325) — replaced by membership
- `removeUsers()` (private helper) — replaced by membership

**Tests:**
- Happy path, not-member error
- Last owner rejection
- Verify both `group#owner` and `group#member` relations are deleted on remove

---

## Part 2: Listing

### Task 6: `membership.ListPrincipalsByResource` + migrate

**Context:** Today, listing members OF a resource is scattered: `user.ListByOrg` (policy-based), `user.ListByGroup` (policy-based), `project.ListUsers/ListServiceUsers/ListGroups` (policy-based). On top of that, 6 handlers directly call `policyService.ListRoles()` to enrich responses with role info — bypassing the service layer. This task consolidates all of it into membership.

**Public function — `ListPrincipalsByResource(ctx, resourceID, resourceType string, filter MemberFilter) ([]Member, error)`:**

```go
type MemberFilter struct {
    PrincipalType string   // optional: filter to app/user, app/serviceuser, app/group
    RoleIDs       []string // optional: filter by specific roles
    WithRoles     bool     // enrich response with role details
}

type Member struct {
    PrincipalID   string
    PrincipalType string
    Roles         []role.Role // populated when WithRoles=true
}
```

Internally: `policyService.List(resourceID, resourceType, filter...)` → extract principal IDs by type → optionally enrich with `policyService.ListRoles()`. Returns `[]Member` with principal IDs + roles. Entity hydration (fetching full User/ServiceUser/Group objects) stays in the handler.

**What it replaces — service-level:**

| Current function | File | What it does |
|---|---|---|
| `user.ListByOrg(orgID, roleFilter)` | `core/user/service.go:157` | `policy.List(orgID)` → extract user IDs → `repo.GetByIDs` |
| `user.ListByGroup(groupID, roleFilter)` | `core/user/service.go:214` | `policy.List(groupID)` → extract user IDs → `repo.GetByIDs` |
| `project.ListUsers(projectID, permFilter)` | `core/project/service.go:247` | `policy.List(projectID)` → extract user IDs → `userService.GetByIDs` |
| `project.ListServiceUsers(projectID, permFilter)` | `core/project/service.go` | `policy.List(projectID)` → extract SU IDs → `suserService.GetByIDs` |
| `project.ListGroups(projectID)` | `core/project/service.go` | `policy.List(projectID)` → extract group IDs → `groupService.GetByIDs` |

**What it replaces — handler-level direct policyService calls:**

| Handler | Current direct call | Purpose |
|---|---|---|
| `ListOrganizationUsers` (with role_filters) | `h.policyService.List(ctx, policy.Filter{OrgID, RoleIDs})` | Filter users by role |
| `ListOrganizationUsers` (with_roles) | `h.policyService.ListRoles(ctx, ...)` per user | Enrich user with roles |
| `ListProjectUsers` (with_roles) | `h.policyService.ListRoles(ctx, ...)` per user | Enrich user with roles |
| `ListProjectServiceUsers` (with_roles) | `h.policyService.ListRoles(ctx, ...)` per SU | Enrich SU with roles |
| `ListProjectGroups` (with_roles) | `h.policyService.ListRoles(ctx, ...)` per group | Enrich group with roles |
| `ListGroupUsers` (with_roles) | `h.policyService.ListRoles(ctx, ...)` per user | Enrich user with roles |

All replaced by `membershipService.ListPrincipalsByResource(resourceID, resourceType, {WithRoles: true})`.

**Handler rewiring:**
- `ListOrganizationUsers` → `membershipService.ListPrincipalsByResource(orgID, "app/organization", {PrincipalType: "app/user", WithRoles: true, RoleIDs: ...})` then hydrate users
- `ListProjectUsers` → same pattern with `"app/project"`
- `ListProjectServiceUsers` → same pattern with `PrincipalType: "app/serviceuser"`
- `ListProjectGroups` → same pattern with `PrincipalType: "app/group"`
- `ListGroupUsers` → same pattern with `"app/group"`

**Note:** `ListOrganizationServiceUsers` currently uses `serviceUserService.List(orgID)` which queries Postgres by `org_id` column — NOT policy-based. Decide: switch to policy-based listing (consistent) or leave as-is (it works because serviceusers always belong to one org). If Task 1 adds a policy for serviceuser→org membership, then policy-based listing will work for serviceusers too.

**Delete:**
- `user.ListByOrg()` (`core/user/service.go:157`) — replaced
- `user.ListByGroup()` (`core/user/service.go:214`) — replaced
- `user.getUserIDsFromPolicies()` (helper) — replaced
- `project.ListUsers()` (`core/project/service.go:247`) — replaced
- `project.ListServiceUsers()` — replaced
- `project.ListGroups()` — replaced
- Remove `PolicyService` interface from user service if no other user functions use it
- Remove handler-level `policyService` dependency from handlers that only used it for ListRoles

**Tests:**
- List users of org with role filter
- List users of project with role enrichment
- List serviceusers of project
- List groups of project
- Empty results when no members

---

### Task 7: `membership.ListResourcesByPrincipal` + migrate

**Context:** Today, `org.ListByUser`, `project.ListByUser`, `group.ListByUser` all use `relation.LookupResources` (SpiceDB) which reads the `membership` permission — this reads stale direct relations, the root cause of the permission bug. This task switches to policy-based listing so the single source of truth is policies.

**Public function — `ListResourcesByPrincipal(ctx, principalID, principalType, resourceType string) ([]string, error)`:**

Internally: `policyService.List({PrincipalID, PrincipalType, ResourceType})` → extract unique resource IDs. Returns list of resource IDs (org IDs, project IDs, or group IDs).

**What it replaces:**

| Current function | File | What it does (buggy) |
|---|---|---|
| `org.ListByUser(principal, filter)` | `core/organization/service.go:314` | `relation.LookupResources("app/organization", principal, "membership")` — reads SpiceDB, includes stale relations |
| `project.ListByUser(principal, filter)` | `core/project/service.go:151` | `relation.LookupResources("app/project", principal, "membership")` — same |
| `group.ListByUser(principal, filter)` | `core/group/service.go:129` | `relation.LookupResources("app/group", principal, "membership")` — same |

**Special cases to handle:**
- **PAT scope:** `org.ListByUser` intersects with `principal.PAT.OrgID`. `project.ListByUser` and `group.ListByUser` call `intersectPATScope`. This logic moves into membership or stays in the caller — decide during implementation.
- **NonInherited filter:** `project.ListByUser` has a `NonInherited` flag that calls `listNonInheritedProjectIDs` to list only directly-added projects (not inherited via org). Policy-based listing naturally handles this since policies are explicit per resource.

**Callers to update:**

These service functions are called by aggregate handlers and other internal callers:
- `ListOrganizationsByUser` / `ListOrganizationsByCurrentUser` RPC handlers
- `ListProjectsByUser` / `ListProjectsByCurrentUser` RPC handlers
- `ListUserGroups` / `ListCurrentUserGroups` RPC handlers
- `SearchUserOrganizations`, `SearchUserProjects` aggregate handlers
- `deleterService.DeleteUser()` calls `orgService.ListByUser` to find user's orgs — update to call membership

**Delete:**
- `org.ListByUser()` (`core/organization/service.go:314`) — replaced
- `project.ListByUser()` (`core/project/service.go:151`) — replaced
- `project.listNonInheritedProjectIDs()` (helper) — moves into membership or removed
- `group.ListByUser()` (`core/group/service.go:129`) — replaced
- Remove `RelationService.LookupResources` from org/project/group service interfaces if no longer needed

**This is the change that makes policies the single source of truth for "what resources does this principal have access to."** After this, stale SpiceDB relations no longer affect listing.

**Tests:**
- List orgs for user — returns only orgs where user has policy
- List projects for user — returns only projects where user has direct policy
- PAT scoping — only returns resources within PAT's org
- Verify the bug fix: user with stale relation but no policy → NOT returned

---

## After all tasks — RPCs to delete

Separate cleanup task or tracked in #1478. Only delete after all consumers (including SDK) have migrated.

| Delete from proto | Replaced by | Safe to delete when |
|---|---|---|
| `AddOrganizationUsers` | `AddOrganizationMember` | All consumers migrated (Task 1) |
| `RemoveOrganizationUser` | `RemoveOrganizationMember` | All consumers migrated (Task 2) |
| `AddGroupUsers` | `SetGroupMemberRole` | All consumers migrated (Task 4) |
| `RemoveGroupUser` | `RemoveGroupMember` | All consumers migrated (Task 5) |

## Service functions deleted after all tasks

| Package | Deleted functions | Replaced by |
|---|---|---|
| `core/organization` | `AddMember`, `AddUsers`, `SetMemberRole`, `RemoveUsers`, `ListByUser`, `validateSetMemberRoleRequest`, `getUserOrgPolicies`, `validateMinOwnerConstraint`, `replaceUserOrgPolicies` | membership package |
| `core/project` | `SetMemberRole`, `RemoveMember`, `ListByUser`, `ListUsers`, `ListServiceUsers`, `ListGroups`, `validatePrincipal`, `validateProjectRole`, `listNonInheritedProjectIDs` | membership package |
| `core/group` | `AddMember`, `addOwner`, `AddUsers`, `addMemberPolicy`, `RemoveUsers`, `removeUsers`, `ListByUser` | membership package |
| `core/user` | `ListByOrg`, `ListByGroup`, `getUserIDsFromPolicies` | membership package |
| `core/deleter` | `RemoveUsersFromOrg` cascade logic | membership package |
