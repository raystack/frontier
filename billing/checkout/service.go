package checkout

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stripe/stripe-go/v79"

	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/internal/metrics"

	"github.com/raystack/frontier/pkg/metadata"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/spf13/cast"

	"github.com/raystack/frontier/core/authenticate"

	"github.com/raystack/frontier/billing/credit"

	"github.com/google/uuid"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/billing/subscription"
	"go.uber.org/zap"

	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"

	"github.com/raystack/frontier/billing/customer"
	"github.com/stripe/stripe-go/v79/client"
)

const (
	SessionValidity = time.Hour * 24

	MinimumProductQuantity = 1
	MaximumProductQuantity = 100000 // max: 999999

	// ProductQuantityMetadataKey is the metadata key for the quantity of the product
	// it's necessary to cast as this properly because while storing metadata, it's serialized as json
	// and when retrieved, it's always an interface{} of float64 type
	ProductQuantityMetadataKey = "product_quantity"
	// AmountTotalMetadataKey is the metadata key for the total amount of the checkout
	// same goes for this as well, it's always an interface{} of float64 type
	AmountTotalMetadataKey = "amount_total"
	// ProcessedMetadataKey is the metadata key to indicate that the checkout has been processed
	// in the system
	ProcessedMetadataKey = "processed"

	CurrencyMetadataKey               = "currency"
	ProviderIDSubscriptionMetadataKey = "provider_subscription_id"
	InitiatorIDMetadataKey            = "initiated_by"
	CheckoutIDMetadataKey             = "checkout_id"
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
	RegisterToProviderIfRequired(ctx context.Context, customerID string) (customer.Customer, error)
}

type PlanService interface {
	List(ctx context.Context, filter plan.Filter) ([]plan.Plan, error)
	GetByID(ctx context.Context, id string) (plan.Plan, error)
}

type SubscriptionService interface {
	List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error)
	Create(ctx context.Context, sub subscription.Subscription) (subscription.Subscription, error)
	GetByProviderID(ctx context.Context, id string) (subscription.Subscription, error)
	Cancel(ctx context.Context, id string, immediate bool) (subscription.Subscription, error)
	HasUserSubscribedBefore(ctx context.Context, customerID string, planID string) (bool, error)
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

type AuthnService interface {
	GetPrincipal(ctx context.Context, assertions ...authenticate.ClientAssertion) (authenticate.Principal, error)
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
	authnService        AuthnService

	syncJob   *cron.Cron
	mu        sync.Mutex
	syncDelay time.Duration
}

func NewService(stripeClient *client.API, cfg billing.Config, repository Repository,
	customerService CustomerService, planService PlanService,
	subscriptionService SubscriptionService, productService ProductService,
	creditService CreditService, orgService OrganizationService,
	authnService AuthnService) *Service {
	s := &Service{
		stripeClient:        stripeClient,
		stripeAutoTax:       cfg.StripeAutoTax,
		repository:          repository,
		customerService:     customerService,
		planService:         planService,
		subscriptionService: subscriptionService,
		creditService:       creditService,
		productService:      productService,
		orgService:          orgService,
		authnService:        authnService,
		syncDelay:           cfg.RefreshInterval.Checkout,
	}
	return s
}

func (s *Service) Init(ctx context.Context) error {
	if s.syncJob != nil {
		<-s.syncJob.Stop().Done()
	}

	s.syncJob = cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))
	_, err := s.syncJob.AddFunc(fmt.Sprintf("@every %s", s.syncDelay.String()), func() {
		s.backgroundSync(ctx)
	})
	if err != nil {
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
	if metrics.BillingSyncLatency != nil {
		record := metrics.BillingSyncLatency("checkout")
		defer record()
	}

	logger := grpczap.Extract(ctx)
	customers, err := s.customerService.List(ctx, customer.Filter{
		State: customer.ActiveState,
	})
	if err != nil {
		logger.Error("checkout.backgroundSync", zap.Error(err))
		return
	}

	for _, customer := range customers {
		if ctx.Err() != nil {
			// stop processing if context is done
			break
		}

		if !customer.IsActive() || customer.IsOffline() {
			continue
		}
		if err := s.SyncWithProvider(ctx, customer.ID); err != nil {
			logger.Error("checkout.SyncWithProvider", zap.Error(err), zap.String("customer_id", customer.ID))
		}
	}
}

