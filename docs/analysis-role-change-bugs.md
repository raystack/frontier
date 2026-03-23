# Analysis: Organization Role Change Bugs

## Summary

Role changes in Frontier are performed by deleting all existing policies for a user and creating a new policy with the desired role. This is a non-atomic, multi-RPC flow with no server-side integrity checks and flawed client-side error handling. The result is that organizations can end up with zero owners, demoted users can appear to still have elevated privileges, and errors are silently swallowed.

---

## QA Results

| # | Case | Expected | Result |
|---|------|----------|--------|
| 1 | New org, remove the only owner | Error | Pass |
| 2 | New org, change role of only owner | Error | **Fail** |
| 3 | New org, change only owner to member and remove | Error | **Fail** |
| 4 | Old org, owner → member → owner | Should fail at step 1 | **Fail** |
| 5 | New org, member → admin via curl to gateway | Error | Pass |
| 6 | 2 owners + 1 member: demoted viewer assigns owner | Error | **Fail** |
| 7 | Three dots still visible after demotion + refresh | Should not be visible | **Fail** |

---

## Current State

### Backend: How Role Changes Work

The UI calls two separate RPCs sequentially:

```
1. DeletePolicy(policyId)                        ← for each existing policy
2. CreatePolicy(roleId, resource, principal)      ← with the new role
```

These are independent, non-atomic operations. Each goes through the authorization interceptor individually.

#### Authorization Flow

```
Request → AuthorizationInterceptor → Handler → Service → Repository
              │
              ├── Extracts resource from request
              ├── Queries SpiceDB in real-time (no cache)
              └── Checks "policymanage" permission on the resource
```

- **DeletePolicy**: interceptor fetches the policy, extracts the resource, checks `policymanage`
- **CreatePolicy**: interceptor extracts the resource from the request body, checks `policymanage`
- SpiceDB is queried live for every check — no caching

#### Who Has `policymanage` Permission

From `internal/bootstrap/schema/base_schema.zed`:

```
permission policymanage = platform->superuser
                        + granted->app_organization_administer
                        + granted->app_organization_policymanage
                        + owner
```

| Role | Has `policymanage`? |
|------|:-------------------:|
| Organization Owner | Yes |
| Organization Access Manager | Yes |
| Organization Manager | No |
| Organization Viewer | No |

#### Minimum Owner Check — Where It Exists Today

The "at least one owner" check exists in **only one place**:

`internal/api/v1beta1connect/organization.go:515-524` — `RemoveOrganizationUser` handler:

```go
admins, err := h.userService.ListByOrg(ctx, orgResp.ID, organization.AdminRole)
if len(admins) == 1 && admins[0].ID == request.Msg.GetUserId() {
    return nil, connect.NewError(connect.CodePermissionDenied, ErrMinAdminCount)
}
```

This check does **NOT** exist in:
- `DeletePolicy` handler (`internal/api/v1beta1connect/policy.go:110-141`)
- `CreatePolicy` handler (`internal/api/v1beta1connect/policy.go:21-78`)
- Policy service layer (`core/policy/service.go`)

#### `UpdatePolicy` RPC — Exists but Unimplemented

The proto defines `UpdatePolicy` with `{ id, body: PolicyRequestBody }`, and the connect server has a route for it, but the handler returns `CodeUnimplemented`:

```go
// proto/v1beta1/frontierv1beta1connect/frontier.connect.go:4783
func (UnimplementedFrontierServiceHandler) UpdatePolicy(...) {
    return nil, connect.NewError(connect.CodeUnimplemented, errors.New("...is not implemented"))
}
```

---

### Frontend: How the Members Page Works

#### Page Load — API Calls

When `MembersPage` mounts, it fires 4 API calls in parallel:

```
┌── BatchCheckPermission ─────────────────────────────────────┐
│   Checks for the CURRENT user:                               │
│     "invitationcreate" on app/organization:{orgId}           │
│     "update" on app/organization:{orgId}                     │
│                                                              │
│   Result:                                                    │
│     canCreateInvite → controls "Invite people" button        │
│     canDeleteUser   → controls three-dots menu visibility    │
└──────────────────────────────────────────────────────────────┘

┌── ListOrganizationUsers (withRoles: true) ───────────────────┐
│   Returns: users[] + rolePairs[] (userId → Role[] mapping)    │
└──────────────────────────────────────────────────────────────┘

┌── ListRoles (scopes: ["app/organization"]) ──────────────────┐
│   Returns: all org-level roles (owner, admin, member, etc.)   │
└──────────────────────────────────────────────────────────────┘

┌── ListOrganizationInvitations (if canCreateInvite) ──────────┐
│   Returns: pending invitations                                │
└──────────────────────────────────────────────────────────────┘
```

**File**: `web/sdk/react/views/members/members-page.tsx:38-79`

