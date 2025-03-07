package orgbilling

import (
	"context"
	"time"

	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/salt/rql"
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
	ID                     string             `rql:"name=id,type=string"`
	Name                   string             `rql:"name=name,type=string"`
	Title                  string             `rql:"name=title,type=string"`
	CreatedBy              string             `rql:"name=created_by,type=string"`
	PlanName               string             `rql:"name=plan_name,type=string"`
	PaymentMode            string             `rql:"name=payment_mode,type=string"`
	Country                string             `rql:"name=country,type=string"`
	Avatar                 string             `rql:"name=avatar,type=string"`
	State                  organization.State `rql:"name=state,type=string"`
	CreatedAt              time.Time          `rql:"name=created_at,type=datetime"`
	UpdatedAt              time.Time          `rql:"name=updated_at,type=datetime"`
	SubscriptionCycleEndAt time.Time          `rql:"name=subscription_cycle_end_at,type=datetime"`
	SubscriptionState      string             `rql:"name=subscription_state,type=string"`
	PlanInterval           string             `rql:"name=plan_interval,type=string"`
	PlanID                 string             `rql:"name=plan_id,type=string"`
}

func (s Service) Search(ctx context.Context, query *rql.Query) (OrgBilling, error) {
	return s.repository.Search(ctx, query)
}
