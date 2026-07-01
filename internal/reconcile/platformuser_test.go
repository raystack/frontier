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
			[]PlatformUserSpec{{Type: "serviceuser", Ref: "su-1", Relation: member}},
			nil,
		)
		assert.NoError(t, err)
		assert.Equal(t, []Op{{Action: opAdd, Type: "serviceuser", Ref: "su-1", Relation: member}}, ops)
	})

	t.Run("already-correct principal is a no-op (matched by email)", func(t *testing.T) {
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
				principal("user", "keep-id", "keep@x.com", admin), // matches keep -> no-op
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

	t.Run("rejects the server-managed bootstrap service account", func(t *testing.T) {
		_, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "serviceuser", Ref: schema.BootstrapServiceUserID, Relation: admin}}, nil)
		assert.Error(t, err)
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

	// end-to-end through the diff: a user referenced by id and a service user by id
	// both converge to no-ops when already in the desired relation.
	t.Run("diff converges: user by id + serviceuser by id already correct", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{
				{Type: "user", Ref: "user-uuid-1", Relation: admin},
				{Type: "serviceuser", Ref: "su-uuid-1", Relation: admin},
			},
			[]platformPrincipal{
				principal("user", "user-uuid-1", "alice@x.com", admin),
				principal("serviceuser", "su-uuid-1", "", admin),
			},
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})
}
