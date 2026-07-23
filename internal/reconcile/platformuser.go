package reconcile

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/google/uuid"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

// KindPlatformUser is the desired-state document kind for platform users.
const KindPlatformUser = "PlatformUser"

const (
	principalTypeUser        = "user"
	principalTypeServiceUser = "serviceuser"
)

// PlatformUserSpec is one desired platform-user entry from the YAML spec.
// Relation is "admin" or "member" — a SpiceDB relation, not an RBAC "role"
// (a separate concept in Frontier), hence the field name.
type PlatformUserSpec struct {
	Type     string `yaml:"type"`     // "user" | "serviceuser"
	Ref      string `yaml:"ref"`      // email or uuid for a user; id for a service user
	Relation string `yaml:"relation"` // "admin" | "member"
}

// platformPrincipal is the current state of one platform principal, as returned
// by ListPlatformUsers.
type platformPrincipal struct {
	Type      string
	ID        string
	Email     string // users only
	Relations map[string]struct{}
}

type opAction string

const (
	opAdd    opAction = "add"
	opRemove opAction = "remove"
)

// Op is a single planned change to one (principal, relation). Ref is the desired
// entry's ref for an add (email or id) and the current principal's id for a remove.
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
// user_id, otherwise serviceuser_id. Exactly one is non-empty.
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
	ref := strings.TrimSpace(s.Ref)
	if ref == "" {
		return fmt.Errorf("empty ref")
	}
	// The ref must be a form the server resolves to exactly one principal, so the
	// diff can match it to a live principal. A user is a uuid or an email; a
	// service user is a uuid. An email ref is matched by its address, so a
	// display-name or differently-cased form still matches the address the server
	// stored. A slug (neither a uuid nor an email) is rejected.
	if s.Type == principalTypeUser {
		if _, isEmail := emailAddress(ref); !isUUID(ref) && !isEmail {
			return fmt.Errorf("ref %q must be a user id (uuid) or an email address", s.Ref)
		}
	} else if !isUUID(ref) {
		return fmt.Errorf("ref %q must be a service user id (uuid)", s.Ref)
	}
	// The bootstrap SA is server-managed; reject it here too, not just skip it on
	// the current side. It is a service-user id, and users are a separate id space
	// (the server's own guard checks only the service-user id), so this is scoped
	// to serviceuser entries. IsBootstrapServiceUser matches the id in any UUID
	// form, so a non-canonical spelling cannot slip past the guard.
	if s.Type == principalTypeServiceUser && schema.IsBootstrapServiceUser(ref) {
		return fmt.Errorf("ref %q is the bootstrap service account, which the server manages and cannot be reconciled", s.Ref)
	}
	return nil
}

// isUUID reports whether s parses as a UUID in any form uuid.Parse accepts
// (canonical, uppercase, urn:uuid, braces, or no dashes).
func isUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// emailAddress returns the lowercased address part of an email and whether s
// parsed as one. "Alice <A@X.com>", "a@x.com", and "A@X.com" all yield
// "a@x.com", so a display-name or differently-cased ref matches the address the
// server stored. A value that is not an email (a uuid, a slug) does not parse
// and returns ok=false.
func emailAddress(s string) (addr string, ok bool) {
	a, err := mail.ParseAddress(strings.TrimSpace(s))
	if err != nil {
		return "", false
	}
	got := strings.ToLower(a.Address)
	// A quoted local part unquotes to an address ParseAddress cannot read back
	// (`"john smith"@x.com` -> `john smith@x.com`). Treat it as not a usable email,
	// so callers fall back to the id and export never emits a ref that fails its
	// own re-validation.
	if _, err := mail.ParseAddress(got); err != nil {
		return "", false
	}
	return got, true
}

// canonicalRef normalizes a ref to the form the server stores. A UUID in any
// accepted form becomes the canonical lowercase id; an email becomes its
// lowercased address with any display name dropped, which is the form an add
// sends and the form matching compares. Anything else is only trimmed.
func canonicalRef(ref string) string {
	ref = strings.TrimSpace(ref)
	if u, err := uuid.Parse(ref); err == nil {
		return u.String()
	}
	// An add for an unmatched email spec sends this ref, so normalize it to the
	// plain address the server resolves, not a display-name form it would treat as
	// a new user.
	if addr, ok := emailAddress(ref); ok {
		return addr
	}
	return ref
}

// specMatchesPrincipal reports whether a desired spec refers to a current
// principal: a service user matches by id, a user by id or by email address.
// Emails compare by their parsed address (lowercased), so a display-name or
// differently-cased ref matches the plain address the server stored. A uuid ref
// does not parse as an email, so it only ever matches by id, never against a
// user whose email happens to be another user's id.
func specMatchesPrincipal(s PlatformUserSpec, p platformPrincipal) bool {
	if s.Type != p.Type {
		return false
	}
	if s.Ref == p.ID {
		return true
	}
	if p.Type != principalTypeUser {
		return false
	}
	refAddr, refOK := emailAddress(s.Ref)
	pAddr, pOK := emailAddress(p.Email)
	return refOK && pOK && refAddr == pAddr
}

var platformRelationOrder = []string{schema.AdminRelationName, schema.MemberRelationName}

// diffPlatformUsers returns the ops that make the current platform principals
// match the desired spec, per (principal, relation). Order is stable.
func diffPlatformUsers(desired []PlatformUserSpec, current []platformPrincipal) ([]Op, error) {
	// Validate every entry, then canonicalize its ref so a valid-but-non-canonical
	// id (uppercase, urn:uuid, braces, no dashes) matches the principal's stored id
	// instead of planning a spurious remove of the access it means to keep.
	canonical := make([]PlatformUserSpec, len(desired))
	for i, s := range desired {
		if err := validateSpec(s); err != nil {
			return nil, fmt.Errorf("invalid platform-user spec %+v: %w", s, err)
		}
		s.Ref = canonicalRef(s.Ref)
		canonical[i] = s
	}
	desired = canonical

	// Emit adds before removes so a relation change (admin -> member) never drops
	// all access if a later op fails.
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
