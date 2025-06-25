package postgres

import (
	"testing"

	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

func TestBillingInvoiceRepository_prepareDataQuery(t *testing.T) {
	tests := []struct {
		name       string
		rql        *rql.Query
		wantSQL    string
		wantParams []interface{}
		wantErr    bool
	}{
		{
			name: "basic query with offset",
			rql: &rql.Query{
				Limit:  10,
				Offset: 20,
			},
			wantSQL:    `SELECT "billing_invoices"."id" AS "id", "billing_invoices"."amount" AS "amount", "billing_invoices"."currency" AS "currency", "billing_invoices"."state" AS "state", "billing_invoices"."hosted_url" AS "hosted_url", "billing_invoices"."created_at" AS "created_at", "organizations"."id" AS "org_id", "organizations"."name" AS "org_name", "organizations"."title" AS "org_title" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") INNER JOIN "organizations" ON ("billing_customers"."org_id" = "organizations"."id") LIMIT $1 OFFSET $2`,
			wantParams: []interface{}{int64(10), int64(20)},
			wantErr:    false,
		},
		{
			name: "query with amount filter and offset",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "amount",
						Operator: "gte",
						Value:    1000,
					},
				},
				Limit:  10,
				Offset: 50,
			},
			wantSQL:    `SELECT "billing_invoices"."id" AS "id", "billing_invoices"."amount" AS "amount", "billing_invoices"."currency" AS "currency", "billing_invoices"."state" AS "state", "billing_invoices"."hosted_url" AS "hosted_url", "billing_invoices"."created_at" AS "created_at", "organizations"."id" AS "org_id", "organizations"."name" AS "org_name", "organizations"."title" AS "org_title" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") INNER JOIN "organizations" ON ("billing_customers"."org_id" = "organizations"."id") WHERE ("billing_invoices"."amount" >= $1) LIMIT $2 OFFSET $3`,
			wantParams: []interface{}{int64(1000), int64(10), int64(50)},
			wantErr:    false,
		},
		{
			name: "query with state filter, search and offset",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "state",
						Operator: "eq",
						Value:    "paid",
					},
				},
				Search: "test",
				Limit:  10,
				Offset: 30,
			},
			wantSQL:    `SELECT "billing_invoices"."id" AS "id", "billing_invoices"."amount" AS "amount", "billing_invoices"."currency" AS "currency", "billing_invoices"."state" AS "state", "billing_invoices"."hosted_url" AS "hosted_url", "billing_invoices"."created_at" AS "created_at", "organizations"."id" AS "org_id", "organizations"."name" AS "org_name", "organizations"."title" AS "org_title" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") INNER JOIN "organizations" ON ("billing_customers"."org_id" = "organizations"."id") WHERE (("billing_invoices"."state" = $1) AND ((CAST("billing_invoices"."state" AS TEXT) ILIKE $2) OR (CAST("billing_invoices"."currency" AS TEXT) ILIKE $3) OR (CAST("billing_invoices"."amount" AS TEXT) ILIKE $4) OR (CAST("organizations"."name" AS TEXT) ILIKE $5) OR (CAST("organizations"."title" AS TEXT) ILIKE $6))) LIMIT $7 OFFSET $8`,
			wantParams: []interface{}{"paid", "%test%", "%test%", "%test%", "%test%", "%test%", int64(10), int64(30)},
			wantErr:    false,
		},
		{
			name: "query with sort and offset",
			rql: &rql.Query{
				Sort: []rql.Sort{
					{
						Name:  "state",
						Order: "desc",
					},
				},
				Limit:  10,
				Offset: 40,
			},
			wantSQL:    `SELECT "billing_invoices"."id" AS "id", "billing_invoices"."amount" AS "amount", "billing_invoices"."currency" AS "currency", "billing_invoices"."state" AS "state", "billing_invoices"."hosted_url" AS "hosted_url", "billing_invoices"."created_at" AS "created_at", "organizations"."id" AS "org_id", "organizations"."name" AS "org_name", "organizations"."title" AS "org_title" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") INNER JOIN "organizations" ON ("billing_customers"."org_id" = "organizations"."id") ORDER BY "billing_invoices"."state" DESC LIMIT $1 OFFSET $2`,
			wantParams: []interface{}{int64(10), int64(40)},
			wantErr:    false,
		},
		{
			name: "query with org name sort and offset",
			rql: &rql.Query{
				Sort: []rql.Sort{
					{
						Name:  "org_name",
						Order: "asc",
					},
				},
				Limit:  10,
				Offset: 40,
			},
			wantSQL:    `SELECT "billing_invoices"."id" AS "id", "billing_invoices"."amount" AS "amount", "billing_invoices"."currency" AS "currency", "billing_invoices"."state" AS "state", "billing_invoices"."hosted_url" AS "hosted_url", "billing_invoices"."created_at" AS "created_at", "organizations"."id" AS "org_id", "organizations"."name" AS "org_name", "organizations"."title" AS "org_title" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") INNER JOIN "organizations" ON ("billing_customers"."org_id" = "organizations"."id") ORDER BY "organizations"."name" ASC LIMIT $1 OFFSET $2`,
			wantParams: []interface{}{int64(10), int64(40)},
			wantErr:    false,
		},
		{
			name: "query with invalid sort field and offset",
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
		{
			name: "query with empty value filter",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "state",
						Operator: "empty",
					},
				},
				Limit:  10,
				Offset: 1,
			},
			wantSQL:    `SELECT "billing_invoices"."id" AS "id", "billing_invoices"."amount" AS "amount", "billing_invoices"."currency" AS "currency", "billing_invoices"."state" AS "state", "billing_invoices"."hosted_url" AS "hosted_url", "billing_invoices"."created_at" AS "created_at", "organizations"."id" AS "org_id", "organizations"."name" AS "org_name", "organizations"."title" AS "org_title" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") INNER JOIN "organizations" ON ("billing_customers"."org_id" = "organizations"."id") WHERE (("billing_invoices"."state" IS NULL) OR ("billing_invoices"."state" = $1)) LIMIT $2 OFFSET $3`,
			wantParams: []interface{}{"", int64(10), int64(1)},
			wantErr:    false,
		},
		{
			name: "query with not empty value filter",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "state",
						Operator: "notempty",
					},
				},
				Limit:  10,
				Offset: 1,
			},
			wantSQL:    `SELECT "billing_invoices"."id" AS "id", "billing_invoices"."amount" AS "amount", "billing_invoices"."currency" AS "currency", "billing_invoices"."state" AS "state", "billing_invoices"."hosted_url" AS "hosted_url", "billing_invoices"."created_at" AS "created_at", "organizations"."id" AS "org_id", "organizations"."name" AS "org_name", "organizations"."title" AS "org_title" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") INNER JOIN "organizations" ON ("billing_customers"."org_id" = "organizations"."id") WHERE (("billing_invoices"."state" IS NOT NULL) AND ("billing_invoices"."state" != $1)) LIMIT $2 OFFSET $3`,
			wantParams: []interface{}{"", int64(10), int64(1)},
			wantErr:    false,
		},
		{
			name: "query with like operator",
			rql: &rql.Query{
				Filters: []rql.Filter{
					{
						Name:     "state",
						Operator: "like",
						Value:    "paid",
					},
				},
				Limit:  10,
				Offset: 1,
			},
			wantSQL:    `SELECT "billing_invoices"."id" AS "id", "billing_invoices"."amount" AS "amount", "billing_invoices"."currency" AS "currency", "billing_invoices"."state" AS "state", "billing_invoices"."hosted_url" AS "hosted_url", "billing_invoices"."created_at" AS "created_at", "organizations"."id" AS "org_id", "organizations"."name" AS "org_name", "organizations"."title" AS "org_title" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") INNER JOIN "organizations" ON ("billing_customers"."org_id" = "organizations"."id") WHERE ("billing_invoices"."state" LIKE $1) LIMIT $2 OFFSET $3`,
			wantParams: []interface{}{"%paid%", int64(10), int64(1)},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &BillingInvoiceRepository{}
			gotSQL, gotParams, err := r.prepareDataQuery(tt.rql)

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