func (s *Service) Create(ctx context.Context, ch Checkout) (Checkout, error) {
	// need to make it register itself to provider first if needed
	billingCustomer, err := s.customerService.RegisterToProviderIfRequired(ctx, ch.CustomerID)
	if err != nil {
		return Checkout{}, err
	}

	checkoutID := uuid.New().String()
	ch, err = s.templatizeUrls(ch, checkoutID)
	if err != nil {
		return Checkout{}, err
	}

	currentPrincipal, err := s.authnService.GetPrincipal(ctx)
	if err != nil {
		return Checkout{}, err
	}

	// checkout could be for a plan or a product
	if ch.PlanID != "" {
		plan, err := s.planService.GetByID(ctx, ch.PlanID)
		if err != nil {
			return Checkout{}, err
		}
		// ensure we use uuid
		ch.PlanID = plan.ID

		// if already subscribed to the plan, return
		if subID, err := s.checkIfAlreadySubscribed(ctx, ch); err != nil {
			return Checkout{}, err
		} else if subID != "" {
			return Checkout{}, fmt.Errorf("already subscribed to the plan")
		}

		// create subscription items
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
			if planProduct.IsSeatLimitBreached(userCount) {
				return Checkout{}, fmt.Errorf("member count exceeds allowed limit of the plan: %w", product.ErrPerSeatLimitReached)
			}

			for _, productPrice := range planProduct.Prices {
				// only work with plan interval prices
				if productPrice.Interval != plan.Interval {
					continue
				}

				var quantity int64 = 1
				if productPrice.IsLicensed() && planProduct.HasPerSeatBehavior() {
					quantity = userCount
				}
				itemParams := &stripe.CheckoutSessionLineItemParams{
					Price:    stripe.String(productPrice.ProviderID),
					Quantity: stripe.Int64(quantity),
				}
				subsItems = append(subsItems, itemParams)
			}
		}

		var trialDays *int64 = nil
		// if trial is enabled and user has not trialed before, set trial days
		userHasTrialedBefore, err := s.subscriptionService.HasUserSubscribedBefore(ctx, billingCustomer.ID, plan.ID)
		if err != nil {
			return Checkout{}, err
		}
		if plan.TrialDays > 0 && !ch.SkipTrial && !userHasTrialedBefore {
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
				"org_id":               billingCustomer.OrgID,
				"plan_id":              ch.PlanID,
				CheckoutIDMetadataKey:  checkoutID,
				InitiatorIDMetadataKey: currentPrincipal.ID,
				"managed_by":           "frontier",
			},
			CustomerUpdate: &stripe.CheckoutSessionCustomerUpdateParams{
				Address: stripe.String(string(stripe.CheckoutSessionBillingAddressCollectionAuto)),
			},
			Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
			SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
				Description: stripe.String(fmt.Sprintf("Checkout for %s", plan.Name)),
				Metadata: map[string]string{
					"org_id":                            billingCustomer.OrgID,
					CheckoutIDMetadataKey:               checkoutID,
					subscription.InitiatorIDMetadataKey: currentPrincipal.ID,
					"managed_by":                        "frontier",
				},
				TrialPeriodDays: trialDays,
				TrialSettings: &stripe.CheckoutSessionSubscriptionDataTrialSettingsParams{
					EndBehavior: &stripe.CheckoutSessionSubscriptionDataTrialSettingsEndBehaviorParams{
						MissingPaymentMethod: stripe.String(string(stripe.SubscriptionScheduleEndBehaviorCancel)),
					},
				},
			},
			AllowPromotionCodes:     stripe.Bool(true),
			CancelURL:               stripe.String(ch.CancelUrl),
			SuccessURL:              stripe.String(ch.SuccessUrl),
			ExpiresAt:               stripe.Int64(time.Now().Add(SessionValidity).Unix()),
			PaymentMethodCollection: stripe.String(string(stripe.PaymentLinkPaymentMethodCollectionIfRequired)),
		})
		if err != nil {
			return Checkout{}, fmt.Errorf("failed to create subscription at billing provider: %w", err)
		}

		return s.repository.Create(ctx, Checkout{
			ID:               checkoutID,
			ProviderID:       stripeCheckout.ID,
			CustomerID:       billingCustomer.ID,
			PlanID:           plan.ID,
			SkipTrial:        ch.SkipTrial,
			CancelAfterTrial: ch.CancelAfterTrial,
			CancelUrl:        ch.CancelUrl,
			SuccessUrl:       ch.SuccessUrl,
			CheckoutUrl:      stripeCheckout.URL,
			State:            string(stripeCheckout.Status),
			PaymentStatus:    string(stripeCheckout.PaymentStatus),
			Metadata: map[string]any{
				"plan_name":            plan.Name,
				InitiatorIDMetadataKey: currentPrincipal.ID,
			},
			ExpireAt: utils.AsTimeFromEpoch(stripeCheckout.ExpiresAt),
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
		var minQ int64 = MinimumProductQuantity
		var maxQ int64 = MaximumProductQuantity
		var adjustableQuantity bool = true
		if chProduct.Config.MinQuantity > 0 {
			minQ = chProduct.Config.MinQuantity
		}
		if chProduct.Config.MaxQuantity > 0 {
			maxQ = chProduct.Config.MaxQuantity
		}
		if maxQ == 1 {
			adjustableQuantity = false
		}
		var defaultQ int64 = 1
		if ch.Quantity > 0 && ch.Quantity <= maxQ && ch.Quantity >= minQ {
			defaultQ = ch.Quantity
		}
		for _, productPrice := range chProduct.Prices {
			itemParams := &stripe.CheckoutSessionLineItemParams{
				Price: stripe.String(productPrice.ProviderID),
				AdjustableQuantity: &stripe.CheckoutSessionLineItemAdjustableQuantityParams{
					Enabled: stripe.Bool(adjustableQuantity),
					Minimum: stripe.Int64(minQ),
					Maximum: stripe.Int64(maxQ),
				},
			}
			if productPrice.UsageType == product.PriceUsageTypeLicensed {
				itemParams.Quantity = stripe.Int64(defaultQ)
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
			Currency: stripe.String(billingCustomer.Currency),
			Customer: stripe.String(billingCustomer.ProviderID),
			InvoiceCreation: &stripe.CheckoutSessionInvoiceCreationParams{
				Enabled: stripe.Bool(true),
			},
			LineItems: subsItems,
			Mode:      stripe.String(string(stripe.CheckoutSessionModePayment)),
			Metadata: map[string]string{
				"org_id":               billingCustomer.OrgID,
				"product_name":         chProduct.Name,
				"credit_amount":        fmt.Sprintf("%d", chProduct.Config.CreditAmount),
				CheckoutIDMetadataKey:  checkoutID,
				InitiatorIDMetadataKey: currentPrincipal.ID,
				"managed_by":           "frontier",
			},
			CustomerUpdate: &stripe.CheckoutSessionCustomerUpdateParams{
				Address: stripe.String(string(stripe.CheckoutSessionBillingAddressCollectionAuto)),
			},
			AllowPromotionCodes: stripe.Bool(true),
			CancelURL:           stripe.String(ch.CancelUrl),
			SuccessURL:          stripe.String(ch.SuccessUrl),
			ExpiresAt:           stripe.Int64(time.Now().Add(SessionValidity).Unix()),
		})
		if err != nil {
			return Checkout{}, fmt.Errorf("failed to buy product at billing provider: %w", err)
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
				"product_name":         chProduct.Name,
				InitiatorIDMetadataKey: currentPrincipal.ID,
			},
			ExpireAt: utils.AsTimeFromEpoch(stripeCheckout.ExpiresAt),
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

	var errs []error
	// find all checkout sessions of the customer that require a sync
	// and update their state in system
	for idx, ch := range checks {
		if ctx.Err() != nil {
			// stop processing if context is done
			break
		}

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
			Expand: []*string{
				stripe.String("line_items.data.price.product"),
			},
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get checkout session from billing provider: %w", err))
			continue
		}
		if ch.PaymentStatus != string(checkoutSession.PaymentStatus) {
			ch.PaymentStatus = string(checkoutSession.PaymentStatus)
		}
		if ch.State != string(checkoutSession.Status) {
			ch.State = string(checkoutSession.Status)
		}
		if checkoutSession.Subscription != nil {
			ch.Metadata[ProviderIDSubscriptionMetadataKey] = checkoutSession.Subscription.ID
		}
		ch.Metadata[AmountTotalMetadataKey] = checkoutSession.AmountTotal
		ch.Metadata[CurrencyMetadataKey] = checkoutSession.Currency
		if checkoutSession.LineItems != nil {
			for _, lintitem := range checkoutSession.LineItems.Data {
				if lintitem.Price != nil && lintitem.Price.Product != nil &&
					lintitem.Price.Product.ID == ch.ProductID {
					ch.Metadata[ProductQuantityMetadataKey] = lintitem.Quantity
				}
			}
		}
		if checks[idx], err = s.repository.UpdateByID(ctx, ch); err != nil {
			return fmt.Errorf("failed to update checkout session: %w", err)
		}
	}

	// if payment is completed, create subscription for them in system
	for _, ch := range checks {
		if processed, ok := ch.Metadata[ProcessedMetadataKey].(bool); ok && processed {
			continue
		}

		if ch.State == StateComplete.String() &&
			(ch.PaymentStatus == "paid" || ch.PaymentStatus == "no_payment_required") {
			// checkout could be for a plan or a product

			if ch.PlanID != "" {
				// if the checkout was created for subscription
				if _, err := s.ensureSubscription(ctx, ch); err != nil {
					errs = append(errs, fmt.Errorf("ensureSubscription: %w", err))
					continue
				}
			} else if ch.ProductID != "" {
				// if the checkout was created for product
				if err := s.ensureCreditsForProduct(ctx, ch); err != nil {
					errs = append(errs, fmt.Errorf("ensureCreditsForProduct: %w", err))
					continue
				}
			}

			ch.Metadata[ProcessedMetadataKey] = true
			if _, err := s.repository.UpdateByID(ctx, ch); err != nil {
				errs = append(errs, fmt.Errorf("failed to update checkout session: %w", err))
			}
		}
	}

	return errors.Join(errs...)
}

