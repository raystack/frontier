package postgres

import (
	"testing"

	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

func TestProjectUsersRepository_PrepareDataQuery(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		rql       *rql.Query
		wantSQL   string
		wantArgs  []interface{}
		wantErr   bool
	}{
		{
			name:      "should return base query when no search is provided",
			projectID: "project-123",
			rql: &rql.Query{
				Limit:  10,
				Offset: 0,
			},
			wantSQL:  `SELECT "users"."id", "users"."name", "users"."email", "users"."title", "users"."avatar", "users"."state", "policies"."resource_id" AS "project_id", MIN("policies"."created_at") AS "project_joined_at", string_agg(DISTINCT roles.name, ',') AS "role_names", string_agg(DISTINCT roles.title, ',') AS "role_titles", string_agg(DISTINCT roles.id::text, ',') AS "role_ids" FROM "policies" INNER JOIN "users" ON ("policies"."principal_id" = "users"."id") INNER JOIN "roles" ON ("policies"."role_id" = "roles"."id") WHERE (("policies"."principal_type" = $1) AND ("policies"."resource_id" = $2) AND ("policies"."resource_type" = $3)) GROUP BY "users"."id", "users"."name", "users"."email", "users"."title", "users"."state", "policies"."resource_id" LIMIT $4`,
			wantArgs: []interface{}{"app/user", "project-123", "app/project", int64(10)},
			wantErr:  false,
		},
		{
			name:      "should return query with search when search is provided",
			projectID: "project-123",
			rql: &rql.Query{
				Search: "john",
				Limit:  10,
				Offset: 0,
			},
			wantSQL:  `SELECT * FROM (SELECT "users"."id", "users"."name", "users"."email", "users"."title", "users"."avatar", "users"."state", "policies"."resource_id" AS "project_id", MIN("policies"."created_at") AS "project_joined_at", string_agg(DISTINCT roles.name, ',') AS "role_names", string_agg(DISTINCT roles.title, ',') AS "role_titles", string_agg(DISTINCT roles.id::text, ',') AS "role_ids" FROM "policies" INNER JOIN "users" ON ("policies"."principal_id" = "users"."id") INNER JOIN "roles" ON ("policies"."role_id" = "roles"."id") WHERE (("policies"."principal_type" = $1) AND ("policies"."resource_id" = $2) AND ("policies"."resource_type" = $3)) GROUP BY "users"."id", "users"."name", "users"."email", "users"."title", "users"."state", "policies"."resource_id") AS "base" WHERE (("base"."name" ILIKE $4) OR ("base"."email" ILIKE $5) OR ("base"."title" ILIKE $6) OR ("base"."state" ILIKE $7) OR ("base"."role_names" ILIKE $8) OR ("base"."role_titles" ILIKE $9) OR ("base"."role_ids" ILIKE $10)) LIMIT $11`,
			wantArgs: []interface{}{"app/user", "project-123", "app/project", "%john%", "%john%", "%john%", "%john%", "%john%", "%john%", "%john%", int64(10)},
			wantErr:  false,
		},
	}

	repo := NewProjectUsersRepository(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSQL, gotArgs, err := repo.prepareDataQuery(tt.projectID, tt.rql)

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
