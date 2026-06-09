package utils

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

var testDialect = goqu.Dialect("postgres")

func TestAddGroupInQuery(t *testing.T) {
	allowed := []string{"event", "actor_type", "status"}

	tests := []struct {
		name       string
		groupBy    []string
		wantErr    bool
		wantErrMsg string
		wantSQL    string
	}{
		{
			name:    "empty groupBy leaves query unchanged",
			groupBy: nil,
			wantSQL: `SELECT * FROM "audit_records"`,
		},
		{
			name:    "single allowed column",
			groupBy: []string{"event"},
			wantSQL: `SELECT "event" AS "values", COUNT(*) as count FROM "audit_records" GROUP BY "event" ORDER BY "event" ASC`,
		},
		{
			name:    "two allowed columns produce CONCAT",
			groupBy: []string{"event", "actor_type"},
			wantSQL: `SELECT CONCAT("event", ',', "actor_type") AS "values", COUNT(*) as count FROM "audit_records" GROUP BY "event", "actor_type" ORDER BY "event" ASC`,
		},
		{
			name:       "disallowed column returns error",
			groupBy:    []string{"password_hash"},
			wantErr:    true,
			wantErrMsg: "password_hash is not supported in group by",
		},
		{
			name:       "mixed allowed and disallowed returns error",
			groupBy:    []string{"event", "password_hash"},
			wantErr:    true,
			wantErrMsg: "password_hash is not supported in group by",
		},
		{
			name:       "column with semicolon and SQL keywords is rejected",
			groupBy:    []string{"event; DROP TABLE users; --"},
			wantErr:    true,
			wantErrMsg: "event; DROP TABLE users; -- is not supported in group by",
		},
		{
			name:       "column shaped as a subquery is rejected",
			groupBy:    []string{"(SELECT password FROM users LIMIT 1)"},
			wantErr:    true,
			wantErrMsg: "is not supported in group by",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := testDialect.From("audit_records")
			query := &rql.Query{GroupBy: tc.groupBy}

			result, err := AddGroupInQuery(base, query, allowed)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrMsg)
				return
			}

			assert.NoError(t, err)
			sql, _, sqlErr := result.ToSQL()
			assert.NoError(t, sqlErr)
			assert.Equal(t, tc.wantSQL, sql)
		})
	}
}

