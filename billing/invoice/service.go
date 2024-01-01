package invoice

import (
	"context"
	"fmt"
	"time"

	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/client"
)

type Service struct {
	stripeClient    *client.API
	customerService CustomerService
}

type CustomerService interface {
	GetByID(ctx context.Context, id string) (customer.Customer, error)
}

func NewService(stripeClient *client.API, customerService CustomerService) *Service {
	return &Service{
		stripeClient:    stripeClient,
		customerService: customerService,
	}
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Invoice, error) {
	custmr, err := s.customerService.GetByID(ctx, filter.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find customer: %w", err)
	}

	stripeInvoiceItr := s.stripeClient.Invoices.List(&stripe.InvoiceListParams{
		Customer: stripe.String(custmr.ProviderID),
		ListParams: stripe.ListParams{
			Context: ctx,
		},
	})

	var invoices []Invoice
	for stripeInvoiceItr.Next() {
		invoice := stripeInvoiceItr.Invoice()
		invoices = append(invoices, Invoice{
			ID:          "", // TODO: should we persist this?
			ProviderID:  invoice.ID,
			CustomerID:  custmr.ID,
			State:       string(invoice.Status),
			Currency:    string(invoice.Currency),
			Amount:      invoice.Total,
			HostedURL:   invoice.HostedInvoiceURL,
			Metadata:    metadata.FromString(invoice.Metadata),
			EffectiveAt: time.Unix(invoice.EffectiveAt, 0),
			DueDate:     time.Unix(invoice.DueDate, 0),
			CreatedAt:   time.Unix(invoice.Created, 0),
		})
	}
	if err := stripeInvoiceItr.Err(); err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	return invoices, nil
}
