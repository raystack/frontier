package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/raystack/frontier/pkg/server/consts"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	frontierv1beta1connect "github.com/raystack/frontier/proto/v1beta1/frontierv1beta1connect"
)

// WebhookBridgeHandler creates an HTTP handler that bridges raw HTTP webhook requests
// to the ConnectRPC BillingWebhookCallback handler. This is needed because Stripe
// doesn't allow modifying the request body but allows custom URL paths.
// The handler uses the frontierHandler which has all interceptors (auth, logging, audit, etc.) applied.
func WebhookBridgeHandler(frontierHandler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract provider from URL path
		// Expected path: /billing/webhooks/callback/{provider}
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(pathParts) < 4 || pathParts[0] != "billing" ||
			pathParts[1] != "webhooks" || pathParts[2] != "callback" {
			http.Error(w, "invalid path", http.StatusNotFound)
			return
		}
		provider := pathParts[3]
		if provider == "" {
			http.Error(w, "invalid path", http.StatusNotFound)
			return
		}

		// Read raw request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Create ConnectRPC request payload
		requestPayload := &frontierv1beta1.BillingWebhookCallbackRequest{
			Provider: provider,
			Body:     body,
		}

		// Encode request as JSON for ConnectRPC
		requestJSON, err := json.Marshal(requestPayload)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to encode request: %v", err), http.StatusInternalServerError)
			return
		}

		// Create a new HTTP request to the ConnectRPC procedure
		connectReq := httptest.NewRequest(
			http.MethodPost,
			frontierv1beta1connect.FrontierServiceBillingWebhookCallbackProcedure,
			bytes.NewReader(requestJSON),
		)
		connectReq = connectReq.WithContext(r.Context())

		// Copy important headers from the original request
		if contentType := r.Header.Get("Content-Type"); contentType != "" {
			connectReq.Header.Set("X-Original-Content-Type", contentType)
		}
		// Set ConnectRPC content type
		connectReq.Header.Set("Content-Type", "application/json")

		// Copy other important headers (auth, request ID, etc.)
		headersToProxy := []string{
			"Authorization",
			"Cookie",
			consts.RequestIDHeader,
			consts.ProjectRequestKey,
			consts.StripeTestClockRequestKey,
			consts.StripeWebhookSignature,
		}
		for _, header := range headersToProxy {
			if value := r.Header.Get(header); value != "" {
				connectReq.Header.Set(header, value)
			}
		}

		// Forward to the ConnectRPC handler (which has all interceptors)
		frontierHandler.ServeHTTP(w, connectReq)
	}
}
