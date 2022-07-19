package spicedb

import (
	"context"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/internal/store/spicedb/schema_generator"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

func (db *SpiceDB) AddRelation(ctx context.Context, rel relation.Relation) error {
	relationship, err := schema_generator.TransformRelation(rel)
	if err != nil {
		return err
	}
	request := &authzedpb.WriteRelationshipsRequest{
		Updates: []*authzedpb.RelationshipUpdate{
			{
				Operation:    authzedpb.RelationshipUpdate_OPERATION_TOUCH,
				Relationship: relationship,
			},
		},
	}

	_, err = db.client.WriteRelationships(ctx, request)

	if err != nil {
		return err
	}

	return nil
}

func (db *SpiceDB) CheckRelation(ctx context.Context, rel relation.Relation, act action.Action) (bool, error) {
	relationship, err := schema_generator.TransformCheckRelation(rel)
	if err != nil {
		return false, err
	}

	request := &authzedpb.CheckPermissionRequest{
		Resource:   relationship.Resource,
		Subject:    relationship.Subject,
		Permission: act.Id,
	}

	response, err := db.client.CheckPermission(ctx, request)

	if err != nil {
		return false, err
	}

	return response.Permissionship == authzedpb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

func (db *SpiceDB) DeleteRelation(ctx context.Context, rel relation.Relation) error {
	relationship, err := schema_generator.TransformRelation(rel)
	if err != nil {
		return err
	}
	request := &authzedpb.DeleteRelationshipsRequest{
		RelationshipFilter: &authzedpb.RelationshipFilter{
			ResourceType:       relationship.Resource.ObjectType,
			OptionalResourceId: relationship.Resource.ObjectId,
			OptionalRelation:   relationship.Relation,
			OptionalSubjectFilter: &authzedpb.SubjectFilter{
				SubjectType:       relationship.Subject.Object.ObjectType,
				OptionalSubjectId: relationship.Subject.Object.ObjectId,
			},
		},
	}

	_, err = db.client.DeleteRelationships(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

func (db *SpiceDB) DeleteSubjectRelations(ctx context.Context, resourceType, optionalResourceID string) error {
	request := &authzedpb.DeleteRelationshipsRequest{
		RelationshipFilter: &authzedpb.RelationshipFilter{
			ResourceType:       resourceType,
			OptionalResourceId: optionalResourceID,
		},
	}

	_, err := db.client.DeleteRelationships(ctx, request)

	if err != nil {
		return err
	}

	return nil
}
