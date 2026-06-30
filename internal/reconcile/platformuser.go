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

// Op is a single planned change against one (principal, relation): add ensures the
// relation, remove strips just that relation (relation-selective). Ref is the desired
// entry's ref for add (email/id) and the current principal's id for remove.
type Op struct {
	Action   opAction
	Type     string
	Ref      string
	Relation string
}

func (o Op) String() string {
	if o.Action == opRemove {
		return fmt.Sprintf("remove %s %s (%s)", o.Type, o.Ref, o.Relation)
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

// platformRelationOrder gives operations a stable, deterministic order.
var platformRelationOrder = []string{schema.AdminRelationName, schema.MemberRelationName}

// diffPlatformUsers converges current platform principals to the desired spec at
// the (principal, relation) granularity: each current relation no longer desired is
// removed (relation-selectively), each desired relation not present is added. Desired
// entries matching no current principal are new and are added. Output is deterministic.
func diffPlatformUsers(desired []PlatformUserSpec, current []platformPrincipal) ([]Op, error) {
	for _, s := range desired {
		if err := validateSpec(s); err != nil {
			return nil, fmt.Errorf("invalid platform-user spec %+v: %w", s, err)
		}
	}

	var ops []Op
	matched := make([]bool, len(desired))

	for _, p := range current {
		want := map[string]struct{}{}
		for i, s := range desired {
			if specMatchesPrincipal(s, p) {
				matched[i] = true
				want[s.Role] = struct{}{}
			}
		}
		// remove current relations that are no longer desired
		for _, rel := range platformRelationOrder {
			if _, has := p.Relations[rel]; !has {
				continue
			}
			if _, wanted := want[rel]; !wanted {
				ops = append(ops, Op{Action: opRemove, Type: p.Type, Ref: p.ID, Relation: rel})
			}
		}
		// add desired relations not already held
		for _, rel := range platformRelationOrder {
			if _, wanted := want[rel]; !wanted {
				continue
			}
			if _, has := p.Relations[rel]; !has {
				ops = append(ops, Op{Action: opAdd, Type: p.Type, Ref: p.ID, Relation: rel})
			}
		}
	}

	// add desired entries that match no current platform principal
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
