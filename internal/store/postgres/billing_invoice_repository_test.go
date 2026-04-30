package postgres

import (
	"strings"
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// Prepared-mode regression guards. Each test mirrors the goqu chain at its named site.

// Mirror of relation_repository.go::GetByFields query construction.
func TestRelationRepository_GetByFields_PreparedSQLForwardsParams(t *testing.T) {
	rel := relation.Relation{
		Object:       relation.Object{ID: "obj-1", Namespace: "ns-obj"},
		Subject:      relation.Subject{ID: "sub-1", Namespace: "ns-sub"},
		RelationName: "owner",
	}

	stmt := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).Prepared(true)
	if rel.Object.ID != "" {
		stmt = stmt.Where(goqu.Ex{"object_id": rel.Object.ID})
	}
	if rel.Object.Namespace != "" {
		stmt = stmt.Where(goqu.Ex{"object_namespace_name": rel.Object.Namespace})
	}
	if rel.Subject.ID != "" {
		stmt = stmt.Where(goqu.Ex{"subject_id": rel.Subject.ID})
	}
	if rel.Subject.Namespace != "" {
		stmt = stmt.Where(goqu.Ex{"subject_namespace_name": rel.Subject.Namespace})
	}
	if rel.RelationName != "" {
		stmt = stmt.Where(goqu.Ex{"relation_name": rel.RelationName})
	}

	sql, params, err := stmt.ToSQL()
	require.NoError(t, err)

	assert.True(t, strings.Contains(sql, "$1"), "SQL must use $N placeholders, got: %s", sql)
	assert.Equal(t, []interface{}{"obj-1", "ns-obj", "sub-1", "ns-sub", "owner"}, params)
}

// Mirror of relation_repository.go::ListByFields query construction.
func TestRelationRepository_ListByFields_PreparedSQLForwardsParams(t *testing.T) {
	rel := relation.Relation{
		Subject:      relation.Subject{ID: "sub-1", SubRelationName: "member"},
		RelationName: "team",
		Object:       relation.Object{ID: "obj-1"},
	}
	like := "%:" + rel.Subject.SubRelationName

	var exprs []goqu.Expression
	if len(rel.Subject.ID) != 0 {
		exprs = append(exprs, goqu.Ex{"subject_id": rel.Subject.ID})
	}
	if len(rel.RelationName) != 0 {
		exprs = append(exprs, goqu.Ex{"relation_name": goqu.Op{"like": like}})
	}
	if len(rel.Object.ID) != 0 {
		exprs = append(exprs, goqu.Ex{"object_id": rel.Object.ID})
	}

	sql, params, err := dialect.Select(&relationCols{}).From(TABLE_RELATIONS).Prepared(true).Where(exprs...).ToSQL()
	require.NoError(t, err)

	assert.True(t, strings.Contains(sql, "$1"), "SQL must use $N placeholders, got: %s", sql)
	assert.Equal(t, []interface{}{"sub-1", "%:member", "obj-1"}, params)
}

// Mirror of organization_repository.go::List totalCount path.
func TestOrganizationRepository_ListTotalCount_PreparedSQLForwardsParams(t *testing.T) {
	stmt := dialect.From(TABLE_ORGANIZATIONS).Prepared(true).Where(goqu.Ex{"state": "enabled"})
	stmt = stmt.Where(goqu.Ex{"id": goqu.Op{"in": []string{"o-1", "o-2"}}})

	totalCountStmt := stmt.Select(goqu.COUNT("*"))
	sql, params, err := totalCountStmt.ToSQL()
	require.NoError(t, err)

	assert.True(t, strings.Contains(sql, "$1"), "SQL must use $N placeholders, got: %s", sql)
	assert.Equal(t, []interface{}{"enabled", "o-1", "o-2"}, params)
}

// Mirror of project_repository.go::List totalCount path.
func TestProjectRepository_ListTotalCount_PreparedSQLForwardsParams(t *testing.T) {
	stmt := dialect.From(TABLE_PROJECTS).Prepared(true).Where(goqu.Ex{"org_id": "org-1"})
	stmt = stmt.Where(goqu.Ex{"id": goqu.Op{"in": []string{"p-1", "p-2"}}})
	stmt = stmt.Where(goqu.Ex{"state": "enabled"})

	totalCountStmt := stmt.Select(goqu.COUNT("*"))
	sql, params, err := totalCountStmt.ToSQL()
	require.NoError(t, err)

	assert.True(t, strings.Contains(sql, "$1"), "SQL must use $N placeholders, got: %s", sql)
	assert.Equal(t, []interface{}{"org-1", "p-1", "p-2", "enabled"}, params)
}

// Mirror of billing_invoice_repository.go::List totalCount path.
func TestBillingInvoiceRepository_ListTotalCount_PreparedSQLForwardsParams(t *testing.T) {
	stmt := dialect.Select().From(TABLE_BILLING_INVOICES).Prepared(true).Where(goqu.Ex{"customer_id": "cust-1"})
	stmt = stmt.Where(goqu.Ex{"amount": goqu.Op{"gt": 0}})
	stmt = stmt.Where(goqu.Ex{"state": "paid"})

	totalCountStmt := stmt.Select(goqu.COUNT("*"))
	sql, params, err := totalCountStmt.ToSQL()
	require.NoError(t, err)

	assert.True(t, strings.Contains(sql, "$1"), "SQL must use $N placeholders, got: %s", sql)
	assert.Equal(t, []interface{}{"cust-1", int64(0), "paid"}, params)
}
