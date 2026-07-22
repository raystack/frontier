package reconcile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

// KindRole is the desired-state document kind for platform-level roles.
const KindRole = "Role"

const opUpdate opAction = "update"

// RoleSpec is one desired platform-level role. Name is the identity and never
// changes. Every other field uses presence tracking: a pointer is nil when the
// field is omitted, and non-nil when the file lists it (including an explicit
// empty value like `title: ""` or `scopes: []`).
//
// A present field is the whole desired value for that field. An omitted field
// takes the role's default: empty for a custom role, and the shipped
// definition's value for a predefined role. So custom and predefined roles run
// through the same rule; they differ only in their defaults and in that a
// predefined role is never created or deleted by this flow.
type RoleSpec struct {
	Name        string    `yaml:"name"`
	Title       *string   `yaml:"title,omitempty"`
	Description *string   `yaml:"description,omitempty"`
	Permissions *[]string `yaml:"permissions,omitempty"`
	Scopes      *[]string `yaml:"scopes,omitempty"`
	Delete      bool      `yaml:"delete,omitempty"`
}

// roleFields are a role's managed values in normalized form: permissions as
// sorted, deduped slugs and scopes sorted. currentRole, the spec defaults, and
// the desired result all reduce to this one shape, so they compare directly.
type roleFields struct {
	Title       string
	Description string
	Permissions []string // sorted slugs
	Scopes      []string // sorted
}

// currentRole is one platform role as returned by ListRoles. Description is a
// role metadata value, not a top-level field, so it is read from there.
// Metadata holds the role's full current metadata, so an update can merge the
// reconciler-managed keys over it instead of dropping everything else.
type currentRole struct {
	ID          string
	Name        string
	Title       string
	Description string
	Permissions []string // slugs
	Scopes      []string
	Metadata    map[string]any
}

// fields reduces a server role to its normalized managed values.
func (c currentRole) fields() roleFields {
	return roleFields{
		Title:       c.Title,
		Description: c.Description,
		Permissions: normalizePermissions(c.Permissions),
		Scopes:      sortedCopy(c.Scopes),
	}
}

// roleOp is a single planned change. For adds and updates, fields carries the
// resolved values to send. baseMetadata is the role's current metadata for an
// update, so managed keys merge over it.
type roleOp struct {
	action       opAction
	name         string
	fields       roleFields
	id           string         // server role id, set for updates and deletes
	detail       string         // which fields change, for the plan output
	baseMetadata map[string]any // current metadata to merge onto, set for updates
}

func (o roleOp) String() string {
	switch o.action {
	case opRemove:
		return fmt.Sprintf("delete role %s", o.name)
	case opUpdate:
		return fmt.Sprintf("update role %s (%s)", o.name, o.detail)
	default:
		return fmt.Sprintf("add role %s", o.name)
	}
}

// roleDefault returns the default fields for a role name. A predefined role
// defaults to its shipped definition; any other name defaults to empty fields.
// The bool reports whether the name is predefined.
func roleDefault(name string) (roleFields, bool) {
	for _, def := range schema.PredefinedRoles {
		if def.Name == name {
			return roleFields{
				Title:       def.Title,
				Description: def.Description,
				Permissions: normalizePermissions(def.Permissions),
				Scopes:      sortedCopy(def.Scopes),
			}, true
		}
	}
	return roleFields{}, false
}

// resolve lays the spec's present fields over the default. An omitted field
// keeps the default; a present field replaces it, even when it is empty.
func (s RoleSpec) resolve(def roleFields) roleFields {
	want := def
	if s.Title != nil {
		want.Title = *s.Title
	}
	if s.Description != nil {
		want.Description = *s.Description
	}
	if s.Permissions != nil {
		want.Permissions = normalizePermissions(*s.Permissions)
	}
	if s.Scopes != nil {
		want.Scopes = sortedCopy(*s.Scopes)
	}
	return want
}

// diffFields returns the names of the fields that differ between want and have,
// in a stable order. An empty result means the two are already equal.
func diffFields(want, have roleFields) []string {
	var changes []string
	if want.Title != have.Title {
		changes = append(changes, "title")
	}
	if want.Description != have.Description {
		changes = append(changes, "description")
	}
	if !stringSetsEqual(want.Permissions, have.Permissions) {
		changes = append(changes, "permissions")
	}
	if !stringSetsEqual(want.Scopes, have.Scopes) {
		changes = append(changes, "scopes")
	}
	return changes
}

