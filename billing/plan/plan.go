package plan

import (
	"errors"
	"time"

	"github.com/raystack/frontier/billing/feature"

	"github.com/raystack/frontier/pkg/metadata"
)

var (
	ErrNotFound      = errors.New("plan not found")
	ErrInvalidUUID   = errors.New("invalid syntax of uuid")
	ErrInvalidName   = errors.New("plan name is invalid")
	ErrInvalidDetail = errors.New("invalid plan detail")
)

// Plan is a collection of features
// it is a logical grouping of features and doesn't have
// a corresponding billing engine entity
type Plan struct {
	ID string `json:"id" yaml:"id"`

	Name        string            `json:"name" yaml:"name"`   // a machine friendly name for the feature
	Title       string            `json:"title" yaml:"title"` // a human friendly title
	Description string            `json:"description" yaml:"description"`
	Metadata    metadata.Metadata `json:"metadata" yaml:"metadata"`

	// Interval is the interval at which the plan is billed
	// e.g. day, week, month, year
	Interval string `json:"interval" yaml:"interval"`

	// Features for the plan, return only, should not be set when creating a plan
	Features []feature.Feature `json:"features" yaml:"features"`

	State     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Filter struct{}

type File struct {
	Plans    []Plan            `json:"plans" yaml:"plans"`
	Features []feature.Feature `json:"features" yaml:"features"`
}
