package customer

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/core/organization"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/client"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Customer, error)
	List(ctx context.Context, filter Filter) ([]Customer, error)
	Create(ctx context.Context, customer Customer) (Customer, error)
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
	if _, err = s.stripeClient.Customers.Del(customer.ProviderID, &stripe.CustomerParams{
		Params: stripe.Params{
			Context: ctx,
		},
	}); err != nil {
		return fmt.Errorf("failed to delete customer from billing provider: %w", err)
	}
	return s.repository.Delete(ctx, id)
}
