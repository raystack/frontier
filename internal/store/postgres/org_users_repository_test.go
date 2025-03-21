package postgres

import (
	"testing"

	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

func TestOrgUsersRepository_PrepareDataQuery(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		rqlQuery   *rql.Query
		wantSQL    string
		wantParams []interface{}
		wantErr    bool
	}{
		{
			name:  "basic query without filters",
			orgID: "org123",
			rqlQuery: &rql.Query{
				Limit:  10,
				Offset: 0,
			},
			wantSQL:    `SELECT "policies"."resource_id" AS "org_id", "users"."id" AS "id", "users"."name" AS "name", "users"."title" AS "title", "users"."email" AS "email", "users"."state" AS "state", "users"."avatar" AS "avatar", MIN("policies"."created_at") AS "org_joined_at", ARRAY_AGG("roles"."name") AS "role_names", ARRAY_AGG(COALESCE("roles"."title", '')) AS "role_titles", ARRAY_AGG(CAST("roles"."id" AS TEXT)) AS "role_ids" FROM "policies" INNER JOIN "users" ON ("users"."id" = "policies"."principal_id") LEFT JOIN "roles" ON ("roles"."id" = "policies"."role_id") WHERE (("policies"."resource_id" = $1) AND ("policies"."resource_type" = $2) AND ("policies"."principal_type" = $3) AND ("users"."deleted_at" IS NULL) AND ("roles"."deleted_at" IS NULL)) GROUP BY "policies"."resource_id", "users"."id", "users"."name", "users"."title", "users"."email", "users"."state", "users"."created_at", "users"."updated_at" LIMIT $4`,
			wantParams: []interface{}{"org123", "app/organization", "app/user", int64(10)},
		},
		{
			name:  "query with email filter",
			orgID: "org123",
			rqlQuery: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "email",
						Operator: "eq",
						Value:    "test@example.com",
					},
				},
				Limit:  10,
				Offset: 0,
			},
			wantSQL:    `SELECT "policies"."resource_id" AS "org_id", "users"."id" AS "id", "users"."name" AS "name", "users"."title" AS "title", "users"."email" AS "email", "users"."state" AS "state", "users"."avatar" AS "avatar", MIN("policies"."created_at") AS "org_joined_at", ARRAY_AGG("roles"."name") AS "role_names", ARRAY_AGG(COALESCE("roles"."title", '')) AS "role_titles", ARRAY_AGG(CAST("roles"."id" AS TEXT)) AS "role_ids" FROM "policies" INNER JOIN "users" ON ("users"."id" = "policies"."principal_id") LEFT JOIN "roles" ON ("roles"."id" = "policies"."role_id") WHERE (("policies"."resource_id" = $1) AND ("policies"."resource_type" = $2) AND ("policies"."principal_type" = $3) AND ("users"."deleted_at" IS NULL) AND ("roles"."deleted_at" IS NULL) AND ("users"."email" = $4)) GROUP BY "policies"."resource_id", "users"."id", "users"."name", "users"."title", "users"."email", "users"."state", "users"."created_at", "users"."updated_at" LIMIT $5`,
			wantParams: []interface{}{"org123", "app/organization", "app/user", "test@example.com", int64(10)},
		},
		{
			name:  "query with role filter",
			orgID: "org123",
			rqlQuery: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "role_names",
						Operator: "eq",
						Value:    "admin",
					},
				},
				Limit:  10,
				Offset: 0,
			},
			wantSQL:    `SELECT "policies"."resource_id" AS "org_id", "users"."id" AS "id", "users"."name" AS "name", "users"."title" AS "title", "users"."email" AS "email", "users"."state" AS "state", "users"."avatar" AS "avatar", MIN("policies"."created_at") AS "org_joined_at", ARRAY_AGG("roles"."name") AS "role_names", ARRAY_AGG(COALESCE("roles"."title", '')) AS "role_titles", ARRAY_AGG(CAST("roles"."id" AS TEXT)) AS "role_ids" FROM "policies" INNER JOIN "users" ON ("users"."id" = "policies"."principal_id") LEFT JOIN "roles" ON ("roles"."id" = "policies"."role_id") WHERE (("policies"."resource_id" = $1) AND ("policies"."resource_type" = $2) AND ("policies"."principal_type" = $3) AND ("users"."deleted_at" IS NULL) AND ("roles"."deleted_at" IS NULL) AND EXISTS (SELECT 1 FROM "policies" INNER JOIN "roles" ON ("roles"."id" = "policies"."role_id") WHERE (("policies"."principal_id" = "users"."id") AND ("policies"."resource_id" = $4) AND ("policies"."resource_type" = $5) AND ("roles"."name" = $6)) LIMIT $7)) GROUP BY "policies"."resource_id", "users"."id", "users"."name", "users"."title", "users"."email", "users"."state", "users"."created_at", "users"."updated_at" LIMIT $8`,
			wantParams: []interface{}{"org123", "app/organization", "app/user", "org123", "app/organization", "admin", int64(1), int64(10)},
		},
		{
			name:  "query with search",
			orgID: "org123",
			rqlQuery: &rql.Query{
				Search: "john",
				Limit:  10,
				Offset: 0,
			},
			wantSQL:    `SELECT "policies"."resource_id" AS "org_id", "users"."id" AS "id", "users"."name" AS "name", "users"."title" AS "title", "users"."email" AS "email", "users"."state" AS "state", "users"."avatar" AS "avatar", MIN("policies"."created_at") AS "org_joined_at", ARRAY_AGG("roles"."name") AS "role_names", ARRAY_AGG(COALESCE("roles"."title", '')) AS "role_titles", ARRAY_AGG(CAST("roles"."id" AS TEXT)) AS "role_ids" FROM "policies" INNER JOIN "users" ON ("users"."id" = "policies"."principal_id") LEFT JOIN "roles" ON ("roles"."id" = "policies"."role_id") WHERE (("policies"."resource_id" = $1) AND ("policies"."resource_type" = $2) AND ("policies"."principal_type" = $3) AND ("users"."deleted_at" IS NULL) AND ("roles"."deleted_at" IS NULL) AND ((CAST("users"."name" AS TEXT) ILIKE $4) OR (CAST("users"."title" AS TEXT) ILIKE $5) OR (CAST("users"."email" AS TEXT) ILIKE $6) OR (CAST("users"."state" AS TEXT) ILIKE $7))) GROUP BY "policies"."resource_id", "users"."id", "users"."name", "users"."title", "users"."email", "users"."state", "users"."created_at", "users"."updated_at" LIMIT $8`,
			wantParams: []interface{}{"org123", "app/organization", "app/user", "%john%", "%john%", "%john%", "%john%", int64(10)},
		},
		{
			name:  "query with sort",
			orgID: "org123",
			rqlQuery: &rql.Query{
				Sort: []rql.Sort{
					{Name: "name", Order: "asc"},
					{Name: "email", Order: "desc"},
				},
				Limit:  10,
				Offset: 0,
			},
			wantSQL:    `SELECT "policies"."resource_id" AS "org_id", "users"."id" AS "id", "users"."name" AS "name", "users"."title" AS "title", "users"."email" AS "email", "users"."state" AS "state", "users"."avatar" AS "avatar", MIN("policies"."created_at") AS "org_joined_at", ARRAY_AGG("roles"."name") AS "role_names", ARRAY_AGG(COALESCE("roles"."title", '')) AS "role_titles", ARRAY_AGG(CAST("roles"."id" AS TEXT)) AS "role_ids" FROM "policies" INNER JOIN "users" ON ("users"."id" = "policies"."principal_id") LEFT JOIN "roles" ON ("roles"."id" = "policies"."role_id") WHERE (("policies"."resource_id" = $1) AND ("policies"."resource_type" = $2) AND ("policies"."principal_type" = $3) AND ("users"."deleted_at" IS NULL) AND ("roles"."deleted_at" IS NULL)) GROUP BY "policies"."resource_id", "users"."id", "users"."name", "users"."title", "users"."email", "users"."state", "users"."created_at", "users"."updated_at" ORDER BY "name" ASC, "email" DESC LIMIT $4`,
			wantParams: []interface{}{"org123", "app/organization", "app/user", int64(10)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewOrgUsersRepository(nil)
			gotSQL, gotParams, err := r.prepareDataQuery(tt.orgID, tt.rqlQuery)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantParams, gotParams)
			assert.Equal(t, tt.wantSQL, gotSQL)
		})
	}
}

