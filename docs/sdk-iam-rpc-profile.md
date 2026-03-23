# SDK IAM RPC Profile

## Overview

This document profiles all RPCs the Frontier SDK calls for identity and access management (IAM) — roles, policies, relations, and resources. It covers both the React SDK (used by Aurora) and the Admin SDK.

**RPC source**: All RPCs come from `@raystack/proton/frontier`, auto-generated from proto definitions. The SDK re-exports them:
- `web/sdk/src/index.ts` → `export * from '@raystack/proton/frontier'`
- `web/sdk/hooks/index.ts` → re-exports `FrontierServiceQueries`

There is no single curated interface in the SDK listing all RPCs — it uses whatever proton generates. `FrontierServiceQueries` is the auto-generated object where each key is an RPC name (e.g., `FrontierServiceQueries.listPolicies`).

> Group/team management is excluded as it is being deprecated.

---

## The Core Pattern: How Role Changes Work Today

Every role change in the SDK — org members, project members, service accounts — follows the same 3-step pattern:

```
1. listPolicies(resource, userId)       → fetch existing policies
2. deletePolicy(policyId) x N           → delete all existing policies
3. createPolicy(roleId, resource, principal) → create new policy with desired role
```

This pattern appears in **5 places** across the SDK:
- Org member role change (`web/sdk/react/views/members/member-columns.tsx`)
- Project member role change — user (`web/sdk/react/views/projects/details/project-member-columns.tsx`)
- Project member role change — group (`web/sdk/react/views/projects/details/project-member-columns.tsx`)
- Service account project access toggle (`web/sdk/react/views/api-keys/details/manage-service-user-projects-dialog.tsx`)
- Admin role assignment (`web/sdk/admin/components/AssignRole.tsx`)

**No SDK component calls `UpdatePolicy`** — because the backend returns `CodeUnimplemented`.

**No "set role" or "set member role" RPC exists** — the SDK is forced into the delete+create pattern.

---

## React SDK — RPC Flows

### Organization Member Management

#### Load Members Page

| Step | RPC | Input | Purpose | File |
|------|-----|-------|---------|------|
| 1 | `batchCheckPermission` | `update` + `invitationcreate` on org | Gate UI actions | `members-page.tsx` |
| 2 | `listOrganizationUsers` | orgId, withRoles: true | Fetch members + their roles | `useOrganizationMembers.ts` |
| 3 | `listRoles` | scopes: [OrganizationNamespace], state: enabled | Available org roles | `useOrganizationMembers.ts` |
| 4 | `listOrganizationInvitations` | orgId | Pending invitations | `useOrganizationMembers.ts` |

Calls 2-4 are aggregated in the `useOrganizationMembers` hook.

#### Invite Member

| Step | RPC | Input | Purpose | File |
|------|-----|-------|---------|------|
| 1 | `listOrganizationRoles` | orgId, scopes: [OrganizationNamespace] | Org-specific roles | `invite-member-dialog.tsx` |
| 2 | `listRoles` | scopes: [OrganizationNamespace] | Global roles | `invite-member-dialog.tsx` |
| 3 | `createOrganizationInvitation` | orgId, userIds[], roleIds[] | Send invitation | `invite-member-dialog.tsx` |

Roles from steps 1 and 2 are merged into a single list for the role selector.

#### Change Member Role

| Step | RPC | Input | Trigger | File |
|------|-----|-------|---------|------|
| 1 | `listPolicies` | orgId, userId | Lazy — on dropdown open | `member-columns.tsx` |
| 2 | `deletePolicy` x N | policyId (per policy) | On role select | `member-columns.tsx` |
| 3 | `createPolicy` | roleId, resource: `app/organization:{orgId}`, principal: `app/user:{userId}` | After deletes | `member-columns.tsx` |

Uses `Promise.allSettled` for deletes — **errors are swallowed** (see bug analysis doc).

#### Remove Member