// normalizePermissions maps permission references in any accepted form
// (slug, service/resource:verb, service.resource.verb) to slugs, deduped and
// sorted.
func normalizePermissions(refs []string) []string {
	set := map[string]struct{}{}
	for _, ref := range refs {
		set[permission.ParsePermissionName(ref)] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for slug := range set {
		out = append(out, slug)
	}
	sort.Strings(out)
	return out
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

func stringSetsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// validateRoleSpec rejects entries the flow cannot manage: an entry needs a
// name, a predefined role cannot be deleted because bootstrap recreates it on
// the next boot, and a role must resolve to at least one permission.
func validateRoleSpec(s RoleSpec, isPredefined bool) error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if s.Delete && isPredefined {
		return fmt.Errorf("cannot delete a predefined role; bootstrap recreates it on the next boot")
	}
	// The server rejects a role with no permissions, so a spec that resolves to an
	// empty permission set is a plan that can never apply. Fail here with a clear
	// message instead of late at the API. A custom role defaults to empty, so
	// omitting its permissions resolves to empty; writing `permissions: []` on any
	// role resolves to empty too.
	if !s.Delete {
		def, _ := roleDefault(s.Name)
		if len(s.resolve(def).Permissions) == 0 {
			return fmt.Errorf("must list at least one permission; a role with no permissions cannot be applied")
		}
	}
	return nil
}

// diffRoles returns the ops that make the current platform roles match the
// desired spec. Each listed role's desired fields are its own present fields
// laid over its default (empty for custom, the shipped definition for
// predefined). A custom role missing from the server is created; a predefined
// one missing there is an error, since bootstrap owns creating it. A custom
// role can be deleted; a predefined one cannot. Every server role not in the
// file must be accounted for: a predefined one converges back to its
// definition, and a custom one fails the plan so nothing is removed by omission.
func diffRoles(desired []RoleSpec, current []currentRole) ([]roleOp, error) {
	byName := make(map[string]currentRole, len(current))
	for _, c := range current {
		byName[c.Name] = c
	}

	var adds, updates, removes []roleOp
	listed := map[string]struct{}{}
	for _, s := range desired {
		def, isPredefined := roleDefault(s.Name)
		if err := validateRoleSpec(s, isPredefined); err != nil {
			return nil, fmt.Errorf("invalid role spec %q: %w", s.Name, err)
		}
		if _, dup := listed[s.Name]; dup {
			return nil, fmt.Errorf("role %q is listed more than once", s.Name)
		}
		listed[s.Name] = struct{}{}

		cur, exists := byName[s.Name]
		if s.Delete {
			if exists {
				removes = append(removes, roleOp{action: opRemove, name: s.Name, id: cur.ID})
			}
			continue
		}

		want := s.resolve(def)
		if !exists {
			if isPredefined {
				return nil, fmt.Errorf("predefined role %q not found on the server; bootstrap creates it at boot", s.Name)
			}
			adds = append(adds, roleOp{action: opAdd, name: s.Name, fields: want})
			continue
		}
		if changes := diffFields(want, cur.fields()); len(changes) > 0 {
			updates = append(updates, roleOp{action: opUpdate, name: s.Name, fields: want, id: cur.ID,
				detail: strings.Join(changes, ", "), baseMetadata: cur.Metadata})
		}
	}

	// Every server role not in the file. Predefined roles converge back to their
	// definition; custom roles are unaccounted for and fail the plan.
	var names []string
	for name := range byName {
		if _, ok := listed[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	var unaccounted []string
	for _, name := range names {
		cur := byName[name]
		def, isPredefined := roleDefault(name)
		if !isPredefined {
			unaccounted = append(unaccounted, name)
			continue
		}
		if changes := diffFields(def, cur.fields()); len(changes) > 0 {
			updates = append(updates, roleOp{action: opUpdate, name: name, fields: def, id: cur.ID,
				detail: strings.Join(changes, ", ") + "; not in file, reset to default", baseMetadata: cur.Metadata})
		}
	}
	if len(unaccounted) > 0 {
		return nil, fmt.Errorf("roles exist on the server but are not in the file: %s; keep them or mark them 'delete: true'", strings.Join(unaccounted, ", "))
	}

	ops := append(adds, updates...)
	return append(ops, removes...), nil
}
