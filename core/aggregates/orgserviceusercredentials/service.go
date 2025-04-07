package orgserviceusercredentials

import (
	"context"
	"time"

	"github.com/raystack/salt/rql"
)

type Repository interface {
	Search(ctx context.Context, orgID string, query *rql.Query) (OrganizationServiceUserCredentials, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type OrganizationServiceUserCredentials struct {
	Credentials []AggregatedServiceUserCredential `json:"credentials"`
	Pagination  Page                              `json:"pagination"`
}

type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type AggregatedServiceUserCredential struct {
	Title            string    `rql:"name=title,type=string"`
	ServiceUserTitle string    `rql:"name=serviceuser_title,type=string"`
	CreatedAt        time.Time `rql:"name=created_at,type=datetime"`
	OrgID            string    `rql:"name=org_id,type=string"`
}

func (s Service) Search(ctx context.Context, orgID string, query *rql.Query) (OrganizationServiceUserCredentials, error) {
	return s.repository.Search(ctx, orgID, query)
}
