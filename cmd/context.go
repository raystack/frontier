package cmd

import (
	"fmt"
	"strings"

	"connectrpc.com/connect"
)

// newRequest creates a connect.Request and sets a header from a "key:value" string.
func newRequest[T any](msg *T, header string) (*connect.Request[T], error) {
	req := connect.NewRequest(msg)
	if header != "" {
		k, v, err := parseHeader(header)
		if err != nil {
			return nil, err
		}
		req.Header().Set(k, v)
	}
	return req, nil
}

// parseHeader splits a "key:value" header string into key and value.
func parseHeader(header string) (string, string, error) {
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid header format %q: expected key:value", header)
	}
	return parts[0], parts[1], nil
}
