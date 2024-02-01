package invoice

import (
	"context"
	"errors"
	"fmt"
	"time"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
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
		invoices = append(invoices, stripeInvoiceToInvoice(custmr.ID, invoice))
	}
	if err := stripeInvoiceItr.Err(); err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	return invoices, nil
}

func (s *Service) GetUpcoming(ctx context.Context, customerID string) (Invoice, error) {
	logger := grpczap.Extract(ctx)
	custmr, err := s.customerService.GetByID(ctx, customerID)
	if err != nil {
		return Invoice{}, fmt.Errorf("failed to find customer: %w", err)
	}

	stripeInvoice, err := s.stripeClient.Invoices.Upcoming(&stripe.InvoiceUpcomingParams{
		Customer: stripe.String(custmr.ProviderID),
		Params: stripe.Params{
			Context: ctx,
		},
	})
	if err != nil {
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) && stripeErr.Code == stripe.ErrorCodeInvoiceUpcomingNone {
			logger.Debug(fmt.Sprintf("no upcoming invoice: %v", stripeErr))
			return Invoice{}, nil
		}
		return Invoice{}, fmt.Errorf("failed to get upcoming invoice: %w", err)
	}

	return stripeInvoiceToInvoice(customerID, stripeInvoice), nil
}

func stripeInvoiceToInvoice(customerID string, stripeInvoice *stripe.Invoice) Invoice {
	var effectiveAt time.Time
	if stripeInvoice.EffectiveAt != 0 {
		effectiveAt = time.Unix(stripeInvoice.EffectiveAt, 0)
	}
	var dueDate time.Time
	if stripeInvoice.DueDate != 0 {
		dueDate = time.Unix(stripeInvoice.DueDate, 0)
	} else if stripeInvoice.NextPaymentAttempt != 0 {
		dueDate = time.Unix(stripeInvoice.NextPaymentAttempt, 0)
	}
	var createdAt time.Time
	if stripeInvoice.Created != 0 {
		createdAt = time.Unix(stripeInvoice.Created, 0)
	}

	return Invoice{
		ID:          "", // TODO: should we persist this?
		ProviderID:  stripeInvoice.ID,
		CustomerID:  customerID,
		State:       string(stripeInvoice.Status),
		Currency:    string(stripeInvoice.Currency),
		Amount:      stripeInvoice.Total,
		HostedURL:   stripeInvoice.HostedInvoiceURL,
		Metadata:    metadata.FromString(stripeInvoice.Metadata),
		EffectiveAt: effectiveAt,
		DueDate:     dueDate,
		CreatedAt:   createdAt,
	}
}
