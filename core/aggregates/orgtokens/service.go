package orgtokens

import (
	"context"
	"time"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"bytes"
	"encoding/csv"

	"github.com/raystack/salt/rql"
)

var ErrNoContent = errors.New("no content")

const CSVContentType = "text/csv"

type Repository interface {
	Search(ctx context.Context, orgID string, query *rql.Query) (OrganizationTokens, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type OrganizationTokens struct {
	Tokens     []AggregatedToken `json:"tokens"`
	Pagination Page              `json:"pagination"`
}
type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type AggregatedToken struct {
	Amount      int64     `rql:"name=amount,type=number"`
	Type        string    `rql:"name=type,type=string"`
	Description string    `rql:"name=description,type=string"`
	UserID      string    `rql:"name=user_id,type=string"`
	UserTitle   string    `rql:"name=user_title,type=string"`
	UserAvatar  string    `rql:"name=user_avatar,type=string"`
	CreatedAt   time.Time `rql:"name=created_at,type=datetime"`
	OrgID       string    `rql:"name=org_id,type=string"`
}

func (s Service) Search(ctx context.Context, orgID string, query *rql.Query) (OrganizationTokens, error) {
	return s.repository.Search(ctx, orgID, query)
}

// CSVExport represents the structure for CSV export of organization tokens
type CSVExport struct {
	Amount      string `csv:"Amount"`
	Type        string `csv:"Type"`
	Description string `csv:"Description"`
	UserID      string `csv:"User ID"`
	UserTitle   string `csv:"User Title"`
	CreatedAt   string `csv:"Created At"`
	OrgID       string `csv:"Organization ID"`
}

// NewCSVExport converts AggregatedToken to CSVExport
func NewCSVExport(token AggregatedToken) CSVExport {
	return CSVExport{
		Amount:      strconv.FormatInt(token.Amount, 10),
		Type:        token.Type,
		Description: token.Description,
		UserID:      token.UserID,
		UserTitle:   token.UserTitle,
		CreatedAt:   token.CreatedAt.Format(time.RFC3339),
		OrgID:       token.OrgID,
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

// Export generates a CSV file containing organization tokens data
func (s Service) Export(ctx context.Context, orgID string) ([]byte, string, error) {
	orgTokensData, err := s.repository.Search(ctx, orgID, &rql.Query{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to search organization tokens: %w", err)
	}

	if len(orgTokensData.Tokens) == 0 {
		return nil, "", fmt.Errorf("%w: no tokens found for organization %s", ErrNoContent, orgID)
	}

	// Create a buffer to write CSV data
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write headers
	csvExport := NewCSVExport(orgTokensData.Tokens[0])
	headers := csvExport.GetHeaders()
	if err := writer.Write(headers); err != nil {
		return nil, "", fmt.Errorf("error writing CSV headers: %w", err)
	}

	// Write data rows
	for _, token := range orgTokensData.Tokens {
		csvExport := NewCSVExport(token)
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
