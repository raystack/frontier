package reconcile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/raystack/frontier/internal/bootstrap/schema"
)

// KindPermission is the desired-state document kind for custom permissions.
const KindPermission = "Permission"

// PermissionSpec is one desired permission, identified by its key in
// service.resource.verb form (for example compute.order.get). A permission is
// identity only: it is added or deleted, never updated. Deleting needs the
// explicit flag; a permission that just disappears from the file fails the plan
// instead.
type PermissionSpec struct {
	Key    string `yaml:"key"`
	Delete bool   `yaml:"delete,omitempty"`
}

func (s PermissionSpec) String() string { return s.Key }

// namespaceAndName splits the key into its service/resource namespace and verb.
func (s PermissionSpec) namespaceAndName() (string, string) {
	return schema.PermissionNamespaceAndNameFromKey(s.Key)
}

// isBaseNamespace reports whether a namespace belongs to the base schema,
// which the server manages itself.
func isBaseNamespace(ns string) bool {
	return ns == "app" || strings.HasPrefix(ns, "app/")
}

func validatePermissionSpec(s PermissionSpec) error {
	if strings.TrimSpace(s.Key) == "" {
		return fmt.Errorf("key is required")
	}
	ns, name := s.namespaceAndName()
	if ns == "" || name == "" {
		return fmt.Errorf("invalid key %q (must be in service.resource.verb form)", s.Key)
	}
	if isBaseNamespace(ns) {
		return fmt.Errorf("key %q is part of the base schema, which the server manages", s.Key)
	}
	// One shared check so the reconcile plan and the CreatePermission API agree on
	// what a valid custom permission is: SpiceDB grammar, no reserved verb, and a
	// slug that fits SpiceDB's relation-name limit. This stops a plan that passes and
	// then fails when the schema compiles.
	if err := schema.ValidateCustomPermission(ns, name); err != nil {
		return fmt.Errorf("invalid key %q: %w", s.Key, err)
	}
	return nil
}

// currentPermission is one custom permission as returned by ListPermissions,
// identified by its key.
type currentPermission struct {
	ID  string
	Key string
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

// diffPermissions returns the ops that make the current custom permissions match
// the desired spec. Every custom permission on the server must appear in the
// spec, kept or marked delete, so nothing is ever removed by omission. The key
// is the identity: it is one-to-one with the slug the server enforces, because a
// valid key's namespace parts hold no underscores.
func diffPermissions(desired []PermissionSpec, current []currentPermission) ([]permissionOp, error) {
	byKey := make(map[string]currentPermission, len(current))
	for _, c := range current {
		byKey[c.Key] = c
	}

	seen := map[string]PermissionSpec{}
	var adds, removes []permissionOp
	for _, s := range desired {
		if err := validatePermissionSpec(s); err != nil {
			return nil, fmt.Errorf("invalid permission spec %s: %w", s, err)
		}
		if prev, dup := seen[s.Key]; dup {
			if prev.Delete != s.Delete {
				return nil, fmt.Errorf("permission %s is listed both with and without delete", s)
			}
			continue
		}
		seen[s.Key] = s

		cur, exists := byKey[s.Key]
		switch {
		case s.Delete && exists:
			removes = append(removes, permissionOp{action: opRemove, spec: s, id: cur.ID})
		case !s.Delete && !exists:
			adds = append(adds, permissionOp{action: opAdd, spec: s})
		}
	}

	var unaccounted []string
	for key := range byKey {
		if _, ok := seen[key]; !ok {
			unaccounted = append(unaccounted, key)
		}
	}
	if len(unaccounted) > 0 {
		sort.Strings(unaccounted)
		return nil, fmt.Errorf("permissions exist on the server but are not in the file: %s; keep them or mark them 'delete: true'", strings.Join(unaccounted, ", "))
	}

	return append(adds, removes...), nil
}
