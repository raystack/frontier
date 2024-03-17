package usage

import (
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

type Type string

const (
	CreditType  Type = "credit"
	FeatureType Type = "feature"
)

type Usage struct {
	ID         string
	CustomerID string

	// Source is the source app or event that caused the transaction
	Source      string
	Description string
	// UserID is the user who initiated the transaction
	UserID string

	// Type is the type of usage, it can be credit or feature
	// if credit, the amount is the amount of credits that were consumed
	// if feature, the amount is the amount of features that were used
	Type   Type
	Amount int64

	CreatedAt time.Time
	Metadata  metadata.Metadata
}
