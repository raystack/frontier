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
// changes. For custom roles, a field that is present is managed and a field
// that is omitted keeps its server value; permissions must be listed.
// Predefined roles converge to their shipped definition instead: an entry
// overrides the fields it lists, an omitted field takes the definition's
// value, and a predefined role absent from the file resets to the definition.
// Predefined roles cannot be deleted — bootstrap recreates a missing one on
// the next boot.
type RoleSpec struct {
	Name        string   `yaml:"name"`
	Title       string   `yaml:"title,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Permissions []string `yaml:"permissions,omitempty"`
	Scopes      []string `yaml:"scopes,omitempty"`
	Delete      bool     `yaml:"delete,omitempty"`
}

// currentRole is one platform role as returned by ListRoles. Description is a
// role metadata value, not a top-level field, so it is read from there.
type currentRole struct {
	ID          string
	Name        string
	Title       string
	Description string
	Permissions []string // slugs
	Scopes      []string
}

// roleOp is a single planned change. For adds and updates, spec carries the
// final values to send (desired fields merged over current ones).
type roleOp struct {
	action opAction
	spec   RoleSpec
	id     string // server role id, set for updates and deletes
	detail string // which fields change, for the plan output
}

func (o roleOp) String() string {
	switch o.action {
	case opRemove:
		return fmt.Sprintf("delete role %s", o.spec.Name)
	case opUpdate:
		return fmt.Sprintf("update role %s (%s)", o.spec.Name, o.detail)
	default:
		return fmt.Sprintf("add role %s", o.spec.Name)
	}
}

// predefinedRole returns the shipped definition of a predefined role name.
func predefinedRole(name string) (schema.RoleDefinition, bool) {
	for _, def := range schema.PredefinedRoles {
		if def.Name == name {
			return def, true
		}
	}
	return schema.RoleDefinition{}, false
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

// validateRoleSpec rejects entries the flow cannot manage. A managed-empty
// permission list is rejected for both role types: an exported file cannot
// represent it (empty lists are omitted), so it would not survive a round trip.
func validateRoleSpec(s RoleSpec, isPredefined bool) error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if s.Delete {
		if isPredefined {
			return fmt.Errorf("cannot delete a predefined role; bootstrap recreates it on the next boot")
		}
		return nil
	}
	if isPredefined {
		if s.Permissions != nil && len(s.Permissions) == 0 {
			return fmt.Errorf("cannot set a predefined role's permissions to an empty list")
		}
		if s.Scopes != nil && len(s.Scopes) == 0 {
			return fmt.Errorf("cannot set a predefined role's scopes to an empty list")
		}
		return nil
	}
	if len(s.Permissions) == 0 {
		return fmt.Errorf("a custom role must list at least one permission")
	}
	return nil
}

// diffRoles returns the ops that make the current platform roles match the
// desired spec. Custom roles are authoritative: every one on the server must
// appear in the spec — kept, or marked delete — so nothing is removed by
// omission. Predefined roles on the server converge to their shipped
// definition with the file entry, if any, laid over it; boot owns creating
// missing ones, so a predefined role absent from the server is not created
// here (and cannot be represented by an export anyway).
func diffRoles(desired []RoleSpec, current []currentRole) ([]roleOp, error) {
	byName := make(map[string]currentRole, len(current))
	for _, c := range current {
		byName[c.Name] = c
	}

	seen := map[string]struct{}{}
	overrides := map[string]RoleSpec{}
	var customs []RoleSpec
	for _, s := range desired {
		_, isPredefined := predefinedRole(s.Name)
		if err := validateRoleSpec(s, isPredefined); err != nil {
			return nil, fmt.Errorf("invalid role spec %q: %w", s.Name, err)
		}
		if _, dup := seen[s.Name]; dup {
			return nil, fmt.Errorf("role %q is listed more than once", s.Name)
		}
		seen[s.Name] = struct{}{}
		if isPredefined {
			overrides[s.Name] = s
		} else {
			customs = append(customs, s)
		}
	}

	var adds, updates, removes []roleOp

	// Predefined roles, in definition order. A duplicate definition keeps the
	// first one, matching what boot creates.
	seeded := map[string]struct{}{}
	for _, def := range schema.PredefinedRoles {
		if _, dup := seeded[def.Name]; dup {
			continue
		}
		seeded[def.Name] = struct{}{}
		cur, exists := byName[def.Name]
		s, listed := overrides[def.Name]
		if !exists {
			if listed {
				return nil, fmt.Errorf("predefined role %q not found on the server; bootstrap creates it at boot", def.Name)
			}
			continue
		}
		var entry *RoleSpec
		if listed {
			entry = &s
		}
		if op, changed := diffPredefinedRole(def, entry, cur); changed {
			updates = append(updates, op)
		}
	}

	for _, s := range customs {
		cur, exists := byName[s.Name]
		if s.Delete {
			if exists {
				removes = append(removes, roleOp{action: opRemove, spec: s, id: cur.ID})
			}
			continue
		}
		desiredPerms := normalizePermissions(s.Permissions)
		if !exists {
			adds = append(adds, roleOp{action: opAdd, spec: RoleSpec{
				Name:        s.Name,
				Title:       s.Title,
				Description: s.Description,
				Permissions: desiredPerms,
				Scopes:      sortedCopy(s.Scopes),
			}})
			continue
		}
		var changes []string
		merged := RoleSpec{Name: s.Name, Title: cur.Title, Description: cur.Description, Permissions: sortedCopy(cur.Permissions), Scopes: sortedCopy(cur.Scopes)}
		if s.Title != "" && s.Title != cur.Title {
			merged.Title = s.Title
			changes = append(changes, "title")
		}
		if s.Description != "" && s.Description != cur.Description {
			merged.Description = s.Description
			changes = append(changes, "description")
		}
		if !stringSetsEqual(desiredPerms, sortedCopy(cur.Permissions)) {
			merged.Permissions = desiredPerms
			changes = append(changes, "permissions")
		}
		if s.Scopes != nil && !stringSetsEqual(sortedCopy(s.Scopes), sortedCopy(cur.Scopes)) {
			merged.Scopes = sortedCopy(s.Scopes)
			changes = append(changes, "scopes")
		}
		if len(changes) > 0 {
			updates = append(updates, roleOp{action: opUpdate, spec: merged, id: cur.ID, detail: strings.Join(changes, ", ")})
		}
	}

	var unaccounted []string
	for name := range byName {
		if _, ok := seen[name]; ok {
			continue
		}
		if _, isPredefined := predefinedRole(name); isPredefined {
			continue
		}
		unaccounted = append(unaccounted, name)
	}
	if len(unaccounted) > 0 {
		sort.Strings(unaccounted)
		return nil, fmt.Errorf("roles exist on the server but are not in the file: %s; keep them or mark them 'delete: true'", strings.Join(unaccounted, ", "))
	}

	ops := append(adds, updates...)
	return append(ops, removes...), nil
}

// diffPredefinedRole compares one predefined role on the server against its
// shipped definition with the file entry, if any, laid over it. An omitted
// field converges to the definition, not the server value. A field whose
// server value is empty converges only when the entry lists it: an empty
// value cannot be written to or read back from a file, so converging it
// unlisted would break the export round trip.
func diffPredefinedRole(def schema.RoleDefinition, s *RoleSpec, cur currentRole) (roleOp, bool) {
	merged := RoleSpec{Name: def.Name, Title: cur.Title, Description: cur.Description, Permissions: sortedCopy(cur.Permissions), Scopes: sortedCopy(cur.Scopes)}
	var changes []string

	title := def.Title
	titleListed := s != nil && s.Title != ""
	if titleListed {
		title = s.Title
	}
	if title != cur.Title && (titleListed || cur.Title != "") {
		merged.Title = title
		changes = append(changes, "title")
	}

	description := def.Description
	descriptionListed := s != nil && s.Description != ""
	if descriptionListed {
		description = s.Description
	}
	if description != cur.Description && (descriptionListed || cur.Description != "") {
		merged.Description = description
		changes = append(changes, "description")
	}

	perms := normalizePermissions(def.Permissions)
	permsListed := s != nil && s.Permissions != nil
	if permsListed {
		perms = normalizePermissions(s.Permissions)
	}
	if !stringSetsEqual(perms, sortedCopy(cur.Permissions)) && (permsListed || len(cur.Permissions) > 0) {
		merged.Permissions = perms
		changes = append(changes, "permissions")
	}

	scopes := sortedCopy(def.Scopes)
	scopesListed := s != nil && s.Scopes != nil
	if scopesListed {
		scopes = sortedCopy(s.Scopes)
	}
	if !stringSetsEqual(scopes, sortedCopy(cur.Scopes)) && (scopesListed || len(cur.Scopes) > 0) {
		merged.Scopes = scopes
		changes = append(changes, "scopes")
	}

	if len(changes) == 0 {
		return roleOp{}, false
	}
	detail := strings.Join(changes, ", ")
	if s == nil {
		detail += "; not in file, reset to default"
	}
	return roleOp{action: opUpdate, spec: merged, id: cur.ID, detail: detail}, true
}
