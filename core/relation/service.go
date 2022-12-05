package relation

import (
	"context"
	"fmt"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/schema"
)

type Service struct {
	repository      Repository
	authzRepository AuthzRepository
	roleService     RoleService
	userService     UserService
}

func NewService(repository Repository, authzRepository AuthzRepository, roleService RoleService, userService UserService) *Service {
	return &Service{
		repository:      repository,
		authzRepository: authzRepository,
		roleService:     roleService,
		userService:     userService,
	}
}

func (s Service) Get(ctx context.Context, id string) (RelationV2, error) {
	return s.repository.Get(ctx, id)
}

func (s Service) Create(ctx context.Context, rel RelationV2) (RelationV2, error) {
	// If Principal is a user, then we will get ID for that user as Subject.ID
	if rel.Subject.Namespace == schema.UserPrincipal {
		fetchedUser, err := s.userService.GetByEmail(ctx, rel.Subject.ID)
		if err != nil {
			return RelationV2{}, fmt.Errorf("%w: %s", ErrFetchingUser, err.Error())
		}

		rel.Subject.ID = fetchedUser.ID
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

func (s Service) Delete(ctx context.Context, rel Relation) error {
	//fetchedRel, err := s.repository.GetByFields(ctx, rel)
	//if err != nil {
	//	return err
	//}
	//
	//if err = s.authzRepository.Delete(ctx, rel); err != nil {
	//	return err
	//}
	//
	//return s.repository.DeleteByID(ctx, fetchedRel.ID)
	return nil
}

func (s Service) DeleteV2(ctx context.Context, rel RelationV2) error {
	fmt.Printf("rel: %v\n", rel)
	fetchedRel, err := s.repository.GetByFields(ctx, rel)
	if err != nil {
		return err
	}
	fmt.Printf("fetchedRel: %v\n", fetchedRel)
	if err := s.authzRepository.DeleteV2(ctx, fetchedRel); err != nil {
		fmt.Printf("err authzrepo delete: %v\n", err)
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
