// Package stripeprovider implements billing.Provider using the Stripe API.
package stripeprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/billing"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/client"
	"github.com/stripe/stripe-go/v79/webhook"
)

var _ billing.Provider = (*Provider)(nil)

// Provider implements billing.Provider backed by Stripe.
type Provider struct {
	client *client.API
}

// New creates a new Stripe billing provider.
func New(c *client.API) *Provider {
	return &Provider{client: c}
}

// --- Customer ---

func (p *Provider) CreateCustomer(ctx context.Context, params billing.CreateCustomerParams) (*billing.ProviderCustomer, error) {
	var taxIDs []*stripe.CustomerTaxIDDataParams
	for _, t := range params.TaxIDs {
		taxIDs = append(taxIDs, &stripe.CustomerTaxIDDataParams{
			Type:  stripe.String(t.Type),
			Value: stripe.String(t.Value),
		})
	}
	sc, err := p.client.Customers.New(&stripe.CustomerParams{
		Params:  stripe.Params{Context: ctx},
		Email:   &params.Email,
		Name:    &params.Name,
		Phone:   &params.Phone,
		Address: toStripeAddress(params.Address),
		TaxIDData: taxIDs,
		Metadata:  params.Metadata,
		TestClock: params.TestClockID,
	})
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok && stripeErr.Code == stripe.ErrorCodeParameterMissing {
			return nil, fmt.Errorf("missing parameter while registering to biller: %s", stripeErr.Error())
		}
		return nil, fmt.Errorf("failed to register in billing provider: %w", err)
	}
	return fromStripeCustomer(sc), nil
}

func (p *Provider) UpdateCustomer(ctx context.Context, providerID string, params billing.UpdateCustomerParams) (*billing.ProviderCustomer, error) {
	sc, err := p.client.Customers.Update(providerID, &stripe.CustomerParams{
		Params:   stripe.Params{Context: ctx},
		Email:    &params.Email,
		Name:     &params.Name,
		Phone:    &params.Phone,
		Address:  toStripeAddress(params.Address),
		Metadata: params.Metadata,
	})
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok && stripeErr.Code == stripe.ErrorCodeParameterMissing {
			return nil, fmt.Errorf("missing parameter while registering to biller: %s", stripeErr.Error())
		}
		return nil, fmt.Errorf("failed to register in billing provider: %w", err)
	}
	return fromStripeCustomer(sc), nil
}

func (p *Provider) DeleteCustomer(ctx context.Context, providerID string) error {
	_, err := p.client.Customers.Del(providerID, &stripe.CustomerParams{
		Params: stripe.Params{Context: ctx},
	})
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok && stripeErr.Code == stripe.ErrorCodeResourceMissing {
			return nil // already deleted
		}
		return fmt.Errorf("failed to delete customer from billing provider: %w", err)
	}
	return nil
}

func (p *Provider) GetCustomer(ctx context.Context, providerID string) (*billing.ProviderCustomer, error) {
	sc, err := p.client.Customers.Get(providerID, &stripe.CustomerParams{
		Params: stripe.Params{Context: ctx},
	})
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok && stripeErr.Code == stripe.ErrorCodeResourceMissing {
			return nil, billing.ErrNotFoundInProvider
		}
		return nil, fmt.Errorf("failed to get customer from billing provider: %w", err)
	}
	return fromStripeCustomer(sc), nil
}

func (p *Provider) ListPaymentMethods(ctx context.Context, customerProviderID string) ([]billing.ProviderPaymentMethod, error) {
	iter := p.client.PaymentMethods.List(&stripe.PaymentMethodListParams{
		Customer:   stripe.String(customerProviderID),
		ListParams: stripe.ListParams{Context: ctx},
		Expand:     []*string{stripe.String("data.customer")},
	})

	var methods []billing.ProviderPaymentMethod
	for iter.Next() {
		pm := iter.PaymentMethod()
		m := billing.ProviderPaymentMethod{
			ID:        pm.ID,
			Type:      string(pm.Type),
			CreatedAt: pm.Created,
			Metadata:  toAnyMap(pm.Metadata),
		}
		if pm.Type == stripe.PaymentMethodTypeCard {
			m.CardBrand = string(pm.Card.Brand)
			m.CardLast4 = pm.Card.Last4
			m.CardExpiryMonth = pm.Card.ExpMonth
			m.CardExpiryYear = pm.Card.ExpYear
		}
		if pm.Customer.InvoiceSettings != nil &&
			pm.Customer.InvoiceSettings.DefaultPaymentMethod != nil &&
			pm.Customer.InvoiceSettings.DefaultPaymentMethod.ID == pm.ID {
			m.IsDefault = true
		}
		methods = append(methods, m)
	}
	return methods, nil
}

