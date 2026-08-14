package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/pkg/errors"
)

// connectPanicRecovery converts a panic anywhere in the RPC handler chain
// into a CodeInternal error response. The panic value and stack go to the
// server log only, never to the caller.
func connectPanicRecovery(logger *slog.Logger) connect.HandlerOption {
	return connect.WithRecover(func(ctx context.Context, spec connect.Spec, _ http.Header, panicValue any) error {
		logger.ErrorContext(ctx, "rpc handler panic",
			"procedure", spec.Procedure,
			"panic", panicValue,
			"stack", string(debug.Stack()))
		return connect.NewError(connect.CodeInternal, errors.ErrInternalServerError)
	})
}

// committedWriter tracks whether the response status line has gone out, so
// the recovery handler knows whether a 500 can still be sent.
type committedWriter struct {
	http.ResponseWriter
	committed bool
}

func (c *committedWriter) WriteHeader(statusCode int) {
	c.committed = true
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *committedWriter) Write(b []byte) (int, error) {
	c.committed = true
	return c.ResponseWriter.Write(b)
}

// Flush commits the response too. The connect reverse proxy on the UI server
// needs the Flusher interface for streaming.
func (c *committedWriter) Flush() {
	c.committed = true
	if f, ok := c.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap lets http.ResponseController reach the underlying writer's
// Hijacker, deadline, and full-duplex methods.
func (c *committedWriter) Unwrap() http.ResponseWriter {
	return c.ResponseWriter
}

// ReadFrom keeps the underlying writer's optimized copy path (pooled
// buffers, sendfile for OS files), which io.Copy looks for on the
// destination directly without walking Unwrap.
func (c *committedWriter) ReadFrom(r io.Reader) (int64, error) {
	c.committed = true
	if rf, ok := c.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}
	return io.Copy(c.ResponseWriter, r)
}

// httpPanicRecovery responds with a plain 500 instead of dropping the
// connection when a handler behind the wrapped mux panics. If the handler
// already wrote part of a response, the status can no longer be changed, so
// it aborts the connection instead of appending an error to the partial body.
func httpPanicRecovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cw := &committedWriter{ResponseWriter: w}
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
				if cw.committed {
					panic(http.ErrAbortHandler)
				}
				http.Error(cw, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(cw, r)
	})
}
