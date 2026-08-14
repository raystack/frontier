package reconcile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// small controlled default set for the pure tests
var testMetaDefaults = map[string]string{
	"user":         `{"type":"object"}`,
	"organization": `{"type":"object","properties":{}}`,
}

func TestValidateMetaSchemaSpecs(t *testing.T) {
	t.Run("accepts a known built-in", func(t *testing.T) {
		err := validateMetaSchemaSpecs([]MetaSchemaSpec{{Name: "user", Schema: `{"type":"object"}`}}, testMetaDefaults)
		assert.NoError(t, err)
	})
	t.Run("rejects an unknown name", func(t *testing.T) {
		err := validateMetaSchemaSpecs([]MetaSchemaSpec{{Name: "widget", Schema: `{}`}}, testMetaDefaults)
		assert.ErrorContains(t, err, "unknown metaschema")
	})
	t.Run("rejects a duplicate name", func(t *testing.T) {
		err := validateMetaSchemaSpecs([]MetaSchemaSpec{
			{Name: "user", Schema: `{}`},
			{Name: "user", Schema: `{}`},
		}, testMetaDefaults)
		assert.ErrorContains(t, err, "more than once")
	})
	t.Run("rejects an empty schema", func(t *testing.T) {
		err := validateMetaSchemaSpecs([]MetaSchemaSpec{{Name: "user", Schema: "  "}}, testMetaDefaults)
		assert.ErrorContains(t, err, "schema is required")
	})
	t.Run("rejects invalid JSON", func(t *testing.T) {
		err := validateMetaSchemaSpecs([]MetaSchemaSpec{{Name: "user", Schema: "{not json"}}, testMetaDefaults)
		assert.ErrorContains(t, err, "not valid JSON")
	})
	t.Run("rejects a non-object schema", func(t *testing.T) {
		for _, bad := range []string{"123", `"x"`, "[]", "true"} {
			err := validateMetaSchemaSpecs([]MetaSchemaSpec{{Name: "user", Schema: bad}}, testMetaDefaults)
			assert.ErrorContains(t, err, "must be a JSON object", "schema %q should be rejected", bad)
		}
	})
	t.Run("accepts a built-in name in any case", func(t *testing.T) {
		err := validateMetaSchemaSpecs([]MetaSchemaSpec{{Name: "Organization", Schema: `{"type":"object"}`}}, testMetaDefaults)
		assert.NoError(t, err)
	})
}

