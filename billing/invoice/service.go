package invoice

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/raystack/frontier/pkg/db"

	"github.com/google/uuid"
	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/product"

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

type CreditService interface {
	GetBalance(ctx context.Context, accountID string) (int64, error)
	Add(ctx context.Context, cred credit.Credit) error
}

type ProductService interface {
	GetByID(ctx context.Context, id string) (product.Product, error)
}

type Locker interface {
	TryLock(ctx context.Context, id string) (*db.Lock, error)
}

type Service struct {
	stripeClient    *client.API
	repository      Repository
	customerService CustomerService
	creditService   CreditService
	productService  ProductService
	locker          Locker

	syncJob   *cron.Cron
	mu        sync.Mutex
	syncDelay time.Duration

	stripeAutoTax                  bool
	creditOverdraftProduct         string
	creditOverdraftUnitAmount      int64
	creditOverdraftInvoiceCurrency string
	creditOverdraftInvoiceDOM      int
}

func NewService(stripeClient *client.API, invoiceRepository Repository,
	customerService CustomerService, creditService CreditService, productService ProductService,
	locker Locker, cfg billing.Config) *Service {
	return &Service{
		stripeClient:              stripeClient,
		repository:                invoiceRepository,
		customerService:           customerService,
		creditService:             creditService,
		productService:            productService,
		locker:                    locker,
		syncDelay:                 cfg.RefreshInterval.Invoice,
		stripeAutoTax:             cfg.StripeAutoTax,
		creditOverdraftProduct:    cfg.AccountConfig.CreditOverdraftProduct,
		creditOverdraftInvoiceDOM: 1, // 1st day of month
	}
}

func (s *Service) Init(ctx context.Context) error {
	logger := grpczap.Extract(ctx)
	if s.syncJob != nil {
		s.syncJob.Stop()
	}
	s.syncJob = cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))

	if _, err := s.syncJob.AddFunc(fmt.Sprintf("@every %s", s.syncDelay.String()), func() {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		s.backgroundSync(ctx)
	}); err != nil {
		return err
	}
	s.syncJob.Start()

	if s.creditOverdraftProduct != "" {
		creditProduct, err := s.productService.GetByID(ctx, s.creditOverdraftProduct)
		if err != nil {
			return fmt.Errorf("failed to get credit overdraft product: %w", err)
		}
		if creditProduct.Behavior != product.CreditBehavior {
			return errors.New("credit overdraft product must have credit behavior")
		}
		// get first price
		if len(creditProduct.Prices) == 0 {
			return errors.New("credit overdraft product must have at least one price")
		}
		creditPrice := creditProduct.Prices[0]
		if creditPrice.Currency == "" {
			return errors.New("credit overdraft product price must have a currency")
		}
		s.creditOverdraftInvoiceCurrency = creditPrice.Currency
		s.creditOverdraftUnitAmount = int64(float64(creditPrice.Amount) / float64(creditProduct.Config.CreditAmount))
		logger.Info("credit overdraft product details",
			zap.Int64("unit_amount", s.creditOverdraftUnitAmount),
			zap.String("currency", s.creditOverdraftInvoiceCurrency))
	}
	return nil
}

func (s *Service) Close() error {
	if s.syncJob != nil {
		return s.syncJob.Stop().Err()
	}
	return nil
}

