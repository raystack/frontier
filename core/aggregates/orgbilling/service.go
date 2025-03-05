package orgbilling

import (
	"context"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/salt/rql"
	"time"
)

type Repository interface {
	Search(ctx context.Context, query *rql.Query) (OrgBilling, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type OrgBilling struct {
	Organizations []AggregatedOrganization `json:"organization"`
	Group         Group                    `json:"group"`
	Pagination    Page                     `json:"pagination"`
}

type Group struct {
	Name string      `json:"name"`
	Data []GroupData `json:"data"`
}

type GroupData struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type AggregatedOrganization struct {
	ID                string             `rql:"type=string"`
	Name              string             `rql:"type=string"`
	Title             string             `rql:"type=string"`
	CreatedBy         string             `rql:"type=string"`
	Plan              string             `rql:"type=string"`
	PaymentMode       string             `rql:"type=string"`
	Country           string             `rql:"type=string"`
	Avatar            string             `rql:"type=string"`
	State             organization.State `rql:"type=string"`
	CreatedAt         time.Time          `rql:"type=datetime"`
	UpdatedAt         time.Time          `rql:"type=datetime"`
	CycleEndAt        time.Time          `rql:"type=datetime"`
	SubscriptionState string             `rql:"type=string"`
	PlanInterval      string             `rql:"type=string"`
	PlanID            string             `rql:"type=string"`
}

func (s Service) Search(ctx context.Context, query *rql.Query) (OrgBilling, error) {
	return s.repository.Search(ctx, query)
}
