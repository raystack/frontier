package invoice

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stripe/stripe-go/v79"

	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/internal/metrics"

	"github.com/raystack/frontier/pkg/utils"
	"go.uber.org/zap"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/stripe/stripe-go/v79/client"
)

type Repository interface {
	Create(ctx context.Context, invoice Invoice) (Invoice, error)
	GetByID(ctx context.Context, id string) (Invoice, error)
	List(ctx context.Context, filter Filter) ([]Invoice, error)
	UpdateByID(ctx context.Context, invoice Invoice) (Invoice, error)
	Delete(ctx context.Context, id string) error
}

type CustomerService interface {
	GetByID(ctx context.Context, id string) (customer.Customer, error)
	List(ctx context.Context, filter customer.Filter) ([]customer.Customer, error)
}

type Service struct {
	stripeClient    *client.API
	repository      Repository
	customerService CustomerService

	syncJob   *cron.Cron
	mu        sync.Mutex
	syncDelay time.Duration
}

func NewService(stripeClient *client.API, invoiceRepository Repository,
	customerService CustomerService, cfg billing.Config) *Service {
	return &Service{
		stripeClient:    stripeClient,
		repository:      invoiceRepository,
		customerService: customerService,
		syncDelay:       cfg.RefreshInterval.Invoice,
	}
}

func (s *Service) Init(ctx context.Context) error {
	if s.syncJob != nil {
		s.syncJob.Stop()
	}

	s.syncJob = cron.New()
	if _, err := s.syncJob.AddFunc(fmt.Sprintf("@every %s", s.syncDelay.String()), func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		s.backgroundSync(ctx)
	}); err != nil {
		return err
	}
	s.syncJob.Start()
	return nil
}

func (s *Service) Close() error {
	if s.syncJob != nil {
		return s.syncJob.Stop().Err()
	}
	return nil
}

func (s *Service) backgroundSync(ctx context.Context) {
	if metrics.BillingSyncLatency != nil {
		record := metrics.BillingSyncLatency("invoice")
		defer record()
	}
	logger := grpczap.Extract(ctx)
	customers, err := s.customerService.List(ctx, customer.Filter{})
	if err != nil {
		logger.Error("invoice.backgroundSync", zap.Error(err))
		return
	}
	for _, customer := range customers {
		if ctx.Err() != nil {
			// stop processing if context is done
			break
		}

		if !customer.IsActive() || customer.ProviderID == "" {
			continue
		}
		if err := s.SyncWithProvider(ctx, customer); err != nil {
			logger.Error("invoice.SyncWithProvider", zap.Error(err))
		}
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
}

func (s *Service) SyncWithProvider(ctx context.Context, customr customer.Customer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	invoiceObs, err := s.repository.List(ctx, Filter{
		CustomerID: customr.ID,
	})
	if err != nil {
		return err
	}

	var errs []error
	stripeInvoices := s.stripeClient.Invoices.List(&stripe.InvoiceListParams{
		Customer: stripe.String(customr.ProviderID),
		ListParams: stripe.ListParams{
			Context: ctx,
		},
	})
	for stripeInvoices.Next() {
		stripeInvoice := stripeInvoices.Invoice()

		// check if already present, if yes, update else create new
		existingInvoice, ok := utils.FindFirst(invoiceObs, func(i Invoice) bool {
			return i.ProviderID == stripeInvoice.ID
		})
		if ok {
			// already present in our system, update it if needed
			updateNeeded := false
			if existingInvoice.State != string(stripeInvoice.Status) {
				existingInvoice.State = string(stripeInvoice.Status)
				updateNeeded = true
			}

			if updateNeeded {
				if _, err := s.repository.UpdateByID(ctx, existingInvoice); err != nil {
					errs = append(errs, fmt.Errorf("failed to update invoice %s: %w", existingInvoice.ID, err))
				}
			}
		} else {
			if _, err := s.repository.Create(ctx, stripeInvoiceToInvoice(customr.ID, stripeInvoice)); err != nil {
				errs = append(errs, fmt.Errorf("failed to create invoice for customer %s: %w", customr.ID, err))
			}
		}

		// add jitter
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	if err := stripeInvoices.Err(); err != nil {
		return fmt.Errorf("failed to list invoices: %w", err)
	}
	return nil
}

// ListAll should only be called by admin users
func (s *Service) ListAll(ctx context.Context, filter Filter) ([]Invoice, error) {
	return s.repository.List(ctx, filter)
}

// List currently queries stripe for invoices, but it should be refactored to query our own database
func (s *Service) List(ctx context.Context, filter Filter) ([]Invoice, error) {
	if filter.CustomerID == "" {
		return nil, errors.New("customer id is required")
	}
	return s.repository.List(ctx, filter)
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
	var periodStartAt time.Time
	if stripeInvoice.PeriodStart != 0 {
		periodStartAt = time.Unix(stripeInvoice.PeriodStart, 0)
	}
	var periodEndAt time.Time
	if stripeInvoice.PeriodEnd != 0 {
		periodEndAt = time.Unix(stripeInvoice.PeriodEnd, 0)
	}
	return Invoice{
		ID:            "",
		ProviderID:    stripeInvoice.ID,
		CustomerID:    customerID,
		State:         string(stripeInvoice.Status),
		Currency:      string(stripeInvoice.Currency),
		Amount:        stripeInvoice.Total,
		HostedURL:     stripeInvoice.HostedInvoiceURL,
		Metadata:      metadata.FromString(stripeInvoice.Metadata),
		EffectiveAt:   effectiveAt,
		DueAt:         dueDate,
		CreatedAt:     createdAt,
		PeriodStartAt: periodStartAt,
		PeriodEndAt:   periodEndAt,
	}
}

func (s *Service) DeleteByCustomer(ctx context.Context, c customer.Customer) error {
	invoices, err := s.ListAll(ctx, Filter{
		CustomerID: c.ID,
	})
	if err != nil {
		return err
	}
	for _, i := range invoices {
		if err := s.repository.Delete(ctx, i.ID); err != nil {
			return err
		}
	}
	return nil
}
