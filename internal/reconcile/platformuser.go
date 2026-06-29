package reconcile

import (
	"fmt"
	"strings"

	"github.com/raystack/frontier/internal/bootstrap/schema"
)

// KindPlatformUser is the desired-state document kind for platform users.
const KindPlatformUser = "PlatformUser"

const (
	principalTypeUser        = "user"
	principalTypeServiceUser = "serviceuser"
)

// PlatformUserSpec is one desired platform-user entry from the YAML spec.
// Role maps directly to the platform relation (admin -> superuser, member -> check);
// the role label and the relation name are the same string ("admin" / "member").
type PlatformUserSpec struct {
	Type string `yaml:"type"` // "user" | "serviceuser"
	Ref  string `yaml:"ref"`  // email/uuid/slug for a user; id for a service user
	Role string `yaml:"role"` // "admin" | "member"
}

// platformPrincipal is the current state of one platform principal (from ListPlatformUsers),
// with the set of platform relations it currently holds.
type platformPrincipal struct {
	Type      string
	ID        string
	Email     string // users only
	Relations map[string]struct{}
}

// opAction is an apply operation kind.
type opAction string

const (
	opAdd    opAction = "add"
	opRemove opAction = "remove"
)

// Op is a single planned change. Add targets a (ref, relation); remove targets a
// principal by id (current RemovePlatformUser strips all relations — relation-selective
// removal lands with the proton `relation` field, see issue #1710 / task #20).
type Op struct {
	Action   opAction
	Type     string
	Ref      string // ref for add (email/id); principal id for remove
	Relation string // add only
}

func (o Op) String() string {
	if o.Action == opRemove {
		return fmt.Sprintf("remove %s %s (all platform relations)", o.Type, o.Ref)
	}
	return fmt.Sprintf("add %s %s as %s", o.Type, o.Ref, o.Relation)
}

func validateSpec(s PlatformUserSpec) error {
	switch s.Type {
	case principalTypeUser, principalTypeServiceUser:
	default:
		return fmt.Errorf("invalid type %q (want %q or %q)", s.Type, principalTypeUser, principalTypeServiceUser)
	}
	switch s.Role {
	case schema.AdminRelationName, schema.MemberRelationName:
	default:
		return fmt.Errorf("invalid role %q (want %q or %q)", s.Role, schema.AdminRelationName, schema.MemberRelationName)
	}
	if strings.TrimSpace(s.Ref) == "" {
		return fmt.Errorf("empty ref")
	}
	return nil
}

// specMatchesPrincipal reports whether a desired spec refers to a current principal.
// Users match by id or email; service users by id.
func specMatchesPrincipal(s PlatformUserSpec, p platformPrincipal) bool {
	if s.Type != p.Type {
		return false
	}
	if s.Ref == p.ID {
		return true
	}
	return p.Type == principalTypeUser && s.Ref != "" && strings.EqualFold(s.Ref, p.Email)
}

// diffPlatformUsers converges current platform principals to the desired spec.
//
// For each current principal it computes the desired relation set (the specs that match
// it). If that differs from what it currently holds, the principal is reconciled:
// removed, then re-granted each desired relation (the only correct option with today's
// principal-level RemovePlatformUser). Desired entries that match no current principal
// are new and are simply added. The result is deterministic given stable input order.
func diffPlatformUsers(desired []PlatformUserSpec, current []platformPrincipal) ([]Op, error) {
	for _, s := range desired {
		if err := validateSpec(s); err != nil {
			return nil, fmt.Errorf("invalid platform-user spec %+v: %w", s, err)
		}
	}

	var ops []Op
	matched := make([]bool, len(desired))

	// Reconcile each existing principal against its desired relations.
	for _, p := range current {
		want := map[string]struct{}{}
		var addRefs []string // desired relations to (re)grant, in spec order
		for i, s := range desired {
			if specMatchesPrincipal(s, p) {
				matched[i] = true
				if _, dup := want[s.Role]; !dup {
					want[s.Role] = struct{}{}
					addRefs = append(addRefs, s.Role)
				}
			}
		}
		if relationsEqual(p.Relations, want) {
			continue // already converged
		}
		// drift: strip the principal, then re-grant exactly the desired relations.
		ops = append(ops, Op{Action: opRemove, Type: p.Type, Ref: p.ID})
		for _, rel := range addRefs {
			ops = append(ops, Op{Action: opAdd, Type: p.Type, Ref: p.ID, Relation: rel})
		}
	}

	// Add desired entries that don't correspond to any current platform principal.
	seenNewRef := map[string]struct{}{}
	for i, s := range desired {
		if matched[i] {
			continue
		}
		key := s.Type + "\x00" + s.Ref + "\x00" + s.Role
		if _, dup := seenNewRef[key]; dup {
			continue
		}
		seenNewRef[key] = struct{}{}
		ops = append(ops, Op{Action: opAdd, Type: s.Type, Ref: s.Ref, Relation: s.Role})
	}

	return ops, nil
}

func relationsEqual(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}
