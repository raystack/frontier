package webhook

import (
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

type State string

const (
	Enabled  State = "enabled"
	Disabled State = "disabled"
)

type Secret struct {
	ID string
	// Value store the secret value in hex
	Value string
}

type Endpoint struct {
	ID string
	// Description is the description of the webhook
	Description string
	// URL is the URL of the webhook
	URL string
	// SubscribedEvents is the list of events that the webhook is subscribed to
	SubscribedEvents []string
	// Headers is the headers to be sent with the webhook
	Headers map[string]string
	// Secrets is the list of secrets to sign the payload
	Secrets []Secret
	// State is the state of the webhook
	State State

	// Metadata is the metadata of the webhook
	Metadata metadata.Metadata
	// CreatedAt is the creation time of the webhook
	CreatedAt time.Time
	// UpdatedAt is the update time of the webhook
	UpdatedAt time.Time
}

type Event struct {
	ID        string
	Action    string
	Data      metadata.Metadata
	CreatedAt time.Time
}

type EndpointFilter struct {
	State State
}
