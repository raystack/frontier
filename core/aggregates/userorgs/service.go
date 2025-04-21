package userorgs

import (
	"context"
	"time"

	"github.com/raystack/salt/rql"
)

type Repository interface {
	Search(ctx context.Context, userID string, query *rql.Query) (UserOrgs, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type UserOrgs struct {
	Organizations []AggregatedUserOrganization `json:"organizations"`
	Group         Group                        `json:"group"`
	Pagination    Page                         `json:"pagination"`
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

type AggregatedUserOrganization struct {
	OrgID        string `rql:"name=org_id,type=string"`
	OrgTitle     string `rql:"name=org_title,type=string"`
	OrgName      string `rql:"name=org_name,type=string"`
	OrgAvatar    string
	ProjectCount int64 `rql:"name=project_count,type=int"`
	RoleNames    []string
	RoleTitles   []string
	RoleIDs      []string
	OrgJoinedOn  time.Time `rql:"name=org_joined_on,type=datetime"`
	UserID       string
}

func (s Service) Search(ctx context.Context, userID string, query *rql.Query) (UserOrgs, error) {
	return s.repository.Search(ctx, userID, query)
}
