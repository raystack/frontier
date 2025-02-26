package orgaggregation

import (
	"context"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/salt/rql"
	"time"
)

type Repository interface{}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type AggregatedOrganization struct {
	ID          string
	Name        string
	Title       string
	CreatedBy   string
	Plan        string
	PaymentMode string
	Country     string
	Avatar      string
	State       organization.State
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CycleEndOn  time.Time
}

func (s Service) Search(ctx context.Context, query *rql.Query) ([]AggregatedOrganization, error) {
	return nil, nil
}
