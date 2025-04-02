package orginvoices

import (
	"context"
	"time"

	"github.com/raystack/salt/rql"
)

type Repository interface {
	Search(ctx context.Context, orgID string, query *rql.Query) (OrganizationInvoices, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type OrganizationInvoices struct {
	Invoices   []Invoice `json:"invoices"`
	Group      Group     `json:"group"`
	Pagination Page      `json:"pagination"`
}

type Group struct {
	Name string      `json:"name"`
	Data []GroupData `json:"data"`
}

type GroupData struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type Invoice struct {
	ID          string    `rql:"name=id,type=string"`
	Amount      int64     `rql:"name=amount,type=number"`
	Status      string    `rql:"name=status,type=string"`
	InvoiceLink string    `rql:"name=invoice_link,type=string"`
	BilledOn    time.Time `rql:"name=billed_on,type=datetime"`
	OrgID       string    `rql:"name=org_id,type=string"`
	CreatedAt   time.Time `rql:"name=created_at,type=datetime"`
}

func (s Service) Search(ctx context.Context, orgID string, query *rql.Query) (OrganizationInvoices, error) {
	return s.repository.Search(ctx, orgID, query)
}
