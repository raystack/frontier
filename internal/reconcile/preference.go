package reconcile

import (
	"fmt"
	"sort"
	"strings"

	"github.com/raystack/frontier/core/preference"
)

// KindPreference is the desired-state document kind for platform preferences.
const KindPreference = "Preference"

const (
	opSet   opAction = "set"
	opReset opAction = "reset"
)

// PreferenceSpec is one desired platform preference. Name is a trait name the
// server knows; value is the string value to set. Preferences are strings end
// to end, so a boolean-like trait is "true" or "false".
type PreferenceSpec struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// preferenceOp is a single planned change. For a reset, value is the trait
// default that gets written back, because the API has no delete.
type preferenceOp struct {
	action opAction
	name   string
	value  string
}

func (o preferenceOp) String() string {
	if o.action == opReset {
		return fmt.Sprintf("reset preference %s to default", o.name)
	}
	return fmt.Sprintf("set preference %s = %s", o.name, o.value)
}

// diffPreferences returns the ops that make the current platform preferences
// match the desired spec. The file is the full desired state: a preference the
// file lists is set to its value, and a preference the file leaves out is reset
// to its trait default. traits holds every valid platform trait, so it is the
// source of defaults, the set of known names, and the validator each value must
// pass.
func diffPreferences(desired []PreferenceSpec, current map[string]string, traits map[string]preference.Trait) ([]preferenceOp, error) {
	// serverValue is the value in effect on the server: the stored value, or the
	// trait default when nothing (or an empty string) is stored. An empty stored
	// value counts as unset, matching how the server resolves platform
	// preferences, so a file entry equal to the default plans no change.
	serverValue := func(name string) string {
		if v, ok := current[name]; ok && v != "" {
			return v
		}
		return traits[name].Default
	}

	desiredByName := make(map[string]string, len(desired))
	var sets []preferenceOp
	for _, s := range desired {
		if strings.TrimSpace(s.Name) == "" {
			return nil, fmt.Errorf("preference name is required")
		}
		if _, dup := desiredByName[s.Name]; dup {
			return nil, fmt.Errorf("preference %q is listed more than once", s.Name)
		}
		trait, known := traits[s.Name]
		if !known {
			return nil, fmt.Errorf("unknown platform preference %q", s.Name)
		}
		desiredByName[s.Name] = s.Value

		// Only a value that differs from the one in effect becomes a set. Validate
		// just those against the trait, so a set the server would reject fails the
		// plan up front instead of at the API. A value equal to the server's own
		// value plans nothing, so it is left unchecked: that keeps the export of a
		// value stored under an older, looser trait definition round-tripping to
		// zero ops (rule 5), rather than rejecting a state the server can reach.
		if s.Value == serverValue(s.Name) {
			continue
		}
		if !trait.GetValidator().Validate(s.Value) {
			return nil, fmt.Errorf("preference %q: %q is not a valid value for its trait", s.Name, s.Value)
		}
		sets = append(sets, preferenceOp{action: opSet, name: s.Name, value: s.Value})
	}

	var resets []preferenceOp
	for name, value := range current {
		trait, known := traits[name]
		if !known {
			// A stored value whose trait no longer exists: leave it alone. The
			// file cannot name it (unknown names fail), so nothing manages it.
			continue
		}
		if _, listed := desiredByName[name]; listed {
			continue
		}
		// A stored empty value counts as unset, so it is already at its default and
		// needs no reset. Only a stored value that really differs is reset.
		if value == "" || value == trait.Default {
			continue
		}
		// A reset writes the trait default back. If the trait would reject its own
		// default, the default is not reachable and the reset can never apply, so
		// fail the plan rather than emit an op that dies at the API.
		if !trait.GetValidator().Validate(trait.Default) {
			return nil, fmt.Errorf("preference %q: its trait default %q is not a valid value, so it cannot be reset", name, trait.Default)
		}
		resets = append(resets, preferenceOp{action: opReset, name: name, value: trait.Default})
	}

	sort.Slice(sets, func(i, j int) bool { return sets[i].name < sets[j].name })
	sort.Slice(resets, func(i, j int) bool { return resets[i].name < resets[j].name })
	return append(sets, resets...), nil
}
