package interceptors

import (
	"net/http"

	"github.com/gorilla/handlers"
)

func WithCors(h http.Handler, origin string) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins([]string{origin}),
		handlers.AllowedMethods([]string{"GET", "HEAD", "POST"}),
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type", "X-Requested-With"}),
		handlers.AllowCredentials(),
	)(h)
}
