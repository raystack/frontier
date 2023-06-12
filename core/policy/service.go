package policy

import (
	"context"

	"github.com/raystack/shield/core/relation"
	"github.com/raystack/shield/internal/bootstrap/schema"
)

type RelationService interface {
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type Service struct {
	repository      Repository
	relationService RelationService
}

func NewService(repository Repository, relationService RelationService) *Service {
	return &Service{
		repository:      repository,
		relationService: relationService,
	}
}

func (s Service) Get(ctx context.Context, id string) (Policy, error) {
	return s.repository.Get(ctx, id)
}

func (s Service) List(ctx context.Context, f Filter) ([]Policy, error) {
	return s.repository.List(ctx, f)
}

func (s Service) Create(ctx context.Context, policy Policy) (Policy, error) {
	pol, err := s.repository.Upsert(ctx, policy)
	if err != nil {
		return Policy{}, err
	}
	policy.ID = pol

	if err = s.AssignRole(ctx, policy); err != nil {
		return policy, err
	}
	return policy, err
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

// AssignRole Note: ideally this should be in a single transaction
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
