package postgres

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
)

func TestSoftDelete(t *testing.T) {
	tests := []struct {
		name       string
		stmt       *goqu.UpdateDataset
		wantSQL    string
		wantParams []any
	}{
		{
			name:       "marks the live rows of the table",
			stmt:       softDelete(TABLE_DOMAINS),
			wantSQL:    `UPDATE "domains" SET "deleted_at"=now() WHERE ("domains"."deleted_at" IS NULL)`,
			wantParams: []any{},
		},
		{
			name:       "joins the caller's filter with AND",
			stmt:       softDelete(TABLE_DOMAINS).Where(goqu.Ex{"id": "8f2d5b1e-3c4a-4d6b-9e7f-1a2b3c4d5e6f"}),
			wantSQL:    `UPDATE "domains" SET "deleted_at"=now() WHERE (("domains"."deleted_at" IS NULL) AND ("id" = $1))`,
			wantParams: []any{"8f2d5b1e-3c4a-4d6b-9e7f-1a2b3c4d5e6f"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotParams, err := tt.stmt.Prepared(true).ToSQL()
			assert.NoError(t, err)
			assert.Equal(t, tt.wantSQL, gotSQL)
			assert.Equal(t, tt.wantParams, gotParams)
		})
	}
}

func TestLiveConflictTarget(t *testing.T) {
	tests := []struct {
		name    string
		columns string
		wantSQL string
	}{
		{
			name:    "one column",
			columns: "urn",
			wantSQL: `INSERT INTO "resources" ("urn") VALUES ($1) ON CONFLICT (urn) WHERE (deleted_at IS NULL) DO UPDATE SET "name"=$2`,
		},
		{
			name:    "two columns",
			columns: "org_id, name",
			wantSQL: `INSERT INTO "resources" ("urn") VALUES ($1) ON CONFLICT (org_id, name) WHERE (deleted_at IS NULL) DO UPDATE SET "name"=$2`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotParams, err := dialect.Insert(TABLE_RESOURCES).
				Rows(goqu.Record{"urn": "frn:proj:ns:res"}).
				OnConflict(goqu.DoUpdate(liveConflictTarget(tt.columns), goqu.Record{"name": "renamed"})).
				Prepared(true).ToSQL()
			assert.NoError(t, err)
			assert.Equal(t, tt.wantSQL, gotSQL)
			assert.Equal(t, []any{"frn:proj:ns:res", "renamed"}, gotParams)
		})
	}
}
