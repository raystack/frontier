package reconcile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/raystack/frontier/internal/bootstrap/schema"
)

// KindPermission is the desired-state document kind for custom permissions.
const KindPermission = "Permission"

// PermissionSpec is one desired permission. A permission is identity only
// (namespace + name): it is added or deleted, never updated. Deleting needs
// the explicit flag; a permission that just disappears from the file fails
// the plan instead.
type PermissionSpec struct {
	Namespace string `yaml:"namespace"`
	Name      string `yaml:"name"`
	Delete    bool   `yaml:"delete,omitempty"`
}

func (s PermissionSpec) String() string {
	return s.Namespace + ":" + s.Name
}

func (s PermissionSpec) slug() string {
	return schema.FQPermissionNameFromNamespace(s.Namespace, s.Name)
}

// isBaseNamespace reports whether a namespace belongs to the base schema,
// which the server manages itself.
func isBaseNamespace(ns string) bool {
	return ns == "app" || strings.HasPrefix(ns, "app/")
}

func validatePermissionSpec(s PermissionSpec) error {
	if strings.TrimSpace(s.Namespace) == "" || strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("namespace and name are required")
	}
	if !schema.IsValidPermissionName(s.Name) {
		return fmt.Errorf("invalid name %q (alphanumeric only)", s.Name)
	}
	if isBaseNamespace(s.Namespace) {
		return fmt.Errorf("namespace %q is part of the base schema, which the server manages", s.Namespace)
	}
	if parts := strings.Split(s.Namespace, "/"); len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("namespace %q must be in service/resource form", s.Namespace)
	}
	// The slug joins service, resource, and verb with "_", so an underscore inside
	// a namespace part would make two different namespaces flatten to the same slug
	// and be silently treated as one. Require each part to be lowercase alphanumeric
	// so the slug is one-to-one and a plan cannot mistake one namespace for another.
	if !schema.IsValidPermissionNamespace(s.Namespace) {
		return fmt.Errorf("invalid namespace %q (each of service/resource must be lowercase alphanumeric)", s.Namespace)
	}
	return nil
}

// currentPermission is one custom permission as returned by ListPermissions.
type currentPermission struct {
	ID        string
	Namespace string
	Name      string
}

type permissionOp struct {
	action opAction
	spec   PermissionSpec
	id     string // server row id, set for deletes
}

func (o permissionOp) String() string {
	if o.action == opRemove {
		return fmt.Sprintf("delete permission %s", o.spec)
	}
	return fmt.Sprintf("add permission %s", o.spec)
}

// diffPermissions returns the ops that make the current custom permissions
// match the desired spec. Every custom permission on the server must appear in
// the spec — kept, or marked delete — so nothing is ever removed by omission.
func diffPermissions(desired []PermissionSpec, current []currentPermission) ([]permissionOp, error) {
	bySlug := make(map[string]currentPermission, len(current))
	for _, c := range current {
		bySlug[schema.FQPermissionNameFromNamespace(c.Namespace, c.Name)] = c
	}

	// The slug is the identity the server enforces (unique in the database), so
	// the diff is keyed by it. Distinct namespace+name pairs can flatten to the
	// same slug when a namespace part contains underscores; that is a conflict.
	seen := map[string]PermissionSpec{}
	accounted := map[string]struct{}{}
	var adds, removes []permissionOp
	for _, s := range desired {
		if err := validatePermissionSpec(s); err != nil {
			return nil, fmt.Errorf("invalid permission spec %s: %w", s, err)
		}
		slug := s.slug()
		if prev, dup := seen[slug]; dup {
			if prev.Namespace != s.Namespace || prev.Name != s.Name {
				return nil, fmt.Errorf("permissions %s and %s collide on the same slug %q", prev, s, slug)
			}
			if prev.Delete != s.Delete {
				return nil, fmt.Errorf("permission %s is listed both with and without delete", s)
			}
			continue
		}
		seen[slug] = s
		accounted[slug] = struct{}{}

		cur, exists := bySlug[slug]
		switch {
		case s.Delete && exists:
			removes = append(removes, permissionOp{action: opRemove, spec: s, id: cur.ID})
		case !s.Delete && !exists:
			adds = append(adds, permissionOp{action: opAdd, spec: s})
		}
	}

	var unaccounted []string
	for slug, c := range bySlug {
		if _, ok := accounted[slug]; !ok {
			unaccounted = append(unaccounted, c.Namespace+":"+c.Name)
		}
	}
	if len(unaccounted) > 0 {
		sort.Strings(unaccounted)
		return nil, fmt.Errorf("permissions exist on the server but are not in the file: %s; keep them or mark them 'delete: true'", strings.Join(unaccounted, ", "))
	}

	return append(adds, removes...), nil
}
