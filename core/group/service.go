package group

import (
	"context"
	"strings"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/pkg/str"
	"github.com/odpf/shield/pkg/uuid"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.RelationV2) (relation.RelationV2, error)
	Delete(ctx context.Context, rel relation.Relation) error
	CheckPermission(ctx context.Context, usr user.User, resourceNS namespace.Namespace, resourceIdxa string, action action.Action) (bool, error)
}

type UserService interface {
	FetchCurrentUser(ctx context.Context) (user.User, error)
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error)
}

type Service struct {
	repository      Repository
	relationService RelationService
	userService     UserService
}

func NewService(repository Repository, relationService RelationService, userService UserService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationService,
		userService:     userService,
	}
}

func (s Service) Create(ctx context.Context, grp Group) (Group, error) {
	newGroup, err := s.repository.Create(ctx, grp)
	if err != nil {
		return Group{}, err
	}

	if err = s.addTeamToOrg(ctx, newGroup, organization.Organization{ID: grp.OrganizationID}); err != nil {
		return Group{}, err
	}

	return newGroup, nil
}

func (s Service) Get(ctx context.Context, idOrSlug string) (Group, error) {
	if uuid.IsValid(idOrSlug) {
		return s.repository.GetByID(ctx, idOrSlug)
	}
	return s.repository.GetBySlug(ctx, idOrSlug)
}

func (s Service) List(ctx context.Context, flt Filter) ([]Group, error) {
	return s.repository.List(ctx, flt)
}

func (s Service) Update(ctx context.Context, grp Group) (Group, error) {
	if strings.TrimSpace(grp.ID) != "" {
		return s.repository.UpdateByID(ctx, grp)
	}
	return s.repository.UpdateBySlug(ctx, grp)
}

func (s Service) ListUserGroups(ctx context.Context, userId string, roleId string) ([]Group, error) {
	return s.repository.ListUserGroups(ctx, userId, roleId)
}

func (s Service) ListUsers(ctx context.Context, idOrSlug string) ([]user.User, error) {
	if uuid.IsValid(idOrSlug) {
		return s.repository.ListUsersByGroupID(ctx, idOrSlug, role.DefinitionTeamMember.ID)
	}
	return s.repository.ListUsersByGroupSlug(ctx, idOrSlug, role.DefinitionTeamMember.ID)
}

func (s Service) ListAdmins(ctx context.Context, idOrSlug string) ([]user.User, error) {
	if uuid.IsValid(idOrSlug) {
		return s.repository.ListUsersByGroupID(ctx, idOrSlug, role.DefinitionTeamAdmin.ID)
	}
	return s.repository.ListUsersByGroupSlug(ctx, idOrSlug, role.DefinitionTeamAdmin.ID)
}

func (s Service) AddUsers(ctx context.Context, groupIdOrSlug string, userIds []string) ([]user.User, error) {
	// TODO(discussion): To be replaced with create relation
	return []user.User{}, nil
}

func (s Service) RemoveUser(ctx context.Context, groupIdOrSlug string, userId string) ([]user.User, error) {
	// TODO(discussion): To be replaced with remove relation
	return []user.User{}, nil
}

func (s Service) AddAdmins(ctx context.Context, groupIdOrSlug string, userIds []string) ([]user.User, error) {
	// TODO(discussion): To be replaced with create relation
	return []user.User{}, nil
}

func (s Service) RemoveAdmin(ctx context.Context, groupIdOrSlug string, userId string) ([]user.User, error) {
	// TODO(discussion): Can be removed for remove relation
	return []user.User{}, nil
}

func (s Service) addTeamToOrg(ctx context.Context, team Group, org organization.Organization) error {
	orgId := str.DefaultStringIfEmpty(org.ID, team.OrganizationID)
	rel := relation.RelationV2{
		Object: relation.Object{
			ID:          team.ID,
			NamespaceID: schema.GroupNamespace,
		},
		Subject: relation.Subject{
			ID:        orgId,
			Namespace: schema.OrganizationNamespace,
			RoleID:    schema.OrganizationNamespace,
		},
	}

	_, err := s.relationService.Create(ctx, rel)
	if err != nil {
		return err
	}

	return nil
}
