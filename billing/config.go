package billing

type Config struct {
	StripeKey string `yaml:"stripe_key" mapstructure:"stripe_key"`
	// PlansPath is a directory path where plans are defined
	PlansPath string `yaml:"plans_path" mapstructure:"plans_path"`
}
