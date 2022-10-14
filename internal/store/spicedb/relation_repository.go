package spicedb

import (
	"context"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/internal/store/spicedb/schema_generator"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

type RelationRepository struct {
	spiceDB *SpiceDB
}

func NewRelationRepository(spiceDB *SpiceDB) *RelationRepository {
	return &RelationRepository{
		spiceDB: spiceDB,
	}
}

func (r RelationRepository) Add(ctx context.Context, rel relation.Relation) error {
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

	if _, err = r.spiceDB.client.WriteRelationships(ctx, request); err != nil {
		return err
	}

	return nil
}

func getRelation(a string) string {
	if a == "group" {
		return "membership"
	}

	return ""
}

func (r RelationRepository) AddV2(ctx context.Context, rel relation.RelationV2) error {
	relationship := &pb.Relationship{
		Resource: &pb.ObjectReference{
			ObjectType: rel.Object.NamespaceID,
			ObjectId:   rel.Object.ID,
		},
		Relation: rel.Subject.RoleID,
		Subject: &pb.SubjectReference{
			Object: &pb.ObjectReference{
				ObjectType: rel.Subject.Namespace,
				ObjectId:   rel.Subject.ID,
			},
			OptionalRelation: getRelation(rel.Subject.Namespace),
		},
	}
	request := &authzedpb.WriteRelationshipsRequest{
		Updates: []*authzedpb.RelationshipUpdate{
			{
				Operation:    authzedpb.RelationshipUpdate_OPERATION_TOUCH,
				Relationship: relationship,
			},
		},
	}

	if _, err := r.spiceDB.client.WriteRelationships(ctx, request); err != nil {
		return err
	}

	return nil
}

func (r RelationRepository) Check(ctx context.Context, rel relation.Relation, act action.Action) (bool, error) {
	relationship, err := schema_generator.TransformCheckRelation(rel)
	if err != nil {
		return false, err
	}

	request := &authzedpb.CheckPermissionRequest{
		Resource:   relationship.Resource,
		Subject:    relationship.Subject,
		Permission: act.ID,
	}

	response, err := r.spiceDB.client.CheckPermission(ctx, request)
	if err != nil {
		return false, err
	}

	return response.Permissionship == authzedpb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

func (r RelationRepository) Delete(ctx context.Context, rel relation.Relation) error {
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

	if _, err = r.spiceDB.client.DeleteRelationships(ctx, request); err != nil {
		return err
	}

	return nil
}

func (r RelationRepository) DeleteSubjectRelations(ctx context.Context, resourceType, optionalResourceID string) error {
	request := &authzedpb.DeleteRelationshipsRequest{
		RelationshipFilter: &authzedpb.RelationshipFilter{
			ResourceType:       resourceType,
			OptionalResourceId: optionalResourceID,
		},
	}

	if _, err := r.spiceDB.client.DeleteRelationships(ctx, request); err != nil {
		return err
	}

	return nil
}
