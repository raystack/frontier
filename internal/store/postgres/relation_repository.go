package postgres

import (
	"context"
	"errors"
	"fmt"

	"database/sql"

	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/str"
)

type RelationRepository struct {
	dbc *db.Client
}

func NewRelationRepository(dbc *db.Client) *RelationRepository {
	return &RelationRepository{
		dbc: dbc,
	}
}

func (r RelationRepository) Create(ctx context.Context, relationToCreate relation.Relation) (relation.Relation, error) {
	var newRelation Relation

	subjectNamespaceID := str.DefaultStringIfEmpty(relationToCreate.SubjectNamespace.ID, relationToCreate.SubjectNamespaceID)
	objectNamespaceID := str.DefaultStringIfEmpty(relationToCreate.ObjectNamespace.ID, relationToCreate.ObjectNamespaceID)
	roleID := str.DefaultStringIfEmpty(relationToCreate.Role.ID, relationToCreate.RoleID)
	var nsID string

	if relationToCreate.RelationType == relation.RelationTypes.Namespace {
		nsID = roleID
		roleID = ""
	}

	createRelationQuery, err := buildCreateRelationQuery(dialect)
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(
			ctx,
			&newRelation,
			createRelationQuery,
			subjectNamespaceID,
			relationToCreate.SubjectID,
			objectNamespaceID,
			relationToCreate.ObjectID,
			sql.NullString{String: roleID, Valid: roleID != ""},
			sql.NullString{String: nsID, Valid: nsID != ""},
		)
	}); err != nil {
		return relation.Relation{}, err
	}

	transformedRelation, err := transformToRelation(newRelation)
	if err != nil {
		return relation.Relation{}, err
	}

	return transformedRelation, nil
}

func (r RelationRepository) List(ctx context.Context) ([]relation.Relation, error) {
	var fetchedRelations []Relation
	listRelationQuery, err := buildListRelationQuery(dialect)
	if err != nil {
		return []relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedRelations, listRelationQuery)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []relation.Relation{}, relation.ErrNotExist
		}
		return []relation.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedRelations []relation.Relation

	for _, r := range fetchedRelations {
		transformedRelation, err := transformToRelation(r)
		if err != nil {
			return []relation.Relation{}, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedRelations = append(transformedRelations, transformedRelation)
	}

	return transformedRelations, nil
}

func (r RelationRepository) Get(ctx context.Context, id string) (relation.Relation, error) {
	var fetchedRelation Relation
	getRelationsQuery, err := buildGetRelationsQuery(dialect)
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &fetchedRelation, getRelationsQuery, id)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return relation.Relation{}, relation.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return relation.Relation{}, relation.ErrInvalidUUID
		}
		return relation.Relation{}, err
	}

	transformedRelation, err := transformToRelation(fetchedRelation)
	if err != nil {
		return relation.Relation{}, err
	}

	return transformedRelation, nil
}

func (r RelationRepository) DeleteByID(ctx context.Context, id string) error {
	deleteRelationByIDQuery, err := buildDeleteRelationByIDQuery(dialect)
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	return r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		result, err := r.dbc.ExecContext(ctx, deleteRelationByIDQuery, id)
		if err != nil {
			return err
		}

		count, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if count == 1 {
			return nil
		}

		// TODO make this idempotent
		return errors.New("relation id not found")
	})
}

func (r RelationRepository) GetByFields(ctx context.Context, rel relation.Relation) (relation.Relation, error) {
	var fetchedRelation Relation

	subjectNamespaceID := str.DefaultStringIfEmpty(rel.SubjectNamespace.ID, rel.SubjectNamespaceID)
	objectNamespaceID := str.DefaultStringIfEmpty(rel.ObjectNamespace.ID, rel.ObjectNamespaceID)
	roleID := str.DefaultStringIfEmpty(rel.Role.ID, rel.RoleID)
	var nsID string

	if rel.RelationType == relation.RelationTypes.Namespace {
		nsID = roleID
		roleID = ""
	}

	getRelationByFieldsQuery, err := buildGetRelationByFieldsQuery(dialect)
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx,
			&fetchedRelation,
			getRelationByFieldsQuery,
			subjectNamespaceID,
			rel.SubjectID,
			objectNamespaceID,
			rel.ObjectID,
			sql.NullString{String: roleID, Valid: roleID != ""},
			sql.NullString{String: nsID, Valid: nsID != ""},
		)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return relation.Relation{}, relation.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return relation.Relation{}, relation.ErrInvalidUUID
		}
		return relation.Relation{}, err
	}

	transformedRelation, err := transformToRelation(fetchedRelation)
	if err != nil {
		return relation.Relation{}, err
	}

	return transformedRelation, nil
}

func (r RelationRepository) Update(ctx context.Context, id string, toUpdate relation.Relation) (relation.Relation, error) {
	var updatedRelation Relation

	subjectNamespaceID := str.DefaultStringIfEmpty(toUpdate.SubjectNamespace.ID, toUpdate.SubjectNamespaceID)
	objectNamespaceID := str.DefaultStringIfEmpty(toUpdate.ObjectNamespace.ID, toUpdate.ObjectNamespaceID)
	roleID := str.DefaultStringIfEmpty(toUpdate.Role.ID, toUpdate.RoleID)
	var nsID string

	if toUpdate.RelationType == relation.RelationTypes.Namespace {
		nsID = roleID
		roleID = ""
	}

	updateRelationQuery, err := buildUpdateRelationQuery(dialect)
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(
			ctx,
			&updatedRelation,
			updateRelationQuery,
			id,
			subjectNamespaceID,
			toUpdate.SubjectID,
			objectNamespaceID,
			toUpdate.ObjectID,
			sql.NullString{String: roleID, Valid: roleID != ""},
			sql.NullString{String: nsID, Valid: nsID != ""},
		)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return relation.Relation{}, relation.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return relation.Relation{}, fmt.Errorf("%w: %s", relation.ErrInvalidUUID, err)
		}
		return relation.Relation{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	toUpdate, err = transformToRelation(updatedRelation)
	if err != nil {
		return relation.Relation{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}
