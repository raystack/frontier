package connectinterceptors

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/pkg/server/consts"
)

type SessionMetadataInterceptor struct{}

func NewSessionMetadataInterceptor() connect.Interceptor {
	return &SessionMetadataInterceptor{}
}

func (s *SessionMetadataInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (s *SessionMetadataInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

func (s *SessionMetadataInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		metadata := make(map[string]any)

		// Get User-Agent
		userAgent := req.Header().Get("User-Agent")
		if userAgent != "" {
			metadata["browser"] = extractBrowser(userAgent)
			metadata["operating_system"] = extractOS(userAgent)
		}

		// Get IP Address
		ip := req.Header().Get("X-Forwarded-For")
		if ip == "" {
			ip = req.Header().Get("X-Real-IP")
		}
		if ip != "" {
			metadata["ip_address"] = strings.Split(ip, ",")[0] // Get first IP if multiple
		}

		ctx = consts.WithSessionMetadata(ctx, metadata)
		return next(ctx, req)
	})
}

func extractBrowser(userAgent string) string {
	switch {
	case strings.Contains(userAgent, "Chrome"):
		return "Chrome"
	case strings.Contains(userAgent, "Firefox"):
		return "Firefox"
	case strings.Contains(userAgent, "Safari"):
		return "Safari"
	default:
		return "Unknown"
	}
}

func extractOS(userAgent string) string {
	switch {
	case strings.Contains(userAgent, "Windows"):
		return "Windows"
	case strings.Contains(userAgent, "Mac OS"):
		return "Mac OS"
	case strings.Contains(userAgent, "Linux"):
		return "Linux"
	case strings.Contains(userAgent, "iPhone"):
		return "iOS"
	case strings.Contains(userAgent, "Android"):
		return "Android"
	default:
		return "Unknown"
	}
}
