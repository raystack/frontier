package cmd

import (
	"strings"

	"connectrpc.com/connect"
)

// newRequest creates a connect.Request and optionally sets a header from a "key:value" string.
func newRequest[T any](msg *T, header string) *connect.Request[T] {
	req := connect.NewRequest(msg)
	if header != "" {
		if k, v, ok := parseHeader(header); ok {
			req.Header().Set(k, v)
		}
	}
	return req
}

// parseHeader splits a "key:value" header string into key and value.
func parseHeader(header string) (string, string, bool) {
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}
