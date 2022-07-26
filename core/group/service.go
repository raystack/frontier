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
	"github.com/odpf/shield/pkg/errors"
	"github.com/odpf/shield/pkg/str"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
	CheckPermission(ctx context.Context, usr user.User, resourceNS namespace.Namespace, resourceIdxa string, action action.Action) (bool, error)
}

type UserService interface {
	FetchCurrentUser(ctx context.Context) (user.User, error)
}

type Service struct {
	store           Store
	relationService RelationService
	userService     UserService
}

func NewService(store Store, relationService RelationService, userService UserService) *Service {
	return &Service{
		store:           store,
		relationService: relationService,
		userService:     userService,
	}
}

func (s Service) CreateGroup(ctx context.Context, grp Group) (Group, error) {
	user, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return Group{}, err
	}

	newGroup, err := s.store.CreateGroup(ctx, grp)
	if err != nil {
		return Group{}, err
	}

	err = s.AddTeamToOrg(ctx, newGroup, organization.Organization{ID: grp.OrganizationID})
	if err != nil {
		return Group{}, err
	}

	err = s.AddAdminToTeam(ctx, user, newGroup)
	if err != nil {
		return Group{}, err
	}

	err = s.AddMemberToTeam(ctx, user, newGroup)
	if err != nil {
		return Group{}, err
	}

	return newGroup, nil
}

func (s Service) GetGroup(ctx context.Context, id string) (Group, error) {
	return s.store.GetGroup(ctx, id)
}

func (s Service) ListGroups(ctx context.Context, org organization.Organization) ([]Group, error) {
	return s.store.ListGroups(ctx, org)
}

func (s Service) UpdateGroup(ctx context.Context, grp Group) (Group, error) {
	return s.store.UpdateGroup(ctx, grp)
}

func (s Service) AddUsersToGroup(ctx context.Context, groupId string, userIds []string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	groupId = strings.TrimSpace(groupId)
	group, err := s.store.GetGroup(ctx, groupId)

	if err != nil {
		return []user.User{}, err
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionTeam, group.ID, action.DefinitionManageTeam)
	if err != nil {
		return []user.User{}, err
	}

	if !isAuthorized {
		return []user.User{}, errors.Unauthorized
	}

	users, err := s.store.GetUsersByIDs(ctx, userIds)
	if err != nil {
		return []user.User{}, err
	}

	for _, usr := range users {
		err = s.AddMemberToTeam(ctx, usr, group)
		if err != nil {
			return []user.User{}, err
		}
	}
	return s.ListGroupUsers(ctx, groupId)
}

func (s Service) RemoveUserFromGroup(ctx context.Context, groupId string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	groupId = strings.TrimSpace(groupId)
	group, err := s.store.GetGroup(ctx, groupId)
	if err != nil {
		return []user.User{}, err
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionTeam, group.ID, action.DefinitionManageTeam)

	if err != nil {
		return []user.User{}, err
	}

	if !isAuthorized {
		return []user.User{}, errors.Unauthorized
	}

	usr, err := s.store.GetUser(ctx, userId)
	if err != nil {
		return []user.User{}, err
	}

	relations, err := s.store.ListUserGroupRelations(ctx, usr.ID, group.ID)
	if err != nil {
		return []user.User{}, err
	}

	for _, rel := range relations {
		if err = s.relationService.Delete(ctx, rel); err != nil {
			return []user.User{}, err
		}
	}

	return s.ListGroupUsers(ctx, groupId)
}

func (s Service) ListUserGroups(ctx context.Context, userId string, roleId string) ([]Group, error) {
	return s.store.ListUserGroups(ctx, userId, roleId)
}

func (s Service) ListGroupUsers(ctx context.Context, groupId string) ([]user.User, error) {
	return s.store.ListGroupUsers(ctx, groupId, role.DefinitionTeamMember.ID)
}

func (s Service) ListGroupAdmins(ctx context.Context, groupId string) ([]user.User, error) {
	return s.store.ListGroupUsers(ctx, groupId, role.DefinitionTeamAdmin.ID)
}

func (s Service) AddAdminsToGroup(ctx context.Context, groupId string, userIds []string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	groupId = strings.TrimSpace(groupId)
	group, err := s.store.GetGroup(ctx, groupId)

	if err != nil {
		return []user.User{}, err
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionTeam, group.ID, action.DefinitionManageTeam)
	if err != nil {
		return []user.User{}, err
	}

	if !isAuthorized {
		return []user.User{}, errors.Unauthorized
	}

	users, err := s.store.GetUsersByIDs(ctx, userIds)

	if err != nil {
		return []user.User{}, err
	}

	for _, usr := range users {
		err = s.AddMemberToTeam(ctx, usr, group)
		if err != nil {
			return []user.User{}, err
		}

		err = s.AddAdminToTeam(ctx, usr, group)
		if err != nil {
			return []user.User{}, err
		}
	}
	return s.ListGroupAdmins(ctx, groupId)
}

func (s Service) RemoveAdminFromGroup(ctx context.Context, groupId string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	groupId = strings.TrimSpace(groupId)
	group, err := s.store.GetGroup(ctx, groupId)
	if err != nil {
		return []user.User{}, err
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionTeam, group.ID, action.DefinitionManageTeam)
	if err != nil {
		return []user.User{}, err
	}

	if !isAuthorized {
		return []user.User{}, errors.Unauthorized
	}

	usr, err := s.store.GetUser(ctx, userId)
	if err != nil {
		return []user.User{}, err
	}

	err = s.RemoveAdminFromTeam(ctx, usr, group)
	if err != nil {
		return []user.User{}, err
	}

	return s.ListGroupAdmins(ctx, groupId)
}

func (s Service) AddTeamToOrg(ctx context.Context, team Group, org organization.Organization) error {
	orgId := str.DefaultStringIfEmpty(org.ID, team.OrganizationID)
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionTeam,
		ObjectID:         team.ID,
		SubjectID:        orgId,
		SubjectNamespace: namespace.DefinitionOrg,
		Role: role.Role{
			ID:        namespace.DefinitionOrg.ID,
			Namespace: namespace.DefinitionTeam,
		},
		RelationType: relation.RelationTypes.Namespace,
	}

	_, err := s.relationService.Create(ctx, rel)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) AddAdminToTeam(ctx context.Context, user user.User, team Group) error {
	rel := s.GetTeamAdminRelation(user, team)
	_, err := s.relationService.Create(ctx, rel)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) AddMemberToTeam(ctx context.Context, user user.User, team Group) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionTeam,
		ObjectID:         team.ID,
		SubjectID:        user.ID,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			ID:        role.DefinitionTeamMember.ID,
			Namespace: namespace.DefinitionTeam,
		},
	}
	_, err := s.relationService.Create(ctx, rel)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) RemoveMemberFromTeam(ctx context.Context, user user.User, team Group) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionTeam,
		ObjectID:         team.ID,
		SubjectID:        user.ID,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			ID:        role.DefinitionTeamMember.ID,
			Namespace: namespace.DefinitionTeam,
		},
	}
	return s.relationService.Delete(ctx, rel)
}

func (s Service) GetTeamAdminRelation(user user.User, team Group) relation.Relation {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionTeam,
		ObjectID:         team.ID,
		SubjectID:        user.ID,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			ID:        role.DefinitionTeamAdmin.ID,
			Namespace: namespace.DefinitionTeam,
		},
	}
	return rel
}

func (s Service) RemoveAdminFromTeam(ctx context.Context, user user.User, team Group) error {
	rel := s.GetTeamAdminRelation(user, team)
	return s.relationService.Delete(ctx, rel)
}
