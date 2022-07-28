package group

import (
	"context"
	"time"

	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/user"
)

type Repository interface {
	Create(ctx context.Context, grp Group) (Group, error)
	GetByID(ctx context.Context, id string) (Group, error)
	GetBySlug(ctx context.Context, slug string) (Group, error)
	List(ctx context.Context, flt Filter) ([]Group, error)
	UpdateByID(ctx context.Context, toUpdate Group) (Group, error)
	UpdateBySlug(ctx context.Context, toUpdate Group) (Group, error)
	// GetUsersByIDs(ctx context.Context, userIds []string) ([]user.User, error)
	// GetUser(ctx context.Context, userId string) (user.User, error)
	ListUserGroups(ctx context.Context, userId string, roleId string) ([]Group, error)
	// GetRelationByFields(ctx context.Context, relation relation.Relation) (relation.Relation, error)
	ListUsersByGroupID(ctx context.Context, groupId string, roleId string) ([]user.User, error)
	ListUsersByGroupSlug(ctx context.Context, groupSlug string, roleId string) ([]user.User, error)
	ListUserGroupIDRelations(ctx context.Context, userId string, groupId string) ([]relation.Relation, error)
	ListUserGroupSlugRelations(ctx context.Context, userId string, groupSlug string) ([]relation.Relation, error)
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
