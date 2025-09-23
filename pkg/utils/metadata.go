package utils

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/authenticate/session"
	"github.com/ua-parser/uap-go/uaparser"
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
			metadata.IpAddress = parts[0]
		}
	}

	if country := req.Header().Get(config.ClientCountry); country != "" {
		metadata.Location.Country = country
	}
	if city := req.Header().Get(config.ClientCity); city != "" {
		metadata.Location.City = city
	}

	// OS and Browser (from User-Agent) using uap-go library
	userAgent := req.Header().Get("User-Agent")
	if userAgent != "" {
		parser := uaparser.NewFromSaved()
		client := parser.Parse(userAgent)

		metadata.OperatingSystem = client.Os.Family
		metadata.Browser = client.UserAgent.Family
	}

	return metadata
}
