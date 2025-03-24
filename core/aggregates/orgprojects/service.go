package orgprojects

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/raystack/frontier/core/project"
	"github.com/raystack/salt/rql"
)

var ErrNoContent = errors.New("no content")

const CSVContentType = "text/csv"

type Repository interface {
	Search(ctx context.Context, orgID string, query *rql.Query) (OrgProjects, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type OrgProjects struct {
	Projects   []AggregatedProject `json:"projects"`
	Group      Group               `json:"group"`
	Pagination Page                `json:"pagination"`
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

type AggregatedProject struct {
	ID             string        `rql:"name=id,type=string"`
	Name           string        `rql:"name=name,type=string"`
	Title          string        `rql:"name=title,type=string"`
	State          project.State `rql:"name=state,type=string"`
	MemberCount    int64         `rql:"name=member_count,type=number"`
	CreatedAt      time.Time     `rql:"name=created_at,type=datetime"`
	OrganizationID string        `rql:"name=organization_id,type=string"`
	UserIDs        []string
}

// CSVExport represents the structure for CSV export of organization projects
type CSVExport struct {
	ProjectID      string `csv:"Project ID"`
	Name           string `csv:"Name"`
	Title          string `csv:"Title"`
	State          string `csv:"State"`
	MemberCount    string `csv:"Member Count"`
	UserIDs        string `csv:"User IDs"`
	CreatedAt      string `csv:"Created At"`
	OrganizationID string `csv:"Organization ID"`
}

// NewCSVExport converts AggregatedProject to CSVExport
func NewCSVExport(project AggregatedProject) CSVExport {
	return CSVExport{
		ProjectID:      project.ID,
		Name:           project.Name,
		Title:          project.Title,
		State:          string(project.State),
		MemberCount:    fmt.Sprint(project.MemberCount),
		UserIDs:        strings.Join(project.UserIDs, ","), // Comma-separated list of user IDs
		CreatedAt:      project.CreatedAt.Format(time.RFC3339),
		OrganizationID: project.OrganizationID,
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

func (s Service) Search(ctx context.Context, orgID string, query *rql.Query) (OrgProjects, error) {
	return s.repository.Search(ctx, orgID, query)
}

// Export generates a CSV file containing organization projects data
func (s Service) Export(ctx context.Context, orgID string) ([]byte, string, error) {
	orgProjectsData, err := s.repository.Search(ctx, orgID, &rql.Query{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to search organization projects: %w", err)
	}

	orgProjectsData.Projects = []AggregatedProject{}
	if len(orgProjectsData.Projects) == 0 {
		return nil, "", fmt.Errorf("%w: no projects found for organization %s", ErrNoContent, orgID)
	}

	// Create a buffer to write CSV data
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write headers
	csvExport := NewCSVExport(orgProjectsData.Projects[0])
	headers := csvExport.GetHeaders()
	if err := writer.Write(headers); err != nil {
		return nil, "", fmt.Errorf("error writing CSV headers: %w", err)
	}

	// Write data rows
	for _, project := range orgProjectsData.Projects {
		csvExport := NewCSVExport(project)
		if err := writer.Write(csvExport.ToRow()); err != nil {
			return nil, "", fmt.Errorf("error writing CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", fmt.Errorf("error flushing CSV writer: %w", err)
	}

	return buf.Bytes(), CSVContentType, nil
}
