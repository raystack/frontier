package checkout

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/pkg/debounce"
	"go.uber.org/zap"

	"github.com/raystack/frontier/billing/feature"
	"github.com/raystack/frontier/billing/plan"

	"github.com/raystack/frontier/billing/customer"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/client"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Checkout, error)
	Create(ctx context.Context, ch Checkout) (Checkout, error)
	UpdateByID(ctx context.Context, ch Checkout) (Checkout, error)
	List(ctx context.Context, filter Filter) ([]Checkout, error)
}

type CustomerService interface {
	GetByID(ctx context.Context, id string) (customer.Customer, error)
}

type PlanService interface {
	List(ctx context.Context, filter plan.Filter) ([]plan.Plan, error)
	GetByID(ctx context.Context, id string) (plan.Plan, error)
}

type SubscriptionService interface {
	List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error)
	Create(ctx context.Context, sub subscription.Subscription) (subscription.Subscription, error)
	GetByProviderID(ctx context.Context, id string) (subscription.Subscription, error)
}

type Service struct {
	repository          Repository
	stripeClient        *client.API
	customerService     CustomerService
	planService         PlanService
	subscriptionService SubscriptionService

	stripeAutomaticTaxEnabled bool
	syncLimiter               *debounce.Limiter
}

func NewService(stripeClient *client.API, repository Repository,
	customerService CustomerService, planService PlanService, subscriptionService SubscriptionService) *Service {
	return &Service{
		stripeClient:        stripeClient,
		repository:          repository,
		customerService:     customerService,
		planService:         planService,
		subscriptionService: subscriptionService,
		syncLimiter:         debounce.New(5 * time.Second),
	}
}

func (s *Service) Create(ctx context.Context, ch Checkout) (Checkout, error) {
	// get billing
	billingCustomer, err := s.customerService.GetByID(ctx, ch.CustomerID)
	if err != nil {
		return Checkout{}, err
	}

	if ch.PlanID != "" {
		// if already subscribed to the plan, return
		if subID, err := s.checkIfAlreadySubscribed(ctx, ch); err != nil {
			return Checkout{}, err
		} else if subID != "" {
			return Checkout{}, fmt.Errorf("already subscribed to the plan")
		}

		// create subscription items
		plan, err := s.planService.GetByID(ctx, ch.PlanID)
		if err != nil {
			return Checkout{}, err
		}
		var subsItems []*stripe.CheckoutSessionLineItemParams
		for _, planFeature := range plan.Features {
			for _, featurePrice := range planFeature.Prices {
				itemParams := &stripe.CheckoutSessionLineItemParams{
					Price: stripe.String(featurePrice.ProviderID),
				}
				if featurePrice.UsageType == feature.PriceUsageTypeLicensed {
					itemParams.Quantity = stripe.Int64(1)
				}
				subsItems = append(subsItems, itemParams)
			}
		}

		// create subscription checkout link
		stripeCheckout, err := s.stripeClient.CheckoutSessions.New(&stripe.CheckoutSessionParams{
			Params: stripe.Params{
				Context: ctx,
			},
			AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
				Enabled: stripe.Bool(s.stripeAutomaticTaxEnabled),
			},
			Currency:  &billingCustomer.Currency,
			Customer:  &billingCustomer.ProviderID,
			LineItems: subsItems,
			Metadata: map[string]string{
				"org_id":     billingCustomer.OrgID,
				"plan_id":    ch.PlanID,
				"managed_by": "frontier",
			},
			Mode: stripe.String("subscription"),
			SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
				Description: stripe.String(fmt.Sprintf("Checkout for %s", plan.Name)),
				Metadata: map[string]string{
					"org_id":     billingCustomer.OrgID,
					"plan_id":    ch.PlanID,
					"plan_name":  plan.Name,
					"interval":   plan.Interval,
					"managed_by": "frontier",
				},
			},
			CancelURL:  &ch.CancelUrl,
			SuccessURL: &ch.SuccessUrl,
		})
		if err != nil {
			return Checkout{}, fmt.Errorf("failed to create subscription at billing provider: %w", err)
		}

		return s.repository.Create(ctx, Checkout{
			ID:            uuid.New().String(),
			ProviderID:    stripeCheckout.ID,
			CustomerID:    billingCustomer.ID,
			PlanID:        plan.ID,
			CancelUrl:     ch.CancelUrl,
			SuccessUrl:    ch.SuccessUrl,
			CheckoutUrl:   stripeCheckout.URL,
			State:         string(stripeCheckout.Status),
			PaymentStatus: string(stripeCheckout.PaymentStatus),
			Metadata: map[string]any{
				"plan_name": plan.Name,
			},
			ExpireAt: time.Unix(stripeCheckout.ExpiresAt, 0),
		})
	}

	return Checkout{}, fmt.Errorf("invalid checkout request")
}

