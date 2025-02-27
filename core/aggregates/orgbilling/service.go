package orgbilling

import (
	"context"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/salt/rql"
	"time"
)

type Repository interface {
	Search(ctx context.Context, query *rql.Query) ([]AggregatedOrganization, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type AggregatedOrganization struct {
	ID          string             `rql:"type=string"`
	Name        string             `rql:"type=string"`
	Title       string             `rql:"type=string"`
	CreatedBy   string             `rql:"type=string"`
	Plan        string             `rql:"type=string"`
	PaymentMode string             `rql:"type=string"`
	Country     string             `rql:"type=string"`
	Avatar      string             `rql:"type=string"`
	State       organization.State `rql:"type=string"`
	CreatedAt   time.Time          `rql:"type=datetime"`
	UpdatedAt   time.Time          `rql:"type=datetime"`
	CycleEndOn  time.Time          `rql:"type=datetime"`
}

func (s Service) Search(ctx context.Context, query *rql.Query) ([]AggregatedOrganization, error) {
	return s.repository.Search(ctx, query)
}
