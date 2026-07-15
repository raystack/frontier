package reconcile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiffPreferences(t *testing.T) {
	defaults := map[string]string{
		"disable_orgs_on_create": "false",
		"disable_orgs_listing":   "false",
		"invite_with_roles":      "true",
	}

	t.Run("no changes when the file matches the server", func(t *testing.T) {
		current := map[string]string{"disable_orgs_on_create": "true"}
		ops, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: "true"},
		}, current, defaults)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("sets a value that differs from the server", func(t *testing.T) {
		current := map[string]string{"disable_orgs_on_create": "false"}
		ops, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: "true"},
		}, current, defaults)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "set preference disable_orgs_on_create = true", ops[0].String())
			assert.Equal(t, "true", ops[0].value)
		}
	})

	t.Run("sets a value that is stored nowhere but differs from the default", func(t *testing.T) {
		// not in the DB yet: the effective value is the default "false"
		ops, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: "true"},
		}, map[string]string{}, defaults)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opSet, ops[0].action)
		}
	})

	t.Run("a value equal to the default and stored nowhere is a no-op", func(t *testing.T) {
		ops, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: "false"},
		}, map[string]string{}, defaults)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("resets a stored preference that the file leaves out", func(t *testing.T) {
		current := map[string]string{
			"disable_orgs_on_create": "true",
			"invite_with_roles":      "false",
		}
		ops, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: "true"},
		}, current, defaults)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "reset preference invite_with_roles to default", ops[0].String())
			assert.Equal(t, "true", ops[0].value) // the default is written back
		}
	})

	t.Run("sets run before resets, each sorted by name", func(t *testing.T) {
		current := map[string]string{
			"disable_orgs_listing": "true", // not listed -> reset
			"invite_with_roles":    "true", // not listed but equals default -> untouched
		}
		ops, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: "true"}, // set
		}, current, defaults)
		assert.NoError(t, err)
		if assert.Len(t, ops, 2) {
			assert.Equal(t, "set preference disable_orgs_on_create = true", ops[0].String())
			assert.Equal(t, "reset preference disable_orgs_listing to default", ops[1].String())
		}
	})

	t.Run("an unknown preference name fails the plan", func(t *testing.T) {
		_, err := diffPreferences([]PreferenceSpec{
			{Name: "not_a_trait", Value: "x"},
		}, map[string]string{}, defaults)
		assert.ErrorContains(t, err, `unknown platform preference "not_a_trait"`)
	})

	t.Run("a duplicate name fails the plan", func(t *testing.T) {
		_, err := diffPreferences([]PreferenceSpec{
			{Name: "invite_with_roles", Value: "true"},
			{Name: "invite_with_roles", Value: "false"},
		}, map[string]string{}, defaults)
		assert.ErrorContains(t, err, "listed more than once")
	})

	t.Run("an empty name fails the plan", func(t *testing.T) {
		_, err := diffPreferences([]PreferenceSpec{{Name: "", Value: "x"}}, map[string]string{}, defaults)
		assert.ErrorContains(t, err, "name is required")
	})

	t.Run("a stored value whose trait no longer exists is left alone", func(t *testing.T) {
		current := map[string]string{"gone_trait": "whatever"}
		ops, err := diffPreferences(nil, current, defaults)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})
}