func (s *Service) ensureCreditsForProduct(ctx context.Context, ch Checkout) error {
	chProduct, err := s.productService.GetByID(ctx, ch.ProductID)
	if err != nil {
		return err
	}
	if chProduct.Behavior != product.CreditBehavior {
		return fmt.Errorf("invalid product, not a credit product")
	}
	creditAmount := chProduct.Config.CreditAmount
	if quantity, ok := ch.Metadata[ProductQuantityMetadataKey]; ok {
		creditAmount = cast.ToInt64(quantity) * chProduct.Config.CreditAmount
	}

	description := fmt.Sprintf("addition of %d credits for %s", creditAmount, chProduct.Title)
	if price, pok := ch.Metadata[AmountTotalMetadataKey]; pok {
		if currency, cok := ch.Metadata[CurrencyMetadataKey].(string); cok {
			description = fmt.Sprintf("addition of %d credits for %s at %d[%s]", creditAmount, chProduct.Title, price, currency)
		}
	}
	initiatorID := ""
	if id, ok := ch.Metadata[InitiatorIDMetadataKey].(string); ok {
		initiatorID = id
	}

	md := metadata.Build(ch.Metadata)
	md[CheckoutIDMetadataKey] = ch.ID
	if err := s.creditService.Add(ctx, credit.Credit{
		ID:          ch.ID,
		CustomerID:  ch.CustomerID,
		Amount:      creditAmount,
		Metadata:    md,
		Description: description,
		Source:      credit.SourceSystemBuyEvent,
		UserID:      initiatorID,
	}); err != nil && !errors.Is(err, credit.ErrAlreadyApplied) {
		return err
	}
	return nil
}

