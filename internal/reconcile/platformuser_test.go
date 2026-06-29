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
			[]PlatformUserSpec{{Type: "user", Ref: "alice@x.com", Role: admin}},
			nil,
		)
		assert.NoError(t, err)
		assert.Equal(t, []Op{{Action: opAdd, Type: "user", Ref: "alice@x.com", Relation: admin}}, ops)
	})

	t.Run("new service user is added", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "serviceuser", Ref: "su-1", Role: member}},
			nil,
		)
		assert.NoError(t, err)
		assert.Equal(t, []Op{{Action: opAdd, Type: "serviceuser", Ref: "su-1", Relation: member}}, ops)
	})

	t.Run("already-correct principal is a no-op (matched by email)", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "user", Ref: "alice@x.com", Role: admin}},
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

	t.Run("role change removes then re-adds the desired relation", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{{Type: "user", Ref: "alice@x.com", Role: member}},
			[]platformPrincipal{principal("user", "alice-id", "alice@x.com", admin)},
		)
		assert.NoError(t, err)
		assert.Equal(t, []Op{
			{Action: opRemove, Type: "user", Ref: "alice-id", Relation: admin},
			{Action: opAdd, Type: "user", Ref: "alice-id", Relation: member},
		}, ops)
	})

	t.Run("mixed: keep one, add one, remove one", func(t *testing.T) {
		ops, err := diffPlatformUsers(
			[]PlatformUserSpec{
				{Type: "user", Ref: "keep@x.com", Role: admin}, // already correct
				{Type: "user", Ref: "new@x.com", Role: member}, // new -> add
			},
			[]platformPrincipal{
				principal("user", "keep-id", "keep@x.com", admin), // matches keep -> no-op
				principal("user", "drop-id", "drop@x.com", admin), // not desired -> remove
			},
		)
		assert.NoError(t, err)
		assert.Equal(t, []Op{
			{Action: opRemove, Type: "user", Ref: "drop-id", Relation: admin},
			{Action: opAdd, Type: "user", Ref: "new@x.com", Relation: member},
		}, ops)
	})

	t.Run("rejects an invalid role", func(t *testing.T) {
		_, err := diffPlatformUsers([]PlatformUserSpec{{Type: "user", Ref: "a@x.com", Role: "owner"}}, nil)
		assert.Error(t, err)
	})

	t.Run("rejects an invalid type", func(t *testing.T) {
		_, err := diffPlatformUsers([]PlatformUserSpec{{Type: "group", Ref: "a@x.com", Role: admin}}, nil)
		assert.Error(t, err)
	})

	t.Run("rejects an empty ref", func(t *testing.T) {
		_, err := diffPlatformUsers([]PlatformUserSpec{{Type: "user", Ref: "", Role: admin}}, nil)
		assert.Error(t, err)
	})
}