func (s *Service) GetByID(ctx context.Context, id string) (Checkout, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) SyncWithProvider(ctx context.Context, customerID string) error {
	checks, err := s.repository.List(ctx, Filter{
		CustomerID: customerID,
	})
	if err != nil {
		return err
	}

	for idx, ch := range checks {
		if ch.State == StateExpired.String() || ch.State == StateComplete.String() {
			continue
		}
		if ch.ExpireAt.Before(time.Now()) {
			ch.State = "expired"
			if _, err := s.repository.UpdateByID(ctx, ch); err != nil {
				return err
			}
			continue
		}

		checkoutSession, err := s.stripeClient.CheckoutSessions.Get(ch.ProviderID, &stripe.CheckoutSessionParams{
			Params: stripe.Params{
				Context: ctx,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to get checkout session from billing provider: %w", err)
		}
		if ch.PaymentStatus != string(checkoutSession.PaymentStatus) {
			ch.PaymentStatus = string(checkoutSession.PaymentStatus)
		}
		if ch.State != string(checkoutSession.Status) {
			ch.State = string(checkoutSession.Status)
		}
		if checkoutSession.Subscription != nil {
			ch.Metadata["provider_subscription_id"] = checkoutSession.Subscription.ID
		}
		if checks[idx], err = s.repository.UpdateByID(ctx, ch); err != nil {
			return fmt.Errorf("failed to update checkout session: %w", err)
		}
	}

	// if payment is completed, create subscription
	for _, ch := range checks {
		if ch.State == StateComplete.String() && (ch.PaymentStatus == "paid" || ch.PaymentStatus == "no_payment_required") {
			// if the checkout was created for subscription
			if ch.PlanID != "" {
				if _, err := s.ensureSubscription(ctx, ch); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Service) checkIfAlreadySubscribed(ctx context.Context, ch Checkout) (string, error) {
	// check if subscription exists
	subs, err := s.subscriptionService.List(ctx, subscription.Filter{
		CustomerID: ch.CustomerID,
	})
	if err != nil {
		return "", err
	}

	for _, sub := range subs {
		if sub.PlanID == ch.PlanID {
			// subscription already exists
			if sub.State == "canceled" || sub.State == "ended" {
				continue
			}
			return sub.ID, nil
		}
	}

	return "", nil
}

func (s *Service) ensureSubscription(ctx context.Context, ch Checkout) (string, error) {
	if ch.Metadata["provider_subscription_id"] == nil {
		return "", fmt.Errorf("invalid checkout session, provider_subscription_id is missing")
	}

	// check if already created in frontier
	_, err := s.subscriptionService.GetByProviderID(ctx, ch.Metadata["provider_subscription_id"].(string))
	if err != nil && !errors.Is(err, subscription.ErrNotFound) {
		return "", err
	}
	if err == nil {
		// already created
		return "", nil
	}

	// create subscription
	sub, err := s.subscriptionService.Create(ctx, subscription.Subscription{
		ID:         uuid.New().String(),
		CustomerID: ch.CustomerID,
		PlanID:     ch.PlanID,
		ProviderID: ch.Metadata["provider_subscription_id"].(string),
	})
	if err != nil {
		return "", err
	}

	return sub.ID, nil
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Checkout, error) {
	logger := grpczap.Extract(ctx)
	s.syncLimiter.Call(func() {
		// fix context as the List ctx will get cancelled after call finishes
		if err := s.SyncWithProvider(context.Background(), filter.CustomerID); err != nil {
			logger.Error("checkout.SyncWithProvider", zap.Error(err))
		}
	})

	return s.repository.List(ctx, filter)
}
