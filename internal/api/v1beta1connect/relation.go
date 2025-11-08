package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RelationService interface {
	Get(ctx context.Context, id string) (relation.Relation, error)
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	List(ctx context.Context, f relation.Filter) ([]relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

func (h *ConnectHandler) ListRelations(ctx context.Context, request *connect.Request[frontierv1beta1.ListRelationsRequest]) (*connect.Response[frontierv1beta1.ListRelationsResponse], error) {
	errorLogger := NewErrorLogger()
	var err error
	var subject relation.Subject
	var object relation.Object

	if request.Msg.GetSubject() != "" {
		subject.Namespace, subject.ID, err = schema.SplitNamespaceAndResourceID(request.Msg.GetSubject())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation)
		}
	}
	if request.Msg.GetObject() != "" {
		object.Namespace, object.ID, err = schema.SplitNamespaceAndResourceID(request.Msg.GetObject())
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation)
		}
	}

	var relations []*frontierv1beta1.Relation
	relationsList, err := h.relationService.List(ctx, relation.Filter{
		Subject: subject,
		Object:  object,
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "ListRelations", err,
			zap.String("subject", request.Msg.GetSubject()),
			zap.String("object", request.Msg.GetObject()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, r := range relationsList {
		relationPB, err := transformRelationV2ToPB(r)
		if err != nil {
			errorLogger.LogTransformError(ctx, request, "ListRelations", r.ID, err)
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		relations = append(relations, relationPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListRelationsResponse{
		Relations: relations,
	}), nil
}

func (h *ConnectHandler) CreateRelation(ctx context.Context, request *connect.Request[frontierv1beta1.CreateRelationRequest]) (*connect.Response[frontierv1beta1.CreateRelationResponse], error) {
	errorLogger := NewErrorLogger()

	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	subjectNamespace, subjectID, err := schema.SplitNamespaceAndResourceID(request.Msg.GetBody().GetSubject())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation)
	}
	objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(request.Msg.GetBody().GetObject())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation)
	}

	// If Principal is a user, then we will get ID for that user as Subject.ID
	if subjectNamespace == schema.UserPrincipal {
		if !utils.IsValidUUID(subjectID) {
			// could be email
			fetchedUser, err := h.userService.GetByEmail(ctx, subjectID)
			if err != nil {
				errorLogger.LogServiceError(ctx, request, "CreateRelation.GetUserByEmail", err,
					zap.String("subject_id", subjectID))
				return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
			}
			subjectID = fetchedUser.ID
		}
	}

	newRelation, err := h.relationService.Create(ctx, relation.Relation{
		Object: relation.Object{
			ID:        objectID,
			Namespace: objectNamespace,
		},
		Subject: relation.Subject{
			ID:              subjectID,
			Namespace:       subjectNamespace,
			SubRelationName: request.Msg.GetBody().GetSubjectSubRelation(),
		},
		RelationName: request.Msg.GetBody().GetRelation(),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "CreateRelation", err,
			zap.String("subject", request.Msg.GetBody().GetSubject()),
			zap.String("object", request.Msg.GetBody().GetObject()),
			zap.String("relation", request.Msg.GetBody().GetRelation()),
			zap.String("subject_sub_relation", request.Msg.GetBody().GetSubjectSubRelation()))

		switch {
		case errors.Is(err, relation.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "CreateRelation", err,
				zap.String("subject", request.Msg.GetBody().GetSubject()),
				zap.String("object", request.Msg.GetBody().GetObject()),
				zap.String("relation", request.Msg.GetBody().GetRelation()),
				zap.String("subject_sub_relation", request.Msg.GetBody().GetSubjectSubRelation()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	relationPB, err := transformRelationV2ToPB(newRelation)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "CreateRelation", newRelation.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreateRelationResponse{
		Relation: relationPB,
	}), nil
}

func (h *ConnectHandler) GetRelation(ctx context.Context, request *connect.Request[frontierv1beta1.GetRelationRequest]) (*connect.Response[frontierv1beta1.GetRelationResponse], error) {
	errorLogger := NewErrorLogger()
	relationID := request.Msg.GetId()

	fetchedRelation, err := h.relationService.Get(ctx, relationID)
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "GetRelation", err,
			zap.String("relation_id", relationID))

		switch {
		case errors.Is(err, relation.ErrNotExist),
			errors.Is(err, relation.ErrInvalidUUID),
			errors.Is(err, relation.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "GetRelation", err,
				zap.String("relation_id", relationID))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	relationPB, err := transformRelationV2ToPB(fetchedRelation)
	if err != nil {
		errorLogger.LogTransformError(ctx, request, "GetRelation", fetchedRelation.ID, err)
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.GetRelationResponse{
		Relation: relationPB,
	}), nil
}

func (h *ConnectHandler) DeleteRelation(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteRelationRequest]) (*connect.Response[frontierv1beta1.DeleteRelationResponse], error) {
	errorLogger := NewErrorLogger()

	subjectNamespace, subjectID, err := schema.SplitNamespaceAndResourceID(request.Msg.GetSubject())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation)
	}
	objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(request.Msg.GetObject())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation)
	}

	err = h.relationService.Delete(ctx, relation.Relation{
		Object: relation.Object{
			Namespace: objectNamespace,
			ID:        objectID,
		},
		Subject: relation.Subject{
			Namespace: subjectNamespace,
			ID:        subjectID,
		},
		RelationName: request.Msg.GetRelation(),
	})
	if err != nil {
		errorLogger.LogServiceError(ctx, request, "DeleteRelation", err,
			zap.String("subject", request.Msg.GetSubject()),
			zap.String("object", request.Msg.GetObject()),
			zap.String("relation", request.Msg.GetRelation()))

		switch {
		case errors.Is(err, relation.ErrNotExist),
			errors.Is(err, relation.ErrInvalidUUID),
			errors.Is(err, relation.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			errorLogger.LogUnexpectedError(ctx, request, "DeleteRelation", err,
				zap.String("subject", request.Msg.GetSubject()),
				zap.String("object", request.Msg.GetObject()),
				zap.String("relation", request.Msg.GetRelation()))
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.DeleteRelationResponse{}), nil
}

func transformRelationV2ToPB(relation relation.Relation) (*frontierv1beta1.Relation, error) {
	rel := &frontierv1beta1.Relation{
		Id:                 relation.ID,
		Object:             schema.JoinNamespaceAndResourceID(relation.Object.Namespace, relation.Object.ID),
		Subject:            schema.JoinNamespaceAndResourceID(relation.Subject.Namespace, relation.Subject.ID),
		SubjectSubRelation: relation.Subject.SubRelationName,
		Relation:           relation.RelationName,
	}
	if !relation.CreatedAt.IsZero() {
		rel.CreatedAt = timestamppb.New(relation.CreatedAt)
	}
	if !relation.UpdatedAt.IsZero() {
		rel.UpdatedAt = timestamppb.New(relation.UpdatedAt)
	}
	return rel, nil
}
