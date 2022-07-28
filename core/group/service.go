package group

import (
	"context"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/errors"
	"github.com/odpf/shield/pkg/str"
	"github.com/odpf/shield/pkg/uuid"
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

	if err = s.addAdminToTeam(ctx, user.ID, newGroup.ID); err != nil {
		return Group{}, err
	}

	if err = s.addMemberToTeam(ctx, user.ID, newGroup.ID); err != nil {
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
	if grp.ID != "" {
		return s.repository.UpdateByID(ctx, grp)
	}
	return s.repository.UpdateBySlug(ctx, grp)
}

func (s Service) AddUsers(ctx context.Context, groupIdOrSlug string, userIds []string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	var groupID = groupIdOrSlug
	if !uuid.IsValid(groupIdOrSlug) {
		grp, err := s.repository.GetBySlug(ctx, groupIdOrSlug)
		if err != nil {
			return []user.User{}, err
		}
		groupID = grp.ID
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionTeam, groupID, action.DefinitionManageTeam)
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
		err = s.addMemberToTeam(ctx, usr.ID, groupID)
		if err != nil {
			return []user.User{}, err
		}
	}

	return s.ListUsers(ctx, groupID)
}

func (s Service) RemoveUser(ctx context.Context, groupIdOrSlug string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	var groupID = groupIdOrSlug
	if !uuid.IsValid(groupIdOrSlug) {
		grp, err := s.repository.GetBySlug(ctx, groupIdOrSlug)
		if err != nil {
			return []user.User{}, err
		}
		groupID = grp.ID
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionTeam, groupID, action.DefinitionManageTeam)
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

	relations, err := s.repository.ListUserGroupIDRelations(ctx, usr.ID, groupID)
	if err != nil {
		return []user.User{}, err
	}

	for _, rel := range relations {
		if err = s.relationService.Delete(ctx, rel); err != nil {
			return []user.User{}, err
		}
	}

	return s.ListUsers(ctx, groupID)
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

func (s Service) AddAdmins(ctx context.Context, groupIdOrSlug string, userIds []string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	var groupID = groupIdOrSlug
	if !uuid.IsValid(groupIdOrSlug) {
		grp, err := s.repository.GetBySlug(ctx, groupIdOrSlug)
		if err != nil {
			return []user.User{}, err
		}
		groupID = grp.ID
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionTeam, groupID, action.DefinitionManageTeam)
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
		err = s.addMemberToTeam(ctx, usr.ID, groupID)
		if err != nil {
			return []user.User{}, err
		}

		err = s.addAdminToTeam(ctx, usr.ID, groupID)
		if err != nil {
			return []user.User{}, err
		}
	}
	return s.ListAdmins(ctx, groupID)
}

func (s Service) RemoveAdmin(ctx context.Context, groupIdOrSlug string, userId string) ([]user.User, error) {
	currentUser, err := s.userService.FetchCurrentUser(ctx)
	if err != nil {
		return []user.User{}, err
	}

	var groupID = groupIdOrSlug
	if !uuid.IsValid(groupIdOrSlug) {
		grp, err := s.repository.GetBySlug(ctx, groupIdOrSlug)
		if err != nil {
			return []user.User{}, err
		}
		groupID = grp.ID
	}

	isAuthorized, err := s.relationService.CheckPermission(ctx, currentUser, namespace.DefinitionTeam, groupID, action.DefinitionManageTeam)
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

	err = s.removeAdminFromTeam(ctx, usr.ID, groupID)
	if err != nil {
		return []user.User{}, err
	}

	return s.ListAdmins(ctx, groupID)
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

func (s Service) addAdminToTeam(ctx context.Context, userID, groupID string) error {
	rel := s.getTeamAdminRelation(userID, groupID)
	_, err := s.relationService.Create(ctx, rel)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) addMemberToTeam(ctx context.Context, userID, groupID string) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionTeam,
		ObjectID:         groupID,
		SubjectID:        userID,
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

func (s Service) removeMemberFromTeam(ctx context.Context, userID, groupID string) error {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionTeam,
		ObjectID:         groupID,
		SubjectID:        userID,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			ID:        role.DefinitionTeamMember.ID,
			Namespace: namespace.DefinitionTeam,
		},
	}
	return s.relationService.Delete(ctx, rel)
}

func (s Service) getTeamAdminRelation(userID, groupID string) relation.Relation {
	rel := relation.Relation{
		ObjectNamespace:  namespace.DefinitionTeam,
		ObjectID:         groupID,
		SubjectID:        userID,
		SubjectNamespace: namespace.DefinitionUser,
		Role: role.Role{
			ID:        role.DefinitionTeamAdmin.ID,
			Namespace: namespace.DefinitionTeam,
		},
	}
	return rel
}

func (s Service) removeAdminFromTeam(ctx context.Context, userID, groupID string) error {
	rel := s.getTeamAdminRelation(userID, groupID)
	return s.relationService.Delete(ctx, rel)
}