| Step | RPC | Input | File |
|------|-----|-------|------|
| 1a | `removeOrganizationUser` | orgId, userId | `remove-member-dialog.tsx` (active member) |
| 1b | `deleteOrganizationInvitation` | orgId, invitationId | `remove-member-dialog.tsx` (pending invite) |

`removeOrganizationUser` cascades — deletes all policies across org, projects, groups, and resources.

---

### Project Member Management

#### Load Project Detail

| Step | RPC | Input | Purpose | File |
|------|-----|-------|---------|------|
| 1 | `getProject` | projectId | Project metadata | `project-detail-page.tsx` |
| 2 | `listProjectUsers` | projectId, withRoles: true | Project members + roles | `project-detail-page.tsx` |
| 3 | `listProjectGroups` | projectId, withRoles: true | Project teams + roles | `project-detail-page.tsx` |
| 4 | `listRoles` | scopes: [ProjectNamespace], state: enabled | Available project roles | `project-detail-page.tsx` |

#### Add Project Member

| Step | RPC | Input | Purpose | File |
|------|-----|-------|---------|------|
| 1 | `batchCheckPermission` | `update` on project | Gate add button | `project-members.tsx` |
| 2 | `listOrganizationUsers` | orgId | Populate member picker | `project-members.tsx` |
| 3 | `createPolicyForProject` | projectId, principal: `app/user:{userId}`, roleId: ProjectViewer | Grant access | `project-members.tsx` |

`createPolicyForProject` is a separate RPC from `createPolicy` — it takes `projectId` directly instead of a resource string.

#### Change Project Member Role

| Step | RPC | Input | Trigger | File |
|------|-----|-------|---------|------|
| 1 | `listPolicies` | projectId, userId or groupId | Lazy — on dropdown open | `project-member-columns.tsx` |
| 2 | `deletePolicy` x N | policyId | On role select | `project-member-columns.tsx` |
| 3 | `createPolicy` | roleId, resource: `app/project:{projectId}`, principal: `app/user:{userId}` or `app/group:{groupId}` | After deletes | `project-member-columns.tsx` |

Handles both users and groups in the same dropdown. Principal format differs by type.

#### Remove Project Member

| Step | RPC | Input | File |
|------|-----|-------|------|
| 1 | `listPolicies` | projectId, userId | `remove-project-member-dialog.tsx` |
| 2 | `deletePolicy` x N | policyId | `remove-project-member-dialog.tsx` |

There is **no `removeProjectUser` RPC** — project member removal is done entirely via policy deletion on the client side.

---

### Service Account / API Key Management

#### List Service Accounts

| Step | RPC | Input | File |
|------|-----|-------|------|
| 1 | `batchCheckPermission` | `update` on org | `api-keys-list-page.tsx` |
| 2 | `listOrganizationServiceUsers` | orgId | `api-keys-list-page.tsx` |

#### Create Service Account (multi-step, sequential)

| Step | RPC | Input | Purpose | File |
|------|-----|-------|---------|------|
| 1 | `listOrganizationProjects` | orgId | Populate project selector | `add-service-account-dialog.tsx` |
| 2 | `createServiceUser` | orgId, title | Create service account | `add-service-account-dialog.tsx` |
| 3 | `createPolicyForProject` | projectId, principal: `serviceuser:{id}`, roleId: ProjectOwner | Grant project access | `add-service-account-dialog.tsx` |
| 4 | `createServiceUserToken` | orgId, serviceUserId, title: "Initial Generated Key" | Generate API key | `add-service-account-dialog.tsx` |

Steps 2→3→4 execute sequentially — each depends on the previous result.

#### Service Account Detail

| Step | RPC | Input | File |
|------|-----|-------|------|
| 1 | `getServiceUser` | serviceUserId, orgId | `service-user-detail-page.tsx` |
| 2 | `listServiceUserTokens` | serviceUserId, orgId | `service-user-detail-page.tsx` |

#### Manage Service Account Project Access

