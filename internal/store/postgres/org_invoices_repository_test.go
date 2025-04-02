package postgres

import (
	"testing"

	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

func TestOrgInvoicesRepository_prepareDataQuery(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		rql        *rql.Query
		wantSQL    string
		wantParams []interface{}
		wantErr    bool
	}{
		{
			name:  "basic query with offset",
			orgID: "org123",
			rql: &rql.Query{
				Limit:  10,
				Offset: 20,
			},
			wantSQL:    `SELECT "billing_invoices"."id" AS "invoice_id", "billing_invoices"."amount" AS "invoice_amount", "billing_invoices"."currency" AS "invoice_currency", "billing_invoices"."state" AS "invoice_state", "billing_invoices"."hosted_url" AS "invoice_hosted_url", "billing_invoices"."created_at" AS "invoice_created_at", "billing_customers"."org_id" AS "org_id" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") WHERE ("billing_customers"."org_id" = $1) LIMIT $2 OFFSET $3`,
			wantParams: []interface{}{"org123", int64(10), int64(20)},
			wantErr:    false,
		},
		{
			name:  "query with amount filter and offset",
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
				Offset: 50,
			},
			wantSQL:    `SELECT "billing_invoices"."id" AS "invoice_id", "billing_invoices"."amount" AS "invoice_amount", "billing_invoices"."currency" AS "invoice_currency", "billing_invoices"."state" AS "invoice_state", "billing_invoices"."hosted_url" AS "invoice_hosted_url", "billing_invoices"."created_at" AS "invoice_created_at", "billing_customers"."org_id" AS "org_id" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") WHERE (("billing_customers"."org_id" = $1) AND ("billing_invoices"."amount" >= $2)) LIMIT $3 OFFSET $4`,
			wantParams: []interface{}{"org123", int64(1000), int64(10), int64(50)},
			wantErr:    false,
		},
		{
			name:  "query with state filter, search and offset",
			orgID: "org123",
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
			wantSQL:    `SELECT "billing_invoices"."id" AS "invoice_id", "billing_invoices"."amount" AS "invoice_amount", "billing_invoices"."currency" AS "invoice_currency", "billing_invoices"."state" AS "invoice_state", "billing_invoices"."hosted_url" AS "invoice_hosted_url", "billing_invoices"."created_at" AS "invoice_created_at", "billing_customers"."org_id" AS "org_id" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") WHERE (("billing_customers"."org_id" = $1) AND ("billing_invoices"."state" = $2) AND ((CAST("billing_invoices"."state" AS TEXT) ILIKE $3) OR (CAST("billing_invoices"."hosted_url" AS TEXT) ILIKE $4) OR (CAST("billing_invoices"."amount" AS TEXT) ILIKE $5))) LIMIT $6 OFFSET $7`,
			wantParams: []interface{}{"org123", "paid", "%test%", "%test%", "%test%", int64(10), int64(30)},
			wantErr:    false,
		},
		{
			name:  "query with sort and offset",
			orgID: "org123",
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
			wantSQL:    `SELECT "billing_invoices"."id" AS "invoice_id", "billing_invoices"."amount" AS "invoice_amount", "billing_invoices"."currency" AS "invoice_currency", "billing_invoices"."state" AS "invoice_state", "billing_invoices"."hosted_url" AS "invoice_hosted_url", "billing_invoices"."created_at" AS "invoice_created_at", "billing_customers"."org_id" AS "org_id" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") WHERE ("billing_customers"."org_id" = $1) ORDER BY "invoice_state" DESC LIMIT $2 OFFSET $3`,
			wantParams: []interface{}{"org123", int64(10), int64(40)},
			wantErr:    false,
		},
		{
			name:  "query with group by state and offset",
			orgID: "org123",
			rql: &rql.Query{
				GroupBy: []string{"state"},
				Sort: []rql.Sort{
					{
						Name:  "amount",
						Order: "desc",
					},
				},
				Limit:  10,
				Offset: 25,
			},
			wantSQL:    `SELECT "billing_invoices"."id" AS "invoice_id", "billing_invoices"."amount" AS "invoice_amount", "billing_invoices"."currency" AS "invoice_currency", "billing_invoices"."state" AS "invoice_state", "billing_invoices"."hosted_url" AS "invoice_hosted_url", "billing_invoices"."created_at" AS "invoice_created_at", "billing_customers"."org_id" AS "org_id" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") WHERE ("billing_customers"."org_id" = $1) ORDER BY "invoice_state" ASC, "invoice_amount" DESC LIMIT $2 OFFSET $3`,
			wantParams: []interface{}{"org123", int64(10), int64(25)},
			wantErr:    false,
		},
		{
			name:  "query with invalid sort field and offset",
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
			r := &OrgInvoicesRepository{}
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

func TestOrgInvoicesRepository_prepareGroupByQuery(t *testing.T) {
	tests := []struct {
		name       string
		orgID      string
		rql        *rql.Query
		wantSQL    string
		wantParams []interface{}
		wantErr    bool
	}{
		{
			name:  "basic group by state",
			orgID: "org123",
			rql: &rql.Query{
				GroupBy: []string{"state"},
			},
			wantSQL:    `SELECT COUNT(*) AS "count", "billing_invoices"."state" AS "values" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") WHERE ("billing_customers"."org_id" = $1) GROUP BY "billing_invoices"."state"`,
			wantParams: []interface{}{"org123"},
			wantErr:    false,
		},
		{
			name:  "group by state with filter",
			orgID: "org123",
			rql: &rql.Query{
				GroupBy: []string{"state"},
				Filters: []rql.Filter{
					{
						Name:     "amount",
						Operator: "gte",
						Value:    1000,
					},
				},
			},
			wantSQL:    `SELECT COUNT(*) AS "count", "billing_invoices"."state" AS "values" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") WHERE (("billing_customers"."org_id" = $1) AND ("billing_invoices"."amount" >= $2)) GROUP BY "billing_invoices"."state"`,
			wantParams: []interface{}{"org123", int64(1000)},
			wantErr:    false,
		},
		{
			name:  "group by state with search",
			orgID: "org123",
			rql: &rql.Query{
				GroupBy: []string{"state"},
				Search:  "test",
			},
			wantSQL:    `SELECT COUNT(*) AS "count", "billing_invoices"."state" AS "values" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") WHERE (("billing_customers"."org_id" = $1) AND ((CAST("billing_invoices"."state" AS TEXT) ILIKE $2) OR (CAST("billing_invoices"."hosted_url" AS TEXT) ILIKE $3) OR (CAST("billing_invoices"."amount" AS TEXT) ILIKE $4))) GROUP BY "billing_invoices"."state"`,
			wantParams: []interface{}{"org123", "%test%", "%test%", "%test%"},
			wantErr:    false,
		},
		{
			name:  "group by state with filter and search",
			orgID: "org123",
			rql: &rql.Query{
				GroupBy: []string{"state"},
				Filters: []rql.Filter{
					{
						Name:     "amount",
						Operator: "gte",
						Value:    1000,
					},
				},
				Search: "test",
			},
			wantSQL:    `SELECT COUNT(*) AS "count", "billing_invoices"."state" AS "values" FROM "billing_invoices" INNER JOIN "billing_customers" ON ("billing_invoices"."customer_id" = "billing_customers"."id") WHERE (("billing_customers"."org_id" = $1) AND ("billing_invoices"."amount" >= $2) AND ((CAST("billing_invoices"."state" AS TEXT) ILIKE $3) OR (CAST("billing_invoices"."hosted_url" AS TEXT) ILIKE $4) OR (CAST("billing_invoices"."amount" AS TEXT) ILIKE $5))) GROUP BY "billing_invoices"."state"`,
			wantParams: []interface{}{"org123", int64(1000), "%test%", "%test%", "%test%"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &OrgInvoicesRepository{}
			gotSQL, gotParams, err := r.prepareGroupByQuery(tt.orgID, tt.rql)

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
