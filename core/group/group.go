package group

import (
	"context"
	"time"

	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/schema"

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
	GetBySlug(ctx context.Context, slug string) (Group, error)
	List(ctx context.Context, flt Filter) ([]Group, error)
	UpdateByID(ctx context.Context, toUpdate Group) (Group, error)
	UpdateBySlug(ctx context.Context, toUpdate Group) (Group, error)
	ListUserGroups(ctx context.Context, userId string, roleId string) ([]Group, error)
	ListGroupRelations(ctx context.Context, objectId, subjectType, role string) ([]relation.RelationV2, error)
	SetState(ctx context.Context, id string, state State) error
	Delete(ctx context.Context, id string) error
}

type Group struct {
	ID             string
	Name           string
	Slug           string
	OrganizationID string `json:"orgId"`
	Metadata       metadata.Metadata
	State          State
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func BuildUserGroupAdminSubject(user user.User) relation.Subject {
	return relation.Subject{
		ID:        user.ID,
		Namespace: schema.UserPrincipal,
		RoleID:    schema.ManagerRole,
	}
}
