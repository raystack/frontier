package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/raystack/frontier/billing/invoice"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

type Invoice struct {
	ID         string `db:"id"`
	ProviderID string `db:"provider_id"`
	CustomerID string `db:"customer_id"`
	State      string `db:"state"`
	Currency   string `db:"currency"`
	Amount     int64  `db:"amount"`
	HostedURL  string `db:"hosted_url"`

	Items    Items              `db:"items"`
	Metadata types.NullJSONText `db:"metadata"`

	PeriodStartAt *time.Time `db:"period_start_at"`
	PeriodEndAt   *time.Time `db:"period_end_at"`
	DueAt         *time.Time `db:"due_at"`
	EffectiveAt   *time.Time `db:"effective_at"`
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

type Items struct {
	Data []invoice.Item `json:"data"`
}

func (t *Items) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, t)
	case string:
		return json.Unmarshal([]byte(src), t)
	case nil:
		return nil
	}
	return fmt.Errorf("cannot convert %T to JsonB", src)
}

func (t Items) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (i Invoice) transform() (invoice.Invoice, error) {
	var unmarshalledMetadata map[string]any
	if i.Metadata.Valid {
		if err := i.Metadata.Unmarshal(&unmarshalledMetadata); err != nil {
			return invoice.Invoice{}, err
		}
	}
	dueAt := time.Time{}
	if i.DueAt != nil {
		dueAt = *i.DueAt
	}
	effectiveAt := time.Time{}
	if i.EffectiveAt != nil {
		effectiveAt = *i.EffectiveAt
	}
	var periodStartAt time.Time
	if i.PeriodStartAt != nil {
		periodStartAt = *i.PeriodStartAt
	}
	var periodEndAt time.Time
	if i.PeriodEndAt != nil {
		periodEndAt = *i.PeriodEndAt
	}
	return invoice.Invoice{
		ID:            i.ID,
		ProviderID:    i.ProviderID,
		CustomerID:    i.CustomerID,
		State:         invoice.State(i.State),
		Currency:      i.Currency,
		Amount:        i.Amount,
		HostedURL:     i.HostedURL,
		Items:         i.Items.Data,
		Metadata:      unmarshalledMetadata,
		DueAt:         dueAt,
		EffectiveAt:   effectiveAt,
		CreatedAt:     i.CreatedAt,
		PeriodStartAt: periodStartAt,
		PeriodEndAt:   periodEndAt,
	}, nil
}

type InvoiceWithOrgModel struct {
	ID        string    `db:"id"`
	Amount    int64     `db:"amount"`
	Currency  string    `db:"currency"`
	State     string    `db:"state"`
	HostedURL string    `db:"hosted_url"`
	CreatedAt time.Time `db:"created_at"`
	OrgID     string    `db:"org_id"`
	OrgName   string    `db:"org_name"`
	OrgTitle  string    `db:"org_title"`
}

func (i InvoiceWithOrgModel) transform() invoice.InvoiceWithOrganization {
	return invoice.InvoiceWithOrganization{
		ID:          i.ID,
		Amount:      i.Amount,
		Currency:    i.Currency,
		State:       invoice.State(i.State),
		InvoiceLink: i.HostedURL,
		CreatedAt:   i.CreatedAt,
		OrgID:       i.OrgID,
		OrgName:     i.OrgName,
		OrgTitle:    i.OrgTitle,
	}
}

type BillingInvoiceRepository struct {
	dbc *db.Client
}

func NewBillingInvoiceRepository(dbc *db.Client) *BillingInvoiceRepository {
	return &BillingInvoiceRepository{
		dbc: dbc,
	}
}

