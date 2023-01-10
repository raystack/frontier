package obeservability

import (
	"context"
	"net/http"
	"strings"

	"github.com/odpf/shield/internal/proxy/middleware"

	"github.com/odpf/salt/log"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

const (
	headerRequestID     = "X-Request-Id"
	requestIDContextKey = "requestIDContextKey"
)

type Ware struct {
	log  *log.Zap
	next http.Handler
}

func New(log *log.Zap, next http.Handler) *Ware {
	return &Ware{
		log:  log,
		next: next,
	}
}

func (m Ware) Info() *middleware.MiddlewareInfo {
	return &middleware.MiddlewareInfo{
		Name:        "_observability",
		Description: "to handle observability",
	}
}

func (m *Ware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	reqID := GetIncomingRequestID(req)
	ctx := SetRequestID(req.Context(), reqID)
	ctx = m.log.NewContext(req.Context())
	ctx = log.ZapContextWithFields(ctx,
		zap.String("host", req.Host),
		zap.String("path", req.URL.String()),
		zap.String("method", req.Method),
		zap.String("remote_address", req.RemoteAddr),
		zap.String("scheme", req.Proto),
		zap.String("request_id", reqID),
	)
	req = req.WithContext(ctx)
	m.next.ServeHTTP(rw, req)
}

func GetIncomingRequestID(req *http.Request) string {
	reqID := strings.TrimSpace(req.Header.Get(headerRequestID))
	if reqID == "" {
		reqID = xid.New().String()
	}
	req.Header.Set(headerRequestID, reqID)

	return reqID
}

func SetRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey, reqID)
}

func GetRequestID(ctx context.Context) string {
	reqID, ok := ctx.Value(requestIDContextKey).(string)
	if !ok {
		return ""
	}
	return reqID
}
