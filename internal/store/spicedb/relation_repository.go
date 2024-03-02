package spicedb

import (
	"context"
	"errors"
	"fmt"
	"io"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	newrelic "github.com/newrelic/go-agent"
	"github.com/raystack/frontier/core/relation"
)

type RelationRepository struct {
	spiceDB *SpiceDB

	// fullyConsistent makes sure all APIs are highly consistent on their responses
	// turning it on will result in slower API calls but useful in tests
	fullyConsistent bool

	// TODO(kushsharma): after every call, check if the response returns a relationship
	// snapshot(zedtoken/zookie), if it does, store it in a cache/db, and use it for subsequent calls
	// this will make the calls faster and avoid the use of fully consistent spiceDB
}

const nrProductName = "spicedb"

func NewRelationRepository(spiceDB *SpiceDB, fullyConsistent bool) *RelationRepository {
	return &RelationRepository{
		spiceDB:         spiceDB,
		fullyConsistent: fullyConsistent,
	}
}

func (r RelationRepository) Add(ctx context.Context, rel relation.Relation) error {
	relationship := &authzedpb.Relationship{
		Resource: &authzedpb.ObjectReference{
			ObjectType: rel.Object.Namespace,
			ObjectId:   rel.Object.ID,
		},
		Relation: rel.RelationName,
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
				ObjectType: rel.Subject.Namespace,
				ObjectId:   rel.Subject.ID,
			},
			OptionalRelation: rel.Subject.SubRelationName,
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
				"relation":          rel.Subject.SubRelationName,
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

func (r RelationRepository) Check(ctx context.Context, rel relation.Relation) (bool, error) {
	request := &authzedpb.CheckPermissionRequest{
		Consistency: r.getConsistency(),
		Resource: &authzedpb.ObjectReference{
			ObjectId:   rel.Object.ID,
			ObjectType: rel.Object.Namespace,
		},
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
				ObjectId:   rel.Subject.ID,
				ObjectType: rel.Subject.Namespace,
			},
			OptionalRelation: rel.Subject.SubRelationName,
		},
		Permission: rel.RelationName,
	}

	nrCtx := newrelic.FromContext(ctx)
	if nrCtx != nil {
		nr := newrelic.DatastoreSegment{
			Product:    nrProductName,
			Collection: fmt.Sprintf("object:%s::subject:%s", request.GetResource().GetObjectType(), request.GetSubject().GetObject().GetObjectType()),
			Operation:  "Check",
			StartTime:  nrCtx.StartSegmentNow(),
		}
		defer nr.End()
	}

	response, err := r.spiceDB.client.CheckPermission(ctx, request)
	if err != nil {
		return false, err
	}

	return response.GetPermissionship() == authzedpb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

