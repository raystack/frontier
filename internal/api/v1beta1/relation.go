package v1beta1

import (
	"context"
	"errors"
	"fmt"
	"strings"

	shielduuid "github.com/odpf/shield/pkg/uuid"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/schema"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/core/relation"
	errpkg "github.com/odpf/shield/pkg/errors"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockery --name=RelationService -r --case underscore --with-expecter --structname RelationService --filename relation_service.go --output=./mocks
type RelationService interface {
	Get(ctx context.Context, id string) (relation.RelationV2, error)
	Create(ctx context.Context, rel relation.RelationV2) (relation.RelationV2, error)
	List(ctx context.Context) ([]relation.RelationV2, error)
	Delete(ctx context.Context, rel relation.RelationV2) error
	GetRelationByFields(ctx context.Context, rel relation.RelationV2) (relation.RelationV2, error)
}

var (
	grpcRelationNotFoundErr = status.Errorf(codes.NotFound, "relation doesn't exist")
	ErrSubjectSplitNotation = errors.New("subject/object should be provided as 'namespace:uuid'")
)

func (h Handler) ListRelations(ctx context.Context, request *shieldv1beta1.ListRelationsRequest) (*shieldv1beta1.ListRelationsResponse, error) {
	logger := grpczap.Extract(ctx)
	var relations []*shieldv1beta1.Relation

	relationsList, err := h.relationService.List(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, r := range relationsList {
		relationPB, err := transformRelationV2ToPB(r)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		relations = append(relations, relationPB)
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

	subjectNamespace, subjectID, err := extractSubjectFromPrincipal(request.GetBody().GetSubject())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	result, err := h.resourceService.CheckAuthz(ctx, resource.Resource{
		Name:        request.GetBody().GetObjectId(),
		NamespaceID: request.GetBody().ObjectNamespace,
	}, action.Action{ID: schema.EditPermission})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		default:
			formattedErr := fmt.Errorf("%s: %w", ErrInternalServer, err)
			logger.Error(formattedErr.Error())
			return nil, status.Errorf(codes.Internal, ErrInternalServer.Error())
		}
	}
	if !result {
		return nil, status.Errorf(codes.PermissionDenied, errpkg.ErrForbidden.Error())
	}

	// If Principal is a user, then we will get ID for that user as Subject.ID
	if subjectNamespace == schema.UserPrincipal || subjectNamespace == "user" {
		if !shielduuid.IsValid(subjectID) {
			// could be email
			fetchedUser, err := h.userService.GetByEmail(ctx, subjectID)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, err.Error())
			}
			subjectID = fetchedUser.ID
		}
		subjectNamespace = schema.UserPrincipal
	}

	newRelation, err := h.relationService.Create(ctx, relation.RelationV2{
		Object: relation.Object{
			ID:        request.GetBody().GetObjectId(),
			Namespace: request.GetBody().GetObjectNamespace(),
		},
		Subject: relation.Subject{
			ID:        subjectID,
			Namespace: subjectNamespace,
			RoleID:    request.GetBody().GetRoleName(),
		},
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, relation.ErrInvalidDetail):
			return nil, grpcBadBodyError
		default:
			formattedErr := fmt.Errorf("%s: %w", ErrInternalServer, err)
			logger.Error(formattedErr.Error())
			return nil, status.Errorf(codes.Internal, ErrInternalServer.Error())
		}
	}

	relationPB, err := transformRelationV2ToPB(newRelation)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateRelationResponse{
		Relation: relationPB,
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
			return nil, grpcRelationNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	relationPB, err := transformRelationV2ToPB(fetchedRelation)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetRelationResponse{
		Relation: relationPB,
	}, nil
}

func (h Handler) DeleteRelation(ctx context.Context, request *shieldv1beta1.DeleteRelationRequest) (*shieldv1beta1.DeleteRelationResponse, error) {
	logger := grpczap.Extract(ctx)

	subjectNamespace, subjectID, err := extractSubjectFromPrincipal(request.GetSubjectId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}
	objectNamespace, objectID, err := extractSubjectFromPrincipal(request.GetObjectId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	result, err := h.resourceService.CheckAuthz(ctx, resource.Resource{
		Name:        objectID,
		NamespaceID: objectNamespace,
	}, action.Action{ID: schema.EditPermission})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		default:
			formattedErr := fmt.Errorf("%s: %w", ErrInternalServer, err)
			logger.Error(formattedErr.Error())
			return nil, status.Errorf(codes.Internal, ErrInternalServer.Error())
		}
	}

	if !result {
		return nil, status.Errorf(codes.PermissionDenied, errpkg.ErrForbidden.Error())
	}

	err = h.relationService.Delete(ctx, relation.RelationV2{
		Object: relation.Object{
			Namespace: objectNamespace,
			ID:        objectID,
		},
		Subject: relation.Subject{
			Namespace: subjectNamespace,
			ID:        subjectID,
			RoleID:    request.GetRole(),
		},
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, relation.ErrNotExist),
			errors.Is(err, relation.ErrInvalidUUID),
			errors.Is(err, relation.ErrInvalidID):
			return nil, grpcRelationNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	return &shieldv1beta1.DeleteRelationResponse{
		Message: "relation deleted",
	}, nil
}

func transformRelationV2ToPB(relation relation.RelationV2) (*shieldv1beta1.Relation, error) {
	rel := &shieldv1beta1.Relation{
		Id:              relation.ID,
		ObjectId:        relation.Object.ID,
		ObjectNamespace: relation.Object.Namespace,
		Subject:         GenerateSubject(relation.Subject.Namespace, relation.Subject.ID),
		RoleName:        relation.Subject.RoleID,
	}
	if !relation.CreatedAt.IsZero() {
		rel.CreatedAt = timestamppb.New(relation.CreatedAt)
	}
	if !relation.UpdatedAt.IsZero() {
		rel.UpdatedAt = timestamppb.New(relation.UpdatedAt)
	}
	return rel, nil
}

func extractSubjectFromPrincipal(principal string) (string, string, error) {
	splits := strings.Split(principal, ":")
	if len(splits) < 2 {
		return "", "", ErrSubjectSplitNotation
	}
	return splits[0], splits[1], nil
}

func GenerateSubject(principal, Id string) string {
	return fmt.Sprintf("%s:%s", principal, Id)
}
