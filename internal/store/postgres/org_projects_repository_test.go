package postgres

import (
	"testing"

	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

func TestOrgProjectsRepository_prepareDataQuery(t *testing.T) {
	tests := []struct {
		name     string
		orgID    string
		rqlQuery *rql.Query
		wantSQL  string
		wantArgs []interface{}
		wantErr  bool
	}{
		{
			name:  "basic query without filters",
			orgID: "org123",
			rqlQuery: &rql.Query{
				Limit:  10,
				Offset: 0,
			},
			wantSQL:  `SELECT "projects"."id", "projects"."name", "projects"."title", "projects"."state", "projects"."created_at", "projects"."org_id", COUNT(DISTINCT("policies"."principal_id")) AS "member_count", array_agg(DISTINCT users.id) AS "user_ids" FROM "policies" INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") INNER JOIN "users" ON ("policies"."principal_id" = "users"."id") WHERE (("principal_type" = $1) AND ("projects"."org_id" = $2)) GROUP BY "projects"."id", "projects"."name", "projects"."title", "projects"."state", "projects"."created_at", "projects"."org_id" LIMIT $3`,
			wantArgs: []interface{}{"app/user", "org123", int64(10)},
			wantErr:  false,
		},
		{
			name:  "query with string filter",
			orgID: "org123",
			rqlQuery: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "name",
						Operator: "eq",
						Value:    "test-project",
					},
				},
				Limit:  10,
				Offset: 0,
			},
			wantSQL:  `SELECT "projects"."id", "projects"."name", "projects"."title", "projects"."state", "projects"."created_at", "projects"."org_id", COUNT(DISTINCT("policies"."principal_id")) AS "member_count", array_agg(DISTINCT users.id) AS "user_ids" FROM "policies" INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") INNER JOIN "users" ON ("policies"."principal_id" = "users"."id") WHERE ((("principal_type" = $1) AND ("projects"."org_id" = $2)) AND ("projects"."name" = $3)) GROUP BY "projects"."id", "projects"."name", "projects"."title", "projects"."state", "projects"."created_at", "projects"."org_id" LIMIT $4`,
			wantArgs: []interface{}{"app/user", "org123", "test-project", int64(10)},
			wantErr:  false,
		},
		{
			name:  "query with datetime filter",
			orgID: "org123",
			rqlQuery: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "created_at",
						Operator: "gt",
						Value:    "2023-11-02T12:10:21.470756Z",
					},
				},
				Limit:  10,
				Offset: 0,
			},
			wantSQL:  `SELECT "projects"."id", "projects"."name", "projects"."title", "projects"."state", "projects"."created_at", "projects"."org_id", COUNT(DISTINCT("policies"."principal_id")) AS "member_count", array_agg(DISTINCT users.id) AS "user_ids" FROM "policies" INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") INNER JOIN "users" ON ("policies"."principal_id" = "users"."id") WHERE ((("principal_type" = $1) AND ("projects"."org_id" = $2)) AND ("projects"."created_at" > timestamp '2023-11-02T12:10:21.470756Z')) GROUP BY "projects"."id", "projects"."name", "projects"."title", "projects"."state", "projects"."created_at", "projects"."org_id" LIMIT $3`,
			wantArgs: []interface{}{"app/user", "org123", int64(10)},
			wantErr:  false,
		},
		{
			name:  "query with search",
			orgID: "org123",
			rqlQuery: &rql.Query{
				Search: "test",
				Limit:  10,
			},
			wantSQL:  `SELECT "projects"."id", "projects"."name", "projects"."title", "projects"."state", "projects"."created_at", "projects"."org_id", COUNT(DISTINCT("policies"."principal_id")) AS "member_count", array_agg(DISTINCT users.id) AS "user_ids" FROM "policies" INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") INNER JOIN "users" ON ("policies"."principal_id" = "users"."id") WHERE ((("principal_type" = $1) AND ("projects"."org_id" = $2)) AND (("projects"."title" ILIKE $3) OR ("projects"."name" ILIKE $4) OR ("projects"."state" ILIKE $5))) GROUP BY "projects"."id", "projects"."name", "projects"."title", "projects"."state", "projects"."created_at", "projects"."org_id" LIMIT $6`,
			wantArgs: []interface{}{"app/user", "org123", "%test%", "%test%", "%test%", int64(10)},
			wantErr:  false,
		},
		{
			name:  "query with sorting",
			orgID: "org123",
			rqlQuery: &rql.Query{
				Sort: []rql.Sort{
					{
						Name:  "created_at",
						Order: "desc",
					},
				},
				Limit: 10,
			},
			wantSQL:  `SELECT "projects"."id", "projects"."name", "projects"."title", "projects"."state", "projects"."created_at", "projects"."org_id", COUNT(DISTINCT("policies"."principal_id")) AS "member_count", array_agg(DISTINCT users.id) AS "user_ids" FROM "policies" INNER JOIN "projects" ON ("policies"."resource_id" = "projects"."id") INNER JOIN "users" ON ("policies"."principal_id" = "users"."id") WHERE (("principal_type" = $1) AND ("projects"."org_id" = $2)) GROUP BY "projects"."id", "projects"."name", "projects"."title", "projects"."state", "projects"."created_at", "projects"."org_id" ORDER BY "created_at" DESC LIMIT $3`,
			wantArgs: []interface{}{"app/user", "org123", int64(10)},
			wantErr:  false,
		},
		{
			name:  "invalid filter field",
			orgID: "org123",
			rqlQuery: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "invalid_field",
						Operator: "eq",
						Value:    "test",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &OrgProjectsRepository{
				dbc: nil, // not needed for this test
			}

			gotSQL, gotArgs, err := r.prepareDataQuery(tt.orgID, tt.rqlQuery)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantSQL, gotSQL)
			assert.Equal(t, tt.wantArgs, gotArgs)
		})
	}
}
