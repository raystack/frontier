package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	svc "github.com/raystack/frontier/core/aggregates/orginvoices"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	COLUMN_AMOUNT     = "amount"
	COLUMN_HOSTED_URL = "hosted_url"
)

type OrgInvoicesRepository struct {
	dbc *db.Client
}

type OrgInvoice struct {
	ID        sql.NullString `db:"invoice_id"`
	Amount    sql.NullInt64  `db:"invoice_amount"`
	State     sql.NullString `db:"invoice_state"`
	HostedURL sql.NullString `db:"invoice_hosted_url"`
	CreatedAt sql.NullTime   `db:"invoice_created_at"`
	OrgID     sql.NullString `db:"org_id"`
}

type OrgInvoicesGroup struct {
	Name sql.NullString         `db:"name"`
	Data []OrgInvoicesGroupData `db:"data"`
}

type OrgInvoicesGroupData struct {
	Name  sql.NullString `db:"values"`
	Count int            `db:"count"`
}

func (i *OrgInvoice) transformToInvoice() svc.AggregatedInvoice {
	return svc.AggregatedInvoice{
		ID:          i.ID.String,
		Amount:      i.Amount.Int64,
		State:       i.State.String,
		InvoiceLink: i.HostedURL.String,
		BilledOn:    i.CreatedAt.Time,
		OrgID:       i.OrgID.String,
	}
}

func NewOrgInvoicesRepository(dbc *db.Client) *OrgInvoicesRepository {
	return &OrgInvoicesRepository{
		dbc: dbc,
	}
}

func (r OrgInvoicesRepository) Search(ctx context.Context, orgID string, rql *rql.Query) (svc.OrganizationInvoices, error) {
	dataQuery, params, err := r.prepareDataQuery(orgID, rql)
	fmt.Println(dataQuery)
	if err != nil {
		return svc.OrganizationInvoices{}, err
	}

	var orgInvoices []OrgInvoice

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	err = r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "GetOrgInvoices", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &orgInvoices, dataQuery, params...)
		})
	})

	if err != nil {
		return svc.OrganizationInvoices{}, err
	}

	res := make([]svc.AggregatedInvoice, 0)
	for _, invoice := range orgInvoices {
		res = append(res, invoice.transformToInvoice())
	}

	return svc.OrganizationInvoices{
		Invoices: res,
		Group:    svc.Group{},
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

func (r OrgInvoicesRepository) prepareDataQuery(orgID string, rql *rql.Query) (string, []interface{}, error) {
	query := r.buildBaseQuery(orgID)

	for _, filter := range rql.Filters {
		query = r.addFilter(query, filter)
	}

	if rql.Search != "" {
		query = r.addSearch(query, rql.Search)
	}

	// Add sorting
	query = r.addSort(query, rql.Sort)

	return query.Offset(uint(rql.Offset)).Limit(uint(rql.Limit)).ToSQL()
}

func (r OrgInvoicesRepository) buildBaseQuery(orgID string) *goqu.SelectDataset {
	return dialect.From(TABLE_BILLING_INVOICES).Prepared(false).
		Select(
			goqu.I(TABLE_BILLING_INVOICES+"."+COLUMN_ID).As("invoice_id"),
			goqu.I(TABLE_BILLING_INVOICES+"."+COLUMN_AMOUNT).As("invoice_amount"),
			goqu.I(TABLE_BILLING_INVOICES+"."+COLUMN_STATE).As("invoice_state"),
			goqu.I(TABLE_BILLING_INVOICES+"."+COLUMN_HOSTED_URL).As("invoice_hosted_url"),
			goqu.I(TABLE_BILLING_INVOICES+"."+COLUMN_CREATED_AT).As("invoice_created_at"),
			goqu.I(TABLE_BILLING_CUSTOMERS+"."+COLUMN_ORG_ID).As("org_id"),
		).
		InnerJoin(
			goqu.T(TABLE_BILLING_CUSTOMERS),
			goqu.On(goqu.I(TABLE_BILLING_INVOICES+".customer_id").Eq(goqu.I(TABLE_BILLING_CUSTOMERS+".id"))),
		).
		Where(goqu.Ex{
			TABLE_BILLING_CUSTOMERS + "." + COLUMN_ORG_ID: orgID,
		})
}

func (r OrgInvoicesRepository) addFilter(query *goqu.SelectDataset, filter rql.Filter) *goqu.SelectDataset {
	field := TABLE_BILLING_INVOICES + "." + filter.Name

	switch filter.Operator {
	case "empty":
		return query.Where(goqu.Or(
			goqu.I(field).IsNull(),
			goqu.I(field).Eq(""),
		))
	case "notempty":
		return query.Where(goqu.And(
			goqu.I(field).IsNotNull(),
			goqu.I(field).Neq(""),
		))
	case "like", "notlike":
		value := "%" + filter.Value.(string) + "%"
		return query.Where(goqu.Ex{field: goqu.Op{filter.Operator: value}})
	default:
		return query.Where(goqu.Ex{field: goqu.Op{filter.Operator: filter.Value}})
	}
}

func (r OrgInvoicesRepository) addSearch(query *goqu.SelectDataset, search string) *goqu.SelectDataset {
	searchSupportedColumns := []string{
		TABLE_BILLING_INVOICES + "." + COLUMN_STATE,
		TABLE_BILLING_INVOICES + "." + COLUMN_HOSTED_URL,
		TABLE_BILLING_INVOICES + "." + COLUMN_AMOUNT,
	}

	searchPattern := "%" + search + "%"

	searchExpressions := make([]goqu.Expression, 0)
	for _, col := range searchSupportedColumns {
		searchExpressions = append(searchExpressions,
			goqu.Cast(goqu.I(col), "TEXT").ILike(searchPattern),
		)
	}

	return query.Where(goqu.Or(searchExpressions...))
}

func (r OrgInvoicesRepository) addSort(query *goqu.SelectDataset, sorts []rql.Sort) *goqu.SelectDataset {
	for _, sort := range sorts {
		switch sort.Order {
		case "asc":
			query = query.OrderAppend(goqu.I(TABLE_BILLING_INVOICES + "." + sort.Name).Asc())
		case "desc":
			query = query.OrderAppend(goqu.I(TABLE_BILLING_INVOICES + "." + sort.Name).Desc())
		}
	}

	return query
}
