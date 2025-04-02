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
	COLUMN_CURRENCY   = "currency"
	COLUMN_HOSTED_URL = "hosted_url"
)

type OrgInvoicesRepository struct {
	dbc *db.Client
}

type OrgInvoice struct {
	ID        sql.NullString `db:"invoice_id"`
	Amount    sql.NullInt64  `db:"invoice_amount"`
	Currency  sql.NullString `db:"invoice_currency"`
	State     sql.NullString `db:"invoice_state"`
	HostedURL sql.NullString `db:"invoice_hosted_url"`
	CreatedAt sql.NullTime   `db:"invoice_created_at"`
	OrgID     sql.NullString `db:"org_id"`
}

type OrgInvoicesGroup struct {
	Name string                 `db:"name"`
	Data []OrgInvoicesGroupData `db:"data"`
}

type OrgInvoicesGroupData struct {
	Name  string `db:"values"`
	Count int    `db:"count"`
}

func (i *OrgInvoice) transformToAggregatedInvoice() svc.AggregatedInvoice {
	return svc.AggregatedInvoice{
		ID:          i.ID.String,
		Amount:      i.Amount.Int64,
		Currency:    i.Currency.String,
		State:       i.State.String,
		InvoiceLink: i.HostedURL.String,
		BilledOn:    i.CreatedAt.Time,
		OrgID:       i.OrgID.String,
	}
}

func (g OrgInvoicesGroup) transformToOrgInvoiceGroup() svc.Group {
	data := make([]svc.GroupData, 0)
	for _, d := range g.Data {
		data = append(data, svc.GroupData{
			Name:  d.Name,
			Count: d.Count,
		})
	}
	return svc.Group{
		Name: g.Name,
		Data: data,
	}
}

func NewOrgInvoicesRepository(dbc *db.Client) *OrgInvoicesRepository {
	return &OrgInvoicesRepository{
		dbc: dbc,
	}
}

func (r OrgInvoicesRepository) Search(ctx context.Context, orgID string, rql *rql.Query) (svc.OrganizationInvoices, error) {
	dataQuery, params, err := r.prepareDataQuery(orgID, rql)
	if err != nil {
		return svc.OrganizationInvoices{}, err
	}

	var orgInvoices []OrgInvoice
	var groupData []OrgInvoicesGroupData
	var group OrgInvoicesGroup

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	err = r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "GetOrgInvoices", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &orgInvoices, dataQuery, params...)
		})

		if err != nil {
			return err
		}

		if len(rql.GroupBy) > 0 {
			groupByKey := rql.GroupBy[0]
			if groupByKey != COLUMN_STATE {
				return fmt.Errorf("grouping only allowed by state field")
			}

			groupQuery, groupParams, err := r.prepareGroupByQuery(orgID, rql)
			if err != nil {
				return err
			}

			err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "GetOrgInvoicesGroup", func(ctx context.Context) error {
				return tx.SelectContext(ctx, &groupData, groupQuery, groupParams...)
			})

			if err != nil {
				return err
			}
			group.Name = rql.GroupBy[0]
			group.Data = groupData
		}
		return nil
	})

	if err != nil {
		return svc.OrganizationInvoices{}, err
	}

	res := make([]svc.AggregatedInvoice, 0)
	for _, invoice := range orgInvoices {
		res = append(res, invoice.transformToAggregatedInvoice())
	}

	return svc.OrganizationInvoices{
		Invoices: res,
		Group:    group.transformToOrgInvoiceGroup(),
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
	query, err := r.addSort(query, rql.Sort, rql.GroupBy)
	if err != nil {
		return "", nil, err
	}

	return query.Offset(uint(rql.Offset)).Limit(uint(rql.Limit)).ToSQL()
}

func (r OrgInvoicesRepository) prepareGroupByQuery(orgID string, rql *rql.Query) (string, []interface{}, error) {
	query := dialect.From(TABLE_BILLING_INVOICES).Prepared(true).
		Select(
			goqu.COUNT("*").As("count"),
			goqu.I(TABLE_BILLING_INVOICES+"."+COLUMN_STATE).As("values"),
		).
		InnerJoin(
			goqu.T(TABLE_BILLING_CUSTOMERS),
			goqu.On(goqu.I(TABLE_BILLING_INVOICES+".customer_id").Eq(goqu.I(TABLE_BILLING_CUSTOMERS+".id"))),
		).
		Where(goqu.Ex{
			TABLE_BILLING_CUSTOMERS + "." + COLUMN_ORG_ID: orgID,
		})

	// Apply the same filters as the main query
	for _, filter := range rql.Filters {
		query = r.addFilter(query, filter)
	}

	// Apply the same search as the main query
	if rql.Search != "" {
		query = r.addSearch(query, rql.Search)
	}

	// Group by state
	query = query.GroupBy(TABLE_BILLING_INVOICES + "." + COLUMN_STATE)

	return query.ToSQL()
}

func (r OrgInvoicesRepository) buildBaseQuery(orgID string) *goqu.SelectDataset {
	return dialect.From(TABLE_BILLING_INVOICES).Prepared(true).
		Select(
			goqu.I(TABLE_BILLING_INVOICES+"."+COLUMN_ID).As("invoice_id"),
			goqu.I(TABLE_BILLING_INVOICES+"."+COLUMN_AMOUNT).As("invoice_amount"),
			goqu.I(TABLE_BILLING_INVOICES+"."+COLUMN_CURRENCY).As("invoice_currency"),
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
		return query.Where(goqu.Or(goqu.I(field).IsNull(), goqu.I(field).Eq("")))
	case "notempty":
		return query.Where(goqu.And(goqu.I(field).IsNotNull(), goqu.I(field).Neq("")))
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

func (r OrgInvoicesRepository) addSort(query *goqu.SelectDataset, sorts []rql.Sort, groupBy []string) (*goqu.SelectDataset, error) {
	// Map of allowed sort fields to their aliases
	allowedSortFields := map[string]string{
		COLUMN_STATE:      "invoice_state",
		COLUMN_AMOUNT:     "invoice_amount",
		COLUMN_CREATED_AT: "invoice_created_at",
	}

	// If there is a group by parameter added then sort the result
	// by group_by first key in asc order by default before any other sort column
	if len(groupBy) > 0 {
		query = query.OrderAppend(goqu.C("invoice_state").Asc())
	}

	// Apply any additional sort conditions
	for _, sort := range sorts {
		aliasedField, allowed := allowedSortFields[sort.Name]
		if !allowed {
			return nil, fmt.Errorf("sorting not allowed on field: %s", sort.Name)
		}

		switch sort.Order {
		case "asc":
			query = query.OrderAppend(goqu.C(aliasedField).Asc())
		case "desc":
			query = query.OrderAppend(goqu.C(aliasedField).Desc())
		}
	}

	return query, nil
}
