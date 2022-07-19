package v1beta1

import (
	"context"
	"errors"

	"github.com/odpf/shield/core/relation"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

type RelationService interface {
	Get(ctx context.Context, id string) (relation.Relation, error)
	List(ctx context.Context) ([]relation.Relation, error)
	Create(ctx context.Context, relation relation.Relation) (relation.Relation, error)
	Update(ctx context.Context, id string, relation relation.Relation) (relation.Relation, error)
}

var grpcRelationNotFoundErr = status.Errorf(codes.NotFound, "relation doesn't exist")

func (v Dep) ListRelations(ctx context.Context, request *shieldv1beta1.ListRelationsRequest) (*shieldv1beta1.ListRelationsResponse, error) {
	logger := grpczap.Extract(ctx)
	var relations []*shieldv1beta1.Relation

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

	return &shieldv1beta1.ListRelationsResponse{
		Relations: relations,
	}, nil
}

func (v Dep) CreateRelation(ctx context.Context, request *shieldv1beta1.CreateRelationRequest) (*shieldv1beta1.CreateRelationResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	newRelation, err := v.RelationService.Create(ctx, relation.Relation{
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

	return &shieldv1beta1.CreateRelationResponse{
		Relation: &relationPB,
	}, nil
}

func (v Dep) GetRelation(ctx context.Context, request *shieldv1beta1.GetRelationRequest) (*shieldv1beta1.GetRelationResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedRelation, err := v.RelationService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, relation.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "relation not found")
		case errors.Is(err, relation.ErrInvalidUUID):
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

	return &shieldv1beta1.GetRelationResponse{
		Relation: &relationPB,
	}, nil
}

func (v Dep) UpdateRelation(ctx context.Context, request *shieldv1beta1.UpdateRelationRequest) (*shieldv1beta1.UpdateRelationResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	updatedRelation, err := v.RelationService.Update(ctx, request.GetId(), relation.Relation{
		SubjectNamespaceId: request.GetBody().SubjectType,
		SubjectId:          request.GetBody().SubjectId,
		ObjectNamespaceId:  request.GetBody().ObjectType,
		ObjectId:           request.GetBody().ObjectId,
		RoleId:             request.GetBody().RoleId,
	})

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, relation.ErrNotExist):
			return nil, grpcRelationNotFoundErr
		case errors.Is(err, relation.ErrInvalidUUID):
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

	return &shieldv1beta1.UpdateRelationResponse{
		Relation: &relationPB,
	}, nil
}

func transformRelationToPB(relation relation.Relation) (shieldv1beta1.Relation, error) {
	subjectType, err := transformNamespaceToPB(relation.SubjectNamespace)

	if err != nil {
		return shieldv1beta1.Relation{}, err
	}

	objectType, err := transformNamespaceToPB(relation.ObjectNamespace)

	if err != nil {
		return shieldv1beta1.Relation{}, err
	}

	role, err := transformRoleToPB(relation.Role)

	if err != nil {
		return shieldv1beta1.Relation{}, err
	}

	return shieldv1beta1.Relation{
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
