package spicedb

type Config struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port" default:"50051"`
	PreSharedKey string `yaml:"pre_shared_key" mapstructure:"pre_shared_key"`

	// Deprecated: Use Consistency instead
	// FullyConsistent ensures APIs although slower than usual will result in responses always most consistent
	FullyConsistent bool `yaml:"fully_consistent" mapstructure:"fully_consistent" default:"false"`

	// Consistency ensures Authz server consistency guarantees for various operations
	// Possible values are:
	// - "full": Guarantees that the data is always fresh
	// - "best_effort": Guarantees that the data is the best effort fresh
	// - "minimize_latency": Tries to prioritise minimal latency
	Consistency string `yaml:"consistency" mapstructure:"consistency" default:"best_effort"`

	// CheckTrace enables tracing in check api for spicedb, it adds considerable
	// latency to the check calls and shouldn't be enabled in production
	CheckTrace bool `yaml:"check_trace" mapstructure:"check_trace" default:"false"`
}
