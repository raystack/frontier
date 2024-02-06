package checkout

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"text/template"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/raystack/frontier/billing/credit"

	"github.com/google/uuid"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/pkg/debounce"
	"go.uber.org/zap"

	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"

	"github.com/raystack/frontier/billing/customer"
	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/client"
)

const (
	SessionValidity = time.Hour * 24
	SyncDelay       = time.Second * 60

	AmountSubscriptionMetadataKey     = "amount_total"
	CurrencySubscriptionMetadataKey   = "currency"
	ProviderIDSubscriptionMetadataKey = "provider_subscription_id"
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

type ProductService interface {
	GetByID(ctx context.Context, id string) (product.Product, error)
}

type CreditService interface {
	Add(ctx context.Context, cred credit.Credit) error
}

type OrganizationService interface {
	MemberCount(ctx context.Context, orgID string) (int64, error)
}

type Service struct {
	stripeAutoTax       bool
	stripeClient        *client.API
	repository          Repository
	customerService     CustomerService
	planService         PlanService
	subscriptionService SubscriptionService
	creditService       CreditService
	productService      ProductService
	orgService          OrganizationService

	syncLimiter *debounce.Limiter
	syncJob     *cron.Cron
	mu          sync.Mutex
}

func NewService(stripeClient *client.API, stripeAutoTax bool, repository Repository,
	customerService CustomerService, planService PlanService,
	subscriptionService SubscriptionService, productService ProductService,
	creditService CreditService, orgService OrganizationService) *Service {
	s := &Service{
		stripeClient:        stripeClient,
		stripeAutoTax:       stripeAutoTax,
		repository:          repository,
		customerService:     customerService,
		planService:         planService,
		subscriptionService: subscriptionService,
		creditService:       creditService,
		productService:      productService,
		orgService:          orgService,
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

	checkoutID := uuid.New().String()
	ch, err = s.templatizeUrls(ch, checkoutID)
	if err != nil {
		return Checkout{}, err
	}

	// checkout could be for a plan or a product
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

		userCount, err := s.orgService.MemberCount(ctx, billingCustomer.OrgID)
		if err != nil {
			return Checkout{}, fmt.Errorf("failed to get member count: %w", err)
		}

		for _, planProduct := range plan.Products {
			// if it's credit, skip
			if planProduct.Behavior == product.CreditBehavior {
				continue
			}

			// if per seat, check if there is a limit of seats, if it breaches limit, fail
			if planProduct.Behavior == product.PerSeatBehavior {
				if planProduct.Config.SeatLimit > 0 && userCount > planProduct.Config.SeatLimit {
					return Checkout{}, fmt.Errorf("member count exceeds allowed limit of the plan: %w", product.ErrPerSeatLimitReached)
				}
			}

			for _, productPrice := range planProduct.Prices {
				// only work with plan interval prices
				if productPrice.Interval != plan.Interval {
					continue
				}

				var quantity int64 = 1
				if productPrice.UsageType == product.PriceUsageTypeLicensed {
					if planProduct.Behavior == product.PerSeatBehavior {
						quantity = userCount
					}
				}
				itemParams := &stripe.CheckoutSessionLineItemParams{
					Price:    stripe.String(productPrice.ProviderID),
					Quantity: stripe.Int64(quantity),
				}
				subsItems = append(subsItems, itemParams)
			}
		}

		var trialDays *int64 = nil
		if plan.TrialDays > 0 {
			trialDays = stripe.Int64(plan.TrialDays)
		}

		// create subscription checkout link
		stripeCheckout, err := s.stripeClient.CheckoutSessions.New(&stripe.CheckoutSessionParams{
			Params: stripe.Params{
				Context: ctx,
			},
			AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
				Enabled: stripe.Bool(s.stripeAutoTax),
			},
			Currency:  stripe.String(billingCustomer.Currency),
			Customer:  stripe.String(billingCustomer.ProviderID),
			LineItems: subsItems,
			Metadata: map[string]string{
				"org_id":      billingCustomer.OrgID,
				"plan_id":     ch.PlanID,
				"checkout_id": checkoutID,
				"managed_by":  "frontier",
			},
			Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
				Description: stripe.String(fmt.Sprintf("Checkout for %s", plan.Name)),
				Metadata: map[string]string{
					"org_id":      billingCustomer.OrgID,
					"checkout_id": checkoutID,
					"managed_by":  "frontier",
				},
				TrialPeriodDays: trialDays,
			},
			AllowPromotionCodes: stripe.Bool(true),
			CancelURL:           stripe.String(ch.CancelUrl),
			SuccessURL:          stripe.String(ch.SuccessUrl),
			ExpiresAt:           stripe.Int64(time.Now().Add(SessionValidity).Unix()),
		})
		if err != nil {
			return Checkout{}, fmt.Errorf("failed to create subscription at billing provider: %w", err)
		}

		return s.repository.Create(ctx, Checkout{
			ID:            checkoutID,
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

	if ch.ProductID != "" {
		chProduct, err := s.productService.GetByID(ctx, ch.ProductID)
		if err != nil {
			return Checkout{}, fmt.Errorf("failed to get product: %w", err)
		}
		if len(chProduct.Prices) == 0 {
			return Checkout{}, fmt.Errorf("invalid product, no prices found")
		}

		var subsItems []*stripe.CheckoutSessionLineItemParams
		for _, productPrice := range chProduct.Prices {
			itemParams := &stripe.CheckoutSessionLineItemParams{
				Price: stripe.String(productPrice.ProviderID),
			}
			if productPrice.UsageType == product.PriceUsageTypeLicensed {
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
			Currency:  stripe.String(billingCustomer.Currency),
			Customer:  stripe.String(billingCustomer.ProviderID),
			LineItems: subsItems,
			Mode:      stripe.String(string(stripe.CheckoutSessionModePayment)),
			Metadata: map[string]string{
				"org_id":        billingCustomer.OrgID,
				"product_name":  chProduct.Name,
				"credit_amount": fmt.Sprintf("%d", chProduct.Config.CreditAmount),
				"checkout_id":   checkoutID,
				"managed_by":    "frontier",
			},
			CancelURL:  stripe.String(ch.CancelUrl),
			SuccessURL: stripe.String(ch.SuccessUrl),
			ExpiresAt:  stripe.Int64(time.Now().Add(SessionValidity).Unix()),
		})
		if err != nil {
			return Checkout{}, fmt.Errorf("failed to create subscription at billing provider: %w", err)
		}

		return s.repository.Create(ctx, Checkout{
			ID:            checkoutID,
			ProviderID:    stripeCheckout.ID,
			CustomerID:    billingCustomer.ID,
			ProductID:     chProduct.ID,
			CancelUrl:     ch.CancelUrl,
			SuccessUrl:    ch.SuccessUrl,
			CheckoutUrl:   stripeCheckout.URL,
			State:         string(stripeCheckout.Status),
			PaymentStatus: string(stripeCheckout.PaymentStatus),
			Metadata: map[string]any{
				"product_name": chProduct.Name,
			},
			ExpireAt: time.Unix(stripeCheckout.ExpiresAt, 0),
		})
	}

	return Checkout{}, fmt.Errorf("invalid checkout request")
}

// templatizeUrls replaces the checkout id in the urls with the actual checkout id
func (s *Service) templatizeUrls(ch Checkout, checkoutID string) (Checkout, error) {
	tpl := template.New("success")
	t, err := tpl.Parse(ch.SuccessUrl)
	if err != nil {
		return Checkout{}, fmt.Errorf("failed to parse success url: %w", err)
	}
	var tplBuffer bytes.Buffer
	if err = t.Execute(&tplBuffer, map[string]string{
		"CheckoutID": checkoutID,
	}); err != nil {
		return Checkout{}, fmt.Errorf("failed to parse success url: %w", err)
	}
	ch.SuccessUrl = tplBuffer.String()

	tpl = template.New("cancel")
	t, err = tpl.Parse(ch.CancelUrl)
	if err != nil {
		return Checkout{}, fmt.Errorf("failed to parse cancel url: %w", err)
	}
	tplBuffer.Reset()
	if err = t.Execute(&tplBuffer, map[string]string{
		"CheckoutID": checkoutID,
	}); err != nil {
		return Checkout{}, fmt.Errorf("failed to parse cancel url: %w", err)
	}
	ch.CancelUrl = tplBuffer.String()
	return ch, nil
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
			ch.Metadata[ProviderIDSubscriptionMetadataKey] = checkoutSession.Subscription.ID
			ch.Metadata[AmountSubscriptionMetadataKey] = checkoutSession.AmountTotal
			ch.Metadata[CurrencySubscriptionMetadataKey] = checkoutSession.Currency
		}
		if checks[idx], err = s.repository.UpdateByID(ctx, ch); err != nil {
			return fmt.Errorf("failed to update checkout session: %w", err)
		}
	}

	// if payment is completed, create subscription for them in system
	for _, ch := range checks {
		if ch.State == StateComplete.String() &&
			(ch.PaymentStatus == "paid" || ch.PaymentStatus == "no_payment_required") {
			// checkout could be for a plan or a product

			if ch.PlanID != "" {
				// if the checkout was created for subscription
				if _, err := s.ensureSubscription(ctx, ch); err != nil {
					return err
				}

				// subscription can also be complimented with free credits
				if err := s.ensureCreditsForPlan(ctx, ch); err != nil {
					return fmt.Errorf("ensureCreditsForPlan: %w", err)
				}
			} else if ch.ProductID != "" {
				// if the checkout was created for product
				if err := s.ensureCreditsForProduct(ctx, ch); err != nil {
					return fmt.Errorf("ensureCreditsForProduct: %w", err)
				}
			}
		}
	}

	return nil
}

