package customer

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/robfig/cron/v3"

	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/internal/metrics"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Customer, error)
	List(ctx context.Context, filter Filter) ([]Customer, error)
	Create(ctx context.Context, customer Customer) (Customer, error)
	UpdateByID(ctx context.Context, customer Customer) (Customer, error)
	Delete(ctx context.Context, id string) error
	UpdateCreditMinByID(ctx context.Context, customerID string, limit int64) (Details, error)
	GetDetailsByID(ctx context.Context, customerID string) (Details, error)
	UpdateDetailsByID(ctx context.Context, customerID string, details Details) (Details, error)
}

type CreditService interface {
	GetBalance(ctx context.Context, id string) (int64, error)
}

type Service struct {
	provider      billing.Provider
	repository    Repository
	creditService CreditService

	syncJob   *cron.Cron
	mu        sync.Mutex
	syncDelay time.Duration
}

func NewService(provider billing.Provider, repository Repository, cfg billing.Config,
	creditService CreditService) *Service {
	return &Service{
		provider:      provider,
		repository:    repository,
		mu:            sync.Mutex{},
		syncDelay:     cfg.RefreshInterval.Customer,
		creditService: creditService,
	}
}

func (s *Service) Create(ctx context.Context, customer Customer, offline bool) (Customer, error) {
	// set defaults
	if customer.State == "" {
		customer.State = ActiveState
	}

	// do not allow creating a new customer account if there exists already an active billing account
	existingAccounts, err := s.repository.List(ctx, Filter{
		OrgID: customer.OrgID,
	})
	if err != nil {
		return Customer{}, err
	}
	activeAccounts := utils.Filter(existingAccounts, func(i Customer) bool {
		return i.State == ActiveState
	})
	if len(activeAccounts) > 0 {
		return Customer{}, ErrActiveConflict
	}

	// do not allow creating account if the balance of a previous account within org
	// is less than 0
	for _, account := range existingAccounts {
		if balance, err := s.creditService.GetBalance(ctx, account.ID); err == nil && balance < 0 {
			return Customer{}, ErrExistingAccountWithPendingDues
		}
	}

	// offline mode, we don't need to create the customer in billing provider
	if !offline {
		providerCustomer, err := s.RegisterToProvider(ctx, customer)
		if err != nil {
			return Customer{}, err
		}
		customer.ProviderID = providerCustomer.ID
	}
	return s.repository.Create(ctx, customer)
}

func (s *Service) RegisterToProvider(ctx context.Context, customer Customer) (*billing.ProviderCustomer, error) {
	var taxIDs []billing.ProviderTaxID
	for _, tax := range customer.TaxData {
		taxIDs = append(taxIDs, billing.ProviderTaxID{
			Type:  tax.Type,
			Value: tax.ID,
		})
	}
	return s.provider.CreateCustomer(ctx, billing.CreateCustomerParams{
		Email: customer.Email,
		Name:  customer.Name,
		Phone: customer.Phone,
		Address: billing.ProviderAddress{
			City:       customer.Address.City,
			Country:    customer.Address.Country,
			Line1:      customer.Address.Line1,
			Line2:      customer.Address.Line2,
			PostalCode: customer.Address.PostalCode,
			State:      customer.Address.State,
		},
		TaxIDs: taxIDs,
		Metadata: map[string]string{
			"org_id":     customer.OrgID,
			"managed_by": "frontier",
		},
		TestClockID: customer.StripeTestClockID,
	})
}

func (s *Service) RegisterToProviderIfRequired(ctx context.Context, customerID string) (Customer, error) {
	custmr, err := s.repository.GetByID(ctx, customerID)
	if err != nil {
		return Customer{}, err
	}
	if custmr.IsOffline() {
		providerCustomer, err := s.RegisterToProvider(ctx, custmr)
		if err != nil {
			return Customer{}, err
		}
		custmr.ProviderID = providerCustomer.ID
		return s.repository.UpdateByID(ctx, custmr)
	}
	return custmr, nil
}

