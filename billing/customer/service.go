package customer

import (
	"context"
	"fmt"
	"time"

	"github.com/raystack/frontier/pkg/metadata"

	"github.com/raystack/frontier/core/organization"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/client"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Customer, error)
	List(ctx context.Context, filter Filter) ([]Customer, error)
	Create(ctx context.Context, customer Customer) (Customer, error)
	UpdateByID(ctx context.Context, customer Customer) (Customer, error)
	Delete(ctx context.Context, id string) error
}

type OrgService interface {
	Get(ctx context.Context, id string) (organization.Organization, error)
}

type Service struct {
	stripeClient *client.API
	orgService   OrgService
	repository   Repository
}

func NewService(stripeClient *client.API, repository Repository, orgService OrgService) *Service {
	return &Service{
		stripeClient: stripeClient,
		repository:   repository,
		orgService:   orgService,
	}
}

func (s Service) Create(ctx context.Context, customer Customer) (Customer, error) {
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
		Email: &customer.Email,
		Name:  &customer.Name,
		Phone: &customer.Phone,
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
				return Customer{}, fmt.Errorf("missing parameter while registering to biller: %s", stripeErr.Error())
			}
		}
		return Customer{}, fmt.Errorf("failed to register in billing provider: %w", err)
	}
	customer.ProviderID = stripeCustomer.ID
	return s.repository.Create(ctx, customer)
}

func (s Service) Update(ctx context.Context, customer Customer) (Customer, error) {
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

func (s Service) GetByID(ctx context.Context, id string) (Customer, error) {
	return s.repository.GetByID(ctx, id)
}

func (s Service) List(ctx context.Context, filter Filter) ([]Customer, error) {
	return s.repository.List(ctx, filter)
}

func (s Service) GetByOrgID(ctx context.Context, orgID string) (Customer, error) {
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

func (s Service) Delete(ctx context.Context, id string) error {
	customer, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// TODO: cancel and delete all subscriptions before deleting the customer

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

func (s Service) ListPaymentMethods(ctx context.Context, id string) ([]PaymentMethod, error) {
	customer, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	stripePaymentMethodItr := s.stripeClient.PaymentMethods.List(&stripe.PaymentMethodListParams{
		Customer: stripe.String(customer.ProviderID),
		ListParams: stripe.ListParams{
			Context: ctx,
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
		paymentMethods = append(paymentMethods, pm)
	}
	return paymentMethods, nil
}