func (s *Service) checkIfAlreadySubscribed(ctx context.Context, ch Checkout) (string, error) {
	// check if subscription exists
	subs, err := s.subscriptionService.List(ctx, subscription.Filter{
		CustomerID: ch.CustomerID,
		PlanID:     ch.PlanID,
	})
	if err != nil {
		return "", err
	}

	for _, sub := range subs {
		// don't care about canceled or ended subscriptions
		// trialing subscriptions will be canceled later
		if sub.State == subscription.StateCanceled.String() ||
			sub.State == subscription.StateEnded.String() ||
			sub.State == subscription.StateTrialing.String() {
			continue
		}
		// subscription already exists
		return sub.ID, nil
	}

	return "", nil
}

func (s *Service) cancelTrialingSubscription(ctx context.Context, customerID string, planID string) error {
	// check if subscription exists
	subs, err := s.subscriptionService.List(ctx, subscription.Filter{
		CustomerID: customerID,
		PlanID:     planID,
	})
	if err != nil {
		return err
	}

	for _, sub := range subs {
		// cancel immediately if trialing
		if sub.State == subscription.StateTrialing.String() && !sub.TrialEndsAt.IsZero() {
			if _, err := s.subscriptionService.Cancel(ctx, sub.ID, true); err != nil {
				return fmt.Errorf("failed to cancel trialing subscription: %w", err)
			}
		}
	}

	return nil
}

