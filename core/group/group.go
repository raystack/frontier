package group

import (
	"context"
	"time"

	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/pkg/metadata"
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
	Create(ctx context.Context, grp Group) (Group, error)
	GetByID(ctx context.Context, id string) (Group, error)
	GetByIDs(ctx context.Context, groupIDs []string) ([]Group, error)
	List(ctx context.Context, flt Filter) ([]Group, error)
	UpdateByID(ctx context.Context, toUpdate Group) (Group, error)
	ListUserGroups(ctx context.Context, userId string, roleId string) ([]Group, error)
	ListGroupRelations(ctx context.Context, objectId, subjectType, role string) ([]relation.Relation, error)
	SetState(ctx context.Context, id string, state State) error
	Delete(ctx context.Context, id string) error
}

type Group struct {
	ID             string
	Name           string
	Title          string
	OrganizationID string `json:"orgId"`
	Metadata       metadata.Metadata
	State          State
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
