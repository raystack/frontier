package reconcile

import (
	"testing"

	"github.com/raystack/frontier/core/preference"
	"github.com/stretchr/testify/assert"
)

func TestDiffPreferences(t *testing.T) {
	// All three are checkbox traits, so their validator accepts only "true" or
	// "false" and rejects an empty value.
	traits := map[string]preference.Trait{
		"disable_orgs_on_create": {Input: preference.TraitInputCheckbox, InputHints: "true,false", Default: "false"},
		"disable_orgs_listing":   {Input: preference.TraitInputCheckbox, InputHints: "true,false", Default: "false"},
		"invite_with_roles":      {Input: preference.TraitInputCheckbox, InputHints: "true,false", Default: "true"},
	}

	t.Run("no changes when the file matches the server", func(t *testing.T) {
		current := map[string]string{"disable_orgs_on_create": "true"}
		ops, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: "true"},
		}, current, traits)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("sets a value that differs from the server", func(t *testing.T) {
		current := map[string]string{"disable_orgs_on_create": "false"}
		ops, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: "true"},
		}, current, traits)
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
		}, map[string]string{}, traits)
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opSet, ops[0].action)
		}
	})

	t.Run("a value equal to the default and stored nowhere is a no-op", func(t *testing.T) {
		ops, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: "false"},
		}, map[string]string{}, traits)
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
		}, current, traits)
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
		}, current, traits)
		assert.NoError(t, err)
		if assert.Len(t, ops, 2) {
			assert.Equal(t, "set preference disable_orgs_on_create = true", ops[0].String())
			assert.Equal(t, "reset preference disable_orgs_listing to default", ops[1].String())
		}
	})

	t.Run("an empty stored value counts as the default", func(t *testing.T) {
		// The server treats an empty stored value as unset and falls back to the
		// default, so a file entry equal to the default must plan no change.
		current := map[string]string{"disable_orgs_on_create": ""}
		ops, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: "false"}, // the default
		}, current, traits)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("an unknown preference name fails the plan", func(t *testing.T) {
		_, err := diffPreferences([]PreferenceSpec{
			{Name: "not_a_trait", Value: "x"},
		}, map[string]string{}, traits)
		assert.ErrorContains(t, err, `unknown platform preference "not_a_trait"`)
	})

	t.Run("a duplicate name fails the plan", func(t *testing.T) {
		_, err := diffPreferences([]PreferenceSpec{
			{Name: "invite_with_roles", Value: "true"},
			{Name: "invite_with_roles", Value: "false"},
		}, map[string]string{}, traits)
		assert.ErrorContains(t, err, "listed more than once")
	})

	t.Run("an empty name fails the plan", func(t *testing.T) {
		_, err := diffPreferences([]PreferenceSpec{{Name: "", Value: "x"}}, map[string]string{}, traits)
		assert.ErrorContains(t, err, "name is required")
	})

	t.Run("a stored value whose trait no longer exists is left alone", func(t *testing.T) {
		current := map[string]string{"gone_trait": "whatever"}
		ops, err := diffPreferences(nil, current, traits)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("a stored empty value not in the file is left alone, not reset", func(t *testing.T) {
		// An empty stored value counts as unset, so it is already at its default
		// and needs no reset. Planning a reset here would write the default on
		// every run and break the export round trip.
		current := map[string]string{"disable_orgs_on_create": ""}
		ops, err := diffPreferences(nil, current, traits)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("an empty value the trait does not accept fails the plan", func(t *testing.T) {
		// A checkbox trait accepts only "true" or "false", so an empty value is a
		// set the server would reject. It must fail the plan, not appear in it.
		_, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: ""},
		}, map[string]string{}, traits)
		assert.ErrorContains(t, err, "not a valid value")
	})

	t.Run("a value outside a trait's options fails the plan", func(t *testing.T) {
		_, err := diffPreferences([]PreferenceSpec{
			{Name: "disable_orgs_on_create", Value: "maybe"},
		}, map[string]string{}, traits)
		assert.ErrorContains(t, err, "not a valid value")
	})

	t.Run("a listed value equal to the server value round-trips even if the trait would now reject it", func(t *testing.T) {
		// R5: export lists the server's stored value verbatim. If a trait's options
		// tighten after a value was stored (a changed custom trait, or an upgrade),
		// the server still holds the old value. Reconciling its export must plan
		// zero ops, not reject the value it just exported. So a value equal to the
		// one in effect is never re-validated, only a value that becomes a set is.
		theme := map[string]preference.Trait{
			"theme": {Input: preference.TraitInputSelect, InputHints: "light,dark", Default: "light"},
		}
		current := map[string]string{"theme": "auto"} // stored when "auto" was allowed
		ops, err := diffPreferences([]PreferenceSpec{{Name: "theme", Value: "auto"}}, current, theme)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("a reset to an unreachable default fails the plan", func(t *testing.T) {
		// A misconfigured text trait whose default its own validator rejects
		// (text cannot be empty). Resetting it would write a value the server
		// refuses, so the plan must fail rather than emit that reset.
		broken := map[string]preference.Trait{
			"broken": {Input: preference.TraitInputText, Default: ""},
		}
		current := map[string]string{"broken": "something"} // stored, not listed -> reset
		_, err := diffPreferences(nil, current, broken)
		assert.ErrorContains(t, err, "cannot be reset")
	})
}
