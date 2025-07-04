package orgserviceuser

import (
	"context"
	"errors"
	"time"

	"github.com/raystack/salt/rql"
)

var (
	ErrInvalidDetail = errors.New("invalid service user detail")
)

type Repository interface {
	Search(ctx context.Context, orgID string, query *rql.Query) (OrganizationServiceUsers, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type OrganizationServiceUsers struct {
	ServiceUsers []AggregatedServiceUser `json:"service_users"`
	Pagination   Page                    `json:"pagination"`
}

type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type AggregatedServiceUser struct {
	ID            string    `rql:"filter,sort,column=id"`
	OrgID         string    `rql:"filter,sort,column=org_id"`
	Title         string    `rql:"filter,sort,column=title"`
	ProjectTitles []string  `rql:"filter,sort,column=project_titles"`
	CreatedAt     time.Time `rql:"filter,sort,column=created_at"`
	UpdatedAt     time.Time `rql:"filter,sort,column=updated_at"`
}

func (s Service) Search(ctx context.Context, orgID string, query *rql.Query) (OrganizationServiceUsers, error) {
	return s.repository.Search(ctx, orgID, query)
}
