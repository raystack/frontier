package stripeprovider

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/raystack/frontier/billing"
	"github.com/stripe/stripe-go/v79"
)

func TestVerifyWebhook(t *testing.T) {
	p := New(nil) // webhook verification doesn't need the API client
	secret := "whsec_test_secret"

	payload := buildWebhookPayload(t, string(stripe.EventTypeCustomerSubscriptionUpdated), "sub_123")
	sig := computeWebhookSignature(t, payload, secret)

	evt, err := p.VerifyWebhook(payload, sig, []string{secret})
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if evt.Type != billing.EventSubscriptionUpdated {
		t.Errorf("Type = %q, want %q", evt.Type, billing.EventSubscriptionUpdated)
	}
	if evt.ObjectID != "sub_123" {
		t.Errorf("ObjectID = %q, want %q", evt.ObjectID, "sub_123")
	}
}

func TestVerifyWebhook_SecretRotation(t *testing.T) {
	p := New(nil)
	secret := "whsec_current"
	oldSecret := "whsec_old"

	payload := buildWebhookPayload(t, string(stripe.EventTypeInvoicePaid), "in_456")
	sig := computeWebhookSignature(t, payload, secret)

	// Should succeed when valid secret is second in the list
	evt, err := p.VerifyWebhook(payload, sig, []string{oldSecret, secret})
	if err != nil {
		t.Fatalf("VerifyWebhook with rotated secrets: %v", err)
	}
	if evt.Type != billing.EventInvoicePaid {
		t.Errorf("Type = %q, want %q", evt.Type, billing.EventInvoicePaid)
	}
}

func TestVerifyWebhook_BadSignature(t *testing.T) {
	p := New(nil)
	payload := buildWebhookPayload(t, string(stripe.EventTypeCustomerCreated), "cus_789")

	_, err := p.VerifyWebhook(payload, "t=0,v1=bad", []string{"whsec_test"})
	if err == nil {
		t.Fatal("expected error for bad signature")
	}
}

func TestVerifyWebhook_NoSecrets(t *testing.T) {
	p := New(nil)
	_, err := p.VerifyWebhook([]byte("{}"), "t=0,v1=abc", nil)
	if err == nil {
		t.Fatal("expected error for empty secrets")
	}
}

func TestVerifyWebhook_UnknownEventType(t *testing.T) {
	p := New(nil)
	secret := "whsec_test"

	payload := buildWebhookPayload(t, "some.unknown.event", "obj_1")
	sig := computeWebhookSignature(t, payload, secret)

	evt, err := p.VerifyWebhook(payload, sig, []string{secret})
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if evt.Type != "some.unknown.event" {
		t.Errorf("Type = %q, want passthrough of unknown event", evt.Type)
	}
}

// buildWebhookPayload constructs a minimal Stripe webhook event JSON.
func buildWebhookPayload(t *testing.T, eventType, objectID string) []byte {
	t.Helper()
	evt := map[string]any{
		"id":          "evt_test",
		"type":        eventType,
		"api_version": stripe.APIVersion,
		"data": map[string]any{
			"object": map[string]any{
				"id": objectID,
			},
		},
	}
	b, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal webhook payload: %v", err)
	}
	return b
}

// computeWebhookSignature creates a valid Stripe webhook signature header value.
func computeWebhookSignature(t *testing.T, payload []byte, secret string) string {
	t.Helper()
	ts := time.Now().Unix()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d", ts)))
	mac.Write([]byte("."))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))
	return fmt.Sprintf("t=%d,v1=%s", ts, sig)
}