#### Permission Gating — Three-Dots Menu

The `BatchCheckPermission` call checks the `update` permission (not `policymanage`) on the org. This single boolean gates all member management actions:

```typescript
// members-page.tsx:58-69
const { canDeleteUser } = useMemo(() => ({
    canDeleteUser: shouldShowComponent(
        permissions,
        `${PERMISSIONS.UpdatePermission}::${resource}`
    )
}), [permissions, resource]);
```

```typescript
// member-columns.tsx:224
return canUpdateGroup ? (
    <DropdownMenu> ... </DropdownMenu>     // show three dots
) : null;                                   // show nothing
```

**Two problems**:
1. Checked **once on page load**, never re-evaluated after role changes
2. Checks `update` permission, but the backend gates on `policymanage` — these are different permissions granted to different roles

#### Menu Open — What Happens

When the user clicks the three dots:

```
onOpenChange(true) → refetchPolicies()
                          ↓
                   ListPolicies RPC { orgId, userId: targetMember.id }
                          ↓
                   Returns: policies[] for that member
```

The dropdown shows roles the member doesn't already have (no permission check here):

```typescript
excludedRoles = differenceWith(isEqualById, allOrgRoles, memberCurrentRoles)
```

#### Role Change Click — The `updateRole` Flow

**File**: `web/sdk/react/views/members/member-columns.tsx:177-222`

```
Step 1: Get cached policies from ListPolicies query
        policies = policiesData?.policies || []

Step 2: Delete ALL existing policies
        Promise.allSettled(policies.map(p => deletePolicy(p.id)))
            ├── Backend checks "policymanage" per call
            ├── Each call is independent
            └── Errors are SWALLOWED (console.warn only)

Step 3: Create new policy (ALWAYS runs, even if deletes failed)
        createPolicy({ roleId, resource, principal })
            ├── Backend checks "policymanage"
            ├── onSuccess: refetch() + toast("Member role updated")
            └── onError: toast("Something went wrong")
```

The critical error-swallowing code:

```typescript
const deleteResults = await Promise.allSettled(
    policies.map((p: Policy) => deletePolicy(...))
);

if (deleteErrors.length > 0) {
    console.warn('Some policy deletions failed:', deleteErrors);
    // No early return. No error toast. Continues to createPolicy.
}

await createPolicy(createReq);  // runs regardless
```

For comparison, the admin SDK's `AssignRole.tsx` uses `Promise.all` (correct), which properly propagates errors.

---

## Root Cause Analysis by Bug

### Bugs #2, #3, #4 — Last owner can be demoted

**Root cause**: `DeletePolicy` has no minimum-owner check. The authorization interceptor only checks `policymanage` — it doesn't enforce business rules like "at least one owner must remain." The only place this check exists (`RemoveOrganizationUser`) is a different RPC entirely.

**Bug #3 extra**: After the owner is demoted to member (bug #2), `RemoveOrganizationUser` checks `len(admins) == 1` — but there are now zero admins, so the guard passes and the member can be removed.

### Bug #6 — Demoted user can still assign roles

**Backend**: The authorization is real-time (SpiceDB, no cache). After demotion, the viewer should NOT have `policymanage`. The backend should reject both `DeletePolicy` and `CreatePolicy`.

**Frontend**: Even if the backend rejects the calls, `Promise.allSettled` swallows the `PermissionDenied` error from `deletePolicy`. If `createPolicy` also fails, the outer `catch` fires — but the user may see stale state from the previous render and believe it succeeded.

**Needs verification**: Whether the operation actually persisted or only appeared to succeed in the UI.

### Bug #7 — Three dots visible after demotion + refresh

