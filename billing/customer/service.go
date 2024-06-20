package customer

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/raystack/frontier/pkg/metadata"

	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/client"
)

const (
	SyncDelay = time.Second * 60
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Customer, error)
	List(ctx context.Context, filter Filter) ([]Customer, error)
	Create(ctx context.Context, customer Customer) (Customer, error)
	UpdateByID(ctx context.Context, customer Customer) (Customer, error)
	Delete(ctx context.Context, id string) error
}

type Service struct {
	stripeClient *client.API
	repository   Repository

	syncJob *cron.Cron
	mu      sync.Mutex
}

func NewService(stripeClient *client.API, repository Repository) *Service {
	return &Service{
		stripeClient: stripeClient,
		repository:   repository,
		mu:           sync.Mutex{},
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
		State: ActiveState,
	})
	if err != nil {
		return Customer{}, err
	}
	if len(existingAccounts) > 0 {
		return Customer{}, ErrActiveConflict
	}

	// offline mode, we don't need to create the customer in billing provider
	if !offline {
		stripeCustomer, err := s.RegisterToProvider(ctx, customer)
		if err != nil {
			return Customer{}, err
		}
		customer.ProviderID = stripeCustomer.ID
	}
	return s.repository.Create(ctx, customer)
}

func (s *Service) RegisterToProvider(ctx context.Context, customer Customer) (*stripe.Customer, error) {
	// create a new customer in stripe
	var customerTaxes []*stripe.CustomerTaxIDDataParams = nil
	for _, tax := range customer.TaxData {
		customerTaxes = append(customerTaxes, &stripe.CustomerTaxIDDataParams{
			Type:  stripe.String(tax.Type),
			Value: stripe.String(tax.ID),
		})
	}
	// create a new customer in stripe
	stripeCustomer, err := s.stripeClient.Customers.New(&stripe.CustomerParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Address: &stripe.AddressParams{
			City:       &customer.Address.City,
			Country:    &customer.Address.Country,
			Line1:      &customer.Address.Line1,
			Line2:      &customer.Address.Line2,
			PostalCode: &customer.Address.PostalCode,
			State:      &customer.Address.State,
		},
		Email:     &customer.Email,
		Name:      &customer.Name,
		Phone:     &customer.Phone,
		TaxIDData: customerTaxes,
		Metadata: map[string]string{
			"org_id":     customer.OrgID,
			"managed_by": "frontier",
		},
		TestClock: customer.StripeTestClockID,
	})
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			switch stripeErr.Code {
			case stripe.ErrorCodeParameterMissing:
				// stripe error
				return nil, fmt.Errorf("missing parameter while registering to biller: %s", stripeErr.Error())
			}
		}
		return nil, fmt.Errorf("failed to register in billing provider: %w", err)
	}

	return stripeCustomer, nil
}

func (s *Service) RegisterToProviderIfRequired(ctx context.Context, customerID string) (Customer, error) {
	custmr, err := s.repository.GetByID(ctx, customerID)
	if err != nil {
		return Customer{}, err
	}
	if custmr.IsOffline() {
		stripeCustomer, err := s.RegisterToProvider(ctx, custmr)
		if err != nil {
			return Customer{}, err
		}
		custmr.ProviderID = stripeCustomer.ID
		return s.repository.UpdateByID(ctx, custmr)
	}
	return custmr, nil
}

func (s *Service) Update(ctx context.Context, customer Customer) (Customer, error) {
	existingCustomer, err := s.repository.GetByID(ctx, customer.ID)
	if err != nil {
		return Customer{}, err
	}

	// update a customer in stripe
	stripeCustomer, err := s.stripeClient.Customers.Update(existingCustomer.ProviderID, &stripe.CustomerParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Address: &stripe.AddressParams{
			City:       &customer.Address.City,
			Country:    &customer.Address.Country,
			Line1:      &customer.Address.Line1,
			Line2:      &customer.Address.Line2,
			PostalCode: &customer.Address.PostalCode,
			State:      &customer.Address.State,
		},
		Email: &customer.Email,
		Name:  &customer.Name,
		Phone: &customer.Phone,
		Metadata: map[string]string{
			"org_id":     customer.OrgID,
			"managed_by": "frontier",
		},
	})
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			switch stripeErr.Code {
			case stripe.ErrorCodeParameterMissing:
				// stripe error
				return Customer{}, fmt.Errorf("missing parameter while registering to biller: %s", stripeErr.Error())
			}
		}
		return Customer{}, fmt.Errorf("failed to register in billing provider: %w", err)
	}
	customer.ProviderID = stripeCustomer.ID
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

	// deleting customer cancel all of its plans
	if _, err = s.stripeClient.Customers.Del(customer.ProviderID, &stripe.CustomerParams{
		Params: stripe.Params{
			Context: ctx,
		},
	}); err != nil {
		var throw = true
		// Try to safely cast a generic error to a stripe.Error so that we can get at
		// some additional Stripe-specific information about what went wrong.
		if stripeErr, ok := err.(*stripe.Error); ok {
			// The Code field will contain a basic identifier for the failure.
			if stripeErr.Code == stripe.ErrorCodeResourceMissing {
				// it's ok if the customer is already deleted
				throw = false
			}
		}
		if throw {
			return fmt.Errorf("failed to delete customer from billing provider: %w", err)
		}
	}

	return s.repository.Delete(ctx, id)
}

