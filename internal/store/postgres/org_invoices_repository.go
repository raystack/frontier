package postgres

import (
	"context"
	"database/sql"

	svc "github.com/raystack/frontier/core/aggregates/orginvoices"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

type OrgInvoicesRepository struct {
	dbc *db.Client
}

type OrgInvoice struct {
	ID          sql.NullString `db:"id"`
	Amount      sql.NullInt64  `db:"amount"`
	Status      sql.NullString `db:"status"`
	InvoiceLink sql.NullString `db:"invoice_link"`
	BilledOn    sql.NullTime   `db:"billed_on"`
	OrgID       sql.NullString `db:"org_id"`
	CreatedAt   sql.NullTime   `db:"created_at"`
}

type OrgInvoicesGroup struct {
	Name sql.NullString         `db:"name"`
	Data []OrgInvoicesGroupData `db:"data"`
}

type OrgInvoicesGroupData struct {
	Name  sql.NullString `db:"values"`
	Count int            `db:"count"`
}

func (i *OrgInvoice) transformToInvoice() svc.Invoice {
	return svc.Invoice{
		ID:          i.ID.String,
		Amount:      i.Amount.Int64,
		Status:      i.Status.String,
		InvoiceLink: i.InvoiceLink.String,
		BilledOn:    i.BilledOn.Time,
		OrgID:       i.OrgID.String,
		CreatedAt:   i.CreatedAt.Time,
	}
}

func NewOrgInvoicesRepository(dbc *db.Client) *OrgInvoicesRepository {
	return &OrgInvoicesRepository{
		dbc: dbc,
	}
}

func (r OrgInvoicesRepository) Search(ctx context.Context, orgID string, rql *rql.Query) (svc.OrganizationInvoices, error) {
	return svc.OrganizationInvoices{}, nil
}