func TestOrgUsersRepository_BuildNonRoleFilterCondition(t *testing.T) {
	tests := []struct {
		name    string
		filter  rql.Filter
		wantSQL string
		wantErr bool
	}{
		{
			name: "eq operator",
			filter: rql.Filter{
				Name:     "email",
				Operator: "eq",
				Value:    "test@example.com",
			},
			wantSQL: `("users"."email" = 'test@example.com')`,
		},
		{
			name: "like operator",
			filter: rql.Filter{
				Name:     "name",
				Operator: "like",
				Value:    "john",
			},
			wantSQL: `(CAST("users"."name" AS TEXT) ILIKE '%john%')`,
		},
		{
			name: "notlike operator",
			filter: rql.Filter{
				Name:     "name",
				Operator: "notlike",
				Value:    "john",
			},
			wantSQL: `(CAST("users"."name" AS TEXT) NOT ILIKE '%john%')`,
		},
		{
			name: "in operator",
			filter: rql.Filter{
				Name:     "state",
				Operator: "in",
				Value:    "active,inactive",
			},
			wantSQL: `("users"."state" IN ('active', 'inactive'))`,
		},
		{
			name: "empty operator",
			filter: rql.Filter{
				Name:     "title",
				Operator: "empty",
			},
			wantSQL: `coalesce(users.title, '') = ''`,
		},
		{
			name: "invalid operator",
			filter: rql.Filter{
				Name:     "email",
				Operator: "invalid",
				Value:    "test@example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewOrgUsersRepository(nil)
			got, err := r.buildNonRoleFilterCondition(tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			sql, _, err := dialect.From("dummy").Where(got).ToSQL()
			assert.NoError(t, err)
			// Remove the "SELECT * FROM "dummy" WHERE" part from the generated SQL
			actualSQL := strings.TrimPrefix(sql, `SELECT * FROM "dummy" WHERE `)
			assert.Equal(t, tt.wantSQL, actualSQL)
		})
	}
}

func TestOrgUsersRepository_BuildRoleFilterCondition(t *testing.T) {
	tests := []struct {
		name    string
		orgID   string
		filter  rql.Filter
		wantSQL string
		wantErr bool
	}{
		{
			name:  "eq operator",
			orgID: "org123",
			filter: rql.Filter{
				Name:     "role_names",
				Operator: "eq",
				Value:    "admin",
			},
			wantSQL: `EXISTS (SELECT 1 FROM "policies" INNER JOIN "roles" ON ("roles"."id" = "policies"."role_id") WHERE (("policies"."principal_id" = "users"."id") AND ("policies"."resource_id" = 'org123') AND ("policies"."resource_type" = 'app/organization') AND ("roles"."name" = 'admin')) LIMIT 1)`,
		},
		{
			name:  "neq operator",
			orgID: "org123",
			filter: rql.Filter{
				Name:     "role_names",
				Operator: "neq",
				Value:    "admin",
			},
			wantSQL: `(NOT EXISTS (SELECT 1 FROM "policies" INNER JOIN "roles" ON ("roles"."id" = "policies"."role_id") WHERE (("policies"."principal_id" = "users"."id") AND ("policies"."resource_id" = 'org123') AND ("policies"."resource_type" = 'app/organization') AND ("roles"."name" = 'admin')) LIMIT 1) AND EXISTS (SELECT 1 FROM "policies" INNER JOIN "roles" ON ("roles"."id" = "policies"."role_id") WHERE (("policies"."principal_id" = "users"."id") AND ("policies"."resource_id" = 'org123') AND ("policies"."resource_type" = 'app/organization')) LIMIT 1))`,
		},
		{
			name:  "invalid operator",
			orgID: "org123",
			filter: rql.Filter{
				Name:     "role_names",
				Operator: "like",
				Value:    "admin",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewOrgUsersRepository(nil)
			got, err := r.buildRoleFilterCondition(tt.orgID, tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			sql, _, err := dialect.From("dummy").Where(goqu.And(got...)).ToSQL()
			assert.NoError(t, err)
			// Remove the "SELECT * FROM "dummy" WHERE" part from the generated SQL
			actualSQL := strings.TrimPrefix(sql, `SELECT * FROM "dummy" WHERE `)
			assert.Equal(t, tt.wantSQL, actualSQL)
		})
	}
}
