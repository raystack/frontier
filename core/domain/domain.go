package domain

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, domain Domain) error
	Get(ctx context.Context, id string) (Domain, error)
	Update(ctx context.Context, id string, domain Domain) (Domain, error)
	List(ctx context.Context, flt Filter) ([]Domain, error)
	Delete(ctx context.Context, id string) (string, error)
}

type Domain struct {
	ID         string
	Name       string
	OrgID      string
	Token      string
	Verified   bool
	VerifiedAt time.Time
	CreatedAt  time.Time
}