**Root cause**: The UI gates on `update` permission, but the backend gates policy operations on `policymanage`. After demotion to member, the member role may still have `update` on the org (it's a broader permission), so the three dots appear even though the user can't actually perform policy operations.

---

## Action Items

### ACTION 1: Implement `SetOrganizationMemberRole` RPC (new, recommended)

> **Priority: High | Fixes: #2, #3, #4, #6 (partially)**

Instead of relying on low-level policy RPCs for role changes, introduce a purpose-built RPC that encapsulates the entire operation. Aurora assigns exactly 1 role per user per org — this RPC would enforce that invariant server-side.

**Proto** (in `raystack/proton`):

```protobuf
message SetOrganizationMemberRoleRequest {
    string org_id = 1;
    string user_id = 2;
    string role_id = 3;
}

message SetOrganizationMemberRoleResponse {
    Policy policy = 1;
}
```

**Why this over `UpdatePolicy`**:
- `UpdatePolicy` operates on a policy ID — the caller must first fetch policies, find the right one, then update it. It's still a multi-step flow from the client's perspective.
- `SetOrganizationMemberRole` operates on `(org, user, role)` — the caller says "make this user this role" and the backend handles everything.
- It's the right level of abstraction: the frontend shouldn't need to know about policies at all for this use case.

**Service layer** (`core/organization/service.go` or `core/policy/service.go`):

```go
func (s Service) SetMemberRole(ctx context.Context, orgID, userID, newRoleID string) (policy.Policy, error) {
    // 1. Validate the new role exists and is org-scoped
    newRole, err := s.roleService.Get(ctx, newRoleID)
    if err != nil {
        return policy.Policy{}, err
    }

    // 2. Get existing policies for this user in this org
    existing, err := s.policyService.List(ctx, policy.Filter{
        OrgID:       orgID,
        PrincipalID: userID,
    })
    if err != nil {
        return policy.Policy{}, err
    }

    // 3. Minimum owner check: if demoting from admin role, ensure another admin exists
    for _, p := range existing {
        if isAdminRole(p.RoleID) && !isAdminRole(newRoleID) {
            if err := ensureMinAdmin(ctx, orgID, userID); err != nil {
                return policy.Policy{}, err
            }
        }
    }

    // 4. Delete old policies + SpiceDB relations
    for _, p := range existing {
        if err := s.policyService.Delete(ctx, p.ID); err != nil {
            return policy.Policy{}, err
        }
    }

    // 5. Create new policy + SpiceDB relations
    return s.policyService.Create(ctx, policy.Policy{
        RoleID:       newRoleID,
        ResourceID:   orgID,
        ResourceType: schema.OrganizationNamespace,
        PrincipalID:  userID,
        PrincipalType: schema.UserPrincipal,
    })
}
```

**Authorization**: Check `policymanage` on `app/organization:{orgId}` in the interceptor.

---

### ACTION 2: Implement `UpdatePolicy` as a Fallback Atomic Operation

> **Priority: High | Fixes: #2, #3, #4**

The proto already defines `UpdatePolicy` with `{ id, body: PolicyRequestBody }` but the handler is unimplemented. Even if `SetOrganizationMemberRole` is added, `UpdatePolicy` should be implemented as a safe atomic alternative to delete+create.

**Service layer** (`core/policy/service.go`):

```go
func (s Service) Update(ctx context.Context, id string, newRoleID string) (Policy, error) {
    existing, err := s.repository.Get(ctx, id)
    if err != nil {
        return Policy{}, err
    }

    newRole, err := s.roleService.Get(ctx, newRoleID)
    if err != nil {
        return Policy{}, err
    }

    // Minimum owner check
    if existing.RoleID != newRole.ID {
        if err := s.ensureMinAdmin(ctx, existing); err != nil {
            return Policy{}, err
        }
    }

    // Delete old SpiceDB relations
    if err := s.relationService.Delete(ctx, relation.Relation{
        Object: relation.Object{ID: id, Namespace: schema.RoleBindingNamespace},
    }); err != nil {
        return Policy{}, err
    }

    // Update policy in DB with new role
    existing.RoleID = newRole.ID
    updated, err := s.repository.Upsert(ctx, existing)
    if err != nil {
        return Policy{}, err
    }

    // Create new SpiceDB relations
    if err := s.AssignRole(ctx, updated); err != nil {
        return updated, err
    }

    return updated, nil
}
```

**Files to modify**:
| File | Change |
|------|--------|
| `core/policy/service.go` | Add `Update` method with `ensureMinAdmin` |
| `internal/api/v1beta1connect/policy.go` | Implement `UpdatePolicy` handler |
| `pkg/server/connect_interceptors/authorization.go` | Add authorization entry for `UpdatePolicy` |

---

### ACTION 3: Add Minimum Owner Check to `DeletePolicy`

> **Priority: High | Fixes: #2, #3, #4 (defense in depth)**

Even with `SetOrganizationMemberRole` and `UpdatePolicy`, `DeletePolicy` remains a public RPC. It must not allow deleting the last owner's policy.

**Where**: `core/policy/service.go` — `Delete` method

```go
func (s Service) Delete(ctx context.Context, id string) error {
    existing, err := s.repository.Get(ctx, id)
    if err != nil {
        return err
    }

    // If this policy grants an admin role on an org, check minimum owner count
    if err := s.ensureMinAdmin(ctx, existing); err != nil {
        return err
    }

    // ... existing delete logic (relation delete + repo delete)
}
```

The `ensureMinAdmin` helper:
1. Check if the policy's role is an admin/owner role
2. If yes, count other admin policies on the same org (excluding this one)
3. If this is the last one, return `ErrLastAdmin`

---

### ACTION 4: Fix `RemoveOrganizationUser` Edge Case

> **Priority: Medium | Fixes: #3 edge case**

The current check doesn't handle the case where there are already zero admins:

```go
// Current — misses zero-admin state
if len(admins) == 1 && admins[0].ID == request.Msg.GetUserId() {

// Fixed — catches zero admins and last admin
if len(admins) == 0 {
    return nil, connect.NewError(connect.CodeFailedPrecondition, ErrNoAdmins)
}
if len(admins) == 1 && admins[0].ID == request.Msg.GetUserId() {
    return nil, connect.NewError(connect.CodePermissionDenied, ErrMinAdminCount)
}
```

**Where**: `internal/api/v1beta1connect/organization.go:515-524`

---

### ACTION 5: Frontend — Use `SetOrganizationMemberRole` (or `UpdatePolicy`)

> **Priority: High | Fixes: #6, #7 (partially)**

Replace the delete+create flow with a single RPC call:

```typescript
// member-columns.tsx — replace updateRole()
async function updateRole(role: Role) {
    try {
        await setOrganizationMemberRole({
            orgId: organizationId,
            userId: member?.id,
            roleId: role.id
        });

        refetch();
        toast.success('Member role updated');
    } catch (error: any) {
        toast.error('Something went wrong', {
            description: error?.message || 'Failed to update member role'
        });
    }
}
```

One call. No `Promise.allSettled`. No error swallowing. Backend handles atomicity and validation.

**Interim fix** (if backend isn't ready yet): Replace `Promise.allSettled` with `Promise.all` so errors propagate:

```typescript
await Promise.all(
    policies.map((p: Policy) =>
        deletePolicy(create(DeletePolicyRequestSchema, { id: p.id as string }))
    )
);
```

**Apply the same fix to**:
- `web/sdk/react/views/project/members/project-member-columns.tsx`
- `web/sdk/react/views/teams/members/team-member-columns.tsx`

---

### ACTION 6: Frontend — Fix Permission Check to Use `policymanage`

> **Priority: Medium | Fixes: #7**

The UI gates the three-dots menu on `update` permission, but the backend gates policy operations on `policymanage`. These are different permissions granted to different roles.

```typescript
// members-page.tsx — change this:
{ permission: PERMISSIONS.UpdatePermission, resource }

// to this:
{ permission: PERMISSIONS.PolicyManagePermission, resource }
```

This ensures the menu only appears for users who can actually perform policy operations.

---

### ACTION 7: Frontend — Re-evaluate Permissions After Role Changes

> **Priority: Medium | Fixes: #6, #7**

After any role change, invalidate and refetch the current user's permissions:

```typescript
// In MembersPage: pass permission refetch to columns
const { permissions, refetch: refetchPermissions } = usePermissions(...);

// After successful role change in MembersActions:
onSuccess: () => {
    refetch();              // refetch member list
    refetchPermissions();   // refetch current user's permissions
}
```

---

## Summary of All Files to Modify

### Backend

| File | Change |
|------|--------|
| `core/policy/service.go` | Add `Update` method, add `ensureMinAdmin` helper, add min-owner check to `Delete` |
| `core/organization/service.go` (or new file) | Add `SetMemberRole` method |
| `internal/api/v1beta1connect/policy.go` | Implement `UpdatePolicy` handler |
| `internal/api/v1beta1connect/organization.go` | Add `SetOrganizationMemberRole` handler, fix `RemoveOrganizationUser` edge case |
| `pkg/server/connect_interceptors/authorization.go` | Add authorization entries for `UpdatePolicy` and `SetOrganizationMemberRole` |
| Proto repo (`raystack/proton`) | Add `SetOrganizationMemberRole` RPC definition |

### Frontend

| File | Change |
|------|--------|
| `web/sdk/react/views/members/member-columns.tsx` | Replace delete+create with single RPC, fix error handling |
| `web/sdk/react/views/members/members-page.tsx` | Check `policymanage` instead of `update`, pass permission refetch |
| `web/sdk/react/views/project/members/project-member-columns.tsx` | Same role change + error handling fixes |
| `web/sdk/react/views/teams/members/team-member-columns.tsx` | Same role change + error handling fixes |

---

## QA Test Matrix After Fixes

| # | Case | Expected | Fixed By |
|---|------|----------|----------|
| 1 | Remove only owner from org | Error | Backend: `RemoveOrganizationUser` (already works) |
| 2 | Change role of only owner | Error | ACTION 1/2/3: min-owner check in `SetMemberRole`/`UpdatePolicy`/`DeletePolicy` |
| 3 | Change only owner to member, then remove | Error at role change step | ACTION 1/2/3 + ACTION 4 |
| 4 | Owner → member → owner (last owner) | Error at step 1 | ACTION 1/2/3: min-owner check |
| 5 | Member changes role via API | Error | Backend: authorization interceptor (already works) |
| 6 | Demoted viewer assigns roles | Error from backend, surfaced in UI | ACTION 5: single RPC + proper error handling |
| 7 | Three dots visible after demotion | Not visible after refresh | ACTION 6: check `policymanage` + ACTION 7: re-evaluate permissions |
