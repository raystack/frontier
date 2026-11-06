package reconcile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func curAuditLogs() currentBillingFeature {
	return currentBillingFeature{
		ID:         "f1",
		Name:       "audit_logs",
		Title:      "Audit Logs",
		Metadata:   map[string]any{"info": "Every action is recorded.", "rank": float64(5)},
		ProductIDs: []string{"p1", "p2"},
	}
}

func auditLogsSpec() BillingFeatureSpec {
	return BillingFeatureSpec{
		Name:     "audit_logs",
		Title:    "Audit Logs",
		Metadata: map[string]any{"info": "Every action is recorded.", "rank": 5},
	}
}

func TestDiffBillingFeatures(t *testing.T) {
	t.Run("adds a feature the server does not have", func(t *testing.T) {
		desired := []BillingFeatureSpec{auditLogsSpec(), {Name: "co_pilot", Title: "Co-pilot"}}
		ops, err := diffBillingFeatures(desired, []currentBillingFeature{curAuditLogs()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opAdd, ops[0].action)
			assert.Equal(t, "co_pilot", ops[0].spec.Name)
			assert.Equal(t, "add feature co_pilot", ops[0].String())
		}
	})

	t.Run("updates a feature whose title differs", func(t *testing.T) {
		s := auditLogsSpec()
		s.Title = "Audit logs"
		ops, err := diffBillingFeatures([]BillingFeatureSpec{s}, []currentBillingFeature{curAuditLogs()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opUpdate, ops[0].action)
			assert.Equal(t, "update feature audit_logs (title)", ops[0].String())
			assert.Equal(t, "f1", ops[0].current.ID)
		}
	})

	t.Run("updates a feature whose metadata differs", func(t *testing.T) {
		s := auditLogsSpec()
		s.Metadata["info"] = "See [the docs](https://docs.example.com)."
		ops, err := diffBillingFeatures([]BillingFeatureSpec{s}, []currentBillingFeature{curAuditLogs()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update feature audit_logs (metadata)", ops[0].String())
		}
	})

	t.Run("updates both title and metadata", func(t *testing.T) {
		s := BillingFeatureSpec{Name: "audit_logs", Title: "Logs", Metadata: map[string]any{"info": "new"}}
		ops, err := diffBillingFeatures([]BillingFeatureSpec{s}, []currentBillingFeature{curAuditLogs()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update feature audit_logs (title, metadata)", ops[0].String())
		}
	})

	t.Run("is a no-op when only metadata key order and number form differ", func(t *testing.T) {
		s := BillingFeatureSpec{
			Name:     "audit_logs",
			Title:    "Audit Logs",
			Metadata: map[string]any{"rank": 5.0, "info": "Every action is recorded."},
		}
		ops, err := diffBillingFeatures([]BillingFeatureSpec{s}, []currentBillingFeature{curAuditLogs()})
		assert.NoError(t, err)
		assert.Empty(t, ops)

		ops, err = diffBillingFeatures([]BillingFeatureSpec{auditLogsSpec()}, []currentBillingFeature{curAuditLogs()})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("an omitted title resets the title to empty", func(t *testing.T) {
		s := auditLogsSpec()
		s.Title = ""
		ops, err := diffBillingFeatures([]BillingFeatureSpec{s}, []currentBillingFeature{curAuditLogs()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, "update feature audit_logs (title)", ops[0].String())
			assert.Equal(t, "", ops[0].spec.Title)
		}
	})

	t.Run("omitted or empty metadata resets the metadata to empty", func(t *testing.T) {
		for _, md := range []map[string]any{nil, {}} {
			s := auditLogsSpec()
			s.Metadata = md
			ops, err := diffBillingFeatures([]BillingFeatureSpec{s}, []currentBillingFeature{curAuditLogs()})
			assert.NoError(t, err)
			if assert.Len(t, ops, 1) {
				assert.Equal(t, "update feature audit_logs (metadata)", ops[0].String())
			}
		}
	})

	t.Run("omitted metadata matches a feature with no metadata", func(t *testing.T) {
		cur := curAuditLogs()
		cur.Metadata = nil
		cur.Title = ""
		ops, err := diffBillingFeatures([]BillingFeatureSpec{{Name: "audit_logs"}}, []currentBillingFeature{cur})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("fails when a server feature is missing from the file", func(t *testing.T) {
		other := currentBillingFeature{ID: "f2", Name: "workflows"}
		_, err := diffBillingFeatures([]BillingFeatureSpec{auditLogsSpec()}, []currentBillingFeature{curAuditLogs(), other})
		assert.ErrorContains(t, err, "features exist on the server but are not in the file: workflows")
	})

	t.Run("rejects a delete flag", func(t *testing.T) {
		s := auditLogsSpec()
		s.Delete = true
		_, err := diffBillingFeatures([]BillingFeatureSpec{s}, []currentBillingFeature{curAuditLogs()})
		assert.ErrorContains(t, err, "cannot be deleted")
	})

	t.Run("adds a name that differs from a server feature only by case", func(t *testing.T) {
		upper := auditLogsSpec()
		upper.Name = "Audit_Logs"
		ops, err := diffBillingFeatures([]BillingFeatureSpec{auditLogsSpec(), upper}, []currentBillingFeature{curAuditLogs()})
		assert.NoError(t, err)
		if assert.Len(t, ops, 1) {
			assert.Equal(t, opAdd, ops[0].action)
			assert.Equal(t, "Audit_Logs", ops[0].spec.Name)
		}
	})

	t.Run("fails on a duplicate name", func(t *testing.T) {
		_, err := diffBillingFeatures([]BillingFeatureSpec{auditLogsSpec(), auditLogsSpec()}, []currentBillingFeature{curAuditLogs()})
		assert.ErrorContains(t, err, "listed more than once")
	})

	t.Run("allows server features whose names differ only by case", func(t *testing.T) {
		upper := auditLogsSpec()
		upper.Name = "Audit_Logs"
		curUpper := curAuditLogs()
		curUpper.ID = "f2"
		curUpper.Name = "Audit_Logs"
		ops, err := diffBillingFeatures([]BillingFeatureSpec{auditLogsSpec(), upper}, []currentBillingFeature{curAuditLogs(), curUpper})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("matches a server name with surrounding spaces exactly", func(t *testing.T) {
		spaced := auditLogsSpec()
		spaced.Name = "audit_logs "
		curSpaced := curAuditLogs()
		curSpaced.Name = "audit_logs "
		ops, err := diffBillingFeatures([]BillingFeatureSpec{spaced}, []currentBillingFeature{curSpaced})
		assert.NoError(t, err)
		assert.Empty(t, ops)
	})

	t.Run("fails on an empty name", func(t *testing.T) {
		_, err := diffBillingFeatures([]BillingFeatureSpec{{Name: "  "}}, nil)
		assert.ErrorContains(t, err, "name is required")
	})

	t.Run("orders adds before updates, each by name", func(t *testing.T) {
		current := []currentBillingFeature{
			{ID: "f1", Name: "zeta"},
			{ID: "f2", Name: "alpha"},
		}
		desired := []BillingFeatureSpec{
			{Name: "zeta", Title: "Zeta"},
			{Name: "new_b"},
			{Name: "alpha", Title: "Alpha"},
			{Name: "new_a"},
		}
		ops, err := diffBillingFeatures(desired, current)
		assert.NoError(t, err)
		var got []string
		for _, op := range ops {
			got = append(got, op.String())
		}
		assert.Equal(t, []string{
			"add feature new_a",
			"add feature new_b",
			"update feature alpha (title)",
			"update feature zeta (title)",
		}, got)
	})
}
