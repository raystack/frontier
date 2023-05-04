package relation

import (
	"context"
	"errors"
	"fmt"

	"github.com/odpf/shield/internal/bootstrap/schema"
	shielduuid "github.com/odpf/shield/pkg/uuid"
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

func (s Service) Get(ctx context.Context, id string) (RelationV2, error) {
	return s.repository.Get(ctx, id)
}

func (s Service) Create(ctx context.Context, rel RelationV2) (RelationV2, error) {
	if !isValidID(rel.Object.ID) || !isValidID(rel.Subject.ID) {
		return RelationV2{}, errors.New("subject/object id should be a valid uuid or a wildcard *")
	}

	createdRelation, err := s.repository.Upsert(ctx, rel)
	if err != nil {
		return RelationV2{}, fmt.Errorf("%w: %s", ErrCreatingRelationInStore, err.Error())
	}

	err = s.authzRepository.Add(ctx, createdRelation)
	if err != nil {
		return RelationV2{}, fmt.Errorf("%w: %s", ErrCreatingRelationInAuthzEngine, err.Error())
	}

	return createdRelation, nil
}

func (s Service) List(ctx context.Context) ([]RelationV2, error) {
	return s.repository.List(ctx)
}

func (s Service) GetRelationsByFields(ctx context.Context, rel RelationV2) ([]RelationV2, error) {
	return s.repository.GetByFields(ctx, rel)
}

func (s Service) Delete(ctx context.Context, rel RelationV2) error {
	fetchedRels, err := s.GetRelationsByFields(ctx, rel)
	if err != nil {
		return err
	}

	for _, fetchedRel := range fetchedRels {
		if err = s.authzRepository.Delete(ctx, fetchedRel); err != nil {
			return err
		}
		if err = s.repository.DeleteByID(ctx, fetchedRel.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) CheckPermission(ctx context.Context, subject Subject, object Object, permissionName string) (bool, error) {
	return s.authzRepository.Check(ctx, RelationV2{
		Object:  object,
		Subject: subject,
	}, permissionName)
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

func isValidID(id string) bool {
	if id == "*" || id == schema.PlatformID {
		// check either wildcard or global id
		return true
	}
	if shielduuid.IsValid(id) {
		return true
	}
	return false
}
