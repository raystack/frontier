package relation

import (
	"context"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/user"
)

type Service struct {
	store      Store
	authzStore AuthzStore
}

func NewService(store Store, authzStore AuthzStore) *Service {
	return &Service{
		store:      store,
		authzStore: authzStore,
	}
}

func (s Service) Get(ctx context.Context, id string) (Relation, error) {
	return s.store.GetRelation(ctx, id)
}

func (s Service) Create(ctx context.Context, rel Relation) (Relation, error) {
	rel, err := s.store.CreateRelation(ctx, rel)
	if err != nil {
		return Relation{}, err
	}

	err = s.authzStore.AddRelation(ctx, rel)
	if err != nil {
		return Relation{}, err
	}

	return rel, nil
}

func (s Service) List(ctx context.Context) ([]Relation, error) {
	return s.store.ListRelations(ctx)
}

func (s Service) Update(ctx context.Context, id string, toUpdate Relation) (Relation, error) {
	oldRelation, err := s.store.GetRelation(ctx, id)

	if err != nil {
		return Relation{}, err
	}

	newRelation, err := s.store.UpdateRelation(ctx, id, toUpdate)

	if err != nil {
		return Relation{}, err
	}

	err = s.authzStore.DeleteRelation(ctx, oldRelation)

	if err != nil {
		return Relation{}, err
	}

	err = s.authzStore.AddRelation(ctx, newRelation)

	if err != nil {
		return Relation{}, err
	}

	return newRelation, nil
}

func (s Service) Delete(ctx context.Context, rel Relation) error {
	fetchedRel, err := s.store.GetRelationByFields(ctx, rel)
	if err != nil {
		return err
	}

	err = s.authzStore.DeleteRelation(ctx, rel)
	if err != nil {
		return err
	}

	err = s.store.DeleteRelationByID(ctx, fetchedRel.ID)

	return err
}

func (s Service) CheckPermission(ctx context.Context, usr user.User, resourceNS namespace.Namespace, resourceIdxa string, action action.Action) (bool, error) {
	return s.authzStore.CheckRelation(ctx, Relation{
		ObjectNamespace:  resourceNS,
		ObjectID:         resourceIdxa,
		SubjectID:        usr.ID,
		SubjectNamespace: namespace.DefinitionUser,
	}, action)
}
