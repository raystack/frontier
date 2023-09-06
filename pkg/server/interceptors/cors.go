package interceptors

import (
	"net/http"

	"github.com/gorilla/handlers"
)

type CorsConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins" mapstructure:"allowed_origins" default:""`
	AllowedMethods []string `yaml:"allowed_methods" mapstructure:"allowed_methods" default:"GET POST PUT DELETE PATCH"`
	AllowedHeaders []string `yaml:"allowed_headers" mapstructure:"allowed_headers" default:"Authorization"`
	ExposedHeaders []string `yaml:"exposed_headers" mapstructure:"exposed_headers" default:"Content-Type"`
}

func WithCors(h http.Handler, conf CorsConfig) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins(conf.AllowedOrigins),
		handlers.AllowedMethods(conf.AllowedMethods),
		handlers.AllowedHeaders(conf.AllowedHeaders),
		handlers.ExposedHeaders(conf.ExposedHeaders),
		handlers.AllowCredentials(),
	)(h)
}