// --- Product / Price ---

func (p *Provider) CreateProduct(ctx context.Context, params billing.CreateProductParams) error {
	_, err := p.client.Products.New(&stripe.ProductParams{
		Params:      stripe.Params{Context: ctx},
		ID:          &params.ID,
		Name:        &params.Name,
		Description: &params.Description,
		Metadata:    params.Metadata,
	})
	return err
}

func (p *Provider) UpdateProduct(ctx context.Context, providerID string, params billing.UpdateProductParams) error {
	_, err := p.client.Products.Update(providerID, &stripe.ProductParams{
		Params:      stripe.Params{Context: ctx},
		Name:        &params.Name,
		Description: &params.Description,
		Metadata:    params.Metadata,
	})
	return err
}

func (p *Provider) CreatePrice(ctx context.Context, params billing.CreatePriceParams) (string, error) {
	pp := &stripe.PriceParams{
		Params:        stripe.Params{Context: ctx},
		Product:       &params.ProductID,
		Nickname:      &params.Name,
		BillingScheme: stripe.String(params.BillingScheme),
		Currency:      &params.Currency,
		UnitAmount:    &params.Amount,
		Metadata:      params.Metadata,
	}
	if params.Interval != "" {
		pp.Recurring = &stripe.PriceRecurringParams{
			Interval:  stripe.String(params.Interval),
			UsageType: stripe.String(params.UsageType),
		}
		if params.MeteredAggregate != "" {
			pp.Recurring.AggregateUsage = stripe.String(params.MeteredAggregate)
		}
	}
	sp, err := p.client.Prices.New(pp)
	if err != nil {
		return "", err
	}
	return sp.ID, nil
}

func (p *Provider) UpdatePrice(ctx context.Context, providerID string, params billing.UpdatePriceParams) error {
	_, err := p.client.Prices.Update(providerID, &stripe.PriceParams{
		Params:   stripe.Params{Context: ctx},
		Nickname: &params.Name,
		Metadata: params.Metadata,
	})
	return err
}

// --- Subscription ---

func (p *Provider) CreateSubscription(ctx context.Context, params billing.CreateSubscriptionParams) (*billing.ProviderSubscription, error) {
	sp := &stripe.SubscriptionParams{
		Params:   stripe.Params{Context: ctx},
		Customer: stripe.String(params.CustomerProviderID),
		Currency: stripe.String(params.Currency),
		Metadata: params.Metadata,
		AutomaticTax: &stripe.SubscriptionAutomaticTaxParams{
			Enabled: stripe.Bool(params.AutoTax),
		},
		TrialSettings: &stripe.SubscriptionTrialSettingsParams{
			EndBehavior: &stripe.SubscriptionTrialSettingsEndBehaviorParams{
				MissingPaymentMethod: stripe.String(string(stripe.SubscriptionScheduleEndBehaviorCancel)),
			},
		},
	}
	if params.TrialDays != nil {
		sp.TrialPeriodDays = params.TrialDays
	}
	if params.CouponID != "" {
		sp.Coupon = stripe.String(params.CouponID)
	}
	for _, item := range params.Items {
		sp.Items = append(sp.Items, &stripe.SubscriptionItemsParams{
			Price:    stripe.String(item.PriceProviderID),
			Quantity: stripe.Int64(item.Quantity),
			Metadata: item.Metadata,
		})
	}
	ss, err := p.client.Subscriptions.New(sp)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription at billing provider: %w", err)
	}
	return fromStripeSubscription(ss), nil
}

func (p *Provider) GetSubscription(ctx context.Context, providerID string) (*billing.ProviderSubscription, error) {
	ss, err := p.client.Subscriptions.Get(providerID, &stripe.SubscriptionParams{
		Params: stripe.Params{Context: ctx},
		Expand: []*string{stripe.String("schedule")},
	})
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok && stripeErr.Code == stripe.ErrorCodeResourceMissing {
			return nil, billing.ErrNotFoundInProvider
		}
		return nil, fmt.Errorf("failed to get subscription from billing provider: %w", err)
	}
	return fromStripeSubscription(ss), nil
}

