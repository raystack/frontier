package billing

// Provider-neutral types for billing operations.
// These types carry exactly the data that billing services need,
// without coupling to any specific provider's SDK.

// --- Customer types ---

type ProviderCustomer struct {
	ID       string
	Deleted  bool
	Name     string
	Email    string
	Phone    string
	Currency string
	Address  ProviderAddress
	TaxIDs   []ProviderTaxID
}

type ProviderAddress struct {
	City       string
	Country    string
	Line1      string
	Line2      string
	PostalCode string
	State      string
}

type ProviderTaxID struct {
	Type  string
	Value string
}

type CreateCustomerParams struct {
	Email    string
	Name     string
	Phone    string
	Address  ProviderAddress
	TaxIDs   []ProviderTaxID
	Metadata map[string]string
	// TestClockID is used for provider testing (e.g., Stripe test clocks).
	// May be ignored by providers that don't support it.
	TestClockID *string
}

type UpdateCustomerParams struct {
	Email    string
	Name     string
	Phone    string
	Address  ProviderAddress
	Metadata map[string]string
}

type ProviderPaymentMethod struct {
	ID              string
	Type            string
	CardBrand       string
	CardLast4       string
	CardExpiryMonth int64
	CardExpiryYear  int64
	IsDefault       bool
	Metadata        map[string]any
	CreatedAt       int64
}

// --- Product and price types ---

type CreateProductParams struct {
	ID          string
	Name        string
	Description string
	Metadata    map[string]string
}

type UpdateProductParams struct {
	Name        string
	Description string
	Metadata    map[string]string
}

type CreatePriceParams struct {
	ProductID       string
	Name            string
	Amount          int64
	Currency        string
	BillingScheme   string
	Interval        string
	UsageType       string
	MeteredAggregate string
	Metadata        map[string]string
}

type UpdatePriceParams struct {
	Name     string
	Metadata map[string]string
}

// --- Subscription types ---

type ProviderSubscription struct {
	ID                   string
	Status               string
	CanceledAt           int64
	EndedAt              int64
	TrialEnd             int64
	CurrentPeriodStart   int64
	CurrentPeriodEnd     int64
	BillingCycleAnchor   int64
	Livemode             bool
	AutomaticTaxEnabled  bool
	Items                []ProviderSubscriptionItem
	Schedule             *ProviderScheduleRef
	Metadata             map[string]string
}

// ProviderScheduleRef is the schedule info embedded in a subscription response.
type ProviderScheduleRef struct {
	ID           string
	CurrentPhase *ProviderCurrentPhase
}

type ProviderCurrentPhase struct {
	StartDate int64
	EndDate   int64
}

type ProviderSubscriptionItem struct {
	ID       string
	PriceID  string
	Quantity int64
	Metadata map[string]string
	// Price details needed for plan resolution
	ProductID string
	Interval  string
}

type CreateSubscriptionParams struct {
	CustomerProviderID string
	Currency           string
	Items              []SubscriptionItemInput
	Metadata           map[string]string
	TrialDays          *int64
	AutoTax            bool
	CancelAtTrialEnd   bool
	CouponID           string
}

type SubscriptionItemInput struct {
	PriceProviderID string
	Quantity        int64
	Metadata        map[string]string
}

type CancelSubscriptionParams struct {
	InvoiceNow bool
	Prorate    bool
}

type UpdateSubscriptionItemsParams struct {
	Items                      []SubscriptionItemUpdate
	PendingInvoiceItemInterval *InvoiceItemInterval
}

type SubscriptionItemUpdate struct {
	ID       string
	PriceID  string
	Quantity int64
	Metadata map[string]string
}

type InvoiceItemInterval struct {
	Interval      string
	IntervalCount int64
}

// --- Schedule types ---

type ProviderSchedule struct {
	ID           string
	CurrentPhase *ProviderCurrentPhase
	EndBehavior  string
	Phases       []ProviderPhase
}

type ProviderPhase struct {
	StartDate           int64
	EndDate             int64
	Currency            string
	Description         string
	TrialEnd            int64
	ProrationBehavior   string
	CollectionMethod    string
	AutomaticTaxEnabled bool
	Metadata            map[string]string
	Items               []ProviderPhaseItem
}

type ProviderPhaseItem struct {
	PriceID   string
	ProductID string
	Quantity  int64
	Metadata  map[string]string
}

type UpdateScheduleParams struct {
	Phases            []SchedulePhaseInput
	EndBehavior       string
	ProrationBehavior string
	CollectionMethod  string
}

