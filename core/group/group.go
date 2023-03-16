package group

import (
	"context"
	"time"

	"github.com/goto/shield/core/relation"
	"github.com/goto/shield/pkg/metadata"
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
}

type Group struct {
	ID             string
	Name           string
	Slug           string
	OrganizationID string `json:"orgId"`
	Metadata       metadata.Metadata
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
