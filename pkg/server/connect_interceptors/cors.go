package connectinterceptors

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
)

type CorsConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins" mapstructure:"allowed_origins" default:""`
	MaxAge         int      `yaml:"max_age" mapstructure:"max_age" default:"7200"`
}

// WithConnectCORS adds CORS support to a Connect HTTP handler.
// It uses Connect-specific methods, headers, and exposed headers from the connectcors package.
func WithConnectCORS(connectHandler http.Handler, conf CorsConfig) http.Handler {
	// If no origins are configured, return handler without CORS
	if len(conf.AllowedOrigins) == 0 {
		return connectHandler
	}

	c := cors.New(cors.Options{
		AllowedOrigins: conf.AllowedOrigins,
		AllowedMethods: connectcors.AllowedMethods(),
		// Use wildcard for headers to support all Connect RPC headers
		// Connect can send various headers depending on the request type
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   connectcors.ExposedHeaders(),
		AllowCredentials: true,
		MaxAge:           conf.MaxAge,
		Debug:            false,
	})
	return c.Handler(connectHandler)
}
