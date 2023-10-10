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
	ID string

	Name        string // a machine friendly name for the feature
	Title       string // a human friendly title
	Description string
	Metadata    metadata.Metadata

	// Features for the plan, return only, should not be set when creating a plan
	Features []feature.Feature

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type Filter struct{}