func (p *Provider) CancelSubscription(ctx context.Context, providerID string, params billing.CancelSubscriptionParams) (*billing.ProviderSubscription, error) {
	ss, err := p.client.Subscriptions.Cancel(providerID, &stripe.SubscriptionCancelParams{
		Params:     stripe.Params{Context: ctx},
		InvoiceNow: stripe.Bool(params.InvoiceNow),
		Prorate:    stripe.Bool(params.Prorate),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription at billing provider: %w", err)
	}
	return fromStripeSubscription(ss), nil
}

func (p *Provider) UpdateSubscriptionItems(ctx context.Context, providerID string, params billing.UpdateSubscriptionItemsParams) error {
	sp := &stripe.SubscriptionParams{
		Params: stripe.Params{Context: ctx},
	}
	for _, item := range params.Items {
		sp.Items = append(sp.Items, &stripe.SubscriptionItemsParams{
			ID:       stripe.String(item.ID),
			Price:    stripe.String(item.PriceID),
			Quantity: stripe.Int64(item.Quantity),
			Metadata: item.Metadata,
		})
	}
	if params.PendingInvoiceItemInterval != nil {
		sp.PendingInvoiceItemInterval = &stripe.SubscriptionPendingInvoiceItemIntervalParams{
			Interval:      stripe.String(params.PendingInvoiceItemInterval.Interval),
			IntervalCount: stripe.Int64(params.PendingInvoiceItemInterval.IntervalCount),
		}
	}
	_, err := p.client.Subscriptions.Update(providerID, sp)
	if err != nil {
		return fmt.Errorf("failed to update subscription quantity at billing provider: %w", err)
	}
	return nil
}

// --- Schedule ---

func (p *Provider) GetSchedule(ctx context.Context, scheduleID string) (*billing.ProviderSchedule, error) {
	ss, err := p.client.SubscriptionSchedules.Get(scheduleID, &stripe.SubscriptionScheduleParams{
		Params: stripe.Params{Context: ctx},
		Expand: []*string{stripe.String("phases.items.price.product")},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription schedule from billing provider: %w", err)
	}
	return fromStripeSchedule(ss), nil
}

func (p *Provider) CreateScheduleFromSubscription(ctx context.Context, subscriptionProviderID string) (*billing.ProviderSchedule, error) {
	ss, err := p.client.SubscriptionSchedules.New(&stripe.SubscriptionScheduleParams{
		Params:           stripe.Params{Context: ctx},
		FromSubscription: stripe.String(subscriptionProviderID),
		Expand:           []*string{stripe.String("phases.items.price.product")},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription schedule at billing provider: %w", err)
	}
	return fromStripeSchedule(ss), nil
}

func (p *Provider) UpdateSchedule(ctx context.Context, scheduleID string, params billing.UpdateScheduleParams) (*billing.ProviderSchedule, error) {
	sp := &stripe.SubscriptionScheduleParams{
		Params: stripe.Params{Context: ctx},
	}
	if params.EndBehavior != "" {
		sp.EndBehavior = stripe.String(params.EndBehavior)
	}
	if params.ProrationBehavior != "" {
		sp.ProrationBehavior = stripe.String(params.ProrationBehavior)
	}
	if params.CollectionMethod != "" {
		sp.DefaultSettings = &stripe.SubscriptionScheduleDefaultSettingsParams{
			CollectionMethod: stripe.String(params.CollectionMethod),
		}
	}
	for _, phase := range params.Phases {
		pp := &stripe.SubscriptionSchedulePhaseParams{
			Currency: stripe.String(phase.Currency),
			Metadata: phase.Metadata,
		}
		if phase.StartDate != nil {
			pp.StartDate = phase.StartDate
		}
		if phase.EndDate != nil {
			pp.EndDate = phase.EndDate
		}
		if phase.EndDateNow {
			pp.EndDateNow = stripe.Bool(true)
		}
		if phase.Iterations != nil {
			pp.Iterations = phase.Iterations
		}
		if phase.Description != "" {
			pp.Description = stripe.String(phase.Description)
		}
		if phase.TrialEnd != nil {
			pp.TrialEnd = phase.TrialEnd
		}
		if phase.ProrationBehavior != "" {
			pp.ProrationBehavior = stripe.String(phase.ProrationBehavior)
		}
		if phase.CollectionMethod != "" {
			pp.CollectionMethod = stripe.String(phase.CollectionMethod)
		}
		pp.AutomaticTax = &stripe.SubscriptionSchedulePhaseAutomaticTaxParams{
			Enabled: stripe.Bool(phase.AutoTax),
		}
		for _, item := range phase.Items {
			pp.Items = append(pp.Items, &stripe.SubscriptionSchedulePhaseItemParams{
				Price:    stripe.String(item.PriceID),
				Quantity: stripe.Int64(item.Quantity),
				Metadata: item.Metadata,
			})
		}
		sp.Phases = append(sp.Phases, pp)
	}
	ss, err := p.client.SubscriptionSchedules.Update(scheduleID, sp)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription schedule at billing provider: %w", err)
	}
	return fromStripeSchedule(ss), nil
}

// --- Checkout ---

func (p *Provider) CreateCheckoutSession(ctx context.Context, params billing.CreateCheckoutSessionParams) (*billing.ProviderCheckoutSession, error) {
	sp := &stripe.CheckoutSessionParams{
		Params:   stripe.Params{Context: ctx},
		Customer: stripe.String(params.CustomerProviderID),
		Currency: stripe.String(params.Currency),
		Mode:     stripe.String(params.Mode),
		Metadata: params.Metadata,
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(params.AutoTax),
		},
		CancelURL:          stripe.String(params.CancelURL),
		SuccessURL:         stripe.String(params.SuccessURL),
		AllowPromotionCodes: stripe.Bool(params.AllowPromotionCodes),
	}
	if params.ExpiresAt > 0 {
		sp.ExpiresAt = stripe.Int64(params.ExpiresAt)
	}
	if params.AddressCollection != "" {
		sp.CustomerUpdate = &stripe.CheckoutSessionCustomerUpdateParams{
			Address: stripe.String(params.AddressCollection),
		}
	}
	if params.PaymentMethodCollection != "" {
		sp.PaymentMethodCollection = stripe.String(params.PaymentMethodCollection)
	}
	for _, li := range params.LineItems {
		item := &stripe.CheckoutSessionLineItemParams{
			Price: stripe.String(li.PriceProviderID),
		}
		if li.Quantity > 0 {
			item.Quantity = stripe.Int64(li.Quantity)
		}
		if li.AdjustableQuantity != nil {
			item.AdjustableQuantity = &stripe.CheckoutSessionLineItemAdjustableQuantityParams{
				Enabled: stripe.Bool(li.AdjustableQuantity.Enabled),
			}
			if li.AdjustableQuantity.Enabled {
				item.AdjustableQuantity.Minimum = stripe.Int64(li.AdjustableQuantity.Minimum)
				item.AdjustableQuantity.Maximum = stripe.Int64(li.AdjustableQuantity.Maximum)
			}
		}
		sp.LineItems = append(sp.LineItems, item)
	}
	if params.Mode == "subscription" && params.SubscriptionMetadata != nil {
		sp.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: params.SubscriptionMetadata,
		}
		if desc, ok := params.SubscriptionMetadata["description"]; ok {
			sp.SubscriptionData.Description = stripe.String(desc)
			delete(sp.SubscriptionData.Metadata, "description")
		}
		if params.TrialDays != nil {
			sp.SubscriptionData.TrialPeriodDays = params.TrialDays
			sp.SubscriptionData.TrialSettings = &stripe.CheckoutSessionSubscriptionDataTrialSettingsParams{
				EndBehavior: &stripe.CheckoutSessionSubscriptionDataTrialSettingsEndBehaviorParams{
					MissingPaymentMethod: stripe.String(string(stripe.SubscriptionScheduleEndBehaviorCancel)),
				},
			}
		}
	}
	if params.InvoiceCreation {
		sp.InvoiceCreation = &stripe.CheckoutSessionInvoiceCreationParams{
			Enabled: stripe.Bool(true),
		}
	}
	if len(params.PaymentMethodTypes) > 0 {
		for _, pmt := range params.PaymentMethodTypes {
			sp.PaymentMethodTypes = append(sp.PaymentMethodTypes, stripe.String(pmt))
		}
		sp.PaymentMethodOptions = &stripe.CheckoutSessionPaymentMethodOptionsParams{
			CustomerBalance: &stripe.CheckoutSessionPaymentMethodOptionsCustomerBalanceParams{
				FundingType: stripe.String(string(stripe.CheckoutSessionPaymentMethodOptionsCustomerBalanceFundingTypeBankTransfer)),
				BankTransfer: &stripe.CheckoutSessionPaymentMethodOptionsCustomerBalanceBankTransferParams{
					Type: stripe.String(string(stripe.CheckoutSessionPaymentMethodOptionsCustomerBalanceBankTransferTypeUSBankTransfer)),
				},
			},
		}
	}

	cs, err := p.client.CheckoutSessions.New(sp)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout at billing provider: %w", err)
	}
	return fromStripeCheckoutSession(cs), nil
}

