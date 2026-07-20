# RFC 0001: Declarative management of platform resources

| | |
|---|---|
| **Status** | Accepted. Implemented in the `reconcile` package and the `reconcile` and `export` commands. |
| **Author** | Rohil Surana |
| **Created** | 2026-07-09 |
| **Updated** | 2026-07-20 |

## Summary

Frontier can manage its platform configuration from a file instead of one-off API calls.
You write down what should exist for one kind of resource. `frontier reconcile` compares the
file with the live server, prints a plan, and applies the difference through the admin API.
`frontier export` does the reverse: it prints the live state in the same format, so a file can
be seeded from a running server.

This document sets the rules every kind follows. The kinds are PlatformUser, Permission, Role,
Preference, and Webhook. New kinds plug in under the same rules without changing the commands
or the file format.

## Problem

Platform configuration used to be managed in three ways, and all three had problems.

1. **Superusers came from server config** (`app.admin.users`). Adding one meant a config change
   and a restart. Removing one did not work: the server only promoted, never demoted.
2. **Custom permissions and roles came from a boot-time file** (`resources_config_path`). It
   could add and update, never remove. There was no preview. A bad file stopped the server
   from booting.
3. **Everything else was set by hand** through the admin API. Nothing recorded what should
   exist, so changes drifted with no record.

## Goals

- One reviewable file per kind that says what should exist.
- A plan before every change.
- Removal that works, and that always shows in the plan first.
- Changes that survive restarts and upgrades.
- New kinds that plug in without new commands or a new file format.

## Non-goals

- Managing user data (organizations, projects, groups, users, policies). People manage those
  through the product, and a file would fight them.
- Renaming a role's internal name. The server looks predefined roles up by name in many code
  paths.
- Deciding how a deployment stores its files or triggers runs. This document defines the
  commands and their behavior. Each install wires them into its own process.

## The file

A file holds one or more YAML documents. Each names a version and a kind, and carries a spec:

```yaml
apiVersion: v1
kind: PlatformUser
spec:
  - type: user
    ref: alice@example.org
    relation: admin
```

- A document with no `apiVersion` is read as `v1`. An unknown version is rejected.
- A document with content but no `kind`, or with a missing spec, is rejected. That is usually a
  typo. To mean an empty list, write `spec: []`.
- The whole file is checked before anything applies. Documents then apply in the order they
  appear.

## The commands

- `frontier reconcile -f <file> [--dry-run]` checks the whole file, prints one plan per kind,
  and applies it unless this is a dry run.
- `frontier export <kind>` prints the live state of one kind as a file on stdout.

Both log in as a superuser through the admin API. The bootstrap service account exists so
automation always has one. The `-H` flag passes the token for now. A file or environment input
is planned (see future work).

## The five rules

Every kind, current and future, follows these five rules.

### 1. Scope and identity

A kind manages a set it declares: which entries, and which fields within an entry. Anything
outside that set is invisible. The diff skips it, export leaves it out, and a file entry that
names it is rejected. Values other tools set (webhook headers, extra role metadata, signing
secrets) pass through writes untouched. Every entry has one stable identity, like a name or a
URL. Reconcile never changes an identity. A rename is a delete plus a create.

### 2. The file is the desired state

Each entry is an object or a value. The resource decides which. A kind does not pick from a
menu.

- An **object** is something the flow creates and deletes: a custom role, a permission, a
  webhook. Every object on the server must appear in the file. One that is missing fails the
  plan. Deleting one needs `delete: true`. If the API cannot delete, the flag maps to the
  nearest real action (archive, disable) and the kind says so, or it rejects the flag.
- A **value** always exists and has a default: a platform grant, a preference, a predefined
  role's fields. The file sets it. A value the file leaves out goes back to its default. A
  value has no delete flag. A default must be reachable through the API. If it is not, the plan
  fails for that entry with a clear message.

One field rule covers both. A field you write is the whole value. A field you leave out takes
the default. An empty field (`""` or `[]`) means empty. Nothing merges. For an object the
default is empty, so leaving a field out and writing it empty come to the same thing. They
differ only where the default is not empty: a predefined role's fields, a preference.

A plan never contains a change that cannot apply, and never skips a managed entry in silence.
Drift either shows in the plan or fails it.

### 3. Check the whole file before changing anything

Nothing applies until the file passes: a known version, a registered kind (a missing or
misspelled kind is an error, not a skip), a present spec, correct order, no unknown fields, and
every entry valid. Entry checks must not need the server. A check that needs live state (does
this permission exist, is this trait known) runs during the diff, where it shows in the plan or
fails it.

### 4. Converge, do not transact

Each change is one API call. Adds and updates run before deletes, so a run that stops halfway
never leaves less access than intended. Nothing rolls back. Recovery from any failure is the
same: fix the cause and run again. What already applied shows no change, and the run picks up
the rest. Two runs racing each other, or a run racing a hand change, are last-write-wins, and
the next run corrects it.

### 5. Export is the reverse of reconcile

