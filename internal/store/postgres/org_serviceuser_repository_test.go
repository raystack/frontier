package postgres

import (
	"testing"

	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

func TestOrgServiceUserRepository_prepareDataQuery(t *testing.T) {
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
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE ("serviceusers"."org_id" = $3) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC LIMIT $4 OFFSET $5`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
				int64(10),         // limit
				int64(5),          // offset
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
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE (("serviceusers"."org_id" = $3) AND (CAST("serviceusers"."title" AS TEXT) ILIKE $4)) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC LIMIT $5 OFFSET $6`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
				"%test%",          // search pattern for title
				int64(10),         // limit
				int64(5),          // offset
			},
			wantErr: false,
		},
		{
			name:  "query with title filter",
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
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE (("serviceusers"."org_id" = $3) AND ("serviceusers"."title" = $4)) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC LIMIT $5 OFFSET $6`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
				"test-title",      // filter value
				int64(10),         // limit
				int64(5),          // offset
			},
			wantErr: false,
		},
		{
			name:  "query with like filter",
			orgID: "org1",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "title",
						Operator: "like",
						Value:    "api",
					},
				},
				Limit:  10,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE (("serviceusers"."org_id" = $3) AND ("serviceusers"."title" LIKE $4)) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC LIMIT $5 OFFSET $6`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
				"%api%",           // like pattern for title
				int64(10),         // limit
				int64(5),          // offset
			},
			wantErr: false,
		},
		{
			name:  "query with created_at filter",
			orgID: "org1",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "created_at",
						Operator: "gt",
						Value:    "2023-01-01T00:00:00Z",
					},
				},
				Limit:  10,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE (("serviceusers"."org_id" = $3) AND ("serviceusers"."created_at" > $4)) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC LIMIT $5 OFFSET $6`,
			wantParams: []interface{}{
				"app/serviceuser",      // principal_type
				"app/project",          // resource_type
				"org1",                 // org_id
				"2023-01-01T00:00:00Z", // created_at value
				int64(10),              // limit
				int64(5),               // offset
			},
			wantErr: false,
		},
		{
			name:  "query with valid sort by title desc",
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
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE ("serviceusers"."org_id" = $3) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC, "serviceusers"."title" DESC LIMIT $4 OFFSET $5`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
				int64(10),         // limit
				int64(5),          // offset
			},
			wantErr: false,
		},
		{
			name:  "query with valid sort by created_at asc",
			orgID: "org1",
			rql: &rql.Query{
				Sort: []rql.Sort{
					{
						Name:  "created_at",
						Order: "asc",
					},
				},
				Limit:  10,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE ("serviceusers"."org_id" = $3) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC, "serviceusers"."created_at" ASC LIMIT $4 OFFSET $5`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
				int64(10),         // limit
				int64(5),          // offset
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
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE (("serviceusers"."org_id" = $3) AND (("serviceusers"."title" IS NULL) OR ("serviceusers"."title" = $4))) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC LIMIT $5 OFFSET $6`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
				"",                // empty string for comparison
				int64(10),         // limit
				int64(5),          // offset
			},
			wantErr: false,
		},
		{
			name:  "query with notempty check filter",
			orgID: "org1",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "title",
						Operator: "notempty",
					},
				},
				Limit:  10,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE (("serviceusers"."org_id" = $3) AND (("serviceusers"."title" IS NOT NULL) AND ("serviceusers"."title" != $4))) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC LIMIT $5 OFFSET $6`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
				"",                // empty string for comparison
				int64(10),         // limit
				int64(5),          // offset
			},
			wantErr: false,
		},
		{
			name:  "query with notlike filter",
			orgID: "org1",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "title",
						Operator: "notlike",
						Value:    "test",
					},
				},
				Limit:  10,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE (("serviceusers"."org_id" = $3) AND ("serviceusers"."title" NOT LIKE $4)) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC LIMIT $5 OFFSET $6`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
				"%test%",          // notlike pattern for title
				int64(10),         // limit
				int64(5),          // offset
			},
			wantErr: false,
		},
		{
			name:  "query with unknown filter field (should be ignored)",
			orgID: "org1",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "unknown_field",
						Operator: "eq",
						Value:    "value",
					},
				},
				Limit:  10,
				Offset: 5,
			},
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE ("serviceusers"."org_id" = $3) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC LIMIT $4 OFFSET $5`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
				int64(10),         // limit
				int64(5),          // offset
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
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE (("serviceusers"."org_id" = $3) AND ("serviceusers"."title" LIKE $4) AND ("serviceusers"."created_at" > $5) AND (CAST("serviceusers"."title" AS TEXT) ILIKE $6)) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC, "serviceusers"."created_at" DESC LIMIT $7 OFFSET $8`,
			wantParams: []interface{}{
				"app/serviceuser",      // principal_type
				"app/project",          // resource_type
				"org1",                 // org_id
				"%api%",                // like pattern for title
				"2023-01-01T00:00:00Z", // created_at value
				"%test%",               // search pattern for title
				int64(20),              // limit
				int64(5),               // offset
			},
			wantErr: false,
		},
		{
			name:  "query with no limit or offset",
			orgID: "org1",
			rql: &rql.Query{
				Search: "test",
			},
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE (("serviceusers"."org_id" = $3) AND (CAST("serviceusers"."title" AS TEXT) ILIKE $4)) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
				"%test%",          // search pattern for title
			},
			wantErr: false,
		},
		{
			name:    "query with nil rql",
			orgID:   "org1",
			rql:     nil,
			wantSQL: `SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", STRING_AGG(projects.title, ', ') AS "project_titles" FROM "serviceusers" INNER JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2)) INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") WHERE ("serviceusers"."org_id" = $3) GROUP BY "serviceusers"."id" ORDER BY "serviceusers"."title" ASC`,
			wantParams: []interface{}{
				"app/serviceuser", // principal_type
				"app/project",     // resource_type
				"org1",            // org_id
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &OrgServiceUserRepository{}
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
