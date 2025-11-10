package connectinterceptors

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
)

type CorsConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins" mapstructure:"allowed_origins" default:""`
	AllowedMethods []string `yaml:"allowed_methods" mapstructure:"allowed_methods" default:"GET POST PUT DELETE PATCH"`
	AllowedHeaders []string `yaml:"allowed_headers" mapstructure:"allowed_headers" default:"Authorization"`
	ExposedHeaders []string `yaml:"exposed_headers" mapstructure:"exposed_headers" default:"Content-Type"`
	MaxAge         int      `yaml:"max_age" mapstructure:"max_age" default:"7200"`
}

// WithConnectCORS adds CORS support to a Connect HTTP handler.
func WithConnectCORS(connectHandler http.Handler, conf CorsConfig) http.Handler {
	// If no origins are configured, return handler without CORS
	if len(conf.AllowedOrigins) == 0 {
		return connectHandler
	}

	c := cors.New(cors.Options{
		AllowedOrigins: conf.AllowedOrigins,
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: connectcors.AllowedHeaders(),
		ExposedHeaders: connectcors.ExposedHeaders(),
		MaxAge:         conf.MaxAge,
	})
	return c.Handler(connectHandler)
}
