package billing

import "time"

type Config struct {
	StripeKey            string   `yaml:"stripe_key" mapstructure:"stripe_key"`
	StripeAutoTax        bool     `yaml:"stripe_auto_tax" mapstructure:"stripe_auto_tax"`
	StripeWebhookSecrets []string `yaml:"stripe_webhook_secrets" mapstructure:"stripe_webhook_secrets"`
	// PlansPath is a directory path where plans are defined
	PlansPath       string `yaml:"plans_path" mapstructure:"plans_path"`
	DefaultCurrency string `yaml:"default_currency" mapstructure:"default_currency"`

	AccountConfig      AccountConfig      `yaml:"customer" mapstructure:"customer"`
	PlanChangeConfig   PlanChangeConfig   `yaml:"plan_change" mapstructure:"plan_change"`
	SubscriptionConfig SubscriptionConfig `yaml:"subscription" mapstructure:"subscription"`
	ProductConfig      ProductConfig      `yaml:"product" mapstructure:"product"`

	RefreshInterval RefreshInterval `yaml:"refresh_interval" mapstructure:"refresh_interval"`
}

type RefreshInterval struct {
	Customer     time.Duration `yaml:"customer" mapstructure:"customer" default:"1m"`
	Subscription time.Duration `yaml:"subscription" mapstructure:"subscription" default:"1m"`
	Checkout     time.Duration `yaml:"checkout" mapstructure:"checkout" default:"1m"`
	Invoice      time.Duration `yaml:"invoice" mapstructure:"invoice" default:"5m"`
}

type AccountConfig struct {
	AutoCreateWithOrg     bool   `yaml:"auto_create_with_org" mapstructure:"auto_create_with_org"`
	DefaultPlan           string `yaml:"default_plan" mapstructure:"default_plan"`
	DefaultOffline        bool   `yaml:"default_offline" mapstructure:"default_offline"`
	OnboardCreditsWithOrg int64  `yaml:"onboard_credits_with_org" mapstructure:"onboard_credits_with_org"`

	// CreditOverdraftProduct helps identify the product pricing per unit amount for the overdraft
	// credits being invoiced
	CreditOverdraftProduct string `yaml:"credit_overdraft_product" mapstructure:"credit_overdraft_product"`

	// CreditOverdraftInvoiceDay is the day of the range(month) when the overdraft credits are invoiced
	CreditOverdraftInvoiceDay int `yaml:"credit_overdraft_invoice_day" mapstructure:"credit_overdraft_invoice_day" default:"1"`

	// CreditOverdraftInvoiceRangeShift is the shift in the invoice range for the overdraft credits
	// if positive, the invoice range will be shifted to the future else it will be shifted to the past
	CreditOverdraftInvoiceRangeShift int `yaml:"credit_overdraft_invoice_range_shift" mapstructure:"credit_overdraft_invoice_range_shift" default:"0"`
}

type PlanChangeConfig struct {
	// ProrationBehavior is the behavior of proration when a plan is changed
	// possible values: create_prorations, none, always_invoice
	ProrationBehavior          string `yaml:"proration_behavior" mapstructure:"proration_behavior" default:"create_prorations"`
	ImmediateProrationBehavior string `yaml:"immediate_proration_behavior" mapstructure:"immediate_proration_behavior" default:"create_prorations"`

	// CollectionMethod is the behavior of collection method when a plan is changed
	// possible values: charge_automatically, send_invoice
	CollectionMethod string `yaml:"collection_method" mapstructure:"collection_method" default:"charge_automatically"`
}

type SubscriptionConfig struct {
	// BehaviorAfterTrial as `cancel` will cancel the subscription after trial end and will not generate invoice.
	BehaviorAfterTrial string `yaml:"behavior_after_trial" mapstructure:"behavior_after_trial" default:"release"`
}

type ProductConfig struct {
	// SeatChangeBehavior is the behavior of changes in product per seat cost when number of users
	// in the subscription org changes
	// possible values: exact, incremental
	// "exact" will change the seat count to the exact number of users within the organization
	// "incremental" will change the seat count to the number of users within the organization
	// but will not decrease the seat count if reduced
	SeatChangeBehavior string `yaml:"seat_change_behavior" mapstructure:"seat_change_behavior" default:"exact"`
}
