package connectinterceptors

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/pkg/server/consts"
)

type MetadataConfig struct {
	ViewerAddress string
	ViewerCountry string
	ViewerCity    string
}

type SessionMetadataInterceptor struct {
	config MetadataConfig
}

func NewSessionMetadataInterceptor(config MetadataConfig) connect.Interceptor {
	return &SessionMetadataInterceptor{
		config: config,
	}
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

		// IP Address
		if viewerAddress := req.Header().Get(s.config.ViewerAddress); viewerAddress != "" {
			if parts := strings.Split(viewerAddress, ":"); len(parts) > 0 {
				metadata["ip"] = parts[0]
			}
		}

		// Location
		location := make(map[string]string)
		if country := req.Header().Get(s.config.ViewerCountry); country != "" {
			location["country"] = country
		}
		if city := req.Header().Get(s.config.ViewerCity); city != "" {
			location["city"] = city
		}
		if len(location) > 0 {
			metadata["location"] = location
		}

		// OS and Browser (from User-Agent)
		userAgent := req.Header().Get("User-Agent")
		if userAgent != "" {
			metadata["os"] = extractOS(userAgent)
			metadata["browser"] = extractBrowser(userAgent)
		}

		ctx = consts.WithSessionMetadata(ctx, metadata)
		return next(ctx, req)
	})
}

func extractBrowser(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	switch {
	case strings.Contains(userAgent, "edg/"):
		return "Edge"
	case strings.Contains(userAgent, "chrome/") && !strings.Contains(userAgent, "edg/"):
		return "Chrome"
	case strings.Contains(userAgent, "firefox/"):
		return "Firefox"
	case strings.Contains(userAgent, "safari/") && !strings.Contains(userAgent, "chrome/"):
		return "Safari"
	case strings.Contains(userAgent, "opera/"):
		return "Opera"
	default:
		return "Unknown"
	}
}

func extractOS(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	switch {
	case strings.Contains(userAgent, "windows"):
		return "Windows"
	case strings.Contains(userAgent, "mac os") || strings.Contains(userAgent, "macos"):
		return "macOS"
	case strings.Contains(userAgent, "linux"):
		return "Linux"
	case strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad"):
		return "iOS"
	case strings.Contains(userAgent, "android"):
		return "Android"
	case strings.Contains(userAgent, "chrome os"):
		return "Chrome OS"
	default:
		return "Unknown"
	}
}
