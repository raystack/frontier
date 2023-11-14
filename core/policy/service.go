package policy

import (
	"context"
	"errors"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/role"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type RoleService interface {
	Get(ctx context.Context, id string) (role.Role, error)
	List(ctx context.Context, f role.Filter) ([]role.Role, error)
}

type Service struct {
	repository      Repository
	relationService RelationService
	roleService     RoleService
}

func NewService(repository Repository, relationService RelationService, roleService RoleService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationService,
		roleService:     roleService,
	}
}

func (s Service) Get(ctx context.Context, id string) (Policy, error) {
	return s.repository.Get(ctx, id)
}

func (s Service) List(ctx context.Context, f Filter) ([]Policy, error) {
	return s.repository.List(ctx, f)
}

func (s Service) Create(ctx context.Context, policy Policy) (Policy, error) {
	// check if role exists and get its ID if it was passed by name
	policyRole, err := s.roleService.Get(ctx, policy.RoleID)
	if err != nil {
		return Policy{}, err
	}
	policy.RoleID = policyRole.ID

	createdPolicy, err := s.repository.Upsert(ctx, policy)
	if err != nil {
		return Policy{}, err
	}

	if err = s.AssignRole(ctx, createdPolicy); err != nil {
		return createdPolicy, err
	}
	return createdPolicy, err
}

func (s Service) Delete(ctx context.Context, id string) error {
	if err := s.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			ID:        id,
			Namespace: schema.RoleBindingNamespace,
		},
	}); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}

func (s Service) Replace(ctx context.Context, existingID string, pol Policy) (Policy, error) {
	if err := s.Delete(ctx, existingID); err != nil && !errors.Is(err, ErrNotExist) {
		return Policy{}, err
	}
	pol.ID = existingID
	return s.Create(ctx, pol)
}

// AssignRole Note: ideally this should be in a single transaction
// read more about how user defined roles work in spicedb https://authzed.com/blog/user-defined-roles
func (s Service) AssignRole(ctx context.Context, pol Policy) error {
	// bind role with user
	subjectSubRelation := ""
	if pol.PrincipalType == schema.GroupPrincipal {
		subjectSubRelation = schema.MemberRelationName
	}
	_, err := s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        pol.ID,
			Namespace: schema.RoleBindingNamespace,
		},
		Subject: relation.Subject{
			ID:              pol.PrincipalID,
			Namespace:       pol.PrincipalType,
			SubRelationName: subjectSubRelation,
		},
		RelationName: schema.RoleBearerRelationName,
	})
	if err != nil {
		return err
	}
	_, err = s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        pol.ID,
			Namespace: schema.RoleBindingNamespace,
		},
		Subject: relation.Subject{
			ID:        pol.RoleID,
			Namespace: schema.RoleNamespace,
		},
		RelationName: schema.RoleRelationName,
	})
	if err != nil {
		return err
	}

	// bind policy to resource
	_, err = s.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        pol.ResourceID,
			Namespace: pol.ResourceType,
		},
		Subject: relation.Subject{
			ID:        pol.ID,
			Namespace: schema.RoleBindingNamespace,
		},
		RelationName: schema.RoleGrantRelationName,
	})
	if err != nil {
		return err
	}
	return nil
}

// ListRoles lists roles assigned via policies to a user
func (s Service) ListRoles(ctx context.Context, principalType, principalID, objectNamespace, objectID string) ([]role.Role, error) {
	flt := Filter{
		PrincipalType: principalType,
		PrincipalID:   principalID,
	}
	switch objectNamespace {
	case schema.OrganizationNamespace:
		flt.OrgID = objectID
	case schema.ProjectNamespace:
		flt.ProjectID = objectID
	case schema.GroupNamespace:
		flt.GroupID = objectID
	}
	policies, err := s.List(ctx, flt)
	if err != nil {
		return nil, err
	}

	roleIDs := utils.Map(policies, func(p Policy) string {
		return p.RoleID
	})
	return s.roleService.List(ctx, role.Filter{
		IDs: roleIDs,
	})
}
