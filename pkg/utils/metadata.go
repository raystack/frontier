package utils

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/authenticate/session"
)

// SessionMetadataConfig holds configuration for header names
type SessionMetadataConfig struct {
	ClientIP      string
	ClientCountry string
	ClientCity    string
}

// ExtractSessionMetadata extracts session metadata from HTTP headers
func ExtractSessionMetadata(ctx context.Context, req connect.AnyRequest, config SessionMetadataConfig) session.SessionMetadata {
	metadata := session.SessionMetadata{}

	// IP Address
	if clientIP := req.Header().Get(config.ClientIP); clientIP != "" {
		if parts := strings.Split(clientIP, ":"); len(parts) > 0 {
			metadata.IP = parts[0]
		}
	}

	if country := req.Header().Get(config.ClientCountry); country != "" {
		metadata.Location.Country = country
	}
	if city := req.Header().Get(config.ClientCity); city != "" {
		metadata.Location.City = city
	}

	// OS and Browser (from User-Agent)
	userAgent := req.Header().Get("User-Agent")
	if userAgent != "" {
		metadata.OS = extractOS(userAgent)
		metadata.Browser = extractBrowser(userAgent)
	}

	return metadata
}

func extractBrowser(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	switch {
	// Chrome
	case strings.Contains(userAgent, "chrome/") && !strings.Contains(userAgent, "edg/") && !strings.Contains(userAgent, "opr/") && !strings.Contains(userAgent, "crios/") && !strings.Contains(userAgent, "whale/"):
		return "Chrome"
	// Safari
	case strings.Contains(userAgent, "safari/") && !strings.Contains(userAgent, "chrome/") && !strings.Contains(userAgent, "crios/") && !strings.Contains(userAgent, "version/"):
		return "Safari"
	// Edge
	case strings.Contains(userAgent, "edg/"):
		return "Edge"
	// Firefox
	case strings.Contains(userAgent, "firefox/"):
		return "Firefox"
	// Safari (in-app)
	case strings.Contains(userAgent, "safari/") && strings.Contains(userAgent, "version/") && !strings.Contains(userAgent, "chrome/"):
		return "Safari (in-app)"
	// Android Webview
	case strings.Contains(userAgent, "wv") && strings.Contains(userAgent, "android"):
		return "Android Webview"
	// Opera
	case strings.Contains(userAgent, "opr/") || strings.Contains(userAgent, "opera/"):
		return "Opera"
	// Samsung Internet
	case strings.Contains(userAgent, "samsungbrowser/"):
		return "Samsung Internet"
	// YaBrowser
	case strings.Contains(userAgent, "yabrowser/") || strings.Contains(userAgent, "yandex/"):
		return "YaBrowser"
	// Whale Browser
	case strings.Contains(userAgent, "whale/"):
		return "Whale Browser"
	// Mozilla Compatible Agent
	case strings.Contains(userAgent, "mozilla/") && !strings.Contains(userAgent, "firefox/") && !strings.Contains(userAgent, "chrome/") && !strings.Contains(userAgent, "safari/"):
		return "Mozilla Compatible Agent"
	// Android Browser
	case strings.Contains(userAgent, "android") && !strings.Contains(userAgent, "chrome/") && !strings.Contains(userAgent, "firefox/") && !strings.Contains(userAgent, "samsungbrowser/"):
		return "Android Browser"
	// Mozilla
	case strings.Contains(userAgent, "mozilla/") && !strings.Contains(userAgent, "firefox/") && !strings.Contains(userAgent, "chrome/") && !strings.Contains(userAgent, "safari/") && !strings.Contains(userAgent, "edg/"):
		return "Mozilla"
	// Brave
	case strings.Contains(userAgent, "brave/"):
		return "Brave"
	// Tor
	case strings.Contains(userAgent, "tor/"):
		return "Tor"
	// DuckDuckGo
	case strings.Contains(userAgent, "duckduckgo/"):
		return "DuckDuckGo"
	default:
		return "Unknown"
	}
}

func extractOS(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	switch {
	// Windows
	case strings.Contains(userAgent, "windows"):
		return "Windows"
	// Macintosh
	case strings.Contains(userAgent, "mac os") || strings.Contains(userAgent, "macos"):
		return "Macintosh"
	// Android
	case strings.Contains(userAgent, "android"):
		return "Android"
	// Linux
	case strings.Contains(userAgent, "linux"):
		return "Linux"
	// iOS
	case strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad"):
		return "iOS"
	// Chrome OS
	case strings.Contains(userAgent, "chrome os"):
		return "Chrome OS"
	// OpenBSD
	case strings.Contains(userAgent, "openbsd"):
		return "OpenBSD"
	// Smart TV
	case strings.Contains(userAgent, "smart-tv") || strings.Contains(userAgent, "smarttv") || strings.Contains(userAgent, "smart tv"):
		return "Smart TV"
	// PlayStation
	case strings.Contains(userAgent, "playstation"):
		return "PlayStation"
	default:
		return "Unknown"
	}
}
