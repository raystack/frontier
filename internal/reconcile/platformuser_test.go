package reconcile

import (
	"testing"

	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
)

func principal(typ, id, email string, relations ...string) platformPrincipal {
	rels := make(map[string]struct{}, len(relations))
	for _, r := range relations {
		rels[r] = struct{}{}
	}
	return platformPrincipal{Type: typ, ID: id, Email: email, Relations: rels}
}

func TestDiffPlatformUsers(t *testing.T) {
	admin := schema.AdminRelationName
	member := schema.MemberRelationName

	t.Run("no desired, no current -> no ops", func(t *testing.T) {
		ops, err := diffPlatformUsers(nil, nil)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("new user is added", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "user", Ref: "alice@x.com", Relation: admin}},
			nil,
		)
		assert.NoError(t, err)
		assert.Equal(t, []Op{{Action: opAdd, Type: "user", Ref: "alice@x.com", Relation: admin}}, ops)
	})

	t.Run("new service user is added", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "serviceuser", Ref: "11111111-1111-1111-1111-111111111111", Relation: member}},
			nil,
		)
		assert.NoError(t, err)
		assert.Equal(t, []Op{{Action: opAdd, Type: "serviceuser", Ref: "11111111-1111-1111-1111-111111111111", Relation: member}}, ops)
	})

	t.Run("already-correct principal makes no change (matched by email)", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "user", Ref: "alice@x.com", Relation: admin}},
			[]platformPrincipal{principal("user", "alice-id", "alice@x.com", admin)},
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("undesired principal is removed", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			nil,
			[]platformPrincipal{principal("user", "drop-id", "drop@x.com", admin)},
		)
		assert.NoError(t, err)
		assert.Equal(t, []Op{{Action: opRemove, Type: "user", Ref: "drop-id", Relation: admin}}, ops)
	})

	t.Run("relation change adds the new relation before removing the old", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "user", Ref: "alice@x.com", Relation: member}},
			[]platformPrincipal{principal("user", "alice-id", "alice@x.com", admin)},
		)
		assert.NoError(t, err)
		// add-before-remove so a failed apply never leaves the principal access-less.
		assert.Equal(t, []Op{
			{Action: opAdd, Type: "user", Ref: "alice-id", Relation: member},
			{Action: opRemove, Type: "user", Ref: "alice-id", Relation: admin},
		}, ops)
	})

	t.Run("mixed: keep one, add one, remove one", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{
				{Type: "user", Ref: "keep@x.com", Relation: admin}, // already correct
				{Type: "user", Ref: "new@x.com", Relation: member}, // new -> add
			},
			[]platformPrincipal{
				principal("user", "keep-id", "keep@x.com", admin), // matches keep -> no change
				principal("user", "drop-id", "drop@x.com", admin), // not desired -> remove
			},
		)
		assert.NoError(t, err)
		// adds first, then removes (across principals).
		assert.Equal(t, []Op{
			{Action: opAdd, Type: "user", Ref: "new@x.com", Relation: member},
			{Action: opRemove, Type: "user", Ref: "drop-id", Relation: admin},
		}, ops)
	})

	t.Run("strips the extra relation when a principal holds both but only one is desired", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "user", Ref: "alice@x.com", Relation: admin}},
			[]platformPrincipal{principal("user", "alice-id", "alice@x.com", admin, member)},
		)
		assert.NoError(t, err)
		assert.Equal(t, []Op{{Action: opRemove, Type: "user", Ref: "alice-id", Relation: member}}, ops)
	})

	t.Run("keeps both relations when both are desired", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{
				{Type: "user", Ref: "alice@x.com", Relation: admin},
				{Type: "user", Ref: "alice@x.com", Relation: member},
			},
			[]platformPrincipal{principal("user", "alice-id", "alice@x.com", admin, member)},
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("rejects an invalid relation", func(t *testing.T) {
		_, err := diffPlatformUsers([]PlatformUserSpec{{Type: "user", Ref: "a@x.com", Relation: "owner"}}, nil)
		assert.Error(t, err)
	})

	t.Run("rejects an invalid type", func(t *testing.T) {
		_, err := diffPlatformUsers([]PlatformUserSpec{{Type: "group", Ref: "a@x.com", Relation: admin}}, nil)
		assert.Error(t, err)
	})

	t.Run("rejects an empty ref", func(t *testing.T) {
		_, err := diffPlatformUsers([]PlatformUserSpec{{Type: "user", Ref: "", Relation: admin}}, nil)
		assert.Error(t, err)
	})

	t.Run("rejects the bootstrap service account (the server manages it)", func(t *testing.T) {
		_, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "serviceuser", Ref: schema.BootstrapServiceUserID, Relation: admin}}, nil)
		assert.Error(t, err)
	})

	t.Run("rejects the bootstrap service account in a non-canonical uuid form", func(t *testing.T) {
		_, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "serviceuser", Ref: "urn:uuid:" + schema.BootstrapServiceUserID, Relation: admin}}, nil)
		assert.ErrorContains(t, err, "bootstrap service account")
	})

	t.Run("a user that happens to share the bootstrap uuid is a different principal, not rejected", func(t *testing.T) {
		// The bootstrap id is a service-user id; users are a separate id space, so a
		// user entry with that id is not the bootstrap SA and must not be rejected.
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "user", Ref: schema.BootstrapServiceUserID, Relation: admin}},
			[]platformPrincipal{principal("user", schema.BootstrapServiceUserID, "", admin)},
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("a non-canonical uuid ref matches the stored principal, no spurious remove", func(t *testing.T) {
		// The file lists the same id the server stored, but uppercased. Without
		// canonicalization this fails to match and plans an add + a remove that
		// silently strips the principal's access.
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "user", Ref: "AAAAAAAA-1111-1111-1111-111111111111", Relation: admin}},
			[]platformPrincipal{principal("user", "aaaaaaaa-1111-1111-1111-111111111111", "alice@x.com", admin)},
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("rejects a user ref that is neither a uuid nor an email", func(t *testing.T) {
		_, err := diffPlatformUsers([]PlatformUserSpec{{Type: "user", Ref: "alice-slug", Relation: admin}}, nil)
		assert.ErrorContains(t, err, "must be a user id (uuid) or an email address")
	})

	t.Run("a display-name email ref matches the stored principal by address", func(t *testing.T) {
		// A display-name email is a valid ref; it matches by its address, so it does
		// not plan a spurious remove of access it means to keep.
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "user", Ref: "Alice <alice@x.com>", Relation: admin}},
			[]platformPrincipal{principal("user", "alice-id", "alice@x.com", admin)},
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("an add for a new email ref uses the normalized address, not the ref verbatim", func(t *testing.T) {
		// A display-name or differently-cased ref for a user with no current grant
		// must add by the plain address, so the server grants the real user rather
		// than creating a shadow user whose email is the literal display-name form.
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "user", Ref: "Alice <Alice@X.com>", Relation: admin}},
			nil,
		)
		assert.NoError(t, err)
		assert.Equal(t, []Op{{Action: opAdd, Type: "user", Ref: "alice@x.com", Relation: admin}}, ops)
	})

	t.Run("rejects a service user ref that is not a uuid", func(t *testing.T) {
		_, err := diffPlatformUsers([]PlatformUserSpec{{Type: "serviceuser", Ref: "not-a-uuid", Relation: admin}}, nil)
		assert.ErrorContains(t, err, "must be a service user id (uuid)")
	})
}

