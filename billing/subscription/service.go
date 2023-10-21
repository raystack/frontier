package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/raystack/frontier/billing/feature"
	"github.com/raystack/frontier/billing/plan"

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

type PlanService interface {
	List(ctx context.Context, filter plan.Filter) ([]plan.Plan, error)
	GetByID(ctx context.Context, id string) (plan.Plan, error)
}

type Service struct {
	repository      Repository
	stripeClient    *client.API
	customerService CustomerService
	planService     PlanService
}

func NewService(stripeClient *client.API, repository Repository,
	customerService CustomerService, planService PlanService) *Service {
	return &Service{
		stripeClient:    stripeClient,
		repository:      repository,
		customerService: customerService,
		planService:     planService,
	}
}

func (s *Service) Create(ctx context.Context, sub Subscription) (Subscription, error) {
	// get billing
	billingCustomer, err := s.customerService.GetByID(ctx, sub.CustomerID)
	if err != nil {
		return Subscription{}, err
	}

	// create subscription items
	plan, err := s.planService.GetByID(ctx, sub.PlanID)
	if err != nil {
		return Subscription{}, err
	}
	sub.PlanID = plan.ID // set plan id to the actual plan uuid
	var subsItems []*stripe.CheckoutSessionLineItemParams
	for _, planFeature := range plan.Features {
		itemParams := &stripe.CheckoutSessionLineItemParams{
			Price: stripe.String(planFeature.Price.ProviderID),
		}
		if planFeature.Price.UsageType == feature.PriceUsageTypeLicensed {
			itemParams.Quantity = stripe.Int64(1)
		}
		subsItems = append(subsItems, itemParams)
	}

	// create subscription checkout link
	stripeSubscriptionCheckout, err := s.stripeClient.CheckoutSessions.New(&stripe.CheckoutSessionParams{
		Params: stripe.Params{
			Context: ctx,
		},
		//AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
		//	Enabled: stripe.Bool(true),
		//},
		CancelURL: &sub.CancelUrl,
		Currency:  &billingCustomer.Currency,
		Customer:  &billingCustomer.ProviderID,
		LineItems: subsItems,
		Metadata: map[string]string{
			"org_id":     billingCustomer.OrgID,
			"plan_id":    sub.PlanID,
			"managed_by": "frontier",
		},
		Mode: stripe.String("subscription"),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Description: stripe.String(fmt.Sprintf("Subscription for %s", plan.Name)),
			Metadata: map[string]string{
				"org_id":     billingCustomer.OrgID,
				"plan_id":    sub.PlanID,
				"plan_name":  plan.Name,
				"interval":   plan.Interval,
				"managed_by": "frontier",
			},
		},
		SuccessURL: &sub.SuccessUrl,
	})
	if err != nil {
		return Subscription{}, fmt.Errorf("failed to create subscription at billing provider: %w", err)
	}

	sub.State = "pending"
	sub.Metadata = map[string]any{
		"checkout_session_id": stripeSubscriptionCheckout.ID,
		"checkout_url":        stripeSubscriptionCheckout.URL,
		"payment_status":      stripeSubscriptionCheckout.PaymentStatus,
	}
	return s.repository.Create(ctx, sub)
}

func (s *Service) GetByID(ctx context.Context, id string) (Subscription, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) SyncWithProvider(ctx context.Context, customerID string) error {
	subs, err := s.repository.List(ctx, Filter{
		CustomerID: customerID,
	})
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if sub.ProviderID == "" {
			// not synced yet
			checkoutSession, err := s.stripeClient.CheckoutSessions.Get(sub.Metadata["checkout_session_id"].(string), &stripe.CheckoutSessionParams{
				Params: stripe.Params{
					Context: ctx,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to get checkout session from billing provider: %w", err)
			}
			if checkoutSession.PaymentStatus != stripe.CheckoutSessionPaymentStatusUnpaid {
				sub.Metadata["payment_status"] = checkoutSession.PaymentStatus
			}
			if checkoutSession.Subscription != nil {
				sub.ProviderID = checkoutSession.Subscription.ID
				sub.State = string(checkoutSession.Subscription.Status)
			}
		} else {
			stripeSubscription, err := s.stripeClient.Subscriptions.Get(sub.ProviderID, &stripe.SubscriptionParams{
				Params: stripe.Params{
					Context: ctx,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to get subscription from billing provider: %w", err)
			}
			sub.State = string(stripeSubscription.Status)
			if stripeSubscription.CanceledAt > 0 {
				t := time.Unix(stripeSubscription.CanceledAt, 0)
				sub.CanceledAt = &t
			}
		}
		if _, err := s.repository.UpdateByID(ctx, sub); err != nil {
			return err
		}
	}

	return nil
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
	if err := s.SyncWithProvider(ctx, filter.CustomerID); err != nil {
		return nil, err
	}
	return s.repository.List(ctx, filter)
}
