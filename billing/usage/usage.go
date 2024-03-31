package usage

import (
	"fmt"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrExistingRevertedUsage = fmt.Errorf("a reverted usage cannot be reverted again")
	ErrRevertAmountExceeds   = fmt.Errorf("revert amount is greater than the usage amount")
)

type Type string

func (t Type) String() string {
	return string(t)
}

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
