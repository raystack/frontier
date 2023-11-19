package billing

type Config struct {
	StripeKey     string `yaml:"stripe_key" mapstructure:"stripe_key"`
	StripeAutoTax bool   `yaml:"stripe_auto_tax" mapstructure:"stripe_auto_tax"`
	// PlansPath is a directory path where plans are defined
	PlansPath string `yaml:"plans_path" mapstructure:"plans_path"`
}
