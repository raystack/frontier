package reconcile

import (
	"fmt"
	"sort"
	"strings"
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
// to its trait default. defaults holds every valid platform trait and its
// default, so it is both the source of defaults and the set of known names.
func diffPreferences(desired []PreferenceSpec, current, defaults map[string]string) ([]preferenceOp, error) {
	desiredByName := make(map[string]string, len(desired))
	for _, s := range desired {
		if strings.TrimSpace(s.Name) == "" {
			return nil, fmt.Errorf("preference name is required")
		}
		if _, dup := desiredByName[s.Name]; dup {
			return nil, fmt.Errorf("preference %q is listed more than once", s.Name)
		}
		if _, known := defaults[s.Name]; !known {
			return nil, fmt.Errorf("unknown platform preference %q", s.Name)
		}
		desiredByName[s.Name] = s.Value
	}

	// serverValue is the value in effect on the server: the stored value, or
	// the trait default when nothing is stored. An empty stored value counts as
	// unset, matching how the server resolves platform preferences, so a file
	// entry equal to the default plans no change against it.
	serverValue := func(name string) string {
		if v, ok := current[name]; ok && v != "" {
			return v
		}
		return defaults[name]
	}

	var sets, resets []preferenceOp
	for name, value := range desiredByName {
		if value != serverValue(name) {
			sets = append(sets, preferenceOp{action: opSet, name: name, value: value})
		}
	}
	for name, value := range current {
		def, known := defaults[name]
		if !known {
			// A stored value whose trait no longer exists: leave it alone. The
			// file cannot name it (unknown names fail), so nothing manages it.
			continue
		}
		if _, listed := desiredByName[name]; listed {
			continue
		}
		if value != def {
			resets = append(resets, preferenceOp{action: opReset, name: name, value: def})
		}
	}

	sort.Slice(sets, func(i, j int) bool { return sets[i].name < sets[j].name })
	sort.Slice(resets, func(i, j int) bool { return resets[i].name < resets[j].name })
	return append(sets, resets...), nil
}
