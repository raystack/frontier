package checkout

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/billing/credit"

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

const (
	SessionValidity = time.Hour * 24
	SyncDelay       = time.Second * 60
)

type Repository interface {
	GetByID(ctx context.Context, id string) (Checkout, error)
	Create(ctx context.Context, ch Checkout) (Checkout, error)
	UpdateByID(ctx context.Context, ch Checkout) (Checkout, error)
	List(ctx context.Context, filter Filter) ([]Checkout, error)
}

type CustomerService interface {
	GetByID(ctx context.Context, id string) (customer.Customer, error)
	List(ctx context.Context, filter customer.Filter) ([]customer.Customer, error)
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

type FeatureService interface {
	GetByID(ctx context.Context, id string) (feature.Feature, error)
}

type CreditService interface {
	Add(ctx context.Context, cred credit.Credit) error
}

type Service struct {
	stripeAutoTax       bool
	stripeClient        *client.API
	repository          Repository
	customerService     CustomerService
	planService         PlanService
	subscriptionService SubscriptionService
	creditService       CreditService
	featureService      FeatureService

	syncLimiter *debounce.Limiter
	syncJob     *cron.Cron
	mu          sync.Mutex
}

func NewService(stripeClient *client.API, stripeAutoTax bool, repository Repository,
	customerService CustomerService, planService PlanService,
	subscriptionService SubscriptionService, featureService FeatureService,
	creditService CreditService) *Service {
	s := &Service{
		stripeClient:        stripeClient,
		stripeAutoTax:       stripeAutoTax,
		repository:          repository,
		customerService:     customerService,
		planService:         planService,
		subscriptionService: subscriptionService,
		creditService:       creditService,
		featureService:      featureService,
		syncLimiter:         debounce.New(2 * time.Second),
	}
	return s
}

func (s *Service) Init(ctx context.Context) {
	if s.syncJob != nil {
		s.syncJob.Stop()
	}

	s.syncJob = cron.New()
	s.syncJob.AddFunc(fmt.Sprintf("@every %s", SyncDelay.String()), func() {
		s.backgroundSync(ctx)
	})
	s.syncJob.Start()
}

func (s *Service) Close() error {
	if s.syncJob != nil {
		return s.syncJob.Stop().Err()
	}
	return nil
}

func (s *Service) backgroundSync(ctx context.Context) {
	logger := grpczap.Extract(ctx)
	customers, err := s.customerService.List(ctx, customer.Filter{})
	if err != nil {
		logger.Error("checkout.backgroundSync", zap.Error(err))
		return
	}

	for _, customer := range customers {
		if customer.DeletedAt != nil || customer.ProviderID == "" {
			continue
		}
		if err := s.SyncWithProvider(ctx, customer.ID); err != nil {
			logger.Error("checkout.SyncWithProvider", zap.Error(err))
		}
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
}

func (s *Service) Create(ctx context.Context, ch Checkout) (Checkout, error) {
	// get billing
	billingCustomer, err := s.customerService.GetByID(ctx, ch.CustomerID)
	if err != nil {
		return Checkout{}, err
	}

	// checkout could be for a plan or a feature
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
			// if it's a credit feature, skip, it is added as complimentary
			if planFeature.CreditAmount > 0 {
				continue
			}
			for _, featurePrice := range planFeature.Prices {
				// only work with plan interval prices
				if featurePrice.Interval != plan.Interval {
					continue
				}

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
				Enabled: stripe.Bool(s.stripeAutoTax),
			},
			Currency:  &billingCustomer.Currency,
			Customer:  &billingCustomer.ProviderID,
			LineItems: subsItems,
			Metadata: map[string]string{
				"org_id":     billingCustomer.OrgID,
				"plan_id":    ch.PlanID,
				"managed_by": "frontier",
			},
			Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
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
			ExpiresAt:  stripe.Int64(time.Now().Add(SessionValidity).Unix()),
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
	} else if ch.FeatureID != "" {
		chFeature, err := s.featureService.GetByID(ctx, ch.FeatureID)
		if err != nil {
			return Checkout{}, fmt.Errorf("failed to get feature: %w", err)
		}
		if len(chFeature.Prices) == 0 {
			return Checkout{}, fmt.Errorf("invalid feature, no prices found")
		}

		var subsItems []*stripe.CheckoutSessionLineItemParams
		for _, featurePrice := range chFeature.Prices {
			itemParams := &stripe.CheckoutSessionLineItemParams{
				Price: stripe.String(featurePrice.ProviderID),
			}
			if featurePrice.UsageType == feature.PriceUsageTypeLicensed {
				itemParams.Quantity = stripe.Int64(1)
			}
			subsItems = append(subsItems, itemParams)
		}

		// create one time checkout link
		stripeCheckout, err := s.stripeClient.CheckoutSessions.New(&stripe.CheckoutSessionParams{
			Params: stripe.Params{
				Context: ctx,
			},
			AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
				Enabled: stripe.Bool(s.stripeAutoTax),
			},
			Currency:  &billingCustomer.Currency,
			Customer:  &billingCustomer.ProviderID,
			LineItems: subsItems,
			Mode:      stripe.String(string(stripe.CheckoutSessionModePayment)),
			Metadata: map[string]string{
				"org_id":        billingCustomer.OrgID,
				"feature_name":  chFeature.Name,
				"credit_amount": fmt.Sprintf("%d", chFeature.CreditAmount),
				"managed_by":    "frontier",
			},
			CancelURL:  &ch.CancelUrl,
			SuccessURL: &ch.SuccessUrl,
			ExpiresAt:  stripe.Int64(time.Now().Add(SessionValidity).Unix()),
		})
		if err != nil {
			return Checkout{}, fmt.Errorf("failed to create subscription at billing provider: %w", err)
		}

		return s.repository.Create(ctx, Checkout{
			ID:            uuid.New().String(),
			ProviderID:    stripeCheckout.ID,
			CustomerID:    billingCustomer.ID,
			FeatureID:     chFeature.ID,
			CancelUrl:     ch.CancelUrl,
			SuccessUrl:    ch.SuccessUrl,
			CheckoutUrl:   stripeCheckout.URL,
			State:         string(stripeCheckout.Status),
			PaymentStatus: string(stripeCheckout.PaymentStatus),
			Metadata: map[string]any{
				"feature_name": chFeature.Name,
			},
			ExpireAt: time.Unix(stripeCheckout.ExpiresAt, 0),
		})
	}

	return Checkout{}, fmt.Errorf("invalid checkout request")
}

