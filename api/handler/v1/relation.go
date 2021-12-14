package v1

import (
	"context"
	"errors"
	"github.com/odpf/shield/internal/relation"
	"github.com/odpf/shield/model"
	shieldv1 "github.com/odpf/shield/proto/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

type RelationService interface {
	Get(ctx context.Context, id string) (model.Relation, error)
	List(ctx context.Context) ([]model.Relation, error)
	Create(ctx context.Context, relation model.Relation) (model.Relation, error)
	Update(ctx context.Context, id string, relation model.Relation) (model.Relation, error)
}

var grpcRelationNotFoundErr = status.Errorf(codes.NotFound, "relation doesn't exist")

func (v Dep) ListRelations(ctx context.Context) (*shieldv1.ListRelationsResponse, error) {
	logger := grpczap.Extract(ctx)
	var relations []*shieldv1.Relation

	relationsList, err := v.RelationService.List(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, r := range relationsList {
		relationPB, err := transformRelationToPB(r)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		relations = append(relations, &relationPB)
	}

	return &shieldv1.ListRelationsResponse{
		Relations: relations,
	}, nil
}

func (v Dep) CreateRelation(ctx context.Context, request *shieldv1.CreateRelationRequest) (*shieldv1.CreateRelationResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	newRelation, err := v.RelationService.Create(ctx, model.Relation{
		SubjectNamespaceId: request.GetBody().SubjectType,
		SubjectId:          request.GetBody().SubjectId,
		ObjectNamespaceId:  request.GetBody().ObjectType,
		ObjectId:           request.GetBody().ObjectId,
		RoleId:             request.GetBody().RoleId,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	relationPB, err := transformRelationToPB(newRelation)

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1.CreateRelationResponse{
		Relation: &relationPB,
	}, nil
}

func (v Dep) GetRelation(ctx context.Context, request *shieldv1.GetRelationRequest) (*shieldv1.GetRelationResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedRelation, err := v.RelationService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, relation.RelationDoesntExist):
			return nil, status.Errorf(codes.NotFound, "relation not found")
		case errors.Is(err, relation.InvalidUUID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	relationPB, err := transformRelationToPB(fetchedRelation)

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1.GetRelationResponse{
		Relation: &relationPB,
	}, nil
}

func (v Dep) UpdateRelation(ctx context.Context, request *shieldv1.UpdateRelationRequest) (*shieldv1.UpdateRelationResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	updatedRelation, err := v.RelationService.Update(ctx, request.GetId(), model.Relation{
		SubjectNamespaceId: request.GetBody().SubjectType,
		SubjectId:          request.GetBody().SubjectId,
		ObjectNamespaceId:  request.GetBody().ObjectType,
		ObjectId:           request.GetBody().ObjectId,
		RoleId:             request.GetBody().RoleId,
	})

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, relation.RelationDoesntExist):
			return nil, grpcRelationNotFoundErr
		case errors.Is(err, relation.InvalidUUID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	relationPB, err := transformRelationToPB(updatedRelation)

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1.UpdateRelationResponse{
		Relation: &relationPB,
	}, nil
}

func transformRelationToPB(relation model.Relation) (shieldv1.Relation, error) {
	subjectType, err := transformNamespaceToPB(relation.SubjectNamespace)

	if err != nil {
		return shieldv1.Relation{}, err
	}

	objectType, err := transformNamespaceToPB(relation.ObjectNamespace)

	if err != nil {
		return shieldv1.Relation{}, err
	}

	role, err := transformRoleToPB(relation.Role)

	if err != nil {
		return shieldv1.Relation{}, err
	}

	return shieldv1.Relation{
		Id:          relation.Id,
		SubjectType: &subjectType,
		SubjectId:   relation.SubjectId,
		ObjectType:  &objectType,
		ObjectId:    relation.ObjectId,
		Role:        &role,
		CreatedAt:   timestamppb.New(relation.CreatedAt),
		UpdatedAt:   timestamppb.New(relation.UpdatedAt),
	}, nil
}
