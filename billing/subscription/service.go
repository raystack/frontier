package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/raystack/frontier/billing/customer"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/client"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Subscription, error)
	Create(ctx context.Context, customer Subscription) (Subscription, error)
	UpdateByID(ctx context.Context, customer Subscription) (Subscription, error)
	List(ctx context.Context, filter Filter) ([]Subscription, error)
}

type CustomerService interface {
	GetByID(ctx context.Context, id string) (customer.Customer, error)
}

type Service struct {
	repository      Repository
	stripeClient    *client.API
	customerService CustomerService
}

func NewService(stripeClient *client.API, repository Repository, customerService CustomerService) *Service {
	return &Service{
		stripeClient:    stripeClient,
		repository:      repository,
		customerService: customerService,
	}
}

func (s *Service) Create(ctx context.Context, sub Subscription) (Subscription, error) {
	stripeCustomer := &stripe.Customer{}

	// get billing
	billingCustomer, err := s.customerService.GetByID(ctx, sub.CustomerID)
	if err != nil {
		return Subscription{}, err
	}

	stripeSubscription, err := s.stripeClient.Subscriptions.New(&stripe.SubscriptionParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Currency: &billingCustomer.Currency,
		Customer: &stripeCustomer.ID,
		Metadata: map[string]string{
			"org_id":     billingCustomer.OrgID,
			"managed_by": "frontier",
		},
	})
	if err != nil {
		return Subscription{}, fmt.Errorf("failed to create subscription at billing provider: %w", err)
	}

	sub.ProviderID = stripeSubscription.ID
	return s.repository.Create(ctx, sub)
}

func (s *Service) GetByID(ctx context.Context, id string) (Subscription, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) SyncWithProvider(ctx context.Context, sub Subscription) (Subscription, error) {
	stripeSubscription, err := s.stripeClient.Subscriptions.Get(sub.ProviderID, &stripe.SubscriptionParams{
		Params: stripe.Params{
			Context: ctx,
		},
	})
	if err != nil {
		return Subscription{}, fmt.Errorf("failed to get subscription from billing provider: %w", err)
	}

	sub.State = string(stripeSubscription.Status)
	if stripeSubscription.CanceledAt > 0 {
		t := time.Unix(stripeSubscription.CanceledAt, 0)
		sub.CanceledAt = &t
	}
	return s.repository.UpdateByID(ctx, sub)
}

func (s *Service) Cancel(ctx context.Context, sub Subscription) (Subscription, error) {
	stripeSubscription, err := s.stripeClient.Subscriptions.Cancel(sub.ProviderID, &stripe.SubscriptionCancelParams{
		Params: stripe.Params{
			Context: ctx,
		},
	})
	if err != nil {
		return Subscription{}, fmt.Errorf("failed to cancel subscription at billing provider: %w", err)
	}

	sub.State = string(stripeSubscription.Status)
	if stripeSubscription.CanceledAt > 0 {
		t := time.Unix(stripeSubscription.CanceledAt, 0)
		sub.CanceledAt = &t
	}
	return s.repository.UpdateByID(ctx, sub)
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Subscription, error) {
	return s.repository.List(ctx, filter)
}
