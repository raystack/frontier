package webhook

import (
	"fmt"
	"time"

	"github.com/raystack/frontier/pkg/crypt"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	ErrInvalidEvent       = fmt.Errorf("invalid event")
	ErrInvalidSignature   = fmt.Errorf("invalid signature")
	ErrVerificationFailed = fmt.Errorf("verification failed")
	ErrTimestampExpired   = fmt.Errorf("timestamp expired")

	EventExpiryDuration = time.Minute * 10
)

func ParseAndValidateEvent(data []byte, hexKey string, suppliedHexMAC string) (*frontierv1beta1.WebhookEvent, error) {
	validHmac, err := crypt.VerifyHMACFromHex(data, hexKey, suppliedHexMAC)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", err.Error(), ErrVerificationFailed)
	}
	if !validHmac {
		return nil, ErrInvalidSignature
	}
	var event frontierv1beta1.WebhookEvent
	if err := protojson.Unmarshal(data, &event); err != nil {
		return nil, fmt.Errorf("%s: %w", err.Error(), ErrInvalidEvent)
	}

	// Check if the event is not older than N minutes
	if event.GetCreatedAt() == nil || event.GetCreatedAt().AsTime().Before(time.Now().Add(-EventExpiryDuration)) {
		return nil, ErrTimestampExpired
	}
	return &event, nil
}