func (s *Service) GetByID(ctx context.Context, id string) (Checkout, error) {
	return s.repository.GetByID(ctx, id)
}

// SyncWithProvider syncs the subscription state with the billing provider
func (s *Service) SyncWithProvider(ctx context.Context, customerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	checks, err := s.repository.List(ctx, Filter{
		CustomerID: customerID,
	})
	if err != nil {
		return err
	}

	// find all checkout sessions of the customer that require a sync
	// and update their state in system
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
			ch.Metadata["amount_total"] = checkoutSession.AmountTotal
			ch.Metadata["currency"] = checkoutSession.Currency
		}
		if checks[idx], err = s.repository.UpdateByID(ctx, ch); err != nil {
			return fmt.Errorf("failed to update checkout session: %w", err)
		}
	}

	// if payment is completed, create subscription for them in system
	for _, ch := range checks {
		if ch.State == StateComplete.String() &&
			(ch.PaymentStatus == "paid" || ch.PaymentStatus == "no_payment_required") {
			// checkout could be for a plan or a feature

			if ch.PlanID != "" {
				// if the checkout was created for subscription
				if _, err := s.ensureSubscription(ctx, ch); err != nil {
					return err
				}

				// subscription can also be complimented with free credits
				if err := s.ensureCreditsForPlan(ctx, ch); err != nil {
					return fmt.Errorf("ensureCreditsForPlan: %w", err)
				}
			} else if ch.FeatureID != "" {
				// if the checkout was created for feature
				if err := s.ensureCreditsForFeature(ctx, ch); err != nil {
					return fmt.Errorf("ensureCreditsForFeature: %w", err)
				}
			}
		}
	}

	return nil
}

func (s *Service) ensureCreditsForFeature(ctx context.Context, ch Checkout) error {
	chFeature, err := s.featureService.GetByID(ctx, ch.FeatureID)
	if err != nil {
		return err
	}
	description := fmt.Sprintf("addition of %d credits for %s", chFeature.CreditAmount, chFeature.Title)
	if price, pok := ch.Metadata["amount_total"].(int64); pok {
		if currency, cok := ch.Metadata["currency"].(string); cok {
			description = fmt.Sprintf("addition of %d credits for %s at %d[%s]", chFeature.CreditAmount, chFeature.Title, price, currency)
		}
	}
	if err := s.creditService.Add(ctx, credit.Credit{
		ID:          ch.ID,
		AccountID:   ch.CustomerID,
		Amount:      chFeature.CreditAmount,
		Metadata:    ch.Metadata,
		Description: description,
	}); err != nil && !errors.Is(err, credit.ErrAlreadyApplied) {
		return err
	}
	return nil
}

func (s *Service) ensureCreditsForPlan(ctx context.Context, ch Checkout) error {
	chPlan, err := s.planService.GetByID(ctx, ch.PlanID)
	if err != nil {
		return err
	}

	// find feature with credits
	creditFeatures := utils.Filter(chPlan.Features, func(f feature.Feature) bool {
		return f.CreditAmount > 0
	})
	if len(creditFeatures) == 0 {
		// no such feature
		return nil
	}

	for _, chFeature := range creditFeatures {
		description := fmt.Sprintf("addition of %d credits for %s", chFeature.CreditAmount, chFeature.Title)
		if err := s.creditService.Add(ctx, credit.Credit{
			ID:          ch.ID,
			AccountID:   ch.CustomerID,
			Amount:      chFeature.CreditAmount,
			Metadata:    ch.Metadata,
			Description: description,
		}); err != nil && !errors.Is(err, credit.ErrAlreadyApplied) {
			return err
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
		if err := s.SyncWithProvider(context.Background(), filter.CustomerID); err != nil {
			logger.Error("checkout.SyncWithProvider", zap.Error(err))
		}
	})

	return s.repository.List(ctx, filter)
}
