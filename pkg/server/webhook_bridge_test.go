package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockHandler is a simple mock HTTP handler for testing
type mockHandler struct {
	statusCode int
	response   []byte
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(m.statusCode)
	w.Write(m.response)
}

func TestWebhookBridgeHandler_HTTPMethods(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "invalid method GET",
			method:         "GET",
			path:           "/billing/webhooks/callback/stripe",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "method not allowed",
		},
		{
			name:           "invalid method PUT",
			method:         "PUT",
			path:           "/billing/webhooks/callback/stripe",
			body:           []byte(`{"test":"data"}`),
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "method not allowed",
		},
		{
			name:           "invalid method DELETE",
			method:         "DELETE",
			path:           "/billing/webhooks/callback/stripe",
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFrontierHandler := &mockHandler{
				statusCode: http.StatusOK,
				response:   []byte(`{}`),
			}

			var body io.Reader
			if tt.body != nil {
				body = bytes.NewReader(tt.body)
			}
			req := httptest.NewRequest(tt.method, tt.path, body)
			rr := httptest.NewRecorder()

			bridgeHandler := WebhookBridgeHandler(mockFrontierHandler)
			bridgeHandler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}


func TestWebhookBridgeHandler_PathParsing(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		shouldBeValid bool
	}{
		{
			name:          "missing provider",
			path:          "/billing/webhooks/callback",
			shouldBeValid: false,
		},
		{
			name:          "missing provider with trailing slash",
			path:          "/billing/webhooks/callback/",
			shouldBeValid: false,
		},
		{
			name:          "wrong path - missing webhooks",
			path:          "/billing/wrong/callback/stripe",
			shouldBeValid: false,
		},
		{
			name:          "wrong path - missing callback",
			path:          "/billing/webhooks/wrong/stripe",
			shouldBeValid: false,
		},
		{
			name:          "wrong path - missing billing prefix",
			path:          "/v1beta1/billing/webhooks/callback/stripe",
			shouldBeValid: false,
		},
		{
			name:          "completely wrong path",
			path:          "/v2/api/webhooks",
			shouldBeValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFrontierHandler := &mockHandler{
				statusCode: http.StatusOK,
				response:   []byte(`{}`),
			}
			req := httptest.NewRequest("POST", tt.path, bytes.NewReader([]byte(`{}`)))
			rr := httptest.NewRecorder()

			bridgeHandler := WebhookBridgeHandler(mockFrontierHandler)
			bridgeHandler.ServeHTTP(rr, req)

			if tt.shouldBeValid {
				// Should not get 404 for path issues (might get other errors from handler logic)
				assert.NotEqual(t, http.StatusNotFound, rr.Code, "should not return 404 for valid path")
			} else {
				// Should get 404 for invalid paths
				assert.Equal(t, http.StatusNotFound, rr.Code, "should return 404 for invalid path")
			}
		})
	}
}

func TestWebhookBridgeHandler_SuccessfulRequest(t *testing.T) {
	// Mock handler that verifies the request was transformed correctly
	mockFrontierHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path was changed to ConnectRPC procedure
		assert.Contains(t, r.URL.Path, "BillingWebhookCallback")

		// Verify content type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Read and verify body was encoded as ConnectRPC request
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(body), "provider")
		assert.Contains(t, string(body), "body")

		// Return success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	req := httptest.NewRequest("POST", "/billing/webhooks/callback/stripe", bytes.NewReader([]byte(`{"event":"test"}`)))
	req.Header.Set("Stripe-Signature", "test-signature")
	rr := httptest.NewRecorder()

	bridgeHandler := WebhookBridgeHandler(mockFrontierHandler)
	bridgeHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
