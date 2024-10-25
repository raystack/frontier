package spicedb

type Config struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port" default:"50051"`
	PreSharedKey string `yaml:"pre_shared_key" mapstructure:"pre_shared_key"`

	// FullyConsistent ensures APIs although slower than usual will result in responses always most consistent
	FullyConsistent bool `yaml:"fully_consistent" mapstructure:"fully_consistent" default:"false"`

	// CheckTrace enables tracing in check api for spicedb, it adds considerable
	// latency to the check calls and shouldn't be enabled in production
	CheckTrace bool `yaml:"check_trace" mapstructure:"check_trace" default:"false"`
}