func (s *Service) backgroundSync(ctx context.Context) {
	start := time.Now()
	if metrics.BillingSyncLatency != nil {
		record := metrics.BillingSyncLatency("invoice")
		defer record()
	}
	logger := grpczap.Extract(ctx)
	customers, err := s.customerService.List(ctx, customer.Filter{
		Online: utils.Bool(true),
	})
	if err != nil {
		logger.Error("invoice.backgroundSync", zap.Error(err))
		return
	}
	for _, customr := range customers {
		if ctx.Err() != nil {
			// stop processing if context is done
			break
		}

		if !customr.IsActive() {
			continue
		}
		if err := s.SyncWithProvider(ctx, customr); err != nil {
			logger.Error("invoice.SyncWithProvider", zap.Error(err))
		}
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
	if err := s.Reconcile(ctx); err != nil {
		logger.Error("invoice.Reconcile", zap.Error(err))
	}
	if now := time.Now().UTC(); now.Day() == s.creditOverdraftInvoiceDOM {
		if err := s.GenerateForCredits(ctx); err != nil {
			logger.Error("invoice.GenerateForCredits", zap.Error(err))
		}
	}
	logger.Info("invoice.backgroundSync finished", zap.Duration("duration", time.Since(start)))
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
		Expand: []*string{
			stripe.String("data.lines"),
		},
	})
	for stripeInvoices.Next() {
		stripeInvoice := stripeInvoices.Invoice()

		// check if already present, if yes, update else create new
		existingInvoice, ok := utils.FindFirst(invoiceObs, func(i Invoice) bool {
			return i.ProviderID == stripeInvoice.ID
		})
		if ok {
			err = s.upsert(ctx, customr.ID, &existingInvoice, stripeInvoice)
		} else {
			err = s.upsert(ctx, customr.ID, nil, stripeInvoice)
		}
		if err != nil {
			errs = append(errs, err)
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

func (s *Service) upsert(ctx context.Context, customerID string,
	existingInvoice *Invoice, stripeInvoice *stripe.Invoice) error {
	if existingInvoice != nil {
		// already present in our system, update it if needed
		updateNeeded := false
		if existingInvoice.State != State(stripeInvoice.Status) {
			existingInvoice.State = State(stripeInvoice.Status)
			updateNeeded = true
		}
		if stripeInvoice.EffectiveAt != 0 && existingInvoice.EffectiveAt != utils.AsTimeFromEpoch(stripeInvoice.EffectiveAt) {
			existingInvoice.EffectiveAt = utils.AsTimeFromEpoch(stripeInvoice.EffectiveAt)
			updateNeeded = true
		}
		if stripeInvoice.HostedInvoiceURL != "" && existingInvoice.HostedURL != stripeInvoice.HostedInvoiceURL {
			existingInvoice.HostedURL = stripeInvoice.HostedInvoiceURL
			updateNeeded = true
		}

		if updateNeeded {
			if _, err := s.repository.UpdateByID(ctx, *existingInvoice); err != nil {
				return fmt.Errorf("failed to update invoice %s: %w", existingInvoice.ID, err)
			}
		}
	} else {
		if _, err := s.repository.Create(ctx, stripeInvoiceToInvoice(customerID, stripeInvoice)); err != nil {
			return fmt.Errorf("failed to create invoice for customer %s: %w", customerID, err)
		}
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

// GetUpcoming returns the upcoming invoice for the customer based on the
// active subscription plan. If no upcoming invoice is found, it returns empty.
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
	var items []Item
	if stripeInvoice.Lines != nil {
		for _, line := range stripeInvoice.Lines.Data {
			item := Item{
				ID:         uuid.New().String(),
				ProviderID: line.ID,
				Name:       line.Description,
				Type:       ItemType(line.Metadata[ItemTypeMetadataKey]),
				Quantity:   line.Quantity,
			}
			if line.Price != nil {
				item.UnitAmount = line.Price.UnitAmount
			}
			items = append(items, item)
		}
	}
	return Invoice{
		ID:            "",
		ProviderID:    stripeInvoice.ID,
		CustomerID:    customerID,
		State:         State(stripeInvoice.Status),
		Currency:      string(stripeInvoice.Currency),
		Amount:        stripeInvoice.Total,
		HostedURL:     stripeInvoice.HostedInvoiceURL,
		Metadata:      metadata.FromString(stripeInvoice.Metadata),
		EffectiveAt:   effectiveAt,
		DueAt:         dueDate,
		CreatedAt:     createdAt,
		PeriodStartAt: periodStartAt,
		PeriodEndAt:   periodEndAt,
		Items:         items,
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

// GenerateForCredits finds all customers which has credit min limit lower than
// 0, that is, allows for negative balance and generates an invoice for them.
// Invoices will be paid asynchronously by the customer but system need to
// reconcile the token balance once it's paid.
func (s *Service) GenerateForCredits(ctx context.Context) error {
	var errs []error
	logger := grpczap.Extract(ctx)
	if s.creditOverdraftUnitAmount == 0 || s.creditOverdraftInvoiceCurrency == "" {
		// do not process if credit overdraft details not set
		return nil
	}

	// ensure only one of this job is running at a time
	lock, err := s.locker.TryLock(ctx, GenerateForCreditLockKey)
	if err != nil {
		if errors.Is(err, db.ErrLockBusy) {
			// someone else has the lock, return
			return nil
		}
		return err
	}
	defer func() {
		unlockErr := lock.Unlock(ctx)
		if unlockErr != nil {
			logger.Error("failed to unlock", zap.Error(unlockErr), zap.String("key", GenerateForCreditLockKey))
		}
	}()

	customers, err := s.customerService.List(ctx, customer.Filter{
		Online:           utils.Bool(true),
		AllowedOverdraft: utils.Bool(true),
	})
	if err != nil {
		return err
	}
	for _, c := range customers {
		if ctx.Err() != nil {
			// stop processing if context is done
			break
		}

		balance, err := s.creditService.GetBalance(ctx, c.ID)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get balance for customer %s: %w", c.ID, err))
			continue
		}
		if balance >= 0 {
			continue
		}

		// check if there is already an invoice open for this balance
		invoices, err := s.List(ctx, Filter{
			CustomerID: c.ID,
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to list invoices for customer %s: %w", c.ID, err))
			continue
		}
		// check if invoice line items are of type credit
		// if yes, don't create a new invoice
		var alreadyInvoiced bool
		for _, i := range invoices {
			if i.State == DraftState || i.State == OpenState {
				for _, item := range i.Items {
					if item.Type == CreditItemType {
						alreadyInvoiced = true
					}
				}
			}
		}
		if alreadyInvoiced {
			continue
		}

		// create invoice for the credit overdraft
		items := []Item{
			{
				Name:       "Credit Overdraft",
				Type:       CreditItemType,
				UnitAmount: s.creditOverdraftUnitAmount,
				Quantity:   abs(balance),
			},
		}
		newStripeInvoice, err := s.CreateInProvider(ctx, c, items, s.creditOverdraftInvoiceCurrency)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create invoice for customer %s: %w", c.ID, err))
			continue
		}
		// sync back new invoice
		if err := s.upsert(ctx, c.ID, nil, newStripeInvoice); err != nil {
			errs = append(errs, fmt.Errorf("failed to sync invoice for customer %s: %w", c.ID, err))
			continue
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// CreateInProvider creates a custom invoice with items in the provider.
// Once created the invoice object will be synced back within system using
// regular syncer/webhook loop.
func (s *Service) CreateInProvider(ctx context.Context, custmr customer.Customer,
	items []Item, currency string) (*stripe.Invoice, error) {
	stripeInvoice, err := s.stripeClient.Invoices.New(&stripe.InvoiceParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Customer:    stripe.String(custmr.ProviderID),
		AutoAdvance: stripe.Bool(true),
		Description: stripe.String("Invoice for the underpayment of credit utilization"),
		AutomaticTax: &stripe.InvoiceAutomaticTaxParams{
			Enabled: stripe.Bool(s.stripeAutoTax),
		},
		Currency:                    stripe.String(currency),
		PendingInvoiceItemsBehavior: stripe.String("include"),
		Metadata: map[string]string{
			"org_id":     custmr.OrgID,
			"managed_by": "frontier",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// create line item for the invoice
	for _, item := range items {
		_, err = s.stripeClient.InvoiceItems.New(&stripe.InvoiceItemParams{
			Params: stripe.Params{
				Context: ctx,
			},
			Customer:   stripe.String(custmr.ProviderID),
			Currency:   stripe.String(custmr.Currency),
			Invoice:    stripe.String(stripeInvoice.ID),
			UnitAmount: &item.UnitAmount,
			Quantity:   &item.Quantity,
			Metadata: map[string]string{
				"org_id":     custmr.OrgID,
				"managed_by": "frontier",
				// type is used to identify the item type in the invoice
				// this is useful when reconciling the invoice items for payments and
				// avoid creating duplicate invoices
				ItemTypeMetadataKey: item.Type.String(),
			},
			Description: stripe.String(item.Name),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}
	}
	return stripeInvoice, nil
}

// Reconcile checks all paid invoices and reconciles them with the system.
// If the invoice was created for credit overdraft, it will credit the customer
// account with the amount of the invoice.
func (s *Service) Reconcile(ctx context.Context) error {
	if s.creditOverdraftUnitAmount == 0 {
		// do not process if credit overdraft details not set as currently
		// we only reconcile credit overdraft invoices
		return nil
	}

	invoices, err := s.ListAll(ctx, Filter{
		State:       PaidState,
		NonZeroOnly: true,
	})
	if err != nil {
		return err
	}
	var errs []error
	for _, i := range invoices {
		if ctx.Err() != nil {
			// stop processing if context is done
			break
		}

		// check if already reconciled
		if i.Metadata != nil && i.Metadata[ReconciledMetadataKey] == true {
			continue
		}

		if err := s.reconcileCreditInvoice(ctx, i); err != nil {
			errs = append(errs, fmt.Errorf("failed to reconcile invoice %s: %w", i.ID, err))
			continue
		}

		// mark invoices reconciled to avoid processing them in future
		if i.Metadata == nil {
			i.Metadata = make(map[string]any)
		}
		i.Metadata[ReconciledMetadataKey] = true
		if _, err := s.repository.UpdateByID(ctx, i); err != nil {
			errs = append(errs, fmt.Errorf("failed to update invoice metadata %s: %w", i.ID, err))
			continue
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

func (s *Service) reconcileCreditInvoice(ctx context.Context, i Invoice) error {
	if i.State != PaidState {
		return nil
	}
	var creditItems []Item
	for _, item := range i.Items {
		if item.Type == CreditItemType {
			creditItems = append(creditItems, item)
		}
	}
	if len(creditItems) == 0 {
		return nil
	}
	for _, item := range creditItems {
		// credit the customer account
		if err := s.creditService.Add(ctx, credit.Credit{
			ID:          credit.TxUUID(i.ID, item.ProviderID),
			CustomerID:  i.CustomerID,
			Amount:      item.Quantity,
			Source:      credit.SourceSystemOverdraftEvent,
			Description: "Paid for credit overdraft invoice",
			Metadata: map[string]any{
				"invoice_id": i.ID,
				"overdraft":  true,
				"item":       item.ProviderID,
			},
		}); err != nil {
			if errors.Is(err, credit.ErrAlreadyApplied) {
				continue
			}
			return fmt.Errorf("failed to credit customer %s: %w", i.CustomerID, err)
		}
	}
	return nil
}