func (r BillingInvoiceRepository) Create(ctx context.Context, toCreate invoice.Invoice) (invoice.Invoice, error) {
	if toCreate.ID == "" {
		toCreate.ID = uuid.New().String()
	}
	if toCreate.Metadata == nil {
		toCreate.Metadata = make(map[string]any)
	}
	marshaledMetadata, err := json.Marshal(toCreate.Metadata)
	if err != nil {
		return invoice.Invoice{}, err
	}

	query, params, err := dialect.Insert(TABLE_BILLING_INVOICES).Rows(
		goqu.Record{
			"id":           toCreate.ID,
			"provider_id":  toCreate.ProviderID,
			"customer_id":  toCreate.CustomerID,
			"state":        toCreate.State.String(),
			"currency":     toCreate.Currency,
			"amount":       toCreate.Amount,
			"hosted_url":   toCreate.HostedURL,
			"due_at":       toCreate.DueAt,
			"effective_at": toCreate.EffectiveAt,
			"items": Items{
				Data: toCreate.Items,
			},
			"metadata":        marshaledMetadata,
			"period_start_at": toCreate.PeriodStartAt,
			"period_end_at":   toCreate.PeriodEndAt,
			"created_at":      goqu.L("now()"),
			"updated_at":      goqu.L("now()"),
		}).Returning(&Invoice{}).ToSQL()
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	var invoiceModel Invoice
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&invoiceModel)
	}); err != nil {
		return invoice.Invoice{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return invoiceModel.transform()
}

func (r BillingInvoiceRepository) GetByID(ctx context.Context, id string) (invoice.Invoice, error) {
	stmt := dialect.Select().From(TABLE_BILLING_INVOICES).Where(goqu.Ex{
		"id": id,
	})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	var invoiceModel Invoice
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "GetByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&invoiceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return invoice.Invoice{}, invoice.ErrNotFound
		}
		return invoice.Invoice{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return invoiceModel.transform()
}

func (r BillingInvoiceRepository) List(ctx context.Context, flt invoice.Filter) ([]invoice.Invoice, error) {
	stmt := dialect.Select().From(TABLE_BILLING_INVOICES)
	if flt.CustomerID != "" {
		stmt = stmt.Where(goqu.Ex{
			"customer_id": flt.CustomerID,
		})
	}
	if flt.NonZeroOnly {
		stmt = stmt.Where(goqu.Ex{
			"amount": goqu.Op{"gt": 0},
		})
	}
	if flt.State != "" {
		stmt = stmt.Where(goqu.Ex{
			"state": flt.State.String(),
		})
	}

	if flt.Pagination != nil {
		offset := flt.Pagination.Offset()
		limit := flt.Pagination.PageSize

		// always make this call after all the filters have been applied
		totalCountStmt := stmt.Select(goqu.COUNT("*"))
		totalCountQuery, _, err := totalCountStmt.ToSQL()

		if err != nil {
			return []invoice.Invoice{}, fmt.Errorf("%w: %w", queryErr, err)
		}

		var totalCount int32
		if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "Count", func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &totalCount, totalCountQuery)
		}); err != nil {
			return nil, fmt.Errorf("%w: %w", dbErr, err)
		}

		flt.Pagination.SetCount(totalCount)
		stmt = stmt.Limit(uint(limit)).Offset(uint(offset))
	}

	stmt = stmt.Order(goqu.I("created_at").Desc())
	query, params, err := stmt.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", parseErr, err)
	}

	var invoiceModels []Invoice
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &invoiceModels, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	invoices := make([]invoice.Invoice, 0, len(invoiceModels))
	for _, invoiceModel := range invoiceModels {
		invoice, err := invoiceModel.transform()
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

func (r BillingInvoiceRepository) UpdateByID(ctx context.Context, toUpdate invoice.Invoice) (invoice.Invoice, error) {
	if strings.TrimSpace(toUpdate.ID) == "" {
		return invoice.Invoice{}, invoice.ErrInvalidDetail
	}

	updateRecord := goqu.Record{
		"updated_at": goqu.L("now()"),
	}
	if toUpdate.Metadata != nil {
		marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
		if err != nil {
			return invoice.Invoice{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		updateRecord["metadata"] = marshaledMetadata
	}
	if toUpdate.State != "" {
		updateRecord["state"] = toUpdate.State.String()
	}
	if !toUpdate.EffectiveAt.IsZero() {
		updateRecord["effective_at"] = toUpdate.EffectiveAt
	}
	if toUpdate.HostedURL != "" {
		updateRecord["hosted_url"] = toUpdate.HostedURL
	}

	query, params, err := dialect.Update(TABLE_BILLING_INVOICES).Set(updateRecord).Where(goqu.Ex{
		"id": toUpdate.ID,
	}).Returning(&Invoice{}).ToSQL()
	if err != nil {
		return invoice.Invoice{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var invoiceModel Invoice
	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "UpdateByID", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&invoiceModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return invoice.Invoice{}, invoice.ErrNotFound
		default:
			return invoice.Invoice{}, fmt.Errorf("%w: %w", txnErr, err)
		}
	}

	return invoiceModel.transform()
}

func (r BillingInvoiceRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_BILLING_INVOICES).Where(goqu.Ex{
		"id": id,
	}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "Delete", func(ctx context.Context) error {
		_, err := r.dbc.ExecContext(ctx, query, params...)
		return err
	}); err != nil {
		return fmt.Errorf("%w: %w", txnErr, err)
	}
	return nil
}

func (r BillingInvoiceRepository) Search(ctx context.Context, rqlQuery *rql.Query) ([]invoice.InvoiceWithOrganization, error) {
	dataQuery, params, err := r.prepareDataQuery(rqlQuery)
	if err != nil {
		return nil, err
	}

	var invoiceModels []InvoiceWithOrgModel

	if err = r.dbc.WithTimeout(ctx, TABLE_BILLING_INVOICES, "Search", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &invoiceModels, dataQuery, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %s", dbErr, err)
	}

	// Transform results
	invoices := make([]invoice.InvoiceWithOrganization, 0, len(invoiceModels))
	for _, invoiceModel := range invoiceModels {
		invoices = append(invoices, invoiceModel.transform())
	}

	return invoices, nil
}

func (r BillingInvoiceRepository) prepareDataQuery(rqlQuery *rql.Query) (string, []interface{}, error) {
	query := r.buildBaseQuery()

	// Apply filters
	for _, filter := range rqlQuery.Filters {
		query = r.addFilter(query, filter)
	}

	// Apply search
	if rqlQuery.Search != "" {
		query = r.addSearch(query, rqlQuery.Search)
	}

	// Add sorting
	query, err := r.addSort(query, rqlQuery.Sort)
	if err != nil {
		return "", nil, err
	}

	query = query.Offset(uint(rqlQuery.Offset)).Limit(uint(rqlQuery.Limit))

	return query.ToSQL()
}

func (r BillingInvoiceRepository) buildBaseQuery() *goqu.SelectDataset {
	return dialect.From(TABLE_BILLING_INVOICES).Prepared(true).
		InnerJoin(
			goqu.T(TABLE_BILLING_CUSTOMERS),
			goqu.On(goqu.I(TABLE_BILLING_INVOICES+".customer_id").Eq(goqu.I(TABLE_BILLING_CUSTOMERS+".id"))),
		).
		InnerJoin(
			goqu.T(TABLE_ORGANIZATIONS),
			goqu.On(goqu.I(TABLE_BILLING_CUSTOMERS+".org_id").Eq(goqu.I(TABLE_ORGANIZATIONS+".id"))),
		).
		Select(
			goqu.I(TABLE_BILLING_INVOICES+".id").As("id"),
			goqu.I(TABLE_BILLING_INVOICES+".amount").As("amount"),
			goqu.I(TABLE_BILLING_INVOICES+".currency").As("currency"),
			goqu.I(TABLE_BILLING_INVOICES+".state").As("state"),
			goqu.I(TABLE_BILLING_INVOICES+".hosted_url").As("hosted_url"),
			goqu.I(TABLE_BILLING_INVOICES+".created_at").As("created_at"),
			goqu.I(TABLE_ORGANIZATIONS+".id").As("org_id"),
			goqu.I(TABLE_ORGANIZATIONS+".name").As("org_name"),
			goqu.I(TABLE_ORGANIZATIONS+".title").As("org_title"),
		)
}

func (r BillingInvoiceRepository) addFilter(query *goqu.SelectDataset, filter rql.Filter) *goqu.SelectDataset {
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

func (r BillingInvoiceRepository) addSearch(query *goqu.SelectDataset, search string) *goqu.SelectDataset {
	searchableColumns := []string{
		TABLE_BILLING_INVOICES + ".state",
		TABLE_BILLING_INVOICES + ".hosted_url",
		TABLE_BILLING_INVOICES + ".currency",
		TABLE_ORGANIZATIONS + ".name",
		TABLE_ORGANIZATIONS + ".title",
	}

	searchPattern := "%" + search + "%"

	searchExpressions := make([]goqu.Expression, 0)
	for _, col := range searchableColumns {
		searchExpressions = append(searchExpressions,
			goqu.Cast(goqu.I(col), "TEXT").ILike(searchPattern),
		)
	}

	return query.Where(goqu.Or(searchExpressions...))
}

func (r BillingInvoiceRepository) addSort(query *goqu.SelectDataset, sorts []rql.Sort) (*goqu.SelectDataset, error) {
	// Map of allowed sort fields to their table-qualified names
	allowedSortFields := map[string]string{
		"state":      TABLE_BILLING_INVOICES + ".state",
		"amount":     TABLE_BILLING_INVOICES + ".amount",
		"created_at": TABLE_BILLING_INVOICES + ".created_at",
		"org_name":   TABLE_ORGANIZATIONS + ".name",
		"org_title":  TABLE_ORGANIZATIONS + ".title",
	}

	for _, sort := range sorts {
		field, allowed := allowedSortFields[sort.Name]
		if !allowed {
			return nil, fmt.Errorf("sorting not allowed on field: %s", sort.Name)
		}

		switch sort.Order {
		case "asc":
			query = query.OrderAppend(goqu.I(field).Asc())
		case "desc":
			query = query.OrderAppend(goqu.I(field).Desc())
		}
	}

	return query, nil
}
