package postgres

import (
	"testing"

	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

func TestPrepareDataQuery(t *testing.T) {
	testCases := []struct {
		name           string
		rqlQuery       *rql.Query
		wantSQL        string
		wantParameters []interface{}
		wantErr        bool
	}{
		{
			name: "basic query without filters",
			rqlQuery: &rql.Query{
				Limit:  10,
				Offset: 0,
			},
			wantSQL:        `SELECT "id", "title", "name", "state", "avatar", "updated_at", "created_at", "created_by", "country", "plan_id", "plan_name", "subscription_state", "subscription_cycle_end_at", "plan_interval" FROM (SELECT "organizations"."id" AS "id", "organizations"."title" AS "title", "organizations"."name" AS "name", "organizations"."avatar" AS "avatar", "organizations"."created_at" AS "created_at", "organizations"."updated_at" AS "updated_at", "organizations"."state" AS "state", organizations.metadata->>'country' AS "country", organizations.metadata->>'poc' AS "created_by", "billing_plans"."id" AS "plan_id", "billing_plans"."name" AS "plan_name", "billing_plans"."interval" AS "plan_interval", "billing_subscriptions"."state" AS "subscription_state", "billing_subscriptions"."trial_ends_at", "billing_subscriptions"."current_period_end_at" AS "subscription_cycle_end_at", ROW_NUMBER() OVER (PARTITION BY "organizations"."id" ORDER BY "billing_subscriptions"."created_at" DESC) AS "row_num" FROM "organizations" LEFT JOIN "billing_customers" ON ("organizations"."id" = "billing_customers"."org_id") LEFT JOIN "billing_subscriptions" ON (("billing_subscriptions"."customer_id" = "billing_customers"."id") AND ("billing_subscriptions"."state" != $1)) LEFT JOIN "billing_plans" ON ("billing_plans"."id" = "billing_subscriptions"."plan_id")) AS "ranked_subscriptions" WHERE ("row_num" = $2) LIMIT $3`,
			wantParameters: []interface{}{"canceled", int64(1), int64(10)},
			wantErr:        false,
		},
		{
			name: "query with filter",
			rqlQuery: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "state",
						Operator: "eq",
						Value:    "active",
					},
				},
				Limit:  10,
				Offset: 0,
			},
			wantSQL:        `SELECT "id", "title", "name", "state", "avatar", "updated_at", "created_at", "created_by", "country", "plan_id", "plan_name", "subscription_state", "subscription_cycle_end_at", "plan_interval" FROM (SELECT "organizations"."id" AS "id", "organizations"."title" AS "title", "organizations"."name" AS "name", "organizations"."avatar" AS "avatar", "organizations"."created_at" AS "created_at", "organizations"."updated_at" AS "updated_at", "organizations"."state" AS "state", organizations.metadata->>'country' AS "country", organizations.metadata->>'poc' AS "created_by", "billing_plans"."id" AS "plan_id", "billing_plans"."name" AS "plan_name", "billing_plans"."interval" AS "plan_interval", "billing_subscriptions"."state" AS "subscription_state", "billing_subscriptions"."trial_ends_at", "billing_subscriptions"."current_period_end_at" AS "subscription_cycle_end_at", ROW_NUMBER() OVER (PARTITION BY "organizations"."id" ORDER BY "billing_subscriptions"."created_at" DESC) AS "row_num" FROM "organizations" LEFT JOIN "billing_customers" ON ("organizations"."id" = "billing_customers"."org_id") LEFT JOIN "billing_subscriptions" ON (("billing_subscriptions"."customer_id" = "billing_customers"."id") AND ("billing_subscriptions"."state" != $1)) LEFT JOIN "billing_plans" ON ("billing_plans"."id" = "billing_subscriptions"."plan_id")) AS "ranked_subscriptions" WHERE (("row_num" = $2) AND ("state" = $3)) LIMIT $4`,
			wantParameters: []interface{}{"canceled", int64(1), "active", int64(10)},
			wantErr:        false,
		},
		{
			name: "query with search",
			rqlQuery: &rql.Query{
				Search: "test",
				Limit:  10,
				Offset: 0,
			},
			wantSQL:        `SELECT "id", "title", "name", "state", "avatar", "updated_at", "created_at", "created_by", "country", "plan_id", "plan_name", "subscription_state", "subscription_cycle_end_at", "plan_interval" FROM (SELECT "organizations"."id" AS "id", "organizations"."title" AS "title", "organizations"."name" AS "name", "organizations"."avatar" AS "avatar", "organizations"."created_at" AS "created_at", "organizations"."updated_at" AS "updated_at", "organizations"."state" AS "state", organizations.metadata->>'country' AS "country", organizations.metadata->>'poc' AS "created_by", "billing_plans"."id" AS "plan_id", "billing_plans"."name" AS "plan_name", "billing_plans"."interval" AS "plan_interval", "billing_subscriptions"."state" AS "subscription_state", "billing_subscriptions"."trial_ends_at", "billing_subscriptions"."current_period_end_at" AS "subscription_cycle_end_at", ROW_NUMBER() OVER (PARTITION BY "organizations"."id" ORDER BY "billing_subscriptions"."created_at" DESC) AS "row_num" FROM "organizations" LEFT JOIN "billing_customers" ON ("organizations"."id" = "billing_customers"."org_id") LEFT JOIN "billing_subscriptions" ON (("billing_subscriptions"."customer_id" = "billing_customers"."id") AND ("billing_subscriptions"."state" != $1)) LEFT JOIN "billing_plans" ON ("billing_plans"."id" = "billing_subscriptions"."plan_id")) AS "ranked_subscriptions" WHERE (("row_num" = $2) AND ((CAST("id" AS TEXT) ILIKE $3) OR (CAST("title" AS TEXT) ILIKE $4) OR (CAST("state" AS TEXT) ILIKE $5) OR (CAST("plan_name" AS TEXT) ILIKE $6) OR (CAST("subscription_state" AS TEXT) ILIKE $7) OR (CAST("plan_interval" AS TEXT) ILIKE $8))) LIMIT $9`,
			wantParameters: []interface{}{"canceled", int64(1), "%test%", "%test%", "%test%", "%test%", "%test%", "%test%", int64(10)},
			wantErr:        false,
		},
		{
			name: "query with sort",
			rqlQuery: &rql.Query{
				Sort: []rql.Sort{
					{
						Name:  "created_at",
						Order: "desc",
					},
				},
				Limit:  10,
				Offset: 0,
			},
			wantSQL:        `SELECT "id", "title", "name", "state", "avatar", "updated_at", "created_at", "created_by", "country", "plan_id", "plan_name", "subscription_state", "subscription_cycle_end_at", "plan_interval" FROM (SELECT "organizations"."id" AS "id", "organizations"."title" AS "title", "organizations"."name" AS "name", "organizations"."avatar" AS "avatar", "organizations"."created_at" AS "created_at", "organizations"."updated_at" AS "updated_at", "organizations"."state" AS "state", organizations.metadata->>'country' AS "country", organizations.metadata->>'poc' AS "created_by", "billing_plans"."id" AS "plan_id", "billing_plans"."name" AS "plan_name", "billing_plans"."interval" AS "plan_interval", "billing_subscriptions"."state" AS "subscription_state", "billing_subscriptions"."trial_ends_at", "billing_subscriptions"."current_period_end_at" AS "subscription_cycle_end_at", ROW_NUMBER() OVER (PARTITION BY "organizations"."id" ORDER BY "billing_subscriptions"."created_at" DESC) AS "row_num" FROM "organizations" LEFT JOIN "billing_customers" ON ("organizations"."id" = "billing_customers"."org_id") LEFT JOIN "billing_subscriptions" ON (("billing_subscriptions"."customer_id" = "billing_customers"."id") AND ("billing_subscriptions"."state" != $1)) LEFT JOIN "billing_plans" ON ("billing_plans"."id" = "billing_subscriptions"."plan_id")) AS "ranked_subscriptions" WHERE ("row_num" = $2) ORDER BY "created_at" DESC LIMIT $3`,
			wantParameters: []interface{}{"canceled", int64(1), int64(10)},
			wantErr:        false,
		},
		{
			name: "query with all parameters - filters, search, sort, limit and offset",
			rqlQuery: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "state",
						Operator: "eq",
						Value:    "active",
					},
					{
						Name:     "plan_name",
						Operator: "in",
						Value:    "free,premium",
					},
					{
						Name:     "subscription_state",
						Operator: "notempty",
						Value:    "",
					},
				},
				Search: "test",
				Sort: []rql.Sort{
					{
						Name:  "created_at",
						Order: "desc",
					},
					{
						Name:  "title",
						Order: "asc",
					},
				},
				Limit:  20,
				Offset: 40,
			},
			wantSQL:        `SELECT "id", "title", "name", "state", "avatar", "updated_at", "created_at", "created_by", "country", "plan_id", "plan_name", "subscription_state", "subscription_cycle_end_at", "plan_interval" FROM (SELECT "organizations"."id" AS "id", "organizations"."title" AS "title", "organizations"."name" AS "name", "organizations"."avatar" AS "avatar", "organizations"."created_at" AS "created_at", "organizations"."updated_at" AS "updated_at", "organizations"."state" AS "state", organizations.metadata->>'country' AS "country", organizations.metadata->>'poc' AS "created_by", "billing_plans"."id" AS "plan_id", "billing_plans"."name" AS "plan_name", "billing_plans"."interval" AS "plan_interval", "billing_subscriptions"."state" AS "subscription_state", "billing_subscriptions"."trial_ends_at", "billing_subscriptions"."current_period_end_at" AS "subscription_cycle_end_at", ROW_NUMBER() OVER (PARTITION BY "organizations"."id" ORDER BY "billing_subscriptions"."created_at" DESC) AS "row_num" FROM "organizations" LEFT JOIN "billing_customers" ON ("organizations"."id" = "billing_customers"."org_id") LEFT JOIN "billing_subscriptions" ON (("billing_subscriptions"."customer_id" = "billing_customers"."id") AND ("billing_subscriptions"."state" != $1)) LEFT JOIN "billing_plans" ON ("billing_plans"."id" = "billing_subscriptions"."plan_id")) AS "ranked_subscriptions" WHERE (("row_num" = $2) AND ("state" = $3) AND (CAST("plan_name" AS TEXT) IN ($4, $5)) AND coalesce(subscription_state, '') != '' AND ((CAST("id" AS TEXT) ILIKE $6) OR (CAST("title" AS TEXT) ILIKE $7) OR (CAST("state" AS TEXT) ILIKE $8) OR (CAST("plan_name" AS TEXT) ILIKE $9) OR (CAST("subscription_state" AS TEXT) ILIKE $10) OR (CAST("plan_interval" AS TEXT) ILIKE $11))) ORDER BY "created_at" DESC, "title" ASC LIMIT $12 OFFSET $13`,
			wantParameters: []interface{}{"canceled", int64(1), "active", "free", "premium", "%test%", "%test%", "%test%", "%test%", "%test%", "%test%", int64(20), int64(40)},
			wantErr:        false,
		},
		{
			name: "query with invalid filter",
			rqlQuery: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "invalid_column",
						Operator: "eq",
						Value:    "value",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotSQL, gotParams, err := prepareDataQuery(tc.rqlQuery)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantSQL, gotSQL)
			assert.Equal(t, tc.wantParameters, gotParams)
		})
	}
}