func (s *Service) ensureSubscription(ctx context.Context, ch Checkout) (string, error) {
	if ch.Metadata[ProviderIDSubscriptionMetadataKey] == nil {
		return "", fmt.Errorf("invalid checkout session, provider_subscription_id is missing")
	}
	subProviderID := ch.Metadata[ProviderIDSubscriptionMetadataKey].(string)

	// check if already created in frontier
	_, err := s.subscriptionService.GetByProviderID(ctx, subProviderID)
	if err != nil && !errors.Is(err, subscription.ErrNotFound) {
		return "", err
	}
	if err == nil {
		// already created
		return "", nil
	}

	// cancel existing trials if any
	if err := s.cancelTrialingSubscription(ctx, ch.CustomerID, ch.PlanID); err != nil {
		return "", err
	}

	stripeSubscription, err := s.stripeClient.Subscriptions.Get(subProviderID,
		&stripe.SubscriptionParams{
			Params: stripe.Params{
				Context: ctx,
			},
		})
	if err != nil {
		return "", fmt.Errorf("failed to get subscription from billing provider: %w", err)
	}

	// create subscription
	md := metadata.Build(ch.Metadata)
	md[CheckoutIDMetadataKey] = ch.ID
	md[subscription.ProviderTestResource] = !stripeSubscription.Livemode
	sub, err := s.subscriptionService.Create(ctx, subscription.Subscription{
		ID:          uuid.New().String(),
		ProviderID:  subProviderID,
		CustomerID:  ch.CustomerID,
		PlanID:      ch.PlanID,
		State:       string(stripeSubscription.Status),
		Metadata:    md,
		TrialEndsAt: utils.AsTimeFromEpoch(stripeSubscription.TrialEnd),
	})
	if err != nil {
		return "", err
	}

	// if set to cancel after trial, schedule a phase to cancel the subscription
	if ch.CancelAfterTrial && stripeSubscription.TrialEnd > 0 {
		_, err := s.subscriptionService.Cancel(ctx, sub.ID, false)
		if err != nil {
			return "", fmt.Errorf("failed to schedule cancel of subscription after trial: %w", err)
		}
	}
	return sub.ID, nil
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Checkout, error) {
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
		ExpireAt:    utils.AsTimeFromEpoch(stripeCheckout.ExpiresAt),
		Metadata: map[string]any{
			"mode": "setup",
		},
	})
}