type SchedulePhaseInput struct {
	Items               []SchedulePhaseItemInput
	Currency            string
	StartDate           *int64
	EndDate             *int64
	EndDateNow          bool
	Iterations          *int64
	Metadata            map[string]string
	AutoTax             bool
	Description         string
	TrialEnd            *int64
	ProrationBehavior   string
	CollectionMethod    string
}

type SchedulePhaseItemInput struct {
	PriceID  string
	Quantity int64
	Metadata map[string]string
}

// --- Checkout types ---

type ProviderCheckoutSession struct {
	ID            string
	URL           string
	Status        string
	PaymentStatus string
	ExpiresAt     int64
	AmountTotal   int64
	Currency      string
	SubscriptionID string
	LineItems      []ProviderCheckoutLineItem
}

type ProviderCheckoutLineItem struct {
	ProductID string
	Quantity  int64
}

type CreateCheckoutSessionParams struct {
	CustomerProviderID string
	Currency           string
	Mode               string // "subscription", "payment", "setup"
	SuccessURL         string
	CancelURL          string
	LineItems          []CheckoutLineItemInput
	Metadata           map[string]string
	AutoTax            bool
	ExpiresAt          int64

	// Subscription-specific
	SubscriptionMetadata map[string]string
	TrialDays            *int64
	CancelAtTrialEnd     bool

	// Payment-specific
	InvoiceCreation  bool
	PaymentMethodTypes []string

	// Address collection
	AddressCollection string // "auto" or "never"

	// Promotion
	AllowPromotionCodes bool

	// Payment method
	PaymentMethodCollection string
}

type CheckoutLineItemInput struct {
	PriceProviderID    string
	Quantity           int64
	AdjustableQuantity *AdjustableQuantity
}

type AdjustableQuantity struct {
	Enabled bool
	Minimum int64
	Maximum int64
}

type CreateBillingPortalParams struct {
	CustomerProviderID string
	ReturnURL          string
}

// --- Invoice types ---

type ProviderInvoice struct {
	ID                 string
	Status             string
	EffectiveAt        int64
	HostedURL          string
	Total              int64
	Currency           string
	CreatedAt          int64
	DueDate            int64
	NextPaymentAttempt int64
	PeriodStart        int64
	PeriodEnd          int64
	Metadata           map[string]string
	LineItems          []ProviderInvoiceLineItem
}

type ProviderInvoiceLineItem struct {
	ID          string
	Description string
	Quantity    int64
	UnitAmount  int64
	PeriodStart int64
	PeriodEnd   int64
	Metadata    map[string]string
}

type CreateInvoiceParams struct {
	CustomerProviderID string
	AutoAdvance        bool
	DaysUntilDue       *int64
	CollectionMethod   string
	Description        string
	AutoTax            bool
	Currency           string
	Metadata           map[string]string
	PaymentMethodTypes []string
}

type CreateInvoiceItemParams struct {
	CustomerProviderID string
	InvoiceProviderID  string
	Currency           string
	UnitAmount         int64
	Quantity           int64
	Description        string
	Metadata           map[string]string
	PeriodStart        *int64
	PeriodEnd          *int64
}

// --- Webhook types ---

type WebhookEvent struct {
	Type       string
	ObjectID   string
}

// Well-known webhook event types (provider-neutral).
const (
	EventCheckoutCompleted        = "checkout.completed"
	EventCheckoutPaymentSucceeded = "checkout.async_payment_succeeded"
	EventCustomerCreated          = "customer.created"
	EventCustomerUpdated          = "customer.updated"
	EventCustomerSourceCreated    = "customer.source.created"
	EventCustomerSourceUpdated    = "customer.source.updated"
	EventSubscriptionCreated      = "subscription.created"
	EventSubscriptionUpdated      = "subscription.updated"
	EventSubscriptionDeleted      = "subscription.deleted"
	EventInvoicePaid              = "invoice.paid"
)

// --- Error types ---

// ErrNotFoundInProvider signals that the resource was not found at the provider.
// Services use this to handle "already deleted" or "not yet created" cases.
var ErrNotFoundInProvider = errNotFoundInProvider{}

type errNotFoundInProvider struct{}

func (e errNotFoundInProvider) Error() string {
	return "resource not found in billing provider"
}

// ErrNoUpcomingInvoice signals that there is no upcoming invoice for the customer.
var ErrNoUpcomingInvoice = errNoUpcomingInvoice{}

type errNoUpcomingInvoice struct{}

func (e errNoUpcomingInvoice) Error() string {
	return "no upcoming invoice"
}
