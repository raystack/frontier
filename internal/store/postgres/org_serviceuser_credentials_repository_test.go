package postgres

import (
	"testing"

	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

func TestOrgServiceUserCredentialsRepository_prepareDataQuery(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		rql        *rql.Query
		wantSQL    string
		wantParams []interface{}
		wantErr    bool
	}{
		{
			name:  "basic query without filters",
			orgID: "org1",
			rql: &rql.Query{
				Limit:  10,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceuser_credentials"."title" AS "credential_title", "serviceusers"."title" AS "serviceuser_title", "serviceuser_credentials"."created_at" AS "credential_created_at", "serviceusers"."org_id" AS "org_id" FROM "serviceuser_credentials" INNER JOIN "serviceusers" ON ("serviceuser_credentials"."serviceuser_id" = "serviceusers"."id") WHERE ("serviceusers"."org_id" = $1) LIMIT $2 OFFSET $3`,
			wantParams: []interface{}{
				"org1",    // org_id
				int64(10), // limit
				int64(5),  // offset
			},
			wantErr: false,
		},
		{
			name:  "query with search",
			orgID: "org1",
			rql: &rql.Query{
				Search: "test",
				Limit:  10,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceuser_credentials"."title" AS "credential_title", "serviceusers"."title" AS "serviceuser_title", "serviceuser_credentials"."created_at" AS "credential_created_at", "serviceusers"."org_id" AS "org_id" FROM "serviceuser_credentials" INNER JOIN "serviceusers" ON ("serviceuser_credentials"."serviceuser_id" = "serviceusers"."id") WHERE (("serviceusers"."org_id" = $1) AND ((CAST("serviceuser_credentials"."title" AS TEXT) ILIKE $2) OR (CAST("serviceusers"."title" AS TEXT) ILIKE $3))) LIMIT $4 OFFSET $5`,
			wantParams: []interface{}{
				"org1",    // org_id
				"%test%",  // search pattern for title
				"%test%",  // search pattern for serviceuser_title
				int64(10), // limit
				int64(5),  // offset
			},
			wantErr: false,
		},
		{
			name:  "query with filter",
			orgID: "org1",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "title",
						Operator: "eq",
						Value:    "test-title",
					},
				},
				Limit:  10,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceuser_credentials"."title" AS "credential_title", "serviceusers"."title" AS "serviceuser_title", "serviceuser_credentials"."created_at" AS "credential_created_at", "serviceusers"."org_id" AS "org_id" FROM "serviceuser_credentials" INNER JOIN "serviceusers" ON ("serviceuser_credentials"."serviceuser_id" = "serviceusers"."id") WHERE (("serviceusers"."org_id" = $1) AND ("serviceuser_credentials"."title" = $2)) LIMIT $3 OFFSET $4`,
			wantParams: []interface{}{
				"org1",       // org_id
				"test-title", // filter value
				int64(10),    // limit
				int64(5),     // offset
			},
			wantErr: false,
		},
		{
			name:  "query with valid sort",
			orgID: "org1",
			rql: &rql.Query{
				Sort: []rql.Sort{
					{
						Name:  "title",
						Order: "desc",
					},
				},
				Limit:  10,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceuser_credentials"."title" AS "credential_title", "serviceusers"."title" AS "serviceuser_title", "serviceuser_credentials"."created_at" AS "credential_created_at", "serviceusers"."org_id" AS "org_id" FROM "serviceuser_credentials" INNER JOIN "serviceusers" ON ("serviceuser_credentials"."serviceuser_id" = "serviceusers"."id") WHERE ("serviceusers"."org_id" = $1) ORDER BY "serviceuser_credentials"."title" DESC LIMIT $2 OFFSET $3`,
			wantParams: []interface{}{
				"org1",    // org_id
				int64(10), // limit
				int64(5),  // offset
			},
			wantErr: false,
		},
		{
			name:  "query with invalid sort field",
			orgID: "org1",
			rql: &rql.Query{
				Sort: []rql.Sort{
					{
						Name:  "invalid_field",
						Order: "desc",
					},
				},
				Limit:  10,
				Offset: 5,
			},
			wantSQL:    "",
			wantParams: nil,
			wantErr:    true,
		},
		{
			name:  "query with empty check filter",
			orgID: "org1",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "title",
						Operator: "empty",
					},
				},
				Limit:  10,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceuser_credentials"."title" AS "credential_title", "serviceusers"."title" AS "serviceuser_title", "serviceuser_credentials"."created_at" AS "credential_created_at", "serviceusers"."org_id" AS "org_id" FROM "serviceuser_credentials" INNER JOIN "serviceusers" ON ("serviceuser_credentials"."serviceuser_id" = "serviceusers"."id") WHERE (("serviceusers"."org_id" = $1) AND (("serviceuser_credentials"."title" IS NULL) OR ("serviceuser_credentials"."title" = $2))) LIMIT $3 OFFSET $4`,
			wantParams: []interface{}{
				"org1",    // org_id
				"",        // empty string for comparison
				int64(10), // limit
				int64(5),  // offset
			},
			wantErr: false,
		},
		{
			name:  "complex query with multiple conditions",
			orgID: "org1",
			rql: &rql.Query{
				Search: "test",
				Filters: []rql.Filter{
					{
						Name:     "title",
						Operator: "like",
						Value:    "api",
					},
					{
						Name:     "created_at",
						Operator: "gt",
						Value:    "2023-01-01T00:00:00Z",
					},
				},
				Sort: []rql.Sort{
					{
						Name:  "created_at",
						Order: "desc",
					},
				},
				Limit:  20,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceuser_credentials"."title" AS "credential_title", "serviceusers"."title" AS "serviceuser_title", "serviceuser_credentials"."created_at" AS "credential_created_at", "serviceusers"."org_id" AS "org_id" FROM "serviceuser_credentials" INNER JOIN "serviceusers" ON ("serviceuser_credentials"."serviceuser_id" = "serviceusers"."id") WHERE (("serviceusers"."org_id" = $1) AND ("serviceuser_credentials"."title" LIKE $2) AND ("serviceuser_credentials"."created_at" > $3) AND ((CAST("serviceuser_credentials"."title" AS TEXT) ILIKE $4) OR (CAST("serviceusers"."title" AS TEXT) ILIKE $5))) ORDER BY "serviceuser_credentials"."created_at" DESC LIMIT $6 OFFSET $7`,
			wantParams: []interface{}{
				"org1",                 // org_id
				"%api%",                // like pattern for title
				"2023-01-01T00:00:00Z", // created_at value
				"%test%",               // search pattern for title
				"%test%",               // search pattern for serviceuser_title
				int64(20),              // limit
				int64(5),               // offset
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &OrgServiceUserCredentialsRepository{}
			gotSQL, gotParams, err := r.prepareDataQuery(tt.orgID, tt.rql)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantSQL, gotSQL)
			assert.Equal(t, tt.wantParams, gotParams)
		})
	}
}
