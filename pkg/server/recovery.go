package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"

	"connectrpc.com/connect"
)

var errInternalServer = errors.New("internal server error")

// connectPanicRecovery converts a panic anywhere in the RPC handler chain
// into a CodeInternal error response. The panic value and stack go to the
// server log only, never to the caller.
func connectPanicRecovery(logger *slog.Logger) connect.HandlerOption {
	return connect.WithRecover(func(ctx context.Context, spec connect.Spec, _ http.Header, panicValue any) error {
		logger.ErrorContext(ctx, "rpc handler panic",
			"procedure", spec.Procedure,
			"panic", panicValue,
			"stack", string(debug.Stack()))
		return connect.NewError(connect.CodeInternal, errInternalServer)
	})
}

// uiPanicRecovery responds with a plain 500 instead of dropping the
// connection when a handler behind the UI mux panics.
func uiPanicRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if panicValue := recover(); panicValue != nil {
				// net/http checks for ErrAbortHandler with ==, so we should too.
				if panicValue == http.ErrAbortHandler {
					panic(panicValue)
				}
				logger.ErrorContext(r.Context(), "http handler panic",
					"path", r.URL.Path,
					"panic", panicValue,
					"stack", string(debug.Stack()))
				http.Error(w, errInternalServer.Error(), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
