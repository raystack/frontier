package invoice

import (
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

type Invoice struct {
	ID          string
	CustomerID  string
	ProviderID  string
	State       string
	Currency    string
	Amount      int64
	HostedURL   string
	DueDate     time.Time
	EffectiveAt time.Time
	CreatedAt   time.Time

	Metadata metadata.Metadata
}

type Filter struct {
	CustomerID string
}
