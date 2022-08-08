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

//go:generate mockery --name=RelationService -r --case underscore --with-expecter --structname RelationService --filename relation_service.go --output=./mocks
type RelationService interface {
	Get(ctx context.Context, id string) (relation.Relation, error)
	List(ctx context.Context) ([]relation.Relation, error)
	Create(ctx context.Context, relation relation.Relation) (relation.Relation, error)
	Update(ctx context.Context, relation relation.Relation) (relation.Relation, error)
}

var grpcRelationNotFoundErr = status.Errorf(codes.NotFound, "relation doesn't exist")

func (h Handler) ListRelations(ctx context.Context, request *shieldv1beta1.ListRelationsRequest) (*shieldv1beta1.ListRelationsResponse, error) {
	logger := grpczap.Extract(ctx)
	var relations []*shieldv1beta1.Relation

	relationsList, err := h.relationService.List(ctx)
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

func (h Handler) CreateRelation(ctx context.Context, request *shieldv1beta1.CreateRelationRequest) (*shieldv1beta1.CreateRelationResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	newRelation, err := h.relationService.Create(ctx, relation.Relation{
		SubjectNamespaceID: request.GetBody().GetSubjectType(),
		SubjectID:          request.GetBody().GetSubjectId(),
		ObjectNamespaceID:  request.GetBody().GetObjectType(),
		ObjectID:           request.GetBody().GetObjectId(),
		RoleID:             request.GetBody().GetRoleId(),
	})

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, relation.ErrNotExist),
			errors.Is(err, relation.ErrInvalidDetail):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
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

func (h Handler) GetRelation(ctx context.Context, request *shieldv1beta1.GetRelationRequest) (*shieldv1beta1.GetRelationResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedRelation, err := h.relationService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, relation.ErrNotExist),
			errors.Is(err, relation.ErrInvalidUUID),
			errors.Is(err, relation.ErrInvalidID):
			return nil, status.Errorf(codes.NotFound, "relation not found")
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

func (h Handler) UpdateRelation(ctx context.Context, request *shieldv1beta1.UpdateRelationRequest) (*shieldv1beta1.UpdateRelationResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	updatedRelation, err := h.relationService.Update(ctx, relation.Relation{
		ID:                 request.GetId(),
		SubjectNamespaceID: request.GetBody().GetSubjectType(),
		SubjectID:          request.GetBody().GetSubjectId(),
		ObjectNamespaceID:  request.GetBody().GetObjectType(),
		ObjectID:           request.GetBody().GetObjectId(),
		RoleID:             request.GetBody().GetRoleId(),
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, relation.ErrNotExist):
			return nil, grpcRelationNotFoundErr
		case errors.Is(err, relation.ErrInvalidUUID), errors.Is(err, relation.ErrInvalidID):
			return nil, grpcBadBodyError
		case errors.Is(err, relation.ErrConflict):
			return nil, grpcConflictError
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
		Id:          relation.ID,
		SubjectType: &subjectType,
		SubjectId:   relation.SubjectID,
		ObjectType:  &objectType,
		ObjectId:    relation.ObjectID,
		Role:        &role,
		CreatedAt:   timestamppb.New(relation.CreatedAt),
		UpdatedAt:   timestamppb.New(relation.UpdatedAt),
	}, nil
}
