package metaschema

import "time"

// Config holds runtime configuration for the metaschema service.
type Config struct {
	// RefreshInterval is how often each server reloads the metaschema cache from
	// the database, so a change made on one pod reaches the others. A value of 0
	// disables the background refresh; the cache is still primed once at startup.
	RefreshInterval time.Duration `yaml:"refresh_interval" mapstructure:"refresh_interval" default:"1m"`
}