func TestDiffMetaSchemas(t *testing.T) {
	t.Run("sets a built-in the file overrides", func(t *testing.T) {
		ops, err := diffMetaSchemas(
			[]MetaSchemaSpec{{Name: "user", Schema: `{"type":"object","required":["x"]}`}},
			[]currentMetaSchema{{ID: "user-id", Name: "user", Schema: `{"type":"object"}`}, {ID: "org-id", Name: "organization", Schema: `{"type":"object","properties":{}}`}},
			testMetaDefaults,
		)
		assert.NoError(t, err)
		assert.Equal(t, []string{"set metaschema user"}, planStrings(ops))
		assert.Equal(t, "user-id", ops[0].id)
	})
	t.Run("no op when the file matches the server, ignoring whitespace", func(t *testing.T) {
		ops, err := diffMetaSchemas(
			[]MetaSchemaSpec{{Name: "user", Schema: "{ \"type\":  \"object\" }"}},
			[]currentMetaSchema{{ID: "user-id", Name: "user", Schema: `{"type":"object"}`}, {ID: "org-id", Name: "organization", Schema: `{"type":"object","properties":{}}`}},
			testMetaDefaults,
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})
	t.Run("no op when the file and server differ only in key order and whitespace", func(t *testing.T) {
		ops, err := diffMetaSchemas(
			[]MetaSchemaSpec{{Name: "user", Schema: `{"a":1,"b":2}`}},
			[]currentMetaSchema{{ID: "user-id", Name: "user", Schema: `{ "b": 2, "a": 1 }`}, {ID: "org-id", Name: "organization", Schema: `{"type":"object","properties":{}}`}},
			testMetaDefaults,
		)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})
	t.Run("resets a built-in the file leaves out", func(t *testing.T) {
		ops, err := diffMetaSchemas(
			nil,
			[]currentMetaSchema{{ID: "user-id", Name: "user", Schema: `{"type":"object"}`}, {ID: "org-id", Name: "organization", Schema: `{"type":"string"}`}},
			testMetaDefaults,
		)
		assert.NoError(t, err)
		assert.Equal(t, []string{"reset metaschema organization to default"}, planStrings(ops))
		assert.Equal(t, "org-id", ops[0].id)
	})
	t.Run("creates a built-in missing on the server", func(t *testing.T) {
		ops, err := diffMetaSchemas(
			nil,
			[]currentMetaSchema{{ID: "org-id", Name: "organization", Schema: `{"type":"object","properties":{}}`}},
			testMetaDefaults,
		)
		assert.NoError(t, err)
		assert.Equal(t, []string{"create metaschema user"}, planStrings(ops))
		assert.Equal(t, "", ops[0].id)
	})
	t.Run("matches a built-in name case-insensitively", func(t *testing.T) {
		ops, err := diffMetaSchemas(
			[]MetaSchemaSpec{{Name: "Organization", Schema: `{"type":"object","required":["x"]}`}},
			[]currentMetaSchema{{ID: "user-id", Name: "user", Schema: `{"type":"object"}`}, {ID: "org-id", Name: "organization", Schema: `{"type":"object","properties":{}}`}},
			testMetaDefaults,
		)
		assert.NoError(t, err)
		assert.Equal(t, []string{"set metaschema organization"}, planStrings(ops))
		assert.Equal(t, "org-id", ops[0].id)
	})
}

func TestCanonicalJSON(t *testing.T) {
	t.Run("equal after whitespace normalization", func(t *testing.T) {
		got, err := canonicalJSON(`{ "type" :  "object" }`)
		assert.NoError(t, err)
		want, err := canonicalJSON(`{"type":"object"}`)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("equal after key-order normalization", func(t *testing.T) {
		got, err := canonicalJSON(`{"b":2,"a":1}`)
		assert.NoError(t, err)
		want, err := canonicalJSON(`{"a":1,"b":2}`)
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
	t.Run("rejects invalid JSON", func(t *testing.T) {
		_, err := canonicalJSON("{not json")
		assert.Error(t, err)
	})
	t.Run("keeps a real difference in a large integer", func(t *testing.T) {
		a, err := canonicalJSON(`{"maximum":10000000000000001}`)
		assert.NoError(t, err)
		b, err := canonicalJSON(`{"maximum":10000000000000002}`)
		assert.NoError(t, err)
		assert.NotEqual(t, a, b)
	})
	t.Run("rejects trailing data", func(t *testing.T) {
		_, err := canonicalJSON(`{"a":1} {"b":2}`)
		assert.Error(t, err)
	})
}

func TestExportMetaSchemas(t *testing.T) {
	current := []currentMetaSchema{
		{ID: "user-id", Name: "user", Schema: `{"type":"object","required":["x"]}`},       // overridden
		{ID: "org-id", Name: "organization", Schema: `{"type":"object","properties":{}}`}, // at default
	}
	t.Run("emits only overridden built-ins", func(t *testing.T) {
		specs, err := exportMetaSchemas(current, testMetaDefaults)
		assert.NoError(t, err)
		assert.Equal(t, []MetaSchemaSpec{{Name: "user", Schema: `{"type":"object","required":["x"]}`}}, specs)
	})
	t.Run("reconciling an export plans zero changes", func(t *testing.T) {
		specs, err := exportMetaSchemas(current, testMetaDefaults)
		assert.NoError(t, err)
		ops, err := diffMetaSchemas(specs, current, testMetaDefaults)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})
	t.Run("omits a default missing from the current server state", func(t *testing.T) {
		defaults := map[string]string{
			"user":  `{"type":"object"}`,
			"group": `{"type":"object"}`,
		}
		currentMissingGroup := []currentMetaSchema{
			{ID: "user-id", Name: "user", Schema: `{"type":"object","required":["x"]}`},
		}
		specs, err := exportMetaSchemas(currentMissingGroup, defaults)
		assert.NoError(t, err)
		assert.Equal(t, []MetaSchemaSpec{{Name: "user", Schema: `{"type":"object","required":["x"]}`}}, specs)
	})
	t.Run("emits a stored schema verbatim when it is not valid JSON", func(t *testing.T) {
		badCurrent := []currentMetaSchema{{ID: "user-id", Name: "user", Schema: "{not json"}}
		specs, err := exportMetaSchemas(badCurrent, testMetaDefaults)
		assert.NoError(t, err)
		assert.Equal(t, []MetaSchemaSpec{{Name: "user", Schema: "{not json"}}, specs)
	})
}

// TestMetaSchema_RFCRules checks the metaschema kind against the five rules of
// RFC 0001, at the pure diff/validate/export layer.
func TestMetaSchema_RFCRules(t *testing.T) {
	// R1 Scope and identity: a metaschema outside the managed set is invisible.
	// The diff never touches it and export never emits it.
	t.Run("R1 ignores a metaschema outside the managed set", func(t *testing.T) {
		current := []currentMetaSchema{
			{ID: "user-id", Name: "user", Schema: testMetaDefaults["user"]},                // at default
			{ID: "org-id", Name: "organization", Schema: testMetaDefaults["organization"]}, // at default
			{ID: "cust-id", Name: "custom_widget", Schema: `{"type":"object"}`},            // not a built-in
		}
		ops, err := diffMetaSchemas(nil, current, testMetaDefaults)
		assert.NoError(t, err)
		assert.Empty(t, ops)
		specs, err := exportMetaSchemas(current, testMetaDefaults)
		assert.NoError(t, err)
		assert.Empty(t, specs)
	})

	// R4 Converge, not transact: re-reconciling a state that already matches the
	// file plans nothing, so a re-run after a partial apply converges.
	t.Run("R4 re-reconciling a converged state plans nothing", func(t *testing.T) {
		desired := []MetaSchemaSpec{{Name: "organization", Schema: `{"type":"object","required":["cc"]}`}}
		converged := []currentMetaSchema{
			{ID: "user-id", Name: "user", Schema: testMetaDefaults["user"]},
			{ID: "org-id", Name: "organization", Schema: `{"type":"object","required":["cc"]}`},
		}
		ops, err := diffMetaSchemas(desired, converged, testMetaDefaults)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	// R5 Export inverts reconcile: an all-default server exports nothing, and
	// reconciling that empty export plans nothing.
	t.Run("R5 all-default server exports nothing and round-trips to zero", func(t *testing.T) {
		current := []currentMetaSchema{
			{ID: "user-id", Name: "user", Schema: testMetaDefaults["user"]},
			{ID: "org-id", Name: "organization", Schema: testMetaDefaults["organization"]},
		}
		specs, err := exportMetaSchemas(current, testMetaDefaults)
		assert.NoError(t, err)
		assert.Empty(t, specs)
		ops, err := diffMetaSchemas(specs, current, testMetaDefaults)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	// R5 Export inverts reconcile: an overridden built-in round-trips exactly,
	// including a large integer that float64 would round.
	t.Run("R5 an overridden built-in round-trips, keeping number precision", func(t *testing.T) {
		current := []currentMetaSchema{
			{ID: "user-id", Name: "user", Schema: testMetaDefaults["user"]},
			{ID: "org-id", Name: "organization", Schema: `{"type":"object","properties":{"n":{"maximum":10000000000000001}}}`},
		}
		specs, err := exportMetaSchemas(current, testMetaDefaults)
		assert.NoError(t, err)
		assert.Len(t, specs, 1)
		ops, err := diffMetaSchemas(specs, current, testMetaDefaults)
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})
}

// planStrings is a test helper: the String() of each op, in order.
func planStrings(ops []metaSchemaOp) []string {
	out := make([]string, 0, len(ops))
	for _, op := range ops {
		out = append(out, op.String())
	}
	return out
}