| Step | RPC | Input | File |
|------|-----|-------|------|
| 1 | `listOrganizationProjects` | orgId | `manage-service-user-projects-dialog.tsx` |
| 2 | `listServiceUserProjects` | serviceUserId, orgId | `manage-service-user-projects-dialog.tsx` |
| Grant | `createPolicyForProject` | projectId, principal: `serviceuser:{id}`, roleId: ProjectOwner | (on checkbox check) |
| Revoke | `listPolicies` → `deletePolicy` x N | projectId, userId: serviceUserId | (on checkbox uncheck) |

#### Token Operations

| Flow | RPC | File |
|------|-----|------|
| Generate token | `createServiceUserToken` (orgId, serviceUserId, title) | `add-service-user-token.tsx` |
| Revoke token | `deleteServiceUserToken` (serviceUserId, tokenId, orgId) | `delete-service-user-key-dialog.tsx` |

---

## Admin SDK — Additional RPCs

The Admin SDK shares the same `@raystack/proton/frontier` RPCs but also uses admin-specific search RPCs for paginated lists.

### Admin-Specific Queries

| RPC | Purpose | File |
|-----|---------|------|
| `searchOrganizationUsers` (Admin) | Paginated org members with RQL, search, sort | `admin/views/organizations/details/members/` |
| `searchProjectUsers` (Admin) | Paginated project members | `admin/views/organizations/details/projects/members/` |
| `searchOrganizationProjects` (Admin) | Paginated org projects | `admin/views/organizations/details/projects/` |
| `searchOrganizationServiceUsers` (Admin) | Paginated service accounts | `admin/views/organizations/details/apis/` |
| `searchOrganizations` (Admin) | Org list for invite dialogs | `admin/views/users/list/invite-users.tsx` |
| `searchUserOrganizations` (Admin) | User's org memberships | `admin/views/users/details/layout/side-panel.tsx` |

### Admin Role Assignment

**File**: `web/sdk/admin/components/AssignRole.tsx`

Same 3-step pattern but uses `Promise.all` (correct error handling):

| Step | RPC | Input |
|------|-----|-------|
| 1 | `listPolicies` (via createClient) | orgId, userId |
| 2 | `deletePolicy` x N (`Promise.all`) | policyId per removed role |
| 3 | `createPolicy` x N (`Promise.all`) | roleId per assigned role |

Also used for project role assignment in `admin/views/.../projects/members/assign-role.tsx`.

### Admin User/Org State Management

| RPC | Purpose | File |
|-----|---------|------|
| `disableUser` | Block user | `admin/views/users/details/security/block-user.tsx` |
| `enableUser` | Unblock user | `admin/views/users/details/security/block-user.tsx` |
| `disableOrganization` | Block org | `admin/views/organizations/details/security/block-organization.tsx` |
| `enableOrganization` | Unblock org | `admin/views/organizations/details/security/block-organization.tsx` |

### Admin Member Removal

| Flow | RPC | File |
|------|-----|------|
| Remove org member | `removeOrganizationUser` | `admin/views/.../members/remove-member.tsx` |
| Remove project member | `listPolicies` → `deletePolicy` x N | `admin/views/.../projects/members/remove-member.tsx` |

---

## Permission Check Pattern

**RPC**: `batchCheckPermission`
**Hook**: `usePermissions` (`web/sdk/react/hooks/usePermissions.ts`)

Checks multiple permissions in a single call. Used across pages to gate UI controls:

| Page | Permissions Checked | Controls |
|------|--------------------|----------|
| Members | `invitationcreate`, `update` on org | Invite button, three-dots menu |
| Projects list | `projectcreate`, `update` on org | Create project button |
| Project members | `update` on project | Add member button |
| API keys | `update` on org | Service account management |

Checked **once on page load**, never re-evaluated after role changes.

---

## Principal and Resource Formats

### Principal Format

| Context | Format | Example |
|---------|--------|---------|
| Org member | `app/user:{userId}` | `app/user:abc-123` |
| Project member (user) | `app/user:{userId}` | `app/user:abc-123` |
| Project member (group) | `app/group:{groupId}` | `app/group:def-456` |
| Service account | `serviceuser:{serviceUserId}` | `serviceuser:ghi-789` |

