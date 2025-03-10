package orgbilling

import (
	"bytes"
	"context"
	"encoding/csv"
	"time"

	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/salt/rql"
)

const CSVContentType = "text/csv"

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

func (s Service) Export(ctx context.Context) ([]byte, string, error) {
	orgBillingData, err := s.repository.Search(ctx, &rql.Query{})
	if err != nil {
		return nil, "", err
	}
	if err != nil {
		return nil, "", err
	}
	// Create a buffer to write CSV data
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{
		"Organization ID",
		"Name",
		"Title",
		"Created By",
		"Plan Name",
		"Payment Mode",
		"Country",
		"State",
		"Created At",
		"Updated At",
		"Subscription Cycle End",
		"Subscription State",
		"Plan Interval",
	}

	if err := writer.Write(header); err != nil {
		return nil, "", err
	}

	// Write data rows
	for _, org := range orgBillingData.Organizations {
		row := []string{
			org.ID,
			org.Name,
			org.Title,
			org.CreatedBy,
			org.PlanName,
			org.PaymentMode,
			org.Country,
			string(org.State),
			org.CreatedAt.Format(time.RFC3339),
			org.UpdatedAt.Format(time.RFC3339),
			org.SubscriptionCycleEndAt.Format(time.RFC3339),
			org.SubscriptionState,
			org.PlanInterval,
		}
		if err := writer.Write(row); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), CSVContentType, nil
}
