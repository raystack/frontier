package orgusers

import (
	"context"

	"github.com/raystack/frontier/core/user"
	"github.com/raystack/salt/rql"
	"time"
)

type Repository interface {
	Search(ctx context.Context, query *rql.Query) (OrgUsers, error)
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
	RoleName    string     `rql:"name=role_name,type=string"`
	RoleTitle   string     `rql:"name=role_title,type=string"`
	RoleID      string     `rql:"name=role_id,type=string"`
	OrgID       string     `rql:"name=org_id,type=string"`
	OrgJoinedAt time.Time  `rql:"name=org_joined_at,type=datetime"`
}

func (s Service) Search(ctx context.Context, query *rql.Query) (OrgUsers, error) {
	return s.repository.Search(ctx, query)
}
