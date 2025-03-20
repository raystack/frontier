package orgusers

import (
	"context"

	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"
	"time"
	"errors"

	"github.com/raystack/frontier/core/user"
	"github.com/raystack/salt/rql"
)

const CSVContentType = "text/csv"

type Repository interface {
	Search(ctx context.Context, orgID string, query *rql.Query) (OrgUsers, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type OrgUsers struct {
	Users      []AggregatedUser `json:"users"`
	Group      Group            `json:"group"`
	Pagination Page             `json:"pagination"`
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

type AggregatedUser struct {
	ID          string     `rql:"name=id,type=string"`
	Name        string     `rql:"name=name,type=string"`
	Title       string     `rql:"name=title,type=string"`
	Avatar      string     `rql:"name=avatar,type=string"`
	Email       string     `rql:"name=email,type=string"`
	State       user.State `rql:"name=state,type=string"`
	RoleNames   []string   `rql:"name=role_names,type=string"`
	RoleTitles  []string   `rql:"name=role_titles,type=string"`
	RoleIDs     []string   `rql:"name=role_ids,type=string"`
	OrgID       string     `rql:"name=org_id,type=string"`
	OrgJoinedAt time.Time  `rql:"name=org_joined_at,type=datetime"`
}

func (s Service) Search(ctx context.Context, orgID string, query *rql.Query) (OrgUsers, error) {
	return s.repository.Search(ctx, orgID, query)
}

// CSVExport represents the structure for CSV export of organization users
type CSVExport struct {
	UserID      string `csv:"User ID"`
	Name        string `csv:"Name"`
	Title       string `csv:"Title"`
	Email       string `csv:"Email"`
	State       string `csv:"State"`
	RoleNames   string `csv:"Role Names"`
	RoleTitles  string `csv:"Role Titles"`
	OrgJoinedAt string `csv:"Organization Joined At"`
}

// NewCSVExport converts AggregatedUser to CSVExport
func NewCSVExport(user AggregatedUser) CSVExport {
	return CSVExport{
		UserID:      user.ID,
		Name:        user.Name,
		Title:       user.Title,
		Email:       user.Email,
		State:       string(user.State),
		RoleNames:   strings.Join(user.RoleNames, ", "),
		RoleTitles:  strings.Join(user.RoleTitles, ", "),
		OrgJoinedAt: user.OrgJoinedAt.Format(time.RFC3339),
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

var ErrNoContent = errors.New("no content")

// Export generates a CSV file containing organization users data
func (s Service) Export(ctx context.Context, orgID string) ([]byte, string, error) {
	orgUsersData, err := s.repository.Search(ctx, orgID, &rql.Query{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to search organization users: %w", err)
	}

	if len(orgUsersData.Users) == 0 {
		return nil, "", fmt.Errorf("%w: no users found for organization %s", ErrNoContent, orgID)
	}

	// Create a buffer to write CSV data
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write headers
	csvExport := NewCSVExport(orgUsersData.Users[0])
	headers := csvExport.GetHeaders()
	if err := writer.Write(headers); err != nil {
		return nil, "", fmt.Errorf("error writing CSV headers: %w", err)
	}

	// Write data rows
	for _, user := range orgUsersData.Users {
		csvExport := NewCSVExport(user)
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