func (s *Service) ListPaymentMethods(ctx context.Context, id string) ([]PaymentMethod, error) {
	customer, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	stripePaymentMethodItr := s.stripeClient.PaymentMethods.List(&stripe.PaymentMethodListParams{
		Customer: stripe.String(customer.ProviderID),
		ListParams: stripe.ListParams{
			Context: ctx,
		},
		Expand: []*string{
			stripe.String("data.customer"),
		},
	})
	var paymentMethods []PaymentMethod
	for stripePaymentMethodItr.Next() {
		stripePaymentMethod := stripePaymentMethodItr.PaymentMethod()
		pm := PaymentMethod{
			ID:         "",
			CustomerID: customer.ID,
			ProviderID: stripePaymentMethod.ID,
			Type:       string(stripePaymentMethod.Type),
			Metadata:   metadata.FromString(stripePaymentMethod.Metadata),
			CreatedAt:  time.Unix(stripePaymentMethod.Created, 0),
		}
		if stripePaymentMethod.Type == stripe.PaymentMethodTypeCard {
			pm.CardBrand = string(stripePaymentMethod.Card.Brand)
			pm.CardLast4 = stripePaymentMethod.Card.Last4
			pm.CardExpiryMonth = stripePaymentMethod.Card.ExpMonth
			pm.CardExpiryYear = stripePaymentMethod.Card.ExpYear
		}
		// set default method, if any
		if stripePaymentMethod.Customer.InvoiceSettings != nil && stripePaymentMethod.Customer.InvoiceSettings.DefaultPaymentMethod != nil {
			if stripePaymentMethod.Customer.InvoiceSettings.DefaultPaymentMethod.ID == stripePaymentMethod.ID {
				pm.Metadata["default"] = true
			}
		}

		paymentMethods = append(paymentMethods, pm)
	}
	return paymentMethods, nil
}

func (s *Service) Init(ctx context.Context) error {
	if s.syncJob != nil {
		<-s.syncJob.Stop().Done()
	}

	s.syncJob = cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))
	if _, err := s.syncJob.AddFunc(fmt.Sprintf("@every %s", SyncDelay.String()), func() {
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
	logger := grpczap.Extract(ctx)
	customers, err := s.List(ctx, Filter{
		State: ActiveState,
	})
	if err != nil {
		logger.Error("customer.backgroundSync", zap.Error(err))
		return
	}

	for _, customer := range customers {
		if customer.DeletedAt != nil || customer.IsOffline() {
			continue
		}
		if err := s.SyncWithProvider(ctx, customer); err != nil {
			logger.Error("customer.SyncWithProvider", zap.Error(err), zap.String("customer_id", customer.ID))
		}
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	}
}

// SyncWithProvider syncs the customer state with the billing provider
func (s *Service) SyncWithProvider(ctx context.Context, customr Customer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	stripeCustomer, err := s.stripeClient.Customers.Get(customr.ProviderID, &stripe.CustomerParams{
		Params: stripe.Params{
			Context: ctx,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get customer from billing provider: %w", err)
	}
	var shouldUpdate bool
	if stripeCustomer.Deleted {
		// customer is deleted in the billing provider, we don't enable them back automatically
		if customr.State != DisabledState {
			customr.State = DisabledState
			shouldUpdate = true
		}
	}
	if stripeCustomer.TaxIDs != nil {
		var taxData []Tax
		for _, taxID := range stripeCustomer.TaxIDs.Data {
			taxData = append(taxData, Tax{
				ID:   taxID.Value,
				Type: string(taxID.Type),
			})
		}
		if !slices.EqualFunc(customr.TaxData, taxData, func(a Tax, b Tax) bool {
			return a.ID == b.ID && a.Type == b.Type
		}) {
			customr.TaxData = taxData
			shouldUpdate = true
		}
	}
	if stripeCustomer.Phone != customr.Phone {
		customr.Phone = stripeCustomer.Phone
		shouldUpdate = true
	}
	if stripeCustomer.Email != customr.Email {
		customr.Email = stripeCustomer.Email
		shouldUpdate = true
	}
	if stripeCustomer.Name != customr.Name {
		customr.Name = stripeCustomer.Name
		shouldUpdate = true
	}
	if stripeCustomer.Currency != "" && string(stripeCustomer.Currency) != customr.Currency {
		customr.Currency = string(stripeCustomer.Currency)
		shouldUpdate = true
	}
	if stripeCustomer.Address != nil {
		if stripeCustomer.Address.City != customr.Address.City ||
			stripeCustomer.Address.Country != customr.Address.Country ||
			stripeCustomer.Address.Line1 != customr.Address.Line1 ||
			stripeCustomer.Address.Line2 != customr.Address.Line2 ||
			stripeCustomer.Address.PostalCode != customr.Address.PostalCode ||
			stripeCustomer.Address.State != customr.Address.State {
			customr.Address = Address{
				City:       stripeCustomer.Address.City,
				Country:    stripeCustomer.Address.Country,
				Line1:      stripeCustomer.Address.Line1,
				Line2:      stripeCustomer.Address.Line2,
				PostalCode: stripeCustomer.Address.PostalCode,
				State:      stripeCustomer.Address.State,
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