`frontier export` writes the current state as a file: sorted, and carrying only the fields that
differ from their default. The guarantee is over server states, not files: for every state the
server can reach, reconciling its export must plan zero changes. A check that would reject the
export of a state the server can reach is a bug in the kind.

## The kinds

| Kind | Type | Identity | A missing entry | Delete |
|---|---|---|---|---|
| PlatformUser | value | (principal, relation) | access removed | remove the entry |
| Permission | object | namespace + name | plan fails | `delete: true` |
| Role, custom | object | name | plan fails | `delete: true` |
| Role, predefined | value | name | reset to the shipped definition | not allowed |
| Preference | value | trait name | reset to the trait default | reset, no flag |
| Webhook | object | URL | plan fails | `delete: true` |

Only what is specific to each kind follows.

**PlatformUser.** An entry is `{type, ref, relation}`: a user (by email or id) or a service
user (by id), with relation `admin` or `member`. The file is the full access list, so a missing
entry removes that access. A user with both relations has two entries. Adding a user by an
email that does not exist creates the user. The bootstrap service account is server-managed:
the flow skips it on the server side and rejects it in the file.

**Permission.** An entry is `{namespace, name}`, for example `compute/order` + `get`. A
permission is an identity only, so it is created or deleted, never updated. Creating one also
updates the authorization schema, so it is usable in roles right away, on every pod at once. The
base `app` namespaces are server-managed.

**Role.** An entry is `{name, title, description, permissions, scopes, delete}`. The name is the
identity. Custom roles are objects. Predefined roles are values whose default is the shipped
definition. Both run through the one field rule: a field you write replaces, a field you leave
out takes the default. So an omitted field clears a custom role and returns a predefined role to
its definition. A predefined role you do not list at all resets fully. Predefined roles cannot
be created or deleted here; boot owns that. One caveat: narrowing a predefined role only
restricts principals whose access comes through the role, because the base schema also grants
owners directly. Removing that direct grant is future work.

**Preference.** An entry is `{name, value}`. Values are strings. A preference the file leaves
out resets to its trait default. There is no delete flag, because reset is what removal means
here.

**Webhook.** An entry is `{url, description, subscribed_events, state, delete}`. The URL is the
identity. An empty event set means all events, which is the server default. The signing secret
is server-owned: the server makes it on create and never returns it, so it is never in the
file, a plan, or an export.

## Server-side changes

Two boot behaviors changed to make this flow work.

1. **Superuser promotion from config is gone.** The server no longer reads `app.admin.users`.
   The only config-driven account is the bootstrap service account, which the server guards.
2. **Boot no longer updates existing roles.** It creates a predefined role only when it is
   missing, and never touches an existing one. Updating definitions moves to reconcile: after an
   upgrade, the first run plans the new defaults, a human reads the plan, and applying it
   converges.

### Deprecating the boot-time loader

Keeping `resources_config_path` would give the same objects two owners, so it goes. It is
deprecated, and removed two minor versions later. Migration is one export of permissions and
roles, committed as the desired-state files, then dropping the setting from the server config.

## Key decisions

1. **Object or value, not a menu.** A resource is one or the other, and that alone fixes what a
   missing entry means and whether delete applies. Earlier drafts let each kind pick from three
   behaviors, which was a matrix operators had to memorize.
2. **One field rule.** Omitted means default, empty means empty, written means the whole value.
   The same for every field of every kind. This is what lets the export round trip be
   guaranteed, and it removed a set of per-field special cases the Role kind used to carry.
3. **Convergence, not transactions.** The API cannot group changes into one atomic unit, and
   building that would be a lot of machinery. The file is declarative and applying it is safe to
   repeat, so re-running is the one recovery path for every failure.
4. **Reads and writes use the public API.** The reconciler is an API client like any other, so
   every change is validated, audited, and seen by all pods at once.
5. **Seed files from exports.** Reconciling an export plans no changes, so a kind's first run is
   safe by construction, and an export captures drift a hand-written file would miss.

## Trade-offs

- A predefined role's definition updates arrive on the next reconcile run after an upgrade, not
  at boot. Until someone runs it, the old definitions stay.
- The reset target is compiled into the CLI, so the reconcile image must match the server
  release. Serving definitions from the server would remove this (future work).
- A run can stop between adds and deletes, with no rollback. Adds run first so the failure stays
  on the safe side, and re-running converges.
- The reconciler trusts the file. The dry-run plan is the safety net; the commands do not ask
  for confirmation.

## Future work

- A read-only API that returns the server's own predefined-role definitions, so the reset target
  comes from the running server instead of the CLI's compiled copy. This removes the
  image-version coupling.
- Metaschemas as a kind, once the server stops caching them per pod at boot.
- Passing the auth token without putting it in the process arguments.
- Billing plans as a kind, replacing the boot-time plans loader.
- Removing relation-based ownership from the base schema, so narrowing a predefined role
  restricts everyone who holds it.

## References

- The `reconcile` package and the `reconcile` and `export` commands in this repository.
- Pixxel rollout: `docs/RFC/001-frontier-gitops-reconcile.md` in pixxelhq/IAM.