func TestAddGroupInQuery_QuotesIdentifiers(t *testing.T) {
	tests := []struct {
		name      string
		col       string
		wantInSQL string
	}{
		{
			name:      "column with semicolon and SQL keywords is wrapped as identifier",
			col:       "event; DROP TABLE users; --",
			wantInSQL: `"event; DROP TABLE users; --"`,
		},
		{
			name:      "column shaped as a function call is wrapped as identifier",
			col:       "pg_sleep(60)",
			wantInSQL: `"pg_sleep(60)"`,
		},
		{
			name:      "column containing UNION keywords is wrapped as identifier",
			col:       "event UNION SELECT password FROM users",
			wantInSQL: `"event UNION SELECT password FROM users"`,
		},
		{
			name:      "column shaped as a subquery is wrapped as identifier",
			col:       "(SELECT password FROM users LIMIT 1)",
			wantInSQL: `"(SELECT password FROM users LIMIT 1)"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := testDialect.From("audit_records")
			query := &rql.Query{GroupBy: []string{tc.col}}

			result, err := AddGroupInQuery(base, query, []string{tc.col})
			assert.NoError(t, err)

			sql, _, sqlErr := result.ToSQL()
			assert.NoError(t, sqlErr)
			assert.Contains(t, sql, tc.wantInSQL,
				"column should be wrapped as a quoted identifier, not spliced as raw SQL")
		})
	}
}

func TestAddGroupInQuery_SQLOutput(t *testing.T) {
	auditRecords := []string{"event", "actor_type", "resource_type", "target_type", "org_id", "org_name"}
	prospects := []string{"activity", "status", "source", "verified"}

	tests := []struct {
		name    string
		table   string
		allowed []string
		groupBy []string
		wantSQL string
	}{
		{
			name:    "audit_records groupBy=event",
			table:   "audit_records",
			allowed: auditRecords,
			groupBy: []string{"event"},
			wantSQL: `SELECT "event" AS "values", COUNT(*) as count FROM "audit_records" GROUP BY "event" ORDER BY "event" ASC`,
		},
		{
			name:    "audit_records groupBy=actor_type",
			table:   "audit_records",
			allowed: auditRecords,
			groupBy: []string{"actor_type"},
			wantSQL: `SELECT "actor_type" AS "values", COUNT(*) as count FROM "audit_records" GROUP BY "actor_type" ORDER BY "actor_type" ASC`,
		},
		{
			name:    "audit_records groupBy=resource_type",
			table:   "audit_records",
			allowed: auditRecords,
			groupBy: []string{"resource_type"},
			wantSQL: `SELECT "resource_type" AS "values", COUNT(*) as count FROM "audit_records" GROUP BY "resource_type" ORDER BY "resource_type" ASC`,
		},
		{
			name:    "audit_records groupBy=org_id",
			table:   "audit_records",
			allowed: auditRecords,
			groupBy: []string{"org_id"},
			wantSQL: `SELECT "org_id" AS "values", COUNT(*) as count FROM "audit_records" GROUP BY "org_id" ORDER BY "org_id" ASC`,
		},
		{
			name:    "audit_records two-column groupBy event,actor_type",
			table:   "audit_records",
			allowed: auditRecords,
			groupBy: []string{"event", "actor_type"},
			wantSQL: `SELECT CONCAT("event", ',', "actor_type") AS "values", COUNT(*) as count FROM "audit_records" GROUP BY "event", "actor_type" ORDER BY "event" ASC`,
		},
		{
			name:    "prospects groupBy=activity",
			table:   "prospects",
			allowed: prospects,
			groupBy: []string{"activity"},
			wantSQL: `SELECT "activity" AS "values", COUNT(*) as count FROM "prospects" GROUP BY "activity" ORDER BY "activity" ASC`,
		},
		{
			name:    "prospects groupBy=status",
			table:   "prospects",
			allowed: prospects,
			groupBy: []string{"status"},
			wantSQL: `SELECT "status" AS "values", COUNT(*) as count FROM "prospects" GROUP BY "status" ORDER BY "status" ASC`,
		},
		{
			name:    "prospects groupBy=verified",
			table:   "prospects",
			allowed: prospects,
			groupBy: []string{"verified"},
			wantSQL: `SELECT "verified" AS "values", COUNT(*) as count FROM "prospects" GROUP BY "verified" ORDER BY "verified" ASC`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := testDialect.From(tc.table)
			query := &rql.Query{GroupBy: tc.groupBy}

			result, err := AddGroupInQuery(base, query, tc.allowed)
			assert.NoError(t, err)

			sql, _, sqlErr := result.ToSQL()
			assert.NoError(t, sqlErr)
			assert.Equal(t, tc.wantSQL, sql)
		})
	}
}

func TestBuildGroupByColumns(t *testing.T) {
	tests := []struct {
		name    string
		columns []string
		wantLen int
	}{
		{name: "single column", columns: []string{"event"}, wantLen: 1},
		{name: "multiple columns", columns: []string{"event", "actor_type", "status"}, wantLen: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exprs := buildGroupByColumns(tc.columns)
			assert.Len(t, exprs, tc.wantLen)
		})
	}
}

func TestBuildSelectColumns(t *testing.T) {
	tests := []struct {
		name        string
		columns     []string
		wantValueIn string
	}{
		{
			name:        "single column uses goqu.C",
			columns:     []string{"event"},
			wantValueIn: `"event" AS "values"`,
		},
		{
			name:        "two columns use CONCAT with quoted identifiers",
			columns:     []string{"event", "actor_type"},
			wantValueIn: `CONCAT("event", ',', "actor_type") AS "values"`,
		},
		{
			name:        "three or more columns fall back to first column only",
			columns:     []string{"event", "actor_type", "status"},
			wantValueIn: `"event" AS "values"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exprs := buildSelectColumns(tc.columns)
			assert.Len(t, exprs, 2)

			sql, _, err := testDialect.From("t").Select(exprs...).ToSQL()
			assert.NoError(t, err)
			assert.Contains(t, sql, tc.wantValueIn)
			assert.Contains(t, sql, "COUNT(*) as count")
		})
	}
}
