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
	user, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return Group{}, err
	}

	newGroup, err := s.repository.Create(ctx, grp)
	if err != nil {
		return Group{}, err
	}

	if err = s.addTeamToOrg(ctx, newGroup, organization.Organization{ID: grp.OrganizationID}); err != nil {
		return Group{}, err
	}

	if err = s.addAdminToTeam(ctx, user, newGroup); err != nil {
		return Group{}, err
	}

	if err = s.addMemberToTeam(ctx, user, newGroup); err != nil {
		return Group{}, err
	}

	return newGroup, nil
}

func (s Service) Get(ctx context.Context, id string) (Group, error) {
	return s.repository.Get(ctx, id)
}

func (s Service) List(ctx context.Context, org organization.Organization) ([]Group, error) {
	return s.repository.List(ctx, org)
}

func (s Service) Update(ctx context.Context, grp Group) (Group, error) {
	return s.repository.Update(ctx, grp)
}

func (s Service) AddUsers(ctx context.Context, groupId string, userIds []string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	groupId = strings.TrimSpace(groupId)
	group, err := s.repository.Get(ctx, groupId)

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

	users, err := s.userService.GetByIDs(ctx, userIds)
	if err != nil {
		return []user.User{}, err
	}

	for _, usr := range users {
		err = s.addMemberToTeam(ctx, usr, group)
		if err != nil {
			return []user.User{}, err
		}
	}
	return s.ListUsers(ctx, groupId)
}

func (s Service) RemoveUser(ctx context.Context, groupId string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	groupId = strings.TrimSpace(groupId)
	group, err := s.repository.Get(ctx, groupId)
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

	usr, err := s.userService.GetByID(ctx, userId)
	if err != nil {
		return []user.User{}, err
	}

	relations, err := s.repository.ListUserGroupRelations(ctx, usr.ID, group.ID)
	if err != nil {
		return []user.User{}, err
	}

	for _, rel := range relations {
		if err = s.relationService.Delete(ctx, rel); err != nil {
			return []user.User{}, err
		}
	}

	return s.ListUsers(ctx, groupId)
}

func (s Service) ListUserGroups(ctx context.Context, userId string, roleId string) ([]Group, error) {
	return s.repository.ListUserGroups(ctx, userId, roleId)
}

func (s Service) ListUsers(ctx context.Context, groupId string) ([]user.User, error) {
	return s.repository.ListUsers(ctx, groupId, role.DefinitionTeamMember.ID)
}

func (s Service) ListAdmins(ctx context.Context, groupId string) ([]user.User, error) {
	return s.repository.ListUsers(ctx, groupId, role.DefinitionTeamAdmin.ID)
}

func (s Service) AddAdmins(ctx context.Context, groupId string, userIds []string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	groupId = strings.TrimSpace(groupId)
	group, err := s.repository.Get(ctx, groupId)

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

	users, err := s.userService.GetByIDs(ctx, userIds)
	if err != nil {
		return []user.User{}, err
	}

	for _, usr := range users {
		err = s.addMemberToTeam(ctx, usr, group)
		if err != nil {
			return []user.User{}, err
		}

		err = s.addAdminToTeam(ctx, usr, group)
		if err != nil {
			return []user.User{}, err
		}
	}
	return s.ListAdmins(ctx, groupId)
}

func (s Service) RemoveAdmin(ctx context.Context, groupId string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	groupId = strings.TrimSpace(groupId)
	group, err := s.repository.Get(ctx, groupId)
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

	usr, err := s.userService.GetByID(ctx, userId)
	if err != nil {
		return []user.User{}, err
	}

	err = s.RemoveAdminFromTeam(ctx, usr, group)
	if err != nil {
		return []user.User{}, err
	}

	return s.ListAdmins(ctx, groupId)
}

func (s Service) addTeamToOrg(ctx context.Context, team Group, org organization.Organization) error {
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

func (s Service) addAdminToTeam(ctx context.Context, user user.User, team Group) error {
	rel := s.GetTeamAdminRelation(user, team)
	_, err := s.relationService.Create(ctx, rel)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) addMemberToTeam(ctx context.Context, user user.User, team Group) error {
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

func (s Service) removeMemberFromTeam(ctx context.Context, user user.User, team Group) error {
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
