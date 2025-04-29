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
	ProjectID      string    `rql:"name=project_id,type=string"`
	ProjectTitle   string    `rql:"name=project_title,type=string"`
	ProjectName    string    `rql:"name=project_name,type=string"`
	CreatedOn      time.Time `rql:"name=created_on,type=datetime"`
	UserNames      []string  `rql:"name=user_names,type=string"`
	UserTitles     []string  `rql:"name=user_titles,type=string"`
	UserIDs        []string  `rql:"name=user_ids,type=string"`
	UserAvatars    []string  `rql:"name=user_avatars,type=string"`
	OrgID          string    `rql:"name=org_id,type=string"`
	UserID         string    `rql:"name=user_id,type=string"`
}

func (s Service) Search(ctx context.Context, userID string, orgID string, query *rql.Query) (UserProjects, error) {
	return s.repository.Search(ctx, userID, orgID, query)
}
