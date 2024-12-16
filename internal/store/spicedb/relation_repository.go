package spicedb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync/atomic"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	authzedpb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	newrelic "github.com/newrelic/go-agent"
	"github.com/raystack/frontier/core/relation"
)

type RelationRepository struct {
	spiceDB *SpiceDB

	// Consistency ensures Authz server consistency guarantees for various operations
	// Possible values are:
	// - "full": Guarantees that the data is always fresh
	// - "best_effort": Guarantees that the data is the best effort fresh
	// - "minimize_latency": Tries to prioritise minimal latency
	consistency ConsistencyLevel

	// tracing enables debug traces for check calls
	tracing bool

	// lastToken is the last zookie returned by the server, this is cached at instance level and
	// maybe not be consistent across multiple instances but that is fine in most cases as
	// the token is only used in lookup or list calls, for permission checks we always use the
	// consistency level. Storing it in a shared db/cache will make it consistent across instances.
	// We can also store multiple tokens in the cache based on what kind of resource we are dealing with
	// but that adds complexity.
	lastToken atomic.Pointer[authzedpb.ZedToken]
}

type ConsistencyLevel string

func (c ConsistencyLevel) String() string {
	return string(c)
}

const (
	ConsistencyLevelFull            ConsistencyLevel = "full"
	ConsistencyLevelBestEffort      ConsistencyLevel = "best_effort"
	ConsistencyLevelMinimizeLatency ConsistencyLevel = "minimize_latency"
)

const nrProductName = "spicedb"

func NewRelationRepository(spiceDB *SpiceDB, consistency ConsistencyLevel, tracing bool) *RelationRepository {
	return &RelationRepository{
		spiceDB:     spiceDB,
		consistency: consistency,
		tracing:     tracing,
	}
}

func (r *RelationRepository) Add(ctx context.Context, rel relation.Relation) error {
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

	resp, err := r.spiceDB.client.WriteRelationships(ctx, request)
	if err != nil {
		return err
	}

	r.lastToken.Store(resp.GetWrittenAt())
	return nil
}

func (r *RelationRepository) Check(ctx context.Context, rel relation.Relation) (bool, error) {
	request := &authzedpb.CheckPermissionRequest{
		Consistency: r.getConsistencyForCheck(),
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
		Permission:  rel.RelationName,
		WithTracing: r.tracing,
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
	if response.GetDebugTrace() != nil {
		str, _ := json.Marshal(response.GetDebugTrace())
		grpczap.Extract(ctx).Info("CheckPermission", zap.String("trace", string(str)))
	}

	r.lastToken.Store(response.GetCheckedAt())
	return response.GetPermissionship() == authzedpb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

func (r *RelationRepository) Delete(ctx context.Context, rel relation.Relation) error {
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
	resp, err := r.spiceDB.client.DeleteRelationships(ctx, request)
	if err != nil {
		return err
	}

	r.lastToken.Store(resp.GetDeletedAt())
	return nil
}

func (r *RelationRepository) LookupSubjects(ctx context.Context, rel relation.Relation) ([]string, error) {
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

func (r *RelationRepository) LookupResources(ctx context.Context, rel relation.Relation) ([]string, error) {
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
func (r *RelationRepository) ListRelations(ctx context.Context, rel relation.Relation) ([]relation.Relation, error) {
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

func (r *RelationRepository) BatchCheck(ctx context.Context, relations []relation.Relation) ([]relation.CheckPair, error) {
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
		Consistency: r.getConsistencyForCheck(),
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

	r.lastToken.Store(response.GetCheckedAt())
	return result, respErr
}

func (r *RelationRepository) getConsistency() *authzedpb.Consistency {
	switch r.consistency {
	case ConsistencyLevelMinimizeLatency:
		return &authzedpb.Consistency{Requirement: &authzedpb.Consistency_MinimizeLatency{MinimizeLatency: true}}
	case ConsistencyLevelFull:
		return &authzedpb.Consistency{Requirement: &authzedpb.Consistency_FullyConsistent{FullyConsistent: true}}
	}

	lastToken := r.lastToken.Load()
	if lastToken == nil {
		return &authzedpb.Consistency{Requirement: &authzedpb.Consistency_FullyConsistent{FullyConsistent: true}}
	}
	return &authzedpb.Consistency{
		Requirement: &authzedpb.Consistency_AtLeastAsFresh{
			AtLeastAsFresh: lastToken,
		},
	}
}

func (r *RelationRepository) getConsistencyForCheck() *authzedpb.Consistency {
	if r.consistency == ConsistencyLevelMinimizeLatency {
		return &authzedpb.Consistency{Requirement: &authzedpb.Consistency_MinimizeLatency{MinimizeLatency: true}}
	}
	return &authzedpb.Consistency{Requirement: &authzedpb.Consistency_FullyConsistent{FullyConsistent: true}}
}