func (p *Provider) GetCheckoutSession(ctx context.Context, providerID string) (*billing.ProviderCheckoutSession, error) {
	cs, err := p.client.CheckoutSessions.Get(providerID, &stripe.CheckoutSessionParams{
		Params: stripe.Params{Context: ctx},
		Expand: []*string{stripe.String("line_items.data.price.product")},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get checkout session from billing provider: %w", err)
	}
	return fromStripeCheckoutSession(cs), nil
}

func (p *Provider) CreateBillingPortalSession(ctx context.Context, params billing.CreateBillingPortalParams) (string, error) {
	sp := &stripe.BillingPortalSessionParams{
		Params:   stripe.Params{Context: ctx},
		Customer: stripe.String(params.CustomerProviderID),
	}
	if params.ReturnURL != "" {
		sp.ReturnURL = stripe.String(params.ReturnURL)
	}
	session, err := p.client.BillingPortalSessions.New(sp)
	if err != nil {
		return "", fmt.Errorf("failed to create session for customer portal: %w", err)
	}
	return session.URL, nil
}

// --- Invoice ---

func (p *Provider) ListInvoices(ctx context.Context, customerProviderID string) ([]billing.ProviderInvoice, error) {
	iter := p.client.Invoices.List(&stripe.InvoiceListParams{
		Customer:   stripe.String(customerProviderID),
		ListParams: stripe.ListParams{Context: ctx},
		Expand:     []*string{stripe.String("data.lines")},
	})
	var invoices []billing.ProviderInvoice
	for iter.Next() {
		invoices = append(invoices, *fromStripeInvoice(iter.Invoice()))
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}
	return invoices, nil
}

func (p *Provider) GetUpcomingInvoice(ctx context.Context, customerProviderID string) (*billing.ProviderInvoice, error) {
	si, err := p.client.Invoices.Upcoming(&stripe.InvoiceUpcomingParams{
		Customer: stripe.String(customerProviderID),
		Params:   stripe.Params{Context: ctx},
	})
	if err != nil {
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) && stripeErr.Code == stripe.ErrorCodeInvoiceUpcomingNone {
			return nil, billing.ErrNoUpcomingInvoice
		}
		return nil, fmt.Errorf("failed to get upcoming invoice: %w", err)
	}
	return fromStripeInvoice(si), nil
}

