package group

import (
	"context"
	"errors"
	"time"

	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/user"
)

var (
	ErrNotExist    = errors.New("group doesn't exist")
	ErrInvalidUUID = errors.New("invalid syntax of uuid")
)

type Store interface {
	CreateGroup(ctx context.Context, grp Group) (Group, error)
	GetGroup(ctx context.Context, id string) (Group, error)
	ListGroups(ctx context.Context, org organization.Organization) ([]Group, error)
	UpdateGroup(ctx context.Context, toUpdate Group) (Group, error)
	GetUsersByIDs(ctx context.Context, userIds []string) ([]user.User, error)
	GetUser(ctx context.Context, userId string) (user.User, error)
	ListUserGroups(ctx context.Context, userId string, roleId string) ([]Group, error)
	ListGroupUsers(ctx context.Context, groupId string, roleId string) ([]user.User, error)
	GetRelationByFields(ctx context.Context, relation relation.Relation) (relation.Relation, error)
	ListUserGroupRelations(ctx context.Context, userId string, groupId string) ([]relation.Relation, error)
}

type Group struct {
	ID             string
	Name           string
	Slug           string
	Organization   organization.Organization
	OrganizationID string `json:"orgId"`
	Metadata       map[string]any
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
