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

type Project struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Name  string `json:"name"`
}

// salt resolves an rql field by the `name` tag and reads its type from the `type` tag.
// a `column` tag is ignored, which left created_at and org_id unresolvable.
type AggregatedServiceUser struct {
	ID        string    `rql:"name=id,type=string"`
	OrgID     string    `rql:"name=org_id,type=string"`
	Title     string    `rql:"name=title,type=string"`
	Projects  []Project `rql:"name=projects,type=string"`
	CreatedAt time.Time `rql:"name=created_at,type=datetime"`
	UpdatedAt time.Time `rql:"name=updated_at,type=datetime"`
}

func (s Service) Search(ctx context.Context, orgID string, query *rql.Query) (OrganizationServiceUsers, error) {
	return s.repository.Search(ctx, orgID, query)
}
