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
}

// planStrings is a test helper: the String() of each op, in order.
func planStrings(ops []metaSchemaOp) []string {
	out := make([]string, 0, len(ops))
	for _, op := range ops {
		out = append(out, op.String())
	}
	return out
}
