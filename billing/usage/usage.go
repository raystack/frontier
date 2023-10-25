package usage

import (
	"time"

	"github.com/raystack/frontier/pkg/metadata"
)

type Type string

const (
	TypeCredit  Type = "credit"
	TypeFeature Type = "feature"
)

type Usage struct {
	ID          string
	CustomerID  string
	Source      string
	Description string

	// Type is the type of usage, it can be credit or feature
	// if credit, the amount is the amount of credits that were consumed
	// if feature, the amount is the amount of features that were used
	Type   Type
	Amount int64

	CreatedAt time.Time
	Metadata  metadata.Metadata
}
