package v1beta1

import (
	"context"
	"errors"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/internal/bootstrap/schema"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/raystack/frontier/core/relation"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RelationService interface {
	Get(ctx context.Context, id string) (relation.Relation, error)
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	List(ctx context.Context, f relation.Filter) ([]relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

var (
	grpcRelationNotFoundErr   = status.Errorf(codes.NotFound, "relation doesn't exist")
	ErrNamespaceSplitNotation = errors.New("subject/object should be provided as 'namespace:uuid'")
)

func (h Handler) ListRelations(ctx context.Context, request *frontierv1beta1.ListRelationsRequest) (*frontierv1beta1.ListRelationsResponse, error) {
	var err error
	var subject relation.Subject
	var object relation.Object

	if request.GetSubject() != "" {
		subject.Namespace, subject.ID, err = schema.SplitNamespaceAndResourceID(request.GetSubject())
		if err != nil {
			return nil, ErrNamespaceSplitNotation
		}
	}
	if request.GetObject() != "" {
		object.Namespace, object.ID, err = schema.SplitNamespaceAndResourceID(request.GetObject())
		if err != nil {
			return nil, ErrNamespaceSplitNotation
		}
	}

	var relations []*frontierv1beta1.Relation
	relationsList, err := h.relationService.List(ctx, relation.Filter{
		Subject: subject,
		Object:  object,
	})
	if err != nil {
		return nil, err
	}

	for _, r := range relationsList {
		relationPB, err := transformRelationV2ToPB(r)
		if err != nil {
			return nil, err
		}

		relations = append(relations, relationPB)
	}

	return &frontierv1beta1.ListRelationsResponse{
		Relations: relations,
	}, nil
}

func (h Handler) CreateRelation(ctx context.Context, request *frontierv1beta1.CreateRelationRequest) (*frontierv1beta1.CreateRelationResponse, error) {
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	subjectNamespace, subjectID, err := schema.SplitNamespaceAndResourceID(request.GetBody().GetSubject())
	if err != nil {
		return nil, ErrNamespaceSplitNotation
	}
	objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(request.GetBody().GetObject())
	if err != nil {
		return nil, ErrNamespaceSplitNotation
	}

	// If Principal is a user, then we will get ID for that user as Subject.ID
	if subjectNamespace == schema.UserPrincipal {
		if !utils.IsValidUUID(subjectID) {
			// could be email
			fetchedUser, err := h.userService.GetByEmail(ctx, subjectID)
			if err != nil {
				return nil, grpcUserNotFoundError
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
			SubRelationName: request.GetBody().GetSubjectSubRelation(),
		},
		RelationName: request.GetBody().GetRelation(),
	})
	if err != nil {
		switch {
		case errors.Is(err, relation.ErrInvalidDetail):
			return nil, grpcBadBodyError
		default:
			return nil, err
		}
	}

	relationPB, err := transformRelationV2ToPB(newRelation)
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.CreateRelationResponse{
		Relation: relationPB,
	}, nil
}

func (h Handler) GetRelation(ctx context.Context, request *frontierv1beta1.GetRelationRequest) (*frontierv1beta1.GetRelationResponse, error) {
	fetchedRelation, err := h.relationService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, relation.ErrNotExist),
			errors.Is(err, relation.ErrInvalidUUID),
			errors.Is(err, relation.ErrInvalidID):
			return nil, grpcRelationNotFoundErr
		default:
			return nil, err
		}
	}

	relationPB, err := transformRelationV2ToPB(fetchedRelation)
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.GetRelationResponse{
		Relation: relationPB,
	}, nil
}

func (h Handler) DeleteRelation(ctx context.Context, request *frontierv1beta1.DeleteRelationRequest) (*frontierv1beta1.DeleteRelationResponse, error) {
	subjectNamespace, subjectID, err := schema.SplitNamespaceAndResourceID(request.GetSubject())
	if err != nil {
		return nil, ErrNamespaceSplitNotation
	}
	objectNamespace, objectID, err := schema.SplitNamespaceAndResourceID(request.GetObject())
	if err != nil {
		return nil, ErrNamespaceSplitNotation
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
		RelationName: request.GetRelation(),
	})
	if err != nil {
		switch {
		case errors.Is(err, relation.ErrNotExist),
			errors.Is(err, relation.ErrInvalidUUID),
			errors.Is(err, relation.ErrInvalidID):
			return nil, grpcRelationNotFoundErr
		default:
			return nil, err
		}
	}

	return &frontierv1beta1.DeleteRelationResponse{}, nil
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
