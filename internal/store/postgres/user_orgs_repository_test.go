package postgres

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserOrgsRepository_buildBaseQuery(t *testing.T) {
	tests := []struct {
		name        string
		principalID string
		wantSQL     string
		wantParams  []interface{}
		wantErr     bool
	}{
		{
			name:        "should build query for valid principal id",
			principalID: "test-user-id",
			wantSQL:     `SELECT "policies"."principal_id", "policies"."resource_id" AS "org_id", "organizations"."name" AS "org_name", "organizations"."title" AS "org_title", "organizations"."avatar" AS "org_avatar", MIN("policies"."created_at") AS "org_joined_on", ARRAY_AGG("roles"."name") AS "role_names", ARRAY_AGG("roles"."title") AS "role_titles", ARRAY_AGG("roles"."id") AS "role_ids", COALESCE("project_counts"."project_count", $1) AS "project_count" FROM "policies" INNER JOIN "roles" ON ("policies"."role_id" = "roles"."id") INNER JOIN "organizations" ON ("policies"."resource_id" = "organizations"."id") INNER JOIN "users" ON ("policies"."principal_id" = "users"."id") LEFT JOIN (SELECT "org_id", COUNT("id") AS "project_count" FROM "projects" WHERE (("deleted_at" IS NULL) AND ("state" = $2)) GROUP BY "org_id") AS "project_counts" ON ("project_counts"."org_id" = "organizations"."id") WHERE (("policies"."resource_type" = $3) AND ("policies"."principal_id" = $4)) GROUP BY "policies"."principal_id", "users"."email", "policies"."resource_id", "organizations"."name", "organizations"."title", "organizations"."avatar", "project_counts"."project_count" ORDER BY "organizations"."name" ASC`,
			wantParams:  []interface{}{int64(0), "enabled", "app/organization", "test-user-id"},
			wantErr:     false,
		},
		{
			name:        "should build query for empty principal id",
			principalID: "",
			wantSQL:     `SELECT "policies"."principal_id", "policies"."resource_id" AS "org_id", "organizations"."name" AS "org_name", "organizations"."title" AS "org_title", "organizations"."avatar" AS "org_avatar", MIN("policies"."created_at") AS "org_joined_on", ARRAY_AGG("roles"."name") AS "role_names", ARRAY_AGG("roles"."title") AS "role_titles", ARRAY_AGG("roles"."id") AS "role_ids", COALESCE("project_counts"."project_count", $1) AS "project_count" FROM "policies" INNER JOIN "roles" ON ("policies"."role_id" = "roles"."id") INNER JOIN "organizations" ON ("policies"."resource_id" = "organizations"."id") INNER JOIN "users" ON ("policies"."principal_id" = "users"."id") LEFT JOIN (SELECT "org_id", COUNT("id") AS "project_count" FROM "projects" WHERE (("deleted_at" IS NULL) AND ("state" = $2)) GROUP BY "org_id") AS "project_counts" ON ("project_counts"."org_id" = "organizations"."id") WHERE (("policies"."resource_type" = $3) AND ("policies"."principal_id" = $4)) GROUP BY "policies"."principal_id", "users"."email", "policies"."resource_id", "organizations"."name", "organizations"."title", "organizations"."avatar", "project_counts"."project_count" ORDER BY "organizations"."name" ASC`,
			wantParams:  []interface{}{int64(0), "enabled", "app/organization", ""},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := UserOrgsRepository{}

			gotSQL, gotParams, err := r.buildBaseQuery(tt.principalID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantParams, gotParams)
			assert.Equal(t, strings.Join(strings.Fields(tt.wantSQL), " "), strings.Join(strings.Fields(gotSQL), " "))
		})
	}
}