func (s *Service) Update(ctx context.Context, customer Customer) (Customer, error) {
	existingCustomer, err := s.repository.GetByID(ctx, customer.ID)
	if err != nil {
		return Customer{}, err
	}

	// Always infer org_id from existing customer (ignore from request for security)
	customer.OrgID = existingCustomer.OrgID

	providerCustomer, err := s.provider.UpdateCustomer(ctx, existingCustomer.ProviderID, billing.UpdateCustomerParams{
		Email: customer.Email,
		Name:  customer.Name,
		Phone: customer.Phone,
		Address: billing.ProviderAddress{
			City:       customer.Address.City,
			Country:    customer.Address.Country,
			Line1:      customer.Address.Line1,
			Line2:      customer.Address.Line2,
			PostalCode: customer.Address.PostalCode,
			State:      customer.Address.State,
		},
		Metadata: map[string]string{
			"org_id":     existingCustomer.OrgID,
			"managed_by": "frontier",
		},
	})
	if err != nil {
		return Customer{}, err
	}
	customer.ProviderID = providerCustomer.ID
	return s.repository.UpdateByID(ctx, customer)
}

func (s *Service) GetByID(ctx context.Context, id string) (Customer, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Customer, error) {
	return s.repository.List(ctx, filter)
}

func (s *Service) GetByOrgID(ctx context.Context, orgID string) (Customer, error) {
	if len(orgID) == 0 {
		return Customer{}, ErrInvalidUUID
	}
	custs, err := s.repository.List(ctx, Filter{
		OrgID: orgID,
	})
	if err != nil {
		return Customer{}, err
	}
	if len(custs) == 0 {
		return Customer{}, ErrNotFound
	}
	// Note: maybe we support more than one billing account with the same orgID
	return custs[0], nil
}

func (s *Service) Enable(ctx context.Context, id string) error {
	customer, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if customer.State == ActiveState {
		return nil
	}

	// make sure there doesn't exist an active account for the organization already
	existingAccounts, err := s.repository.List(ctx, Filter{
		OrgID: customer.OrgID,
		State: ActiveState,
	})
	if err != nil {
		return err
	}
	if len(existingAccounts) > 0 {
		return ErrActiveConflict
	}

	customer.State = ActiveState
	_, err = s.repository.UpdateByID(ctx, customer)
	return err
}

func (s *Service) Disable(ctx context.Context, id string) error {
	customer, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if customer.State == DisabledState {
		return nil
	}
	customer.State = DisabledState
	_, err = s.repository.UpdateByID(ctx, customer)
	return err
}

func (s *Service) Delete(ctx context.Context, id string) error {
	customer, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// TODO: cancel and delete all subscriptions before deleting the customer

	if customer.ProviderID != "" {
		if err = s.provider.DeleteCustomer(ctx, customer.ProviderID); err != nil {
			return err
		}
	}

	return s.repository.Delete(ctx, id)
}

func (s *Service) ListPaymentMethods(ctx context.Context, id string) ([]PaymentMethod, error) {
	customer, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var paymentMethods []PaymentMethod

	if customer.ProviderID == "" {
		return paymentMethods, nil
	}

	providerMethods, err := s.provider.ListPaymentMethods(ctx, customer.ProviderID)
	if err != nil {
		return nil, err
	}

	for _, pm := range providerMethods {
		m := PaymentMethod{
			ID:              "",
			CustomerID:      customer.ID,
			ProviderID:      pm.ID,
			Type:            pm.Type,
			CardBrand:       pm.CardBrand,
			CardLast4:       pm.CardLast4,
			CardExpiryMonth: pm.CardExpiryMonth,
			CardExpiryYear:  pm.CardExpiryYear,
			Metadata:        pm.Metadata,
			CreatedAt:       time.Unix(pm.CreatedAt, 0),
		}
		if pm.IsDefault {
			if m.Metadata == nil {
				m.Metadata = make(map[string]any)
			}
			m.Metadata["default"] = true
		}
		paymentMethods = append(paymentMethods, m)
	}
	return paymentMethods, nil
}

