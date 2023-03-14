package telemetry

type Config struct {
	// OpenCensus exporter configurations.
	ServiceName string `yaml:"service_name" mapstructure:"service_name"`
}