func (p *Provider) CreateInvoice(ctx context.Context, params billing.CreateInvoiceParams) (*billing.ProviderInvoice, error) {
	sp := &stripe.InvoiceParams{
		Params:           stripe.Params{Context: ctx},
		Customer:         stripe.String(params.CustomerProviderID),
		AutoAdvance:      stripe.Bool(params.AutoAdvance),
		CollectionMethod: stripe.String(params.CollectionMethod),
		Description:      stripe.String(params.Description),
		Currency:         stripe.String(params.Currency),
		Metadata:         params.Metadata,
		AutomaticTax: &stripe.InvoiceAutomaticTaxParams{
			Enabled: stripe.Bool(params.AutoTax),
		},
		PendingInvoiceItemsBehavior: stripe.String("include"),
	}
	if params.DaysUntilDue != nil {
		sp.DaysUntilDue = params.DaysUntilDue
	}
	if len(params.PaymentMethodTypes) > 0 {
		pmt := make([]*string, len(params.PaymentMethodTypes))
		for i, t := range params.PaymentMethodTypes {
			pmt[i] = stripe.String(t)
		}
		sp.PaymentSettings = &stripe.InvoicePaymentSettingsParams{
			PaymentMethodTypes: pmt,
			PaymentMethodOptions: &stripe.InvoicePaymentSettingsPaymentMethodOptionsParams{
				CustomerBalance: &stripe.InvoicePaymentSettingsPaymentMethodOptionsCustomerBalanceParams{
					FundingType: stripe.String(string(stripe.InvoicePaymentSettingsPaymentMethodOptionsCustomerBalanceFundingTypeBankTransfer)),
					BankTransfer: &stripe.InvoicePaymentSettingsPaymentMethodOptionsCustomerBalanceBankTransferParams{
						Type: stripe.String("us_bank_transfer"),
					},
				},
			},
		}
	}
	si, err := p.client.Invoices.New(sp)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}
	return fromStripeInvoice(si), nil
}

