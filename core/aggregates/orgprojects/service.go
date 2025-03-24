package orgprojects

import (
	"context"
	"time"

	"github.com/raystack/frontier/core/project"
	"github.com/raystack/salt/rql"
)

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

func (s Service) Search(ctx context.Context, orgID string, query *rql.Query) (OrgProjects, error) {
	return s.repository.Search(ctx, orgID, query)
}