func (r RelationRepository) Delete(ctx context.Context, rel relation.Relation) error {
	if rel.Object.Namespace == "" {
		return errors.New("object namespace is required to delete a relation")
	}
	request := &authzedpb.DeleteRelationshipsRequest{
		RelationshipFilter: &authzedpb.RelationshipFilter{
			ResourceType:       rel.Object.Namespace,
			OptionalResourceId: rel.Object.ID,
			OptionalRelation:   rel.RelationName,
			OptionalSubjectFilter: &authzedpb.SubjectFilter{
				SubjectType:       rel.Subject.Namespace,
				OptionalSubjectId: rel.Subject.ID,
				OptionalRelation: &authzedpb.SubjectFilter_RelationFilter{
					Relation: rel.Subject.SubRelationName,
				},
			},
		},
	}

	nrCtx := newrelic.FromContext(ctx)
	if nrCtx != nil {
		nr := newrelic.DatastoreSegment{
			Product: nrProductName,
			QueryParameters: map[string]interface{}{
				"relation":          rel.Subject.SubRelationName,
				"subject_namespace": rel.Subject.Namespace,
				"object_namespace":  rel.Object.Namespace,
			},
			Operation: "Delete_Relation",
			StartTime: nrCtx.StartSegmentNow(),
		}
		defer nr.End()
	}
	_, err := r.spiceDB.client.DeleteRelationships(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

func (r RelationRepository) LookupSubjects(ctx context.Context, rel relation.Relation) ([]string, error) {
	resp, err := r.spiceDB.client.LookupSubjects(ctx, &authzedpb.LookupSubjectsRequest{
		Consistency: r.getConsistency(),
		Resource: &authzedpb.ObjectReference{
			ObjectType: rel.Object.Namespace,
			ObjectId:   rel.Object.ID,
		},
		Permission:        rel.RelationName,
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
		subjects = append(subjects, item.GetSubject().GetSubjectObjectId())
	}
	return subjects, nil
}

func (r RelationRepository) LookupResources(ctx context.Context, rel relation.Relation) ([]string, error) {
	resp, err := r.spiceDB.client.LookupResources(ctx, &authzedpb.LookupResourcesRequest{
		Consistency:        r.getConsistency(),
		ResourceObjectType: rel.Object.Namespace,
		Permission:         rel.RelationName,
		Subject: &authzedpb.SubjectReference{
			Object: &authzedpb.ObjectReference{
				ObjectType: rel.Subject.Namespace,
				ObjectId:   rel.Subject.ID,
			},
			OptionalRelation: rel.Subject.SubRelationName,
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
func (r RelationRepository) ListRelations(ctx context.Context, rel relation.Relation) ([]relation.Relation, error) {
	resp, err := r.spiceDB.client.ReadRelationships(ctx, &authzedpb.ReadRelationshipsRequest{
		Consistency: r.getConsistency(),
		RelationshipFilter: &authzedpb.RelationshipFilter{
			ResourceType:       rel.Object.Namespace,
			OptionalResourceId: rel.Object.ID,
			OptionalRelation:   rel.RelationName,
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
	var rels []relation.Relation
	for {
		item, err := resp.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		pbRel := item.GetRelationship()
		rels = append(rels, relation.Relation{
			Object: relation.Object{
				ID:        pbRel.GetResource().GetObjectId(),
				Namespace: pbRel.GetResource().GetObjectType(),
			},
			Subject: relation.Subject{
				ID:              pbRel.GetSubject().GetObject().GetObjectId(),
				Namespace:       pbRel.GetSubject().GetObject().GetObjectType(),
				SubRelationName: pbRel.GetRelation(),
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

func (r RelationRepository) BatchCheck(ctx context.Context, relations []relation.Relation) ([]relation.CheckPair, error) {
	result := make([]relation.CheckPair, len(relations))
	items := make([]*authzedpb.BulkCheckPermissionRequestItem, 0, len(relations))
	for _, rel := range relations {
		items = append(items, &authzedpb.BulkCheckPermissionRequestItem{
			Resource: &authzedpb.ObjectReference{
				ObjectId:   rel.Object.ID,
				ObjectType: rel.Object.Namespace,
			},
			Subject: &authzedpb.SubjectReference{
				Object: &authzedpb.ObjectReference{
					ObjectId:   rel.Subject.ID,
					ObjectType: rel.Subject.Namespace,
				},
				OptionalRelation: rel.Subject.SubRelationName,
			},
			Permission: rel.RelationName,
		})
	}
	request := &authzedpb.BulkCheckPermissionRequest{
		Consistency: r.getConsistency(),
		Items:       items,
	}

	response, err := r.spiceDB.client.BulkCheckPermission(ctx, request)
	if err != nil {
		return result, err
	}

	var respErr error = nil
	for itemIdx, item := range response.GetPairs() {
		result[itemIdx] = relation.CheckPair{
			Relation: relation.Relation{
				Object: relation.Object{
					ID:        item.GetRequest().GetResource().GetObjectId(),
					Namespace: item.GetRequest().GetResource().GetObjectType(),
				},
				Subject: relation.Subject{
					ID:              item.GetRequest().GetSubject().GetObject().GetObjectId(),
					Namespace:       item.GetRequest().GetSubject().GetObject().GetObjectType(),
					SubRelationName: item.GetRequest().GetSubject().GetOptionalRelation(),
				},
				RelationName: item.GetRequest().GetPermission(),
			},
			Status: false,
		}
		if item.GetError() != nil {
			respErr = errors.Join(respErr, errors.New(item.GetRequest().GetPermission()+": "+item.GetError().GetMessage()))
			continue
		}
		if item.GetItem() != nil {
			result[itemIdx].Status = item.GetItem().GetPermissionship() == authzedpb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION
		}
	}
	return result, respErr
}
