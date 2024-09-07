package user

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/raystack/frontier/core/role"

	"github.com/raystack/frontier/core/policy"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/str"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	BatchCheckPermission(ctx context.Context, relations []relation.Relation) ([]relation.CheckPair, error)
	Delete(ctx context.Context, rel relation.Relation) error
	LookupSubjects(ctx context.Context, rel relation.Relation) ([]string, error)
	LookupResources(ctx context.Context, rel relation.Relation) ([]string, error)
}

type PolicyService interface {
	List(ctx context.Context, f policy.Filter) ([]policy.Policy, error)
}

type RoleService interface {
	List(ctx context.Context, f role.Filter) ([]role.Role, error)
}

type Service struct {
	repository      Repository
	relationService RelationService
	policyService   PolicyService
	roleService     RoleService
	Now             func() time.Time
}

func NewService(repository Repository, relationRepo RelationService,
	policyService PolicyService, roleService RoleService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationRepo,
		policyService:   policyService,
		roleService:     roleService,
		Now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

// GetByID email or slug
func (s Service) GetByID(ctx context.Context, id string) (User, error) {
	if isValidEmail(id) {
		return s.GetByEmail(ctx, id)
	}
	if utils.IsValidUUID(id) {
		return s.repository.GetByID(ctx, id)
	}
	return s.repository.GetByName(ctx, strings.ToLower(id))
}

func (s Service) GetByIDs(ctx context.Context, userIDs []string) ([]User, error) {
	return s.repository.GetByIDs(ctx, userIDs)
}

func (s Service) GetByEmail(ctx context.Context, email string) (User, error) {
	email = strings.ToLower(email)
	return s.repository.GetByEmail(ctx, email)
}

func (s Service) Create(ctx context.Context, user User) (User, error) {
	return s.repository.Create(ctx, User{
		Name:     strings.ToLower(user.Name),
		Email:    strings.ToLower(user.Email),
		State:    Enabled,
		Avatar:   user.Avatar,
		Title:    user.Title,
		Metadata: user.Metadata,
	})
}

func (s Service) List(ctx context.Context, flt Filter) ([]User, error) {
	if flt.OrgID != "" {
		return s.ListByOrg(ctx, flt.OrgID, "")
	}
	if flt.GroupID != "" {
		return s.ListByGroup(ctx, flt.GroupID, "")
	}

	// state gets filtered in db
	return s.repository.List(ctx, flt)
}

// Update by user uuid, email or slug
// Note(kushsharma): we don't actually update email field of the user, if we want to support it
// one security concern is that we need to ensure users can't misuse it to takeover
// invitations created for other users.
func (s Service) Update(ctx context.Context, toUpdate User) (User, error) {
	id := toUpdate.ID
	toUpdate.Email = strings.ToLower(toUpdate.Email)
	toUpdate.Name = strings.ToLower(toUpdate.Name)
	if isValidEmail(id) {
		return s.UpdateByEmail(ctx, toUpdate)
	}
	if utils.IsValidUUID(id) {
		return s.repository.UpdateByID(ctx, toUpdate)
	}
	return s.repository.UpdateByName(ctx, toUpdate)
}

func (s Service) UpdateByEmail(ctx context.Context, toUpdate User) (User, error) {
	toUpdate.Email = strings.ToLower(toUpdate.Email)
	toUpdate.Name = strings.ToLower(toUpdate.Name)
	return s.repository.UpdateByEmail(ctx, toUpdate)
}

func (s Service) Enable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Enabled)
}

func (s Service) Disable(ctx context.Context, id string) error {
	return s.repository.SetState(ctx, id, Disabled)
}

// Delete by user uuid
// don't call this directly, use cascade deleter
func (s Service) Delete(ctx context.Context, id string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{Subject: relation.Subject{
		ID:        id,
		Namespace: schema.UserPrincipal,
	}}); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}

func (s Service) ListByOrg(ctx context.Context, orgID string, roleFilter string) ([]User, error) {
	policies, err := s.policyService.List(ctx, policy.Filter{
		OrgID: orgID,
	})
	if err != nil {
		return nil, err
	}

	userIDs, err := s.getUserIDsFromPolicies(ctx, policies, roleFilter)
	if err != nil {
		return nil, err
	}

	if len(userIDs) == 0 {
		// no users
		return []User{}, nil
	}

	return s.repository.GetByIDs(ctx, userIDs)
}

