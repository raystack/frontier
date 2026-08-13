package metaschema

import (
	"encoding/json"
	"testing"
)

func TestDefaults(t *testing.T) {
	want := []string{NameUser, NameGroup, NameOrg, NameRole, NameProspect}
	if len(Defaults) != len(want) {
		t.Fatalf("Defaults has %d entries, want %d", len(Defaults), len(want))
	}
	for _, name := range want {
		schema, ok := Defaults[name]
		if !ok {
			t.Errorf("Defaults is missing %q", name)
			continue
		}
		if schema == "" {
			t.Errorf("Defaults[%q] is empty", name)
		}
		var v any
		if err := json.Unmarshal([]byte(schema), &v); err != nil {
			t.Errorf("Defaults[%q] is not valid JSON: %v", name, err)
		}
	}
}