func (p *Provider) CreateInvoiceItem(ctx context.Context, params billing.CreateInvoiceItemParams) error {
	sp := &stripe.InvoiceItemParams{
		Params:      stripe.Params{Context: ctx},
		Customer:    stripe.String(params.CustomerProviderID),
		Currency:    stripe.String(params.Currency),
		Invoice:     stripe.String(params.InvoiceProviderID),
		UnitAmount:  &params.UnitAmount,
		Quantity:    &params.Quantity,
		Metadata:    params.Metadata,
		Description: stripe.String(params.Description),
	}
	if params.PeriodStart != nil && params.PeriodEnd != nil {
		sp.Period = &stripe.InvoiceItemPeriodParams{
			Start: params.PeriodStart,
			End:   params.PeriodEnd,
		}
	}
	_, err := p.client.InvoiceItems.New(sp)
	if err != nil {
		return fmt.Errorf("failed to create invoice item: %w", err)
	}
	return nil
}

func (p *Provider) GetInvoice(ctx context.Context, providerID string) (*billing.ProviderInvoice, error) {
	si, err := p.client.Invoices.Get(providerID, &stripe.InvoiceParams{
		Params: stripe.Params{Context: ctx},
		Expand: []*string{stripe.String("lines")},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	return fromStripeInvoice(si), nil
}

// --- Webhook ---

func (p *Provider) VerifyWebhook(payload []byte, signature string, secrets []string) (*billing.WebhookEvent, error) {
	var parseErrs []error
	var evt stripe.Event
	for _, secret := range secrets {
		var err error
		evt, err = webhook.ConstructEvent(payload, signature, secret)
		if err != nil {
			parseErrs = append(parseErrs, err)
			continue
		}
		parseErrs = nil
		break
	}
	if len(parseErrs) > 0 {
		return nil, fmt.Errorf("failed to construct event: %w", errors.Join(parseErrs...))
	}

	return &billing.WebhookEvent{
		Type:     mapStripeEventType(string(evt.Type)),
		ObjectID: evt.GetObjectValue("id"),
	}, nil
}

// mapStripeEventType converts Stripe event types to provider-neutral constants.
func mapStripeEventType(t string) string {
	switch t {
	case string(stripe.EventTypeCheckoutSessionCompleted):
		return billing.EventCheckoutCompleted
	case string(stripe.EventTypeCheckoutSessionAsyncPaymentSucceeded):
		return billing.EventCheckoutPaymentSucceeded
	case string(stripe.EventTypeCustomerCreated):
		return billing.EventCustomerCreated
	case string(stripe.EventTypeCustomerUpdated):
		return billing.EventCustomerUpdated
	case string(stripe.EventTypeCustomerSourceCreated):
		return billing.EventCustomerSourceCreated
	case string(stripe.EventTypeCustomerSourceUpdated):
		return billing.EventCustomerSourceUpdated
	case string(stripe.EventTypeCustomerSubscriptionCreated):
		return billing.EventSubscriptionCreated
	case string(stripe.EventTypeCustomerSubscriptionUpdated):
		return billing.EventSubscriptionUpdated
	case string(stripe.EventTypeCustomerSubscriptionDeleted):
		return billing.EventSubscriptionDeleted
	case string(stripe.EventTypeInvoicePaid):
		return billing.EventInvoicePaid
	default:
		return t
	}
}