func (s *Service) Init(ctx context.Context) error {
	if s.syncDelay == time.Duration(0) {
		return nil
	}
	if s.syncJob != nil {
		<-s.syncJob.Stop().Done()
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
	return nil
}

func (s *Service) Close() error {
	if s.syncJob != nil {
		<-s.syncJob.Stop().Done()
		return s.syncJob.Stop().Err()
	}
	return nil
}

func (s *Service) backgroundSync(ctx context.Context) {
	start := time.Now()
	if metrics.BillingSyncLatency != nil {
		record := metrics.BillingSyncLatency("customer")
		defer record()
	}
	logger := grpczap.Extract(ctx)
	customers, err := s.List(ctx, Filter{
		State: ActiveState,
	})
	if err != nil {
		logger.Error("customer.backgroundSync", zap.Error(err))
		return
	}

	for _, customer := range customers {
		if ctx.Err() != nil {
			// stop processing if context is done
			break
		}

		if customer.DeletedAt != nil || customer.IsOffline() {
			continue
		}
		if err := s.SyncWithProvider(ctx, customer); err != nil {
			logger.Error("customer.SyncWithProvider", zap.Error(err), zap.String("customer_id", customer.ID))
		}
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}
	logger.Info("customer.backgroundSync finished", zap.Duration("duration", time.Since(start)))
}

// SyncWithProvider syncs the customer state with the billing provider
func (s *Service) SyncWithProvider(ctx context.Context, customr Customer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	providerCustomer, err := s.provider.GetCustomer(ctx, customr.ProviderID)
	if err != nil {
		return err
	}

	var shouldUpdate bool
	if providerCustomer.Deleted {
		// customer is deleted in the billing provider, we don't enable them back automatically
		if customr.State != DisabledState {
			customr.State = DisabledState
			shouldUpdate = true
		}
	}
	if customr.IsActive() {
		// don't update for disabled state
		if len(providerCustomer.TaxIDs) > 0 {
			var taxData []Tax
			for _, taxID := range providerCustomer.TaxIDs {
				taxData = append(taxData, Tax{
					ID:   taxID.Value,
					Type: taxID.Type,
				})
			}
			if !slices.EqualFunc(customr.TaxData, taxData, func(a Tax, b Tax) bool {
				return a.ID == b.ID && a.Type == b.Type
			}) {
				customr.TaxData = taxData
				shouldUpdate = true
			}
		}
		if providerCustomer.Phone != customr.Phone {
			customr.Phone = providerCustomer.Phone
			shouldUpdate = true
		}
		if providerCustomer.Email != "" && providerCustomer.Email != customr.Email {
			customr.Email = providerCustomer.Email
			shouldUpdate = true
		}
		if providerCustomer.Name != customr.Name {
			customr.Name = providerCustomer.Name
			shouldUpdate = true
		}
		if providerCustomer.Currency != "" && providerCustomer.Currency != customr.Currency {
			customr.Currency = providerCustomer.Currency
			shouldUpdate = true
		}
		if providerCustomer.Address.City != customr.Address.City ||
			providerCustomer.Address.Country != customr.Address.Country ||
			providerCustomer.Address.Line1 != customr.Address.Line1 ||
			providerCustomer.Address.Line2 != customr.Address.Line2 ||
			providerCustomer.Address.PostalCode != customr.Address.PostalCode ||
			providerCustomer.Address.State != customr.Address.State {
			customr.Address = Address{
				City:       providerCustomer.Address.City,
				Country:    providerCustomer.Address.Country,
				Line1:      providerCustomer.Address.Line1,
				Line2:      providerCustomer.Address.Line2,
				PostalCode: providerCustomer.Address.PostalCode,
				State:      providerCustomer.Address.State,
			}
			shouldUpdate = true
		}
	}
	if shouldUpdate {
		if _, err := s.repository.UpdateByID(ctx, customr); err != nil {
			return fmt.Errorf("failed to update customer in frontier: %w", err)
		}
	}
	return nil
}

func (s *Service) TriggerSyncByProviderID(ctx context.Context, id string) error {
	customrs, err := s.repository.List(ctx, Filter{
		ProviderID: id,
	})
	if err != nil {
		return err
	}
	if len(customrs) == 0 {
		return ErrNotFound
	}
	return s.SyncWithProvider(ctx, customrs[0])
}

func (s *Service) UpdateCreditMinByID(ctx context.Context, customerID string, limit int64) (Details, error) {
	return s.repository.UpdateCreditMinByID(ctx, customerID, limit)
}

func (s *Service) GetDetails(ctx context.Context, customerID string) (Details, error) {
	return s.repository.GetDetailsByID(ctx, customerID)
}

func (s *Service) UpdateDetails(ctx context.Context, customerID string, details Details) (Details, error) {
	return s.repository.UpdateDetailsByID(ctx, customerID, details)
}
