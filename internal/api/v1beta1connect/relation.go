package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RelationService interface {
	Get(ctx context.Context, id string) (relation.Relation, error)
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	List(ctx context.Context, f relation.Filter) ([]relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

func (h *ConnectHandler) ListRelations(ctx context.Context, request *connect.Request[frontierv1beta1.ListRelationsRequest]) (*connect.Response[frontierv1beta1.ListRelationsResponse], error) {
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
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, r := range relationsList {
		relationPB, err := transformRelationV2ToPB(r)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		relations = append(relations, relationPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListRelationsResponse{
		Relations: relations,
	}), nil
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
