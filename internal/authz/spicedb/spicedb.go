package spicedb

import (
	"context"
	"fmt"
	"strings"

	"github.com/odpf/shield/model"

	"github.com/odpf/salt/log"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/odpf/shield/config"
	"google.golang.org/grpc"
)

type SpiceDB struct {
	Policy     *Policy
	Permission *Permission
}

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
	relationship := transformRelation(relation)
	request := &pb.WriteRelationshipsRequest{
		Updates: []*pb.RelationshipUpdate{
			{
				Operation:    pb.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &relationship,
			},
		},
	}

	_, err := p.client.WriteRelationships(ctx, request)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (p Permission) DeleteRelation(ctx context.Context, relation model.Relation) error {
	relationship := transformRelation(relation)
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

	_, err := p.client.DeleteRelationships(ctx, request)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func transformRelation(relation model.Relation) pb.Relationship {
	return pb.Relationship{
		Resource: &pb.ObjectReference{
			ObjectId:   relation.ObjectId,
			ObjectType: strings.ReplaceAll(relation.ObjectNamespaceId, "-", "_"),
		},
		Subject: &pb.SubjectReference{
			Object: &pb.ObjectReference{
				ObjectId:   relation.SubjectId,
				ObjectType: strings.ReplaceAll(relation.SubjectNamespaceId, "-", "_"),
			},
		},
		Relation: strings.ReplaceAll(relation.RoleId, "-", "_"),
	}
}
