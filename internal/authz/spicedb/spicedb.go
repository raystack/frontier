package spicedb

import (
	"context"
	"fmt"

	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/schema_generator"
	"github.com/odpf/shield/model"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/odpf/salt/log"
	"google.golang.org/grpc"
)

type SpiceDB struct {
	Policy     *Policy
	Permission *Permission
}

const nrStoreSpiceDB = "spicedb"

type Policy struct {
	client *authzed.Client
}

type Permission struct {
	client *authzed.Client
}

func (s *SpiceDB) Check() bool {
	return false
}

func (p *Policy) AddPolicy(ctx context.Context, schema string) error {
	request := &pb.WriteSchemaRequest{Schema: schema}
	nr := newrelic.DatastoreSegment{
		Product:    nrStoreSpiceDB,
		Collection: "Policy",
		Operation:  "AddPolicy",
		StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
	}
	defer nr.End()

	_, err := p.client.WriteSchema(ctx, request)
	if err != nil {
		return err
	}
	return nil
}

func New(config config.SpiceDBConfig, logger log.Logger) (*SpiceDB, error) {
	endpoint := fmt.Sprintf("%s:%s", config.Host, config.Port)
	client, err := authzed.NewClient(endpoint, grpc.WithInsecure(), grpcutil.WithInsecureBearerToken(config.PreSharedKey))
	if err != nil {
		return &SpiceDB{}, err
	}

	logger.Info(fmt.Sprintf("Connected to spiceDB: %s", endpoint))

	policy := &Policy{
		client,
	}

	permission := &Permission{
		client,
	}
	return &SpiceDB{
		policy,
		permission,
	}, nil
}

func (p Permission) AddRelation(ctx context.Context, relation model.Relation) error {
	relationship, err := schema_generator.TransformRelation(relation)
	if err != nil {
		return err
	}
	request := &pb.WriteRelationshipsRequest{
		Updates: []*pb.RelationshipUpdate{
			{
				Operation:    pb.RelationshipUpdate_OPERATION_TOUCH,
				Relationship: relationship,
			},
		},
	}

	nr := newrelic.DatastoreSegment{
		Product:    nrStoreSpiceDB,
		Collection: fmt.Sprintf("%s.%s", relationship.Resource.ObjectType, relationship.Subject.Object.ObjectType),
		Operation:  "AddRelation",
		StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
	}
	defer nr.End()

	_, err = p.client.WriteRelationships(ctx, request)

	if err != nil {
		return err
	}

	return nil
}

func (p Permission) CheckRelation(ctx context.Context, relation model.Relation, action model.Action) (bool, error) {
	relationship, err := schema_generator.TransformCheckRelation(relation)
	if err != nil {
		return false, err
	}

	request := &pb.CheckPermissionRequest{
		Resource:   relationship.Resource,
		Subject:    relationship.Subject,
		Permission: action.Id,
	}

	nr := newrelic.DatastoreSegment{
		Product:    nrStoreSpiceDB,
		Collection: fmt.Sprintf("%s.%s", relationship.Resource.ObjectType, relationship.Subject.Object.ObjectType),
		Operation:  "CheckRelation",
		StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
	}
	defer nr.End()

	response, err := p.client.CheckPermission(ctx, request)

	if err != nil {
		return false, err
	}

	return response.Permissionship == pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION, nil
}

func (p Permission) DeleteRelation(ctx context.Context, relation model.Relation) error {
	relationship, err := schema_generator.TransformRelation(relation)
	if err != nil {
		return err
	}
	request := &pb.DeleteRelationshipsRequest{
		RelationshipFilter: &pb.RelationshipFilter{
			ResourceType:       relationship.Resource.ObjectType,
			OptionalResourceId: relationship.Resource.ObjectId,
			OptionalRelation:   relationship.Relation,
			OptionalSubjectFilter: &pb.SubjectFilter{
				SubjectType:       relationship.Subject.Object.ObjectType,
				OptionalSubjectId: relationship.Subject.Object.ObjectId,
			},
		},
	}

	nr := newrelic.DatastoreSegment{
		Product:    nrStoreSpiceDB,
		Collection: fmt.Sprintf("%s.%s", relationship.Resource.ObjectType, relationship.Subject.Object.ObjectType),
		Operation:  "DeleteRelation",
		StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
	}
	defer nr.End()

	_, err = p.client.DeleteRelationships(ctx, request)

	if err != nil {
		return err
	}

	return nil
}

func (p Permission) DeleteSubjectRelations(ctx context.Context, resource model.Resource) error {
	request := &pb.DeleteRelationshipsRequest{
		RelationshipFilter: &pb.RelationshipFilter{
			ResourceType:       resource.NamespaceId,
			OptionalResourceId: resource.Idxa,
		},
	}

	nr := newrelic.DatastoreSegment{
		Product:    nrStoreSpiceDB,
		Collection: resource.NamespaceId,
		Operation:  "DeleteRelationsForResource",
		StartTime:  newrelic.FromContext(ctx).StartSegmentNow(),
	}
	defer nr.End()

	_, err := p.client.DeleteRelationships(ctx, request)

	if err != nil {
		return err
	}

	return nil
}
