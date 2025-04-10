package user

import (
	"context"
	"time"

	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/salt/rql"
)

type State string

func (s State) String() string {
	return string(s)
}

const (
	Enabled  State = "enabled"
	Disabled State = "disabled"
)

type Repository interface {
	GetByID(ctx context.Context, id string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByIDs(ctx context.Context, userIds []string) ([]User, error)
	GetByName(ctx context.Context, name string) (User, error)
	Create(ctx context.Context, user User) (User, error)
	List(ctx context.Context, flt Filter) ([]User, error)
	UpdateByID(ctx context.Context, toUpdate User) (User, error)
	UpdateByName(ctx context.Context, toUpdate User) (User, error)
	UpdateByEmail(ctx context.Context, toUpdate User) (User, error)
	Delete(ctx context.Context, id string) error
	SetState(ctx context.Context, id string, state State) error
	Search(ctx context.Context, query *rql.Query) (SearchUserResponse, error)
}

type User struct {
	ID        string `rql:"name=id,type=string"`
	Name      string `rql:"name=name,type=string"`
	Email     string `rql:"name=email,type=string"`
	State     State  `rql:"name=state,type=string"`
	Avatar    string `rql:"name=avatar,type=string"`
	Title     string `rql:"name=title,type=string"`
	Metadata  metadata.Metadata
	CreatedAt time.Time `rql:"name=created_at,type=datetime"`
	UpdatedAt time.Time `rql:"name=updated_at,type=datetime"`
}

type AccessPair struct {
	User User
	On   string
	Can  []string
}

type SearchUserResponse struct {
	Users      []User `json:"users"`
	Group      Group  `json:"group"`
	Pagination Page   `json:"pagination"`
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
