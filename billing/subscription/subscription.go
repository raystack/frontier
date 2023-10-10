package subscription

import (
	"errors"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound      = errors.New("subscription not found")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
	ErrInvalidID     = errors.New("invalid subscription id")
	ErrInvalidDetail = errors.New("invalid subscription detail")
)

type Subscription struct {
	ID         string
	ProviderID string // identifier set by the billing engine provider
	CustomerID string

	PlanID string
	State  string

	Metadata   metadata.Metadata
	CreatedAt  time.Time
	CanceledAt *time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

type Filter struct {
	CustomerID string
	State      string
}
