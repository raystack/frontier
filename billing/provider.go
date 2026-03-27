package billing

import "context"

// Provider abstracts billing provider operations (Stripe, Polar, etc.).
// Each method maps to a logical billing API operation.
// Implementations translate between provider-specific types and the neutral
// types defined in this package.
type Provider interface {
	// Customer operations
	CreateCustomer(ctx context.Context, params CreateCustomerParams) (*ProviderCustomer, error)
	UpdateCustomer(ctx context.Context, providerID string, params UpdateCustomerParams) (*ProviderCustomer, error)
	DeleteCustomer(ctx context.Context, providerID string) error
	GetCustomer(ctx context.Context, providerID string) (*ProviderCustomer, error)
	ListPaymentMethods(ctx context.Context, customerProviderID string) ([]ProviderPaymentMethod, error)

	// Product and price catalog
	CreateProduct(ctx context.Context, params CreateProductParams) error
	UpdateProduct(ctx context.Context, providerID string, params UpdateProductParams) error
	CreatePrice(ctx context.Context, params CreatePriceParams) (providerID string, err error)
	UpdatePrice(ctx context.Context, providerID string, params UpdatePriceParams) error

	// Subscription lifecycle
	CreateSubscription(ctx context.Context, params CreateSubscriptionParams) (*ProviderSubscription, error)
	GetSubscription(ctx context.Context, providerID string) (*ProviderSubscription, error)
	CancelSubscription(ctx context.Context, providerID string, params CancelSubscriptionParams) (*ProviderSubscription, error)
	UpdateSubscriptionItems(ctx context.Context, providerID string, params UpdateSubscriptionItemsParams) error

	// Subscription scheduling
	GetSchedule(ctx context.Context, scheduleID string) (*ProviderSchedule, error)
	CreateScheduleFromSubscription(ctx context.Context, subscriptionProviderID string) (*ProviderSchedule, error)
	UpdateSchedule(ctx context.Context, scheduleID string, params UpdateScheduleParams) (*ProviderSchedule, error)

	// Checkout and billing portal
	CreateCheckoutSession(ctx context.Context, params CreateCheckoutSessionParams) (*ProviderCheckoutSession, error)
	GetCheckoutSession(ctx context.Context, providerID string) (*ProviderCheckoutSession, error)
	CreateBillingPortalSession(ctx context.Context, params CreateBillingPortalParams) (url string, err error)

	// Invoice management
	ListInvoices(ctx context.Context, customerProviderID string) ([]ProviderInvoice, error)
	GetUpcomingInvoice(ctx context.Context, customerProviderID string) (*ProviderInvoice, error)
	CreateInvoice(ctx context.Context, params CreateInvoiceParams) (*ProviderInvoice, error)
	CreateInvoiceItem(ctx context.Context, params CreateInvoiceItemParams) error
	GetInvoice(ctx context.Context, providerID string) (*ProviderInvoice, error)

	// Webhook verification
	VerifyWebhook(payload []byte, signature string, secrets []string) (*WebhookEvent, error)
}