func (s *Service) CreateSessionForCustomerPortal(ctx context.Context, ch Checkout) (Checkout, error) {
	// get billing
	billingCustomer, err := s.customerService.GetByID(ctx, ch.CustomerID)
	if err != nil {
		return Checkout{}, err
	}

	checkoutID := uuid.New().String()

	sessionParams := &stripe.BillingPortalSessionParams{
		Params: stripe.Params{
			Context: ctx,
		},
		Customer: stripe.String(billingCustomer.ProviderID),
	}

	if ch.CancelUrl != "" {
		sessionParams.ReturnURL = stripe.String(ch.CancelUrl)
	}

	session, err := s.stripeClient.BillingPortalSessions.New(sessionParams)

	if err != nil {
		return Checkout{}, fmt.Errorf("failed to create session for customer portal: %w", err)
	}

	return Checkout{
		ID:          checkoutID,
		ProviderID:  session.ID,
		CustomerID:  billingCustomer.ID,
		CancelUrl:   ch.CancelUrl,
		SuccessUrl:  ch.SuccessUrl,
		CheckoutUrl: session.URL,
		Metadata: map[string]any{
			"mode": "customer_portal",
		},
	}, nil
}

// Apply applies the actual request directly without creating a checkout session
// for example when a request is created for a plan, it will directly subscribe without
// actually paying for it
func (s *Service) Apply(ctx context.Context, ch Checkout) (*subscription.Subscription, *product.Product, error) {
	ch.ID = uuid.New().String()
	// get billing
	billingCustomer, err := s.customerService.GetByID(ctx, ch.CustomerID)
	if err != nil {
		return nil, nil, err
	}

	autoTaxParams := &stripe.SubscriptionAutomaticTaxParams{
		Enabled: stripe.Bool(s.stripeAutoTax),
	}

	// checkout could be for a plan or a product
	if ch.PlanID != "" && !billingCustomer.IsOffline() {
		plan, err := s.planService.GetByID(ctx, ch.PlanID)
		if err != nil {
			return nil, nil, err
		}
		// ensure we use uuid
		ch.PlanID = plan.ID

		// if already subscribed to the plan, return
		if subID, err := s.checkIfAlreadySubscribed(ctx, ch); err != nil {
			return nil, nil, err
		} else if subID != "" {
			return nil, nil, fmt.Errorf("already subscribed to the plan")
		}

		if err := s.cancelTrialingSubscription(ctx, ch.CustomerID, ch.PlanID); err != nil {
			return nil, nil, err
		}

		// create subscription items
		var subsItems []*stripe.SubscriptionItemsParams
		userCount, err := s.orgService.MemberCount(ctx, billingCustomer.OrgID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get member count: %w", err)
		}

		var totalExpectedPrice int64
		for _, planProduct := range plan.Products {
			// if it's credit, skip, they are handled separately
			if planProduct.Behavior == product.CreditBehavior {
				continue
			}
			// if per seat, check if there is a limit of seats, if it breaches limit, fail
			if planProduct.IsSeatLimitBreached(userCount) {
				return nil, nil, fmt.Errorf("member count exceeds allowed limit of the plan: %w", product.ErrPerSeatLimitReached)
			}

			for _, productPrice := range planProduct.Prices {
				// only work with plan interval prices
				if productPrice.Interval != plan.Interval {
					continue
				}

				var quantity int64 = 1
				if productPrice.IsLicensed() && planProduct.HasPerSeatBehavior() {
					quantity = userCount
				}

				itemParams := &stripe.SubscriptionItemsParams{
					Price:    stripe.String(productPrice.ProviderID),
					Quantity: stripe.Int64(quantity),
					Metadata: map[string]string{
						"org_id":     billingCustomer.OrgID,
						"product_id": planProduct.ID,
					},
				}
				subsItems = append(subsItems, itemParams)
				totalExpectedPrice += productPrice.Amount * quantity
			}
		}

		var trialDays *int64 = nil
		if plan.TrialDays > 0 && !ch.SkipTrial {
			trialDays = stripe.Int64(plan.TrialDays)
		}

		if totalExpectedPrice == 0 {
			// if total price is 0, disable auto tax. This ensures that when the subscription is created without
			// user billing details while onboarding, creating 0 amount invoice doesn't fail
			// This will be toggled back on when the user changes it's plan to a paid one
			autoTaxParams.Enabled = stripe.Bool(false)
		}

		var couponID *string
		if ch.ProviderCouponID != "" {
			couponID = stripe.String(ch.ProviderCouponID)
		}
		// create subscription directly
		stripeSubscription, err := s.stripeClient.Subscriptions.New(&stripe.SubscriptionParams{
			Params: stripe.Params{
				Context: ctx,
			},
			AutomaticTax: autoTaxParams,
			Customer:     stripe.String(billingCustomer.ProviderID),
			Currency:     stripe.String(billingCustomer.Currency),
			Items:        subsItems,
			Metadata: map[string]string{
				"org_id":     billingCustomer.OrgID,
				"managed_by": "frontier",
			},
			TrialPeriodDays: trialDays,
			TrialSettings: &stripe.SubscriptionTrialSettingsParams{
				EndBehavior: &stripe.SubscriptionTrialSettingsEndBehaviorParams{
					MissingPaymentMethod: stripe.String(string(stripe.SubscriptionScheduleEndBehaviorCancel)),
				},
			},
			Coupon: couponID,
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
				"org_id":                          billingCustomer.OrgID,
				"delegated":                       "true",
				"checkout_id":                     ch.ID,
				subscription.ProviderTestResource: !stripeSubscription.Livemode,
			},
			State:                string(stripeSubscription.Status),
			TrialEndsAt:          utils.AsTimeFromEpoch(stripeSubscription.TrialEnd),
			BillingCycleAnchorAt: utils.AsTimeFromEpoch(stripeSubscription.BillingCycleAnchor),
			CurrentPeriodStartAt: utils.AsTimeFromEpoch(stripeSubscription.CurrentPeriodStart),
			CurrentPeriodEndAt:   utils.AsTimeFromEpoch(stripeSubscription.CurrentPeriodEnd),
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create subscription: %w", err)
		}
		return &subs, nil, nil
	} else if ch.ProductID != "" {
		chProduct, err := s.productService.GetByID(ctx, ch.ProductID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get product: %w", err)
		}

		if chProduct.Behavior != product.CreditBehavior {
			// if not credit product, we can't apply directly and is supported yet
			return nil, nil, fmt.Errorf("not supported yet")
		}

		var amount = chProduct.Config.CreditAmount
		if quantity, ok := ch.Metadata[ProductQuantityMetadataKey]; ok {
			amount = cast.ToInt64(quantity) * chProduct.Config.CreditAmount
		}
		if ch.Quantity > 0 {
			amount = ch.Quantity * chProduct.Config.CreditAmount
		}

		if err := s.creditService.Add(ctx, credit.Credit{
			ID:          ch.ID,
			CustomerID:  ch.CustomerID,
			Amount:      amount,
			Metadata:    ch.Metadata,
			Source:      credit.SourceSystemAwardedEvent,
			Description: fmt.Sprintf("Awarded %d credits for %s", amount, chProduct.Title),
		}); err != nil {
			return nil, nil, err
		}
		return nil, &chProduct, nil
	}

	return nil, nil, fmt.Errorf("invalid checkout request")
}

func (s *Service) TriggerSyncByProviderID(ctx context.Context, id string) error {
	checkouts, err := s.repository.List(ctx, Filter{
		ProviderID: id,
	})
	if err != nil {
		return err
	}
	if len(checkouts) == 0 {
		return ErrNotFound
	}
	return s.SyncWithProvider(ctx, checkouts[0].CustomerID)
}
