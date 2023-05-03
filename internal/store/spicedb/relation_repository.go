package spicedb

import (
	"context"
	"fmt"
	"io"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/internal/store/spicedb/schema_generator"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	newrelic "github.com/newrelic/go-agent"
)

type RelationRepository struct {
	spiceDB *SpiceDB

	// fullyConsistent makes sure all APIs are highly consistent on their responses
	// turning it on will result in slower API calls but useful in tests
	fullyConsistent bool
}

const nrProductName = "spicedb"

func NewRelationRepository(spiceDB *SpiceDB, fullyConsistent bool) *RelationRepository {
	return &RelationRepository{
		spiceDB:         spiceDB,
		fullyConsistent: fullyConsistent,
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
	if a == schema.GroupPrincipal {
		return "membership"
	}

	return ""
}

func (r RelationRepository) AddV2(ctx context.Context, rel relation.RelationV2) error {
	relationship := &authzedpb.Relationship{
		Resource: &authzedpb.ObjectReference{
			ObjectType: rel.Object.Namespace,
			ObjectId:   rel.Object.ID,
		},
		Relation: schema.GetRoleName(rel.Subject.RoleID),
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
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

	nrCtx := newrelic.FromContext(ctx)
	if nrCtx != nil {
		nr := newrelic.DatastoreSegment{
			Product: nrProductName,
			QueryParameters: map[string]interface{}{
				"relation":          rel.Subject.RoleID,
				"subject_namespace": rel.Subject.Namespace,
				"object_namespace":  rel.Object.Namespace,
			},
			Operation: "Upsert_Relation",
			StartTime: nrCtx.StartSegmentNow(),
		}
		defer nr.End()
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
		Consistency: r.getConsistency(),
		Resource:    relationship.Resource,
		Subject:     relationship.Subject,
		Permission:  act.ID,
	}

	nrCtx := newrelic.FromContext(ctx)
	if nrCtx != nil {
		nr := newrelic.DatastoreSegment{
			Product:    nrProductName,
			Collection: fmt.Sprintf("object:%s::subject:%s", request.Resource.ObjectType, request.Subject.Object.ObjectType),
			Operation:  "Check",
			StartTime:  nrCtx.StartSegmentNow(),
		}
		defer nr.End()
	}

	response, err := r.spiceDB.client.CheckPermission(ctx, request)
	if err != nil {
		return false, err
	}

	return response.Permissionship == authzedpb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

func (r RelationRepository) DeleteV2(ctx context.Context, rel relation.RelationV2) error {
	relationship, err := schema_generator.TransformRelationV2(rel)
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

	nrCtx := newrelic.FromContext(ctx)
	if nrCtx != nil {
		nr := newrelic.DatastoreSegment{
			Product: nrProductName,
			QueryParameters: map[string]interface{}{
				"relation":          rel.Subject.RoleID,
				"subject_namespace": rel.Subject.Namespace,
				"object_namespace":  rel.Object.Namespace,
			},
			Operation: "Delete_Relation",
			StartTime: nrCtx.StartSegmentNow(),
		}
		defer nr.End()
	}
	_, err = r.spiceDB.client.DeleteRelationships(ctx, request)
	if err != nil {
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

	nrCtx := newrelic.FromContext(ctx)
	if nrCtx != nil {
		nr := newrelic.DatastoreSegment{
			Product: nrProductName,
			QueryParameters: map[string]interface{}{
				"object_namespace": resourceType,
				"object_id":        optionalResourceID,
			},
			Operation: "Delete_Subject_Relations",
			StartTime: nrCtx.StartSegmentNow(),
		}
		defer nr.End()
	}

	if _, err := r.spiceDB.client.DeleteRelationships(ctx, request); err != nil {
		return err
	}

	return nil
}

func (r RelationRepository) LookupSubjects(ctx context.Context, rel relation.RelationV2) ([]string, error) {
	resp, err := r.spiceDB.client.LookupSubjects(ctx, &authzedpb.LookupSubjectsRequest{
		Consistency: r.getConsistency(),
		Resource: &authzedpb.ObjectReference{
			ObjectType: rel.Object.Namespace,
			ObjectId:   rel.Object.ID,
		},
		Permission:        rel.Subject.RoleID,
		SubjectObjectType: rel.Subject.Namespace,
	})
	if err != nil {
		return nil, err
	}
	var subjects []string
	for {
		item, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, item.GetSubject().SubjectObjectId)
	}
	return subjects, nil
}

func (r RelationRepository) LookupResources(ctx context.Context, rel relation.RelationV2) ([]string, error) {
	resp, err := r.spiceDB.client.LookupResources(ctx, &authzedpb.LookupResourcesRequest{
		Consistency:        r.getConsistency(),
		ResourceObjectType: rel.Object.Namespace,
		Permission:         rel.Subject.RoleID,
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
				ObjectType: rel.Subject.Namespace,
				ObjectId:   rel.Subject.ID,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	var subjects []string
	for {
		item, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		subjects = append(subjects, item.GetResourceObjectId())
	}
	return subjects, nil
}

// ListRelations shouldn't be used in high TPS flows as consistency requirements are set high
func (r RelationRepository) ListRelations(ctx context.Context, rel relation.RelationV2) ([]relation.RelationV2, error) {
	resp, err := r.spiceDB.client.ReadRelationships(ctx, &authzedpb.ReadRelationshipsRequest{
		Consistency: r.getConsistency(),
		RelationshipFilter: &authzedpb.RelationshipFilter{
			ResourceType:       rel.Object.Namespace,
			OptionalResourceId: rel.Object.ID,
			OptionalRelation:   rel.Subject.RoleID,
			OptionalSubjectFilter: &authzedpb.SubjectFilter{
				SubjectType:       rel.Subject.Namespace,
				OptionalSubjectId: rel.Subject.ID,
				OptionalRelation:  nil,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	var rels []relation.RelationV2
	for {
		item, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		pbRel := item.GetRelationship()
		rels = append(rels, relation.RelationV2{
			Object: relation.Object{
				ID:        pbRel.Resource.ObjectId,
				Namespace: pbRel.Resource.ObjectType,
			},
			Subject: relation.Subject{
				ID:        pbRel.Subject.Object.ObjectId,
				Namespace: pbRel.Subject.Object.ObjectType,
				RoleID:    pbRel.Relation,
			},
		})
	}
	return rels, nil
}

func (r RelationRepository) getConsistency() *authzedpb.Consistency {
	if !r.fullyConsistent {
		return nil
	}
	return &authzedpb.Consistency{Requirement: &authzedpb.Consistency_FullyConsistent{FullyConsistent: true}}
}
