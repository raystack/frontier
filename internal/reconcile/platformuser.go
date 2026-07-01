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
// Relation is the platform relation to grant — admin (-> superuser) or member
// (-> check). It is deliberately "relation", not "role": "role" is a distinct
// RBAC concept in Frontier, whereas these are SpiceDB relation names.
type PlatformUserSpec struct {
	Type     string `yaml:"type"`     // "user" | "serviceuser"
	Ref      string `yaml:"ref"`      // email or uuid for a user; id for a service user
	Relation string `yaml:"relation"` // "admin" | "member"
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

// Op is a single planned change to one (principal, relation). "add" grants the
// relation; "remove" takes away just that one relation. Ref holds the desired
// entry's ref for an add (email or id), and the current principal's id for a remove.
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

// principalIDs maps the op's ref onto the request id fields: a user ref fills
// user_id, anything else fills serviceuser_id. Exactly one is non-empty.
func (o Op) principalIDs() (userID, serviceUserID string) {
	if o.Type == principalTypeUser {
		return o.Ref, ""
	}
	return "", o.Ref
}

func validateSpec(s PlatformUserSpec) error {
	switch s.Type {
	case principalTypeUser, principalTypeServiceUser:
	default:
		return fmt.Errorf("invalid type %q (want %q or %q)", s.Type, principalTypeUser, principalTypeServiceUser)
	}
	switch s.Relation {
	case schema.AdminRelationName, schema.MemberRelationName:
	default:
		return fmt.Errorf("invalid relation %q (want %q or %q)", s.Relation, schema.AdminRelationName, schema.MemberRelationName)
	}
	if strings.TrimSpace(s.Ref) == "" {
		return fmt.Errorf("empty ref")
	}
	// The server manages the bootstrap SA (it seeds it at boot and blocks removal),
	// so it must not be reconciled. Reject it on the desired side too, not just skip
	// it on the current side.
	if s.Type == principalTypeServiceUser && strings.TrimSpace(s.Ref) == schema.BootstrapServiceUserID {
		return fmt.Errorf("ref %q is the bootstrap service account, which the server manages and cannot be reconciled", s.Ref)
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

// platformRelationOrder gives operations a stable, fixed order.
var platformRelationOrder = []string{schema.AdminRelationName, schema.MemberRelationName}

// diffPlatformUsers works out the changes needed to make the current platform
// principals match the desired spec, one (principal, relation) at a time. A current
// relation that is no longer wanted is removed. A wanted relation that is missing is
// added. A desired entry that matches no current principal is new, so it is added.
// The output order is fixed.
func diffPlatformUsers(desired []PlatformUserSpec, current []platformPrincipal) ([]Op, error) {
	for _, s := range desired {
		if err := validateSpec(s); err != nil {
			return nil, fmt.Errorf("invalid platform-user spec %+v: %w", s, err)
		}
	}

	// Collect adds and removes separately so adds come first (see the return).
	// Apply runs in order, so doing adds before removes means a relation change
	// (e.g. admin -> member) never leaves someone with no platform access if a
	// later op fails.
	var adds, removes []Op
	matched := make([]bool, len(desired))

	for _, p := range current {
		want := map[string]struct{}{}
		for i, s := range desired {
			if specMatchesPrincipal(s, p) {
				matched[i] = true
				want[s.Relation] = struct{}{}
			}
		}
		for _, rel := range platformRelationOrder {
			_, has := p.Relations[rel]
			_, wanted := want[rel]
			switch {
			case wanted && !has:
				adds = append(adds, Op{Action: opAdd, Type: p.Type, Ref: p.ID, Relation: rel})
			case has && !wanted:
				removes = append(removes, Op{Action: opRemove, Type: p.Type, Ref: p.ID, Relation: rel})
			}
		}
	}

	// add desired entries that match no current platform principal
	seenNewRef := map[string]struct{}{}
	for i, s := range desired {
		if matched[i] {
			continue
		}
		key := s.Type + "\x00" + s.Ref + "\x00" + s.Relation
		if _, dup := seenNewRef[key]; dup {
			continue
		}
		seenNewRef[key] = struct{}{}
		adds = append(adds, Op{Action: opAdd, Type: s.Type, Ref: s.Ref, Relation: s.Relation})
	}

	return append(adds, removes...), nil
}
