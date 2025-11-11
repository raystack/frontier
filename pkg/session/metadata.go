package session

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/authenticate/session"
	"github.com/ua-parser/uap-go/uaparser"
)

// ExtractSessionMetadata extracts session metadata from HTTP headers
func ExtractSessionMetadata(ctx context.Context, req connect.AnyRequest, headers authenticate.SessionMetadataHeaders) session.SessionMetadata {
	metadata := session.SessionMetadata{}

	// IP Address
	if clientIP := req.Header().Get(headers.ClientIP); clientIP != "" {
		if parts := strings.Split(clientIP, ","); len(parts) > 0 {
			metadata.IpAddress = parts[0]
		}
	}

	if country := req.Header().Get(headers.ClientCountry); country != "" {
		metadata.Location.Country = country
	}
	if city := req.Header().Get(headers.ClientCity); city != "" {
		metadata.Location.City = city
	}
	if latitude := req.Header().Get(headers.ClientLatitude); latitude != "" {
		metadata.Location.Latitude = latitude
	}
	if longitude := req.Header().Get(headers.ClientLongitude); longitude != "" {
		metadata.Location.Longitude = longitude
	}

	// OS and Browser (from User-Agent) using uap-go library
	userAgent := req.Header().Get(headers.ClientUserAgent)
	if userAgent != "" {
		parser := uaparser.NewFromSaved()
		client := parser.Parse(userAgent)

		metadata.OperatingSystem = client.Os.Family
		metadata.Browser = client.UserAgent.Family
	}

	return metadata
}
