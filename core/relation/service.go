package relation

import (
	"context"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/user"
)

type Service struct {
	repository      Repository
	authzRepository AuthzRepository
}

func NewService(repository Repository, authzRepository AuthzRepository) *Service {
	return &Service{
		repository:      repository,
		authzRepository: authzRepository,
	}
}

func (s Service) Get(ctx context.Context, id string) (Relation, error) {
	return s.repository.Get(ctx, id)
}

func (s Service) Create(ctx context.Context, rel Relation) (Relation, error) {
	rel, err := s.repository.Create(ctx, rel)
	if err != nil {
		return Relation{}, err
	}

	if err = s.authzRepository.Add(ctx, rel); err != nil {
		return Relation{}, err
	}

	return rel, nil
}

func (s Service) List(ctx context.Context) ([]Relation, error) {
	return s.repository.List(ctx)
}

func (s Service) Update(ctx context.Context, id string, toUpdate Relation) (Relation, error) {
	oldRelation, err := s.repository.Get(ctx, id)
	if err != nil {
		return Relation{}, err
	}

	newRelation, err := s.repository.Update(ctx, id, toUpdate)
	if err != nil {
		return Relation{}, err
	}

	if err = s.authzRepository.Delete(ctx, oldRelation); err != nil {
		return Relation{}, err
	}

	if err = s.authzRepository.Add(ctx, newRelation); err != nil {
		return Relation{}, err
	}

	return newRelation, nil
}

func (s Service) Delete(ctx context.Context, rel Relation) error {
	fetchedRel, err := s.repository.GetByFields(ctx, rel)
	if err != nil {
		return err
	}

	if err = s.authzRepository.Delete(ctx, rel); err != nil {
		return err
	}

	return s.repository.DeleteByID(ctx, fetchedRel.ID)
}

func (s Service) CheckPermission(ctx context.Context, usr user.User, resourceNS namespace.Namespace, resourceIdxa string, action action.Action) (bool, error) {
	return s.authzRepository.Check(ctx, Relation{
		ObjectNamespace:  resourceNS,
		ObjectID:         resourceIdxa,
		SubjectID:        usr.ID,
		SubjectNamespace: namespace.DefinitionUser,
	}, action)
}

func (s Service) DeleteSubjectRelations(ctx context.Context, resourceType, optionalResourceID string) error {
	return s.authzRepository.DeleteSubjectRelations(ctx, resourceType, optionalResourceID)
}
