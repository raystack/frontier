package postgres

import (
	"testing"

	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

func TestOrgTokensRepository_prepareDataQuery(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		rql        *rql.Query
		wantSQL    string
		wantParams []interface{}
		wantErr    bool
	}{
		{
			name:  "basic query with pagination",
			orgID: "org123",
			rql: &rql.Query{
				Limit:  10,
				Offset: 20,
			},
			wantSQL:    `SELECT "billing_transactions"."amount" AS "token_amount", "billing_transactions"."type" AS "token_type", "billing_transactions"."description" AS "token_description", "billing_transactions"."user_id" AS "token_user_id", "users"."title" AS "user_title", "users"."avatar" AS "user_avatar", "billing_transactions"."created_at" AS "token_created_at", "billing_customers"."org_id" AS "org_id" FROM "billing_transactions" INNER JOIN "billing_customers" ON ("billing_transactions"."account_id" = "billing_customers"."id") LEFT JOIN "users" ON CASE WHEN "billing_transactions"."user_id" IS NOT NULL AND "billing_transactions"."user_id" != '' THEN CAST("billing_transactions"."user_id" AS uuid) = "users"."id" ELSE false END WHERE ("billing_customers"."org_id" = $1) LIMIT $2 OFFSET $3`,
			wantParams: []interface{}{"org123", int64(10), int64(20)},
			wantErr:    false,
		},
		{
			name:  "query with amount filter",
			orgID: "org123",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "amount",
						Operator: "gte",
						Value:    1000,
					},
				},
				Limit:  10,
				Offset: 30,
			},
			wantSQL:    `SELECT "billing_transactions"."amount" AS "token_amount", "billing_transactions"."type" AS "token_type", "billing_transactions"."description" AS "token_description", "billing_transactions"."user_id" AS "token_user_id", "users"."title" AS "user_title", "users"."avatar" AS "user_avatar", "billing_transactions"."created_at" AS "token_created_at", "billing_customers"."org_id" AS "org_id" FROM "billing_transactions" INNER JOIN "billing_customers" ON ("billing_transactions"."account_id" = "billing_customers"."id") LEFT JOIN "users" ON CASE WHEN "billing_transactions"."user_id" IS NOT NULL AND "billing_transactions"."user_id" != '' THEN CAST("billing_transactions"."user_id" AS uuid) = "users"."id" ELSE false END WHERE (("billing_customers"."org_id" = $1) AND ("billing_transactions"."amount" >= $2)) LIMIT $3 OFFSET $4`,
			wantParams: []interface{}{"org123", int64(1000), int64(10), int64(30)},
			wantErr:    false,
		},
		{
			name:  "query with type filter and search",
			orgID: "org123",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "type",
						Operator: "eq",
						Value:    "credit",
					},
				},
				Search: "test",
				Limit:  10,
				Offset: 40,
			},
			wantSQL:    `SELECT "billing_transactions"."amount" AS "token_amount", "billing_transactions"."type" AS "token_type", "billing_transactions"."description" AS "token_description", "billing_transactions"."user_id" AS "token_user_id", "users"."title" AS "user_title", "users"."avatar" AS "user_avatar", "billing_transactions"."created_at" AS "token_created_at", "billing_customers"."org_id" AS "org_id" FROM "billing_transactions" INNER JOIN "billing_customers" ON ("billing_transactions"."account_id" = "billing_customers"."id") LEFT JOIN "users" ON CASE WHEN "billing_transactions"."user_id" IS NOT NULL AND "billing_transactions"."user_id" != '' THEN CAST("billing_transactions"."user_id" AS uuid) = "users"."id" ELSE false END WHERE (("billing_customers"."org_id" = $1) AND ("billing_transactions"."type" = $2) AND ((CAST("billing_transactions"."type" AS TEXT) ILIKE $3) OR (CAST("billing_transactions"."description" AS TEXT) ILIKE $4) OR (CAST("users"."title" AS TEXT) ILIKE $5) OR (CAST("billing_transactions"."amount" AS TEXT) ILIKE $6))) LIMIT $7 OFFSET $8`,
			wantParams: []interface{}{"org123", "credit", "%test%", "%test%", "%test%", "%test%", int64(10), int64(40)},
			wantErr:    false,
		},
		{
			name:  "query with date filter and sort",
			orgID: "org123",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "created_at",
						Operator: "gte",
						Value:    "2024-01-01T00:00:00Z",
					},
				},
				Sort: []rql.Sort{
					{
						Name:  "created_at",
						Order: "desc",
					},
				},
				Limit:  10,
				Offset: 50,
			},
			wantSQL:    `SELECT "billing_transactions"."amount" AS "token_amount", "billing_transactions"."type" AS "token_type", "billing_transactions"."description" AS "token_description", "billing_transactions"."user_id" AS "token_user_id", "users"."title" AS "user_title", "users"."avatar" AS "user_avatar", "billing_transactions"."created_at" AS "token_created_at", "billing_customers"."org_id" AS "org_id" FROM "billing_transactions" INNER JOIN "billing_customers" ON ("billing_transactions"."account_id" = "billing_customers"."id") LEFT JOIN "users" ON CASE WHEN "billing_transactions"."user_id" IS NOT NULL AND "billing_transactions"."user_id" != '' THEN CAST("billing_transactions"."user_id" AS uuid) = "users"."id" ELSE false END WHERE (("billing_customers"."org_id" = $1) AND ("billing_transactions"."created_at" >= $2)) ORDER BY "billing_transactions"."created_at" DESC LIMIT $3 OFFSET $4`,
			wantParams: []interface{}{"org123", "2024-01-01T00:00:00Z", int64(10), int64(50)},
			wantErr:    false,
		},
		{
			name:  "query with multiple sorts",
			orgID: "org123",
			rql: &rql.Query{
				Sort: []rql.Sort{
					{
						Name:  "type",
						Order: "desc",
					},
					{
						Name:  "user_title",
						Order: "asc",
					},
				},
				Limit:  10,
				Offset: 25,
			},
			wantSQL:    `SELECT "billing_transactions"."amount" AS "token_amount", "billing_transactions"."type" AS "token_type", "billing_transactions"."description" AS "token_description", "billing_transactions"."user_id" AS "token_user_id", "users"."title" AS "user_title", "users"."avatar" AS "user_avatar", "billing_transactions"."created_at" AS "token_created_at", "billing_customers"."org_id" AS "org_id" FROM "billing_transactions" INNER JOIN "billing_customers" ON ("billing_transactions"."account_id" = "billing_customers"."id") LEFT JOIN "users" ON CASE WHEN "billing_transactions"."user_id" IS NOT NULL AND "billing_transactions"."user_id" != '' THEN CAST("billing_transactions"."user_id" AS uuid) = "users"."id" ELSE false END WHERE ("billing_customers"."org_id" = $1) ORDER BY "billing_transactions"."type" DESC, "users"."title" ASC LIMIT $2 OFFSET $3`,
			wantParams: []interface{}{"org123", int64(10), int64(25)},
			wantErr:    false,
		},
		{
			name:  "query with empty check",
			orgID: "org123",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "description",
						Operator: "empty",
					},
				},
				Limit:  10,
				Offset: 45,
			},
			wantSQL:    `SELECT "billing_transactions"."amount" AS "token_amount", "billing_transactions"."type" AS "token_type", "billing_transactions"."description" AS "token_description", "billing_transactions"."user_id" AS "token_user_id", "users"."title" AS "user_title", "users"."avatar" AS "user_avatar", "billing_transactions"."created_at" AS "token_created_at", "billing_customers"."org_id" AS "org_id" FROM "billing_transactions" INNER JOIN "billing_customers" ON ("billing_transactions"."account_id" = "billing_customers"."id") LEFT JOIN "users" ON CASE WHEN "billing_transactions"."user_id" IS NOT NULL AND "billing_transactions"."user_id" != '' THEN CAST("billing_transactions"."user_id" AS uuid) = "users"."id" ELSE false END WHERE (("billing_customers"."org_id" = $1) AND (("billing_transactions"."description" IS NULL) OR ("billing_transactions"."description" = $2))) LIMIT $3 OFFSET $4`,
			wantParams: []interface{}{"org123", "", int64(10), int64(45)},
			wantErr:    false,
		},
		{
			name:  "query with invalid sort field",
			orgID: "org123",
			rql: &rql.Query{
				Sort: []rql.Sort{
					{
						Name:  "invalid_field",
						Order: "desc",
					},
				},
				Limit:  10,
				Offset: 15,
			},
			wantSQL:    "",
			wantParams: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &OrgTokensRepository{}
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
