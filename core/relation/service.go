package relation

import (
	"context"
	"errors"
	"fmt"

	shielduuid "github.com/odpf/shield/pkg/uuid"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
)

type Service struct {
	repository      Repository
	authzRepository AuthzRepository
	roleService     RoleService
}

func NewService(repository Repository, authzRepository AuthzRepository, roleService RoleService) *Service {
	return &Service{
		repository:      repository,
		authzRepository: authzRepository,
		roleService:     roleService,
	}
}

func (s Service) Get(ctx context.Context, id string) (RelationV2, error) {
	return s.repository.Get(ctx, id)
}

func (s Service) Create(ctx context.Context, rel RelationV2) (RelationV2, error) {
	if !shielduuid.IsValid(rel.Object.ID) || !shielduuid.IsValid(rel.Subject.ID) {
		return RelationV2{}, errors.New("subject/object id should be a valid uuid")
	}

	createdRelation, err := s.repository.Create(ctx, rel)
	if err != nil {
		return RelationV2{}, fmt.Errorf("%w: %s", ErrCreatingRelationInStore, err.Error())
	}

	err = s.authzRepository.AddV2(ctx, createdRelation)
	if err != nil {
		return RelationV2{}, fmt.Errorf("%w: %s", ErrCreatingRelationInAuthzEngine, err.Error())
	}

	return createdRelation, nil
}

func (s Service) List(ctx context.Context) ([]RelationV2, error) {
	return s.repository.List(ctx)
}

// TODO: Update & Delete planned for v0.6
func (s Service) Update(ctx context.Context, toUpdate Relation) (Relation, error) {
	//oldRelation, err := s.repository.Get(ctx, toUpdate.ID)
	//if err != nil {
	//	return Relation{}, err
	//}
	//
	//newRelation, err := s.repository.Update(ctx, toUpdate)
	//if err != nil {
	//	return Relation{}, err
	//}
	//
	//if err = s.authzRepository.Delete(ctx, oldRelation); err != nil {
	//	return Relation{}, err
	//}
	//
	//if err = s.authzRepository.Add(ctx, newRelation); err != nil {
	//	return Relation{}, err
	//}
	//
	//return newRelation, nil
	return Relation{}, nil
}

func (s Service) GetRelationByFields(ctx context.Context, rel RelationV2) (RelationV2, error) {
	fetchedRel, err := s.repository.GetByFields(ctx, rel)
	if err != nil {
		return RelationV2{}, err
	}

	return fetchedRel, nil
}

func (s Service) Delete(ctx context.Context, rel RelationV2) error {
	fetchedRel, err := s.repository.GetByFields(ctx, rel)
	if err != nil {
		return err
	}
	if err := s.authzRepository.DeleteV2(ctx, fetchedRel); err != nil {
		return err
	}

	return s.repository.DeleteByID(ctx, fetchedRel.ID)
}

func (s Service) CheckPermission(ctx context.Context, userID string, resourceNS namespace.Namespace, resourceIdx string, action action.Action) (bool, error) {
	return s.authzRepository.Check(ctx, Relation{
		ObjectNamespace:  resourceNS,
		ObjectID:         resourceIdx,
		SubjectID:        userID,
		SubjectNamespace: namespace.DefinitionUser,
	}, action)
}

func (s Service) DeleteSubjectRelations(ctx context.Context, resourceType, optionalResourceID string) error {
	return s.authzRepository.DeleteSubjectRelations(ctx, resourceType, optionalResourceID)
}

// LookupSubjects returns all the subjects of a given type that have access whether
// via a computed permission or relation membership.
func (s Service) LookupSubjects(ctx context.Context, rel RelationV2) ([]string, error) {
	return s.authzRepository.LookupSubjects(ctx, rel)
}

// LookupResources returns all the resources of a given type that a subject can access whether
// via a computed permission or relation membership.
func (s Service) LookupResources(ctx context.Context, rel RelationV2) ([]string, error) {
	return s.authzRepository.LookupResources(ctx, rel)
}

// ListRelations lists a set of the relationships matching filter
func (s Service) ListRelations(ctx context.Context, rel RelationV2) ([]RelationV2, error) {
	return s.authzRepository.ListRelations(ctx, rel)
}
