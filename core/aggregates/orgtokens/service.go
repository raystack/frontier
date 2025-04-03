package orgtokens

import (
	"context"
	"time"

	"github.com/raystack/salt/rql"
)

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