### Resource Format

| Context | Format | Example |
|---------|--------|---------|
| Organization | `app/organization:{orgId}` | `app/organization:org-123` |
| Project | `app/project:{projectId}` | `app/project:proj-456` |

---

## Complete RPC Inventory

### Queries (Read)

| RPC | Category | Used By |
|-----|----------|---------|
| `batchCheckPermission` | Permissions | React SDK (4 pages) |
| `listOrganizationUsers` | Members | React SDK, Admin SDK |
| `listOrganizationInvitations` | Invitations | React SDK |
| `listRoles` | Roles | React SDK (5 places), Admin SDK (5 places) |
| `listOrganizationRoles` | Roles | React SDK (1), Admin SDK (2) |
| `listPolicies` | Policies | React SDK (4), Admin SDK (3) |
| `getProject` | Projects | React SDK |
| `listProjectUsers` | Members | React SDK, Admin SDK |
| `listProjectGroups` | Members | React SDK |
| `listOrganizationProjects` | Projects | React SDK (2) |
| `listProjectsByCurrentUser` | Projects | React SDK |
| `listOrganizationServiceUsers` | Service accounts | React SDK |
| `getServiceUser` | Service accounts | React SDK |
| `listServiceUserTokens` | Tokens | React SDK, Admin SDK |
| `listServiceUserProjects` | Service accounts | React SDK, Admin SDK |
| `searchOrganizationUsers` | Members (paginated) | Admin SDK |
| `searchProjectUsers` | Members (paginated) | Admin SDK |
| `searchOrganizationProjects` | Projects (paginated) | Admin SDK |
| `searchOrganizationServiceUsers` | Service accounts (paginated) | Admin SDK |
| `searchOrganizations` | Orgs | Admin SDK |
| `searchUserOrganizations` | User memberships | Admin SDK |

### Mutations (Write)

| RPC | Category | Used By |
|-----|----------|---------|
| `createPolicy` | Role assignment | React SDK (3), Admin SDK (2) |
| `deletePolicy` | Role removal | React SDK (4), Admin SDK (3) |
| `createPolicyForProject` | Project role assignment | React SDK (3), Admin SDK (1) |
| `createOrganizationInvitation` | Invitations | React SDK, Admin SDK |
| `deleteOrganizationInvitation` | Invitations | React SDK |
| `removeOrganizationUser` | Member removal | React SDK, Admin SDK |
| `createServiceUser` | Service accounts | React SDK |
| `createServiceUserToken` | Tokens | React SDK (2) |
| `deleteServiceUserToken` | Tokens | React SDK |
| `disableUser` / `enableUser` | User state | Admin SDK |
| `disableOrganization` / `enableOrganization` | Org state | Admin SDK |

### Not Used by SDK

| RPC | Status |
|-----|--------|
| `UpdatePolicy` | Defined in proto, backend returns `CodeUnimplemented`, SDK never calls it |
| `addOrganizationUsers` | Not called by SDK — members are added via invitations + policy creation |
| `removeProjectUser` | Does not exist — project member removal done via policy deletion |

---

## Known Issues

| Issue | Impact | Details |
|-------|--------|---------|
| `Promise.allSettled` in React SDK role change | Errors silently swallowed | `member-columns.tsx` — admin SDK uses `Promise.all` (correct) |
| Permission check is one-time | Stale UI after role changes | `usePermissions` runs on mount only, never re-evaluated |
| UI gates on `update`, backend on `policymanage` | Three-dots menu visible to unauthorized users | Different permissions granted to different roles |
| No atomic role change RPC | Race conditions, partial failures | SDK forced into delete+create pattern |
| No `removeProjectUser` RPC | Client-side cascading delete | SDK manually does listPolicies + deletePolicy x N |

See `docs/analysis-role-change-bugs.md` for full bug analysis and action items.
