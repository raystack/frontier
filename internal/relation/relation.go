package relation

import (
	"context"
	"errors"

	"github.com/odpf/shield/internal/authz"

	"github.com/odpf/shield/model"
)

type Service struct {
	Store Store
	Authz *authz.Authz
}

var (
	RelationDoesntExist = errors.New("relation doesn't exist")
	InvalidUUID         = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetRelation(ctx context.Context, id string) (model.Relation, error)
	CreateRelation(ctx context.Context, relation model.Relation) (model.Relation, error)
	ListRelations(ctx context.Context) ([]model.Relation, error)
	UpdateRelation(ctx context.Context, id string, toUpdate model.Relation) (model.Relation, error)
}

func (s Service) Get(ctx context.Context, id string) (model.Relation, error) {
	return s.Store.GetRelation(ctx, id)
}

func (s Service) Create(ctx context.Context, relation model.Relation) (model.Relation, error) {
	rel, err := s.Store.CreateRelation(ctx, model.Relation{
		SubjectNamespaceId: relation.SubjectNamespaceId,
		SubjectId:          relation.SubjectId,
		ObjectNamespaceId:  relation.ObjectNamespaceId,
		ObjectId:           relation.ObjectId,
		RoleId:             relation.RoleId,
	})

	if err != nil {
		return model.Relation{}, err
	}

	err = s.Authz.Permission.AddRelation(ctx, rel)

	if err != nil {
		return model.Relation{}, err
	}

	return rel, nil
}

func (s Service) List(ctx context.Context) ([]model.Relation, error) {
	return s.Store.ListRelations(ctx)
}

func (s Service) Update(ctx context.Context, id string, toUpdate model.Relation) (model.Relation, error) {
	oldRelation, err := s.Store.GetRelation(ctx, id)

	if err != nil {
		return model.Relation{}, err
	}

	newRelation, err := s.Store.UpdateRelation(ctx, id, toUpdate)

	if err != nil {
		return model.Relation{}, err
	}

	err = s.Authz.Permission.DeleteRelation(ctx, oldRelation)

	if err != nil {
		return model.Relation{}, err
	}

	err = s.Authz.Permission.AddRelation(ctx, newRelation)

	if err != nil {
		return model.Relation{}, err
	}

	return newRelation, nil
}
