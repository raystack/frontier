package domain

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, domain Domain) (Domain, error)
	Get(ctx context.Context, id string) (Domain, error)
	Update(ctx context.Context, domain Domain) (Domain, error)
	List(ctx context.Context, flt Filter) ([]Domain, error)
	Delete(ctx context.Context, id string) error
	DeleteExpiredDomainRequests(ctx context.Context) error
}

type Status string

func (s Status) String() string {
	return string(s)
}

const (
	Pending  Status = "pending"
	Verified Status = "verified"
)

type Domain struct {
	ID        string
	Name      string
	OrgID     string
	Token     string
	State     string
	UpdatedAt time.Time
	CreatedAt time.Time
}