func (s Service) getUserIDsFromPolicies(ctx context.Context, policies []policy.Policy, roleFilter string) ([]string, error) {
	var roles []role.Role
	var err error

	if roleFilter != "" {
		roleIDs := utils.Map(policies, func(pol policy.Policy) string {
			return pol.RoleID
		})
		roles, err = s.roleService.List(ctx, role.Filter{
			IDs: roleIDs,
		})
		if err != nil {
			return nil, err
		}
	}

	userIDs := make([]string, 0)
	for _, pol := range policies {
		// get only all users with the permission
		if pol.PrincipalType != schema.UserPrincipal {
			continue
		}

		if roleFilter != "" {
			for _, currentRole := range roles {
				if currentRole.ID == pol.RoleID && currentRole.Name == roleFilter {
					userIDs = append(userIDs, pol.PrincipalID)
				}
			}
		} else {
			userIDs = append(userIDs, pol.PrincipalID)
		}
	}
	return userIDs, nil
}

func (s Service) ListByGroup(ctx context.Context, groupID string, roleFilter string) ([]User, error) {
	policies, err := s.policyService.List(ctx, policy.Filter{
		GroupID: groupID,
	})
	if err != nil {
		return nil, err
	}

	userIDs, err := s.getUserIDsFromPolicies(ctx, policies, roleFilter)
	if err != nil {
		return nil, err
	}

	if len(userIDs) == 0 {
		// no users
		return []User{}, nil
	}

	return s.repository.GetByIDs(ctx, userIDs)
}

// Sudo add platform permissions to user
func (s Service) Sudo(ctx context.Context, id string, relationName string) error {
	currentUser, err := s.GetByID(ctx, id)
	if errors.Is(err, ErrNotExist) {
		if isValidEmail(id) {
			// create a new user
			currentUser, err = s.Create(ctx, User{
				Email: id,
				Name:  str.GenerateUserSlug(id),
			})
			if err != nil {
				return err
			}
		} else {
			// skip
			return nil
		}
	}
	if err != nil {
		return err
	}

	// check if already su
	permissionName := ""
	switch relationName {
	case schema.MemberRelationName:
		permissionName = schema.PlatformCheckPermission
	case schema.AdminRelationName:
		permissionName = schema.PlatformSudoPermission
	}
	if permissionName == "" {
		return fmt.Errorf("invalid relation name, possible options are: %s, %s", schema.MemberRelationName, schema.AdminRelationName)
	}

	if ok, err := s.IsSudo(ctx, currentUser.ID, permissionName); err != nil {
		return err
	} else if ok {
		return nil
	}

	// mark su
	_, err = s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
		Subject: relation.Subject{
			ID:        currentUser.ID,
			Namespace: schema.UserPrincipal,
		},
		RelationName: relationName,
	})
	return err
}

// UnSudo remove platform permissions to user
// only remove the 'member' relation if it exists
func (s Service) UnSudo(ctx context.Context, id string) error {
	currentUser, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	relationName := schema.MemberRelationName
	// to check if the user has member relation, we need to check if the user has `check`
	// permission on platform
	if ok, err := s.IsSudo(ctx, currentUser.ID, schema.PlatformCheckPermission); err != nil {
		return err
	} else if !ok {
		// not needed
		return nil
	}

	// unmark su
	err = s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        schema.PlatformID,
			Namespace: schema.PlatformNamespace,
		},
		Subject: relation.Subject{
			ID:        currentUser.ID,
			Namespace: schema.UserPrincipal,
		},
		RelationName: relationName,
	})
	return err
}

// IsSudo checks platform permissions.
// Platform permissions are:
// - superuser
// - check
func (s Service) IsSudo(ctx context.Context, id string, permissionName string) (bool, error) {
	status, err := s.IsSudos(ctx, []string{id}, permissionName)
	if err != nil {
		return false, err
	}
	return len(status) > 0, nil
}

func (s Service) IsSudos(ctx context.Context, ids []string, permissionName string) ([]relation.Relation, error) {
	relations := utils.Map(ids, func(id string) relation.Relation {
		return relation.Relation{
			Subject: relation.Subject{
				ID:        id,
				Namespace: schema.UserPrincipal,
			},
			Object: relation.Object{
				ID:        schema.PlatformID,
				Namespace: schema.PlatformNamespace,
			},
			RelationName: permissionName,
		}
	})
	statusForIDs, err := s.relationService.BatchCheckPermission(ctx, relations)
	if err != nil {
		return nil, err
	}

	successChecks := utils.Filter(statusForIDs, func(pair relation.CheckPair) bool {
		return pair.Status
	})
	return utils.Map(successChecks, func(pair relation.CheckPair) relation.Relation {
		return pair.Relation
	}), nil
}

func isValidEmail(str string) bool {
	_, err := mail.ParseAddress(str)
	return err == nil
}
