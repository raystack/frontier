package invoice

import (
	"fmt"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound      = fmt.Errorf("invoice not found")
	ErrInvalidDetail = fmt.Errorf("invalid invoice detail")
)

type Invoice struct {
	ID            string
	CustomerID    string
	ProviderID    string
	State         string
	Currency      string
	Amount        int64
	HostedURL     string
	DueAt         time.Time
	EffectiveAt   time.Time
	CreatedAt     time.Time
	PeriodStartAt time.Time
	PeriodEndAt   time.Time

	Metadata metadata.Metadata
}

type Filter struct {
	CustomerID  string
	NonZeroOnly bool
}