func (s *Service) ensureCreditsForProduct(ctx context.Context, ch Checkout) error {
	chProduct, err := s.productService.GetByID(ctx, ch.ProductID)
	if err != nil {
		return err
	}
	description := fmt.Sprintf("addition of %d credits for %s", chProduct.Config.CreditAmount, chProduct.Title)
	if price, pok := ch.Metadata[AmountSubscriptionMetadataKey].(int64); pok {
		if currency, cok := ch.Metadata[CurrencySubscriptionMetadataKey].(string); cok {
			description = fmt.Sprintf("addition of %d credits for %s at %d[%s]", chProduct.Config.CreditAmount, chProduct.Title, price, currency)
		}
	}
	if err := s.creditService.Add(ctx, credit.Credit{
		ID:          ch.ID,
		AccountID:   ch.CustomerID,
		Amount:      chProduct.Config.CreditAmount,
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

	if chPlan.OnStartCredits == 0 {
		// no such product
		return nil
	}

	description := fmt.Sprintf("addition of %d credits for %s", chPlan.OnStartCredits, chPlan.Title)
	if err := s.creditService.Add(ctx, credit.Credit{
		ID:          ch.ID,
		AccountID:   ch.CustomerID,
		Amount:      chPlan.OnStartCredits,
		Metadata:    ch.Metadata,
		Description: description,
	}); err != nil && !errors.Is(err, credit.ErrAlreadyApplied) {
		return err
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
	if ch.Metadata[ProviderIDSubscriptionMetadataKey] == nil {
		return "", fmt.Errorf("invalid checkout session, provider_subscription_id is missing")
	}

	// check if already created in frontier
	_, err := s.subscriptionService.GetByProviderID(ctx, ch.Metadata[ProviderIDSubscriptionMetadataKey].(string))
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
		ProviderID: ch.Metadata[ProviderIDSubscriptionMetadataKey].(string),
		Metadata: map[string]any{
			"checkout_id": ch.ID,
		},
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

func (s *Service) CreateSessionForPaymentMethod(ctx context.Context, ch Checkout) (Checkout, error) {
	// get billing
	billingCustomer, err := s.customerService.GetByID(ctx, ch.CustomerID)
	if err != nil {
		return Checkout{}, err
	}

	checkoutID := uuid.New().String()
	ch, err = s.templatizeUrls(ch, checkoutID)
	if err != nil {
		return Checkout{}, err
	}

	// create payment method setup checkout link
	stripeCheckout, err := s.stripeClient.CheckoutSessions.New(&stripe.CheckoutSessionParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Customer:   stripe.String(billingCustomer.ProviderID),
		Currency:   stripe.String(billingCustomer.Currency),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSetup)),
		CancelURL:  stripe.String(ch.CancelUrl),
		SuccessURL: stripe.String(ch.SuccessUrl),
		ExpiresAt:  stripe.Int64(time.Now().Add(SessionValidity).Unix()),
		Metadata: map[string]string{
			"org_id":      billingCustomer.OrgID,
			"checkout_id": checkoutID,
			"managed_by":  "frontier",
		},
	})
	if err != nil {
		return Checkout{}, fmt.Errorf("failed to create checkout at billing provider: %w", err)
	}

	return s.repository.Create(ctx, Checkout{
		ID:          checkoutID,
		ProviderID:  stripeCheckout.ID,
		CustomerID:  billingCustomer.ID,
		CancelUrl:   ch.CancelUrl,
		SuccessUrl:  ch.SuccessUrl,
		CheckoutUrl: stripeCheckout.URL,
		State:       string(stripeCheckout.Status),
		ExpireAt:    time.Unix(stripeCheckout.ExpiresAt, 0),
		Metadata: map[string]any{
			"mode": "setup",
		},
	})
}

// Apply applies the actual request directly without creating a checkout session
// for example when a request is created for a plan, it will be directly subscribe without
// actually paying for it
func (s *Service) Apply(ctx context.Context, ch Checkout) (*subscription.Subscription, *product.Product, error) {
	// get billing
	billingCustomer, err := s.customerService.GetByID(ctx, ch.CustomerID)
	if err != nil {
		return nil, nil, err
	}

	// checkout could be for a plan or a product
	if ch.PlanID != "" {
		// if already subscribed to the plan, return
		if subID, err := s.checkIfAlreadySubscribed(ctx, ch); err != nil {
			return nil, nil, err
		} else if subID != "" {
			return nil, nil, fmt.Errorf("already subscribed to the plan")
		}

		// create subscription items
		plan, err := s.planService.GetByID(ctx, ch.PlanID)
		if err != nil {
			return nil, nil, err
		}
		var subsItems []*stripe.SubscriptionItemsParams
		userCount, err := s.orgService.MemberCount(ctx, billingCustomer.OrgID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get member count: %w", err)
		}

		for _, planProduct := range plan.Products {
			// if it's credit, skip, they are handled separately
			if planProduct.Behavior == product.CreditBehavior {
				continue
			}
			// if per seat, check if there is a limit of seats, if it breaches limit, fail
			if planProduct.Behavior == product.PerSeatBehavior {
				if planProduct.Config.SeatLimit > 0 && userCount > planProduct.Config.SeatLimit {
					return nil, nil, fmt.Errorf("member count exceeds allowed limit of the plan: %w", product.ErrPerSeatLimitReached)
				}
			}

			for _, productPrice := range planProduct.Prices {
				// only work with plan interval prices
				if productPrice.Interval != plan.Interval {
					continue
				}

				var quantity int64 = 1
				if productPrice.UsageType == product.PriceUsageTypeLicensed {
					if planProduct.Behavior == product.PerSeatBehavior {
						quantity = userCount
					}
				}

				itemParams := &stripe.SubscriptionItemsParams{
					Price:    stripe.String(productPrice.ProviderID),
					Quantity: stripe.Int64(quantity),
					Metadata: map[string]string{
						"org_id":     billingCustomer.OrgID,
						"feature_id": planProduct.ID,
					},
				}
				subsItems = append(subsItems, itemParams)
			}
		}

		var trialDays *int64 = nil
		if plan.TrialDays > 0 {
			trialDays = stripe.Int64(plan.TrialDays)
		}

		// create subscription directly
		stripeSubscription, err := s.stripeClient.Subscriptions.New(&stripe.SubscriptionParams{
			Params: stripe.Params{
				Context: ctx,
			},
			Customer: stripe.String(billingCustomer.ProviderID),
			Currency: stripe.String(billingCustomer.Currency),
			Items:    subsItems,
			Metadata: map[string]string{
				"org_id":     billingCustomer.OrgID,
				"managed_by": "frontier",
			},
			TrialPeriodDays: trialDays,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create subscription at billing provider: %w", err)
		}

		// register subscription in frontier
		subs, err := s.subscriptionService.Create(ctx, subscription.Subscription{
			ID:         uuid.New().String(),
			ProviderID: stripeSubscription.ID,
			CustomerID: billingCustomer.ID,
			PlanID:     plan.ID,
			Metadata: map[string]any{
				"org_id":      billingCustomer.OrgID,
				"delegated":   "true",
				"checkout_id": ch.ID,
			},
			TrialEndsAt: time.Unix(stripeSubscription.TrialEnd, 0),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create subscription: %w", err)
		}
		ch.ID = subs.ID

		// subscription can also be complimented with free credits
		if err := s.ensureCreditsForPlan(ctx, ch); err != nil {
			return nil, nil, fmt.Errorf("ensureCreditsForPlan: %w", err)
		}
		return &subs, nil, nil
	} else if ch.ProductID != "" {
		// TODO(kushsharma): not implemented yet
		return nil, nil, fmt.Errorf("not supported yet")
	}

	return nil, nil, fmt.Errorf("invalid checkout request")
}
