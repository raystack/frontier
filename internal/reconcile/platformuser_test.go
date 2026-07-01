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