func TestPrepareGroupByQuery(t *testing.T) {
	testCases := []struct {
		name           string
		rqlQuery       *rql.Query
		wantSQL        string
		wantParameters []interface{}
		wantErr        bool
	}{
		{
			name: "group by state",
			rqlQuery: &rql.Query{
				GroupBy: []string{"state"},
			},
			wantSQL:        `SELECT COUNT(*) AS "count", "state" AS "values" FROM (SELECT "organizations"."id" AS "id", "organizations"."title" AS "title", "organizations"."name" AS "name", "organizations"."avatar" AS "avatar", "organizations"."created_at" AS "created_at", "organizations"."updated_at" AS "updated_at", "organizations"."state" AS "state", organizations.metadata->>'country' AS "country", organizations.metadata->>'poc' AS "created_by", "billing_plans"."id" AS "plan_id", "billing_plans"."name" AS "plan_name", "billing_plans"."interval" AS "plan_interval", "billing_subscriptions"."state" AS "subscription_state", "billing_subscriptions"."trial_ends_at", "billing_subscriptions"."current_period_end_at" AS "subscription_cycle_end_at", ROW_NUMBER() OVER (PARTITION BY "organizations"."id" ORDER BY "billing_subscriptions"."created_at" DESC) AS "row_num" FROM "organizations" LEFT JOIN "billing_customers" ON ("organizations"."id" = "billing_customers"."org_id") LEFT JOIN "billing_subscriptions" ON (("billing_subscriptions"."customer_id" = "billing_customers"."id") AND ("billing_subscriptions"."state" != $1)) LEFT JOIN "billing_plans" ON ("billing_plans"."id" = "billing_subscriptions"."plan_id")) AS "ranked_subscriptions" WHERE ("row_num" = $2) GROUP BY "state"`,
			wantParameters: []interface{}{"canceled", int64(1)},
			wantErr:        false,
		},
		{
			name: "group by plan name",
			rqlQuery: &rql.Query{
				GroupBy: []string{"plan_name"},
			},
			wantSQL:        `SELECT COUNT(*) AS "count", "plan_name" AS "values" FROM (SELECT "organizations"."id" AS "id", "organizations"."title" AS "title", "organizations"."name" AS "name", "organizations"."avatar" AS "avatar", "organizations"."created_at" AS "created_at", "organizations"."updated_at" AS "updated_at", "organizations"."state" AS "state", organizations.metadata->>'country' AS "country", organizations.metadata->>'poc' AS "created_by", "billing_plans"."id" AS "plan_id", "billing_plans"."name" AS "plan_name", "billing_plans"."interval" AS "plan_interval", "billing_subscriptions"."state" AS "subscription_state", "billing_subscriptions"."trial_ends_at", "billing_subscriptions"."current_period_end_at" AS "subscription_cycle_end_at", ROW_NUMBER() OVER (PARTITION BY "organizations"."id" ORDER BY "billing_subscriptions"."created_at" DESC) AS "row_num" FROM "organizations" LEFT JOIN "billing_customers" ON ("organizations"."id" = "billing_customers"."org_id") LEFT JOIN "billing_subscriptions" ON (("billing_subscriptions"."customer_id" = "billing_customers"."id") AND ("billing_subscriptions"."state" != $1)) LEFT JOIN "billing_plans" ON ("billing_plans"."id" = "billing_subscriptions"."plan_id")) AS "ranked_subscriptions" WHERE ("row_num" = $2) GROUP BY "plan_name"`,
			wantParameters: []interface{}{"canceled", int64(1)},
			wantErr:        false,
		},
		{
			name: "invalid group by key",
			rqlQuery: &rql.Query{
				GroupBy: []string{"invalid_column"},
			},
			wantErr: true,
		},
		{
			name: "empty group by",
			rqlQuery: &rql.Query{
				GroupBy: []string{},
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotSQL, _, err := prepareGroupByQuery(tc.rqlQuery)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantSQL, gotSQL)
		})
	}
}
