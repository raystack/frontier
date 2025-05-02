package postgres

import (
	"testing"

	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

func TestUserProjectsRepository_prepareDataQuery(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		orgID     string
		rql       *rql.Query
		wantSQL   string
		wantArgs  []interface{}
		wantError bool
	}{
		{
			name:   "should build valid query with offset and limit",
			userID: "user123",
			orgID:  "org456",
			rql: &rql.Query{
				Offset: 1,
				Limit:  10,
			},
			wantSQL:   `SELECT "p"."id" AS "project_id", "p"."title" AS "project_title", "p"."name" AS "project_name", "p"."created_at" AS "project_created_on", array_agg(DISTINCT u.id ORDER BY u.id) AS "user_ids", array_agg(DISTINCT u.avatar ORDER BY u.avatar) AS "user_avatars", array_agg(DISTINCT u.name ORDER BY u.name) AS "user_names", array_agg(DISTINCT u.title ORDER BY u.title) AS "user_titles" FROM "projects" AS "p" INNER JOIN "policies" AS "pol" ON (("p"."id" = "pol"."resource_id") AND ("pol"."resource_type" = $1) AND ("pol"."deleted_at" IS NULL)) INNER JOIN "users" AS "u" ON (("pol"."principal_id" = "u"."id") AND ("pol"."principal_type" = $2)) WHERE ("p"."id" IN ((SELECT "p2"."id" FROM "projects" AS "p2" INNER JOIN "policies" AS "pol2" ON ("p2"."id" = "pol2"."resource_id") WHERE (("p2"."org_id" = $3) AND ("pol2"."principal_id" = $4) AND ("pol2"."resource_type" = $5) AND ("pol2"."principal_type" = $6) AND ("pol2"."deleted_at" IS NULL))))) GROUP BY "p"."id", "p"."name", "p"."created_at" ORDER BY "p"."name" ASC LIMIT $7 OFFSET $8`,
			wantArgs:  []interface{}{"app/project", "app/user", "org456", "user123", "app/project", "app/user", int64(10), int64(1)},
			wantError: false,
		},
		{
			name:   "should build valid query with search",
			userID: "user123",
			orgID:  "org456",
			rql: &rql.Query{
				Offset: 1,
				Limit:  10,
				Search: "test",
			},
			wantSQL:   `SELECT * FROM (SELECT "p"."id" AS "project_id", "p"."title" AS "project_title", "p"."name" AS "project_name", "p"."created_at" AS "project_created_on", array_agg(DISTINCT u.id ORDER BY u.id) AS "user_ids", array_agg(DISTINCT u.avatar ORDER BY u.avatar) AS "user_avatars", array_agg(DISTINCT u.name ORDER BY u.name) AS "user_names", array_agg(DISTINCT u.title ORDER BY u.title) AS "user_titles" FROM "projects" AS "p" INNER JOIN "policies" AS "pol" ON (("p"."id" = "pol"."resource_id") AND ("pol"."resource_type" = $1) AND ("pol"."deleted_at" IS NULL)) INNER JOIN "users" AS "u" ON (("pol"."principal_id" = "u"."id") AND ("pol"."principal_type" = $2)) WHERE ("p"."id" IN ((SELECT "p2"."id" FROM "projects" AS "p2" INNER JOIN "policies" AS "pol2" ON ("p2"."id" = "pol2"."resource_id") WHERE (("p2"."org_id" = $3) AND ("pol2"."principal_id" = $4) AND ("pol2"."resource_type" = $5) AND ("pol2"."principal_type" = $6) AND ("pol2"."deleted_at" IS NULL))))) GROUP BY "p"."id", "p"."name", "p"."created_at" ORDER BY "p"."name" ASC) AS "base" WHERE (("base"."project_title" ILIKE $7) OR ("base"."project_name" ILIKE $8)) LIMIT $9 OFFSET $10`,
			wantArgs:  []interface{}{"app/project", "app/user", "org456", "user123", "app/project", "app/user", "%test%", "%test%", int64(10), int64(1)},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &UserProjectsRepository{
				dbc: nil,
			}

			query, err := r.prepareDataQuery(tt.userID, tt.orgID, tt.rql)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			sql, args, err := query.ToSQL()
			assert.NoError(t, err)
			assert.Equal(t, tt.wantSQL, sql)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}
