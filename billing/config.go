package billing

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
}

type AccountConfig struct {
	AutoCreateWithOrg     bool   `yaml:"auto_create_with_org" mapstructure:"auto_create_with_org"`
	DefaultPlan           string `yaml:"default_plan" mapstructure:"default_plan"`
	DefaultOffline        bool   `yaml:"default_offline" mapstructure:"default_offline"`
	OnboardCreditsWithOrg int64  `yaml:"onboard_credits_with_org" mapstructure:"onboard_credits_with_org"`
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
