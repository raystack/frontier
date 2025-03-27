package projectusers

import (
	"context"
	"time"

	"github.com/raystack/frontier/core/user"
	"github.com/raystack/salt/rql"
)

type Repository interface {
	Search(ctx context.Context, projectID string, query *rql.Query) (ProjectUsers, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type ProjectUsers struct {
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
	ID              string     `rql:"name=id,type=string"`
	Name            string     `rql:"name=name,type=string"`
	Email           string     `rql:"name=email,type=string"`
	Title           string     `rql:"name=title,type=string"`
	State           user.State `rql:"name=state,type=string"`
	Avatar          string
	RoleNames       []string  `rql:"name=role_names,type=string"`
	RoleTitles      []string  `rql:"name=role_titles,type=string"`
	RoleIDs         []string  `rql:"name=role_ids,type=string"`
	ProjectID       string    `rql:"name=project_id,type=string"`
	ProjectJoinedAt time.Time `rql:"name=project_joined_at,type=datetime"`
}

func (s Service) Search(ctx context.Context, projectID string, query *rql.Query) (ProjectUsers, error) {
	return s.repository.Search(ctx, projectID, query)
}
