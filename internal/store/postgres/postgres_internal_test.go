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
