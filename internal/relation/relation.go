package relation

import (
	"context"
	"errors"
	"github.com/odpf/shield/model"
)

type Service struct {
	Store Store
}

var (
	RelationDoesntExist = errors.New("relation doesn't exist")
	InvalidUUID         = errors.New("invalid syntax of uuid")
)

type Store interface {
	GetRelation(ctx context.Context, id string) (model.Relation, error)
	CreateRelation(ctx context.Context, org model.Relation) (model.Relation, error)
	ListRelations(ctx context.Context) ([]model.Relation, error)
	UpdateRelation(ctx context.Context, toUpdate model.Relation) (model.Relation, error)
}

func (s Service) Get(ctx context.Context, id string) (model.Relation, error) {
	return s.Store.GetRelation(ctx, id)
}

func (s Service) Create(ctx context.Context, relation model.Relation) (model.Relation, error) {
	return s.Store.CreateRelation(ctx, model.Relation{
		SubjectNamespaceId: relation.SubjectNamespaceId,
		SubjectId:          relation.SubjectId,
		ObjectNamespaceId:  relation.ObjectNamespaceId,
		ObjectId:           relation.ObjectId,
		RoleId:             relation.RoleId,
	})
}

func (s Service) List(ctx context.Context) ([]model.Relation, error) {
	return s.Store.ListRelations(ctx)
}

func (s Service) Update(ctx context.Context, toUpdate model.Relation) (model.Relation, error) {
	return s.Store.UpdateRelation(ctx, toUpdate)
}
