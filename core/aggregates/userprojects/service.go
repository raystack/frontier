package userprojects

import (
	"context"
	"time"

	"github.com/raystack/salt/rql"
)

type Repository interface {
	Search(ctx context.Context, userID string, orgID string, query *rql.Query) (UserProjects, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
	}
}

type UserProjects struct {
	Projects   []AggregatedProject `json:"projects"`
	Pagination Page                `json:"pagination"`
}

type Page struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type AggregatedProject struct {
	ProjectID    string `rql:"name=project_id,type=string"`
	ProjectTitle string `rql:"name=project_title,type=string"`
	ProjectName  string `rql:"name=project_name,type=string"`
	CreatedOn    time.Time
	UserNames    []string
	UserTitles   []string
	UserIDs      []string
	UserAvatars  []string
	OrgID        string
	UserID       string
}

func (s Service) Search(ctx context.Context, userID string, orgID string, query *rql.Query) (UserProjects, error) {
	return s.repository.Search(ctx, userID, orgID, query)
}
