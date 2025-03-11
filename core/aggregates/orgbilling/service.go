package orgbilling

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"reflect"

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

type CSVExport struct {
	OrganizationID       string `csv:"Organization ID"`
	Name                 string `csv:"Name"`
	Title                string `csv:"Title"`
	CreatedBy            string `csv:"Created By"`
	PlanName             string `csv:"Plan Name"`
	PaymentMode          string `csv:"Payment Mode"`
	Country              string `csv:"Country"`
	State                string `csv:"State"`
	CreatedAt            string `csv:"Created At"`
	UpdatedAt            string `csv:"Updated At"`
	SubscriptionCycleEnd string `csv:"Subscription Cycle End"`
	SubscriptionState    string `csv:"Subscription State"`
	PlanInterval         string `csv:"Plan Interval"`
}

// FromAggregatedOrganization converts AggregatedOrganization to CSVExport
func NewCSVExport(org AggregatedOrganization) CSVExport {
	return CSVExport{
		OrganizationID:       org.ID,
		Name:                 org.Name,
		Title:                org.Title,
		CreatedBy:            org.CreatedBy,
		PlanName:             org.PlanName,
		PaymentMode:          org.PaymentMode,
		Country:              org.Country,
		State:                string(org.State),
		CreatedAt:            org.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            org.UpdatedAt.Format(time.RFC3339),
		SubscriptionCycleEnd: org.SubscriptionCycleEndAt.Format(time.RFC3339),
		SubscriptionState:    org.SubscriptionState,
		PlanInterval:         org.PlanInterval,
	}
}

// GetHeaders returns the CSV headers based on struct tags
func (c CSVExport) GetHeaders() []string {
	t := reflect.TypeOf(c)
	headers := make([]string, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if tag := field.Tag.Get("csv"); tag != "" {
			headers[i] = tag
		} else {
			headers[i] = field.Name
		}
	}

	return headers
}

// ToRow converts the struct to a string slice for CSV writing
func (c CSVExport) ToRow() []string {
	v := reflect.ValueOf(c)
	row := make([]string, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		row[i] = fmt.Sprint(v.Field(i).Interface())
	}

	return row
}

func (s Service) Search(ctx context.Context, query *rql.Query) (OrgBilling, error) {
	return s.repository.Search(ctx, query)
}

func (s Service) Export(ctx context.Context) ([]byte, string, error) {
	orgBillingData, err := s.repository.Search(ctx, &rql.Query{})
	if err != nil {
		return nil, "", err
	}

	// Create a buffer to write CSV data
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write headers
	if len(orgBillingData.Organizations) > 0 {
		csvExport := NewCSVExport(orgBillingData.Organizations[0])
		headers := csvExport.GetHeaders()
		if err := writer.Write(headers); err != nil {
			return nil, "", err
		}
	}

	// Write data rows
	for _, org := range orgBillingData.Organizations {
		csvExport := NewCSVExport(org)
		if err := writer.Write(csvExport.ToRow()); err != nil {
			return nil, "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), CSVContentType, nil
}
