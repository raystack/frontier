package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/odpf/shield/internal/relation"
	"github.com/odpf/shield/model"
	"time"
)

type Relation struct {
	Id                 string    `db:"id"`
	SubjectNamespaceId string    `db:"subject_namespace_id"`
	SubjectNamespace   Namespace `db:"subject_namespace"`
	SubjectId          string    `db:"subject_id"`
	ObjectNamespaceId  string    `db:"object_namespace_id"`
	ObjectNamespace    Namespace `db:"object_namespace"`
	ObjectId           string    `db:"object_id"`
	RoleId             string    `db:"role_id"`
	Role               Role      `db:"role"`
	CreatedAt          time.Time `db:"created_at"`
	UpdatedAt          time.Time `db:"updated_at"`
}

const (
	createRelationQuery = `
		INSERT INTO relations(
		  subject_namespace_id,
		  subject_id,
		  object_namespace_id,
		  object_id,
		  role_id
		) values (
			  $1,
			  $2,
			  $3,
			  $4,
			  $5
		) RETURNING id, subject_namespace_id,  subject_id, object_namespace_id,  object_id, role_id,  created_at, updated_at`
	listRelationQuery = `
		SELECT 
		       id,
		       subject_namespace_id,
		       subject_id,
		       object_namespace_id,
		       object_id,
		       role_id,
		       created_at,
		       updated_at
		FROM relations`
	getRelationsQuery = `
		SELECT 
		       id, 
		       subject_namespace_id, 
		       subject_id, 
		       object_namespace_id, 
		       object_id, 
		       role_id, 
		       created_at, 
		       updated_at 
		FROM relations 
		WHERE id = $1`
	updateRelationQuery = `
		UPDATE relations SET
			 subject_namespace_id = $2,
			 subject_id = $3,
			 object_namespace_id = $4,
			 object_id = $5,
			 role_id = $6 
		WHERE id = $1`
)

func (s Store) CreateRelation(ctx context.Context, relationToCreate model.Relation) (model.Relation, error) {
	var newRelation Relation

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newRelation, createRelationQuery, relationToCreate.SubjectNamespaceId, relationToCreate.SubjectId, relationToCreate.ObjectNamespaceId, relationToCreate.ObjectId, relationToCreate.RoleId)
	})

	if err != nil {
		return model.Relation{}, err
	}

	transformedRelation, err := transformToRelation(newRelation)

	if err != nil {
		return model.Relation{}, err
	}

	return transformedRelation, nil
}

func (s Store) ListRelations(ctx context.Context) ([]model.Relation, error) {
	var fetchedRelations []Relation
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedRelations, listRelationQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.Relation{}, relation.RelationDoesntExist
	}

	if err != nil {
		return []model.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedRelations []model.Relation

	for _, r := range fetchedRelations {
		transformedRelation, err := transformToRelation(r)
		if err != nil {
			return []model.Relation{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedRelations = append(transformedRelations, transformedRelation)
	}

	return transformedRelations, nil
}

func (s Store) GetRelation(ctx context.Context, id string) (model.Relation, error) {
	var fetchedRelation Relation
	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &fetchedRelation, getRelationsQuery, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Relation{}, relation.RelationDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.Relation{}, relation.InvalidUUID
	} else if err != nil {
		return model.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	if err != nil {
		return model.Relation{}, err
	}

	transformedRelation, err := transformToRelation(fetchedRelation)
	if err != nil {
		return model.Relation{}, err
	}

	return transformedRelation, nil
}

func (s Store) UpdateRelation(ctx context.Context, id string, toUpdate model.Relation) (model.Relation, error) {
	var updatedRelation Relation

	err := s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &updatedRelation, updateRelationQuery, id, toUpdate.SubjectNamespaceId, toUpdate.SubjectId, toUpdate.ObjectNamespaceId, toUpdate.ObjectId, toUpdate.RoleId)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return model.Relation{}, relation.RelationDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.Relation{}, fmt.Errorf("%w: %s", relation.InvalidUUID, err)
	} else if err != nil {
		return model.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	toUpdate, err = transformToRelation(updatedRelation)
	if err != nil {
		return model.Relation{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}

func transformToRelation(from Relation) (model.Relation, error) {

	return model.Relation{
		Id:                 from.Id,
		SubjectNamespaceId: from.SubjectNamespaceId,
		SubjectId:          from.SubjectId,
		ObjectNamespaceId:  from.ObjectNamespaceId,
		ObjectId:           from.ObjectId,
		RoleId:             from.RoleId,
		CreatedAt:          from.CreatedAt,
		UpdatedAt:          from.UpdatedAt,
	}, nil
}
