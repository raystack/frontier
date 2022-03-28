package group

import (
	"context"
	"errors"

	"github.com/odpf/shield/internal/bootstrap/definition"
	shieldError "github.com/odpf/shield/utils/errors"

	"github.com/odpf/shield/internal/authz"
	"github.com/odpf/shield/internal/permission"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store       Store
	Authz       *authz.Authz
	Permissions permission.Permissions
}

type Store interface {
	CreateGroup(ctx context.Context, grp model.Group) (model.Group, error)
	GetGroup(ctx context.Context, id string) (model.Group, error)
	ListGroups(ctx context.Context, org model.Organization) ([]model.Group, error)
	UpdateGroup(ctx context.Context, toUpdate model.Group) (model.Group, error)
	GetUsersByIds(ctx context.Context, userIds []string) ([]model.User, error)
	GetUser(ctx context.Context, userId string) (model.User, error)
	ListGroupUsers(ctx context.Context, groupId string, roleId string) ([]model.User, error)
	GetRelationByFields(ctx context.Context, relation model.Relation) (model.Relation, error)
	ListUserGroupRelations(ctx context.Context, userId string, groupId string) ([]model.Relation, error)
}

var (
	GroupDoesntExist = errors.New("group doesn't exist")
	InvalidUUID      = errors.New("invalid syntax of uuid")
)

func (s Service) CreateGroup(ctx context.Context, grp model.Group) (model.Group, error) {
	user, err := s.Permissions.FetchCurrentUser(ctx)

	if err != nil {
		return model.Group{}, err
	}

	newGroup, err := s.Store.CreateGroup(ctx, grp)

	if err != nil {
		return model.Group{}, err
	}

	err = s.Permissions.AddTeamToOrg(ctx, newGroup, model.Organization{Id: grp.OrganizationId})

	if err != nil {
		return model.Group{}, err
	}

	err = s.Permissions.AddAdminToTeam(ctx, user, newGroup)

	if err != nil {
		return model.Group{}, err
	}

	err = s.Permissions.AddMemberToTeam(ctx, user, newGroup)

	if err != nil {
		return model.Group{}, err
	}

	return newGroup, nil
}

func (s Service) GetGroup(ctx context.Context, id string) (model.Group, error) {
	return s.Store.GetGroup(ctx, id)
}

func (s Service) ListGroups(ctx context.Context, org model.Organization) ([]model.Group, error) {
	return s.Store.ListGroups(ctx, org)
}

func (s Service) UpdateGroup(ctx context.Context, grp model.Group) (model.Group, error) {
	return s.Store.UpdateGroup(ctx, grp)
}

func (s Service) AddUsersToGroup(ctx context.Context, groupId string, userIds []string) ([]model.User, error) {
	currentUser, err := s.Permissions.FetchCurrentUser(ctx)
	if err != nil {
		return []model.User{}, err
	}

	group, err := s.Store.GetGroup(ctx, groupId)

	if err != nil {
		return []model.User{}, err
	}

	isAuthorized, err := s.Permissions.CheckPermission(ctx, currentUser, model.Resource{
		Id:        groupId,
		Namespace: definition.TeamNamespace,
	},
		definition.ManageTeamAction,
	)

	if err != nil {
		return []model.User{}, err
	}

	if !isAuthorized {
		return []model.User{}, shieldError.Unauthorzied
	}

	users, err := s.Store.GetUsersByIds(ctx, userIds)

	if err != nil {
		return []model.User{}, err
	}

	for _, user := range users {
		err = s.Permissions.AddMemberToTeam(ctx, user, group)
		if err != nil {
			return []model.User{}, err
		}
	}
	return s.ListGroupUsers(ctx, groupId)
}

func (s Service) RemoveUserFromGroup(ctx context.Context, groupId string, userId string) ([]model.User, error) {
	currentUser, err := s.Permissions.FetchCurrentUser(ctx)
	if err != nil {
		return []model.User{}, err
	}

	group, err := s.Store.GetGroup(ctx, groupId)

	if err != nil {
		return []model.User{}, err
	}

	isAuthorized, err := s.Permissions.CheckPermission(ctx, currentUser, model.Resource{
		Id:        groupId,
		Namespace: definition.TeamNamespace,
	},
		definition.ManageTeamAction,
	)

	if err != nil {
		return []model.User{}, err
	}

	if !isAuthorized {
		return []model.User{}, shieldError.Unauthorzied
	}

	user, err := s.Store.GetUser(ctx, userId)

	if err != nil {
		return []model.User{}, err
	}

	relations, err := s.Store.ListUserGroupRelations(ctx, user.Id, group.Id)

	if err != nil {
		return []model.User{}, err
	}

	for _, rel := range relations {
		err = s.Permissions.RemoveRelation(ctx, rel)
		if err != nil {
			return []model.User{}, err
		}
	}

	return s.ListGroupUsers(ctx, groupId)
}

func (s Service) ListGroupUsers(ctx context.Context, groupId string) ([]model.User, error) {
	return s.Store.ListGroupUsers(ctx, groupId, definition.TeamMemberRole.Id)
}

func (s Service) ListGroupAdmins(ctx context.Context, groupId string) ([]model.User, error) {
	return s.Store.ListGroupUsers(ctx, groupId, definition.TeamAdminRole.Id)
}

func (s Service) AddAdminsToGroup(ctx context.Context, groupId string, userIds []string) ([]model.User, error) {
	currentUser, err := s.Permissions.FetchCurrentUser(ctx)
	if err != nil {
		return []model.User{}, err
	}

	group, err := s.Store.GetGroup(ctx, groupId)

	if err != nil {
		return []model.User{}, err
	}

	isAuthorized, err := s.Permissions.CheckPermission(ctx, currentUser, model.Resource{
		Id:        groupId,
		Namespace: definition.TeamNamespace,
	},
		definition.ManageTeamAction,
	)

	if err != nil {
		return []model.User{}, err
	}

	if !isAuthorized {
		return []model.User{}, shieldError.Unauthorzied
	}

	users, err := s.Store.GetUsersByIds(ctx, userIds)

	if err != nil {
		return []model.User{}, err
	}

	for _, user := range users {
		err = s.Permissions.AddMemberToTeam(ctx, user, group)
		if err != nil {
			return []model.User{}, err
		}

		err = s.Permissions.AddAdminToTeam(ctx, user, group)
		if err != nil {
			return []model.User{}, err
		}
	}
	return s.ListGroupAdmins(ctx, groupId)
}

func (s Service) RemoveAdminFromGroup(ctx context.Context, groupId string, userId string) ([]model.User, error) {
	currentUser, err := s.Permissions.FetchCurrentUser(ctx)
	if err != nil {
		return []model.User{}, err
	}

	group, err := s.Store.GetGroup(ctx, groupId)

	if err != nil {
		return []model.User{}, err
	}

	isAuthorized, err := s.Permissions.CheckPermission(ctx, currentUser, model.Resource{
		Id:        groupId,
		Namespace: definition.TeamNamespace,
	},
		definition.ManageTeamAction,
	)

	if err != nil {
		return []model.User{}, err
	}

	if !isAuthorized {
		return []model.User{}, shieldError.Unauthorzied
	}

	user, err := s.Store.GetUser(ctx, userId)

	if err != nil {
		return []model.User{}, err
	}

	err = s.Permissions.RemoveAdminFromTeam(ctx, user, group)
	if err != nil {
		return []model.User{}, err
	}

	return s.ListGroupAdmins(ctx, groupId)
}
