package server

import "net/http"

// uiSecurityHeaders are set on every UI server response. img-src limits the
// admin console to images from its own origin, data URLs, and https hosts.
var uiSecurityHeaders = map[string]string{
	"Content-Security-Policy": "img-src 'self' data: https:",
}

// withSecurityHeaders sets the given headers on every response.
func withSecurityHeaders(next http.Handler, headers map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, value := range headers {
			w.Header().Set(key, value)
		}
		next.ServeHTTP(w, r)
	})
}