// TestSpecRefMatching covers the accepted `ref` forms: a user may be referenced by
// id or by email (case-insensitive); a service user only by id.
func TestSpecRefMatching(t *testing.T) {
	admin := schema.AdminRelationName
	userP := principal("user", "user-uuid-1", "alice@x.com")
	suP := principal("serviceuser", "su-uuid-1", "") // service users carry no email

	t.Run("user matches by id", func(t *testing.T) {
		assert.True(t, specMatchesPrincipal(PlatformUserSpec{Type: "user", Ref: "user-uuid-1", Relation: admin}, userP))
	})
	t.Run("user matches by email (case-insensitive)", func(t *testing.T) {
		assert.True(t, specMatchesPrincipal(PlatformUserSpec{Type: "user", Ref: "Alice@X.com", Relation: admin}, userP))
	})
	t.Run("service user matches by id", func(t *testing.T) {
		assert.True(t, specMatchesPrincipal(PlatformUserSpec{Type: "serviceuser", Ref: "su-uuid-1", Relation: admin}, suP))
	})
	t.Run("service user does NOT match by email (email is user-only)", func(t *testing.T) {
		suWithEmail := principal("serviceuser", "su-uuid-1", "svc@x.com")
		assert.False(t, specMatchesPrincipal(PlatformUserSpec{Type: "serviceuser", Ref: "svc@x.com", Relation: admin}, suWithEmail))
	})
	t.Run("no match when the type differs", func(t *testing.T) {
		assert.False(t, specMatchesPrincipal(PlatformUserSpec{Type: "serviceuser", Ref: "user-uuid-1", Relation: admin}, userP))
	})
	t.Run("no match on an unrelated ref", func(t *testing.T) {
		assert.False(t, specMatchesPrincipal(PlatformUserSpec{Type: "user", Ref: "bob@x.com", Relation: admin}, userP))
	})
	t.Run("a uuid ref does not match a user whose email is that uuid", func(t *testing.T) {
		// A user's email can be any string, including another user's id. A uuid ref
		// parses to no address, so it matches only by id, never against that email.
		victim := principal("user", "victim-id", "bbbbbbbb-2222-2222-2222-222222222222")
		assert.False(t, specMatchesPrincipal(
			PlatformUserSpec{Type: "user", Ref: "bbbbbbbb-2222-2222-2222-222222222222", Relation: admin}, victim))
	})

	// end-to-end through the diff: a user referenced by id and a service user by id
	// both make no change when already in the desired relation.
	t.Run("diff makes no change: user by id + serviceuser by id already correct", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{
				{Type: "user", Ref: "11111111-1111-1111-1111-111111111111", Relation: admin},
				{Type: "serviceuser", Ref: "22222222-2222-2222-2222-222222222222", Relation: admin},
			},
			[]platformPrincipal{
				principal("user", "11111111-1111-1111-1111-111111111111", "alice@x.com", admin),
				principal("serviceuser", "22222222-2222-2222-2222-222222222222", "", admin),
			},
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})
}
