package v1beta1connect

import (
	"context"
	"errors"
	"slices"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	IdempotencyReplyHeader = "X-Idempotency-Replayed"
)

type AuditRecordService interface {
	Create(ctx context.Context, record auditrecord.AuditRecord) (auditrecord.AuditRecord, bool, error)
}

func (h *ConnectHandler) CreateAuditRecord(ctx context.Context, request *connect.Request[frontierv1beta1.CreateAuditRecordRequest]) (*connect.Response[frontierv1beta1.CreateAuditRecordResponse], error) {
	// Validate the request parameters
	if err := request.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	actor := request.Msg.GetActor()
	// Validate the actor type for non-system actors. ZeroUUID is a special case for system actors.
	if actor.GetId() != uuid.Nil.String() && !slices.Contains([]string{schema.ServiceUserPrincipal, schema.UserPrincipal}, actor.GetType()) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidActorType)
	}

	resource := request.Msg.GetResource()
	target := request.Msg.GetTarget()

	// Check if the target is provided (not empty)
	var targetPtr *auditrecord.Target
	if target.GetId() != "" || target.GetType() != "" || target.GetName() != "" || target.GetMetadata() != nil {
		targetPtr = &auditrecord.Target{
			ID:       target.GetId(),
			Type:     target.GetType(),
			Name:     target.GetName(),
			Metadata: metadata.BuildFromProto(target.GetMetadata()),
		}
	}

	var requestID *string
	if request.Msg.GetReqId() != "" {
		reqID := request.Msg.GetReqId()
		requestID = &reqID
	}

	record, idempotentReply, err := h.auditRecordService.Create(ctx, auditrecord.AuditRecord{
		Event: request.Msg.GetEvent(),
		Actor: auditrecord.Actor{
			ID:       actor.GetId(),
			Type:     actor.GetType(),
			Name:     actor.GetName(),
			Metadata: metadata.BuildFromProto(actor.GetMetadata()),
		},
		Resource: auditrecord.Resource{
			ID:       resource.GetId(),
			Type:     resource.GetType(),
			Name:     resource.GetName(),
			Metadata: metadata.BuildFromProto(resource.GetMetadata()),
		},
		Target:         targetPtr,
		OccurredAt:     request.Msg.GetOccurredAt().AsTime(),
		OrgID:          request.Msg.GetOrgId(),
		RequestID:      requestID,
		Metadata:       metadata.BuildFromProto(request.Msg.GetMetadata()),
		IdempotencyKey: request.Msg.GetIdempotencyKey(),
	})
	if err != nil {
		switch {
		case errors.Is(err, auditrecord.ErrIdempotencyKeyConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		default:
			return nil, err
		}
	}
	transformedRecord, err := TransformAuditRecordToPB(record)
	if err != nil {
		return nil, err
	}
	response := connect.NewResponse(transformedRecord)
	if idempotentReply {
		response.Header().Set(IdempotencyReplyHeader, "true")
	}
	return response, nil
}

func TransformAuditRecordToPB(record auditrecord.AuditRecord) (*frontierv1beta1.CreateAuditRecordResponse, error) {
	actorMetaData, err := record.Actor.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	resourceMetaData, err := record.Resource.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	var targetMetaData *structpb.Struct
	var target *frontierv1beta1.AuditRecordTarget
	if record.Target != nil {
		targetMetaData, err = record.Target.Metadata.ToStructPB()
		if err != nil {
			return nil, err
		}
		target = &frontierv1beta1.AuditRecordTarget{
			Id:       record.Target.ID,
			Type:     record.Target.Type,
			Name:     record.Target.Name,
			Metadata: targetMetaData,
		}
	}

	var reqId string
	if record.RequestID != nil {
		reqId = *record.RequestID
	}

	metaData, err := record.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.CreateAuditRecordResponse{
		AuditRecord: &frontierv1beta1.AuditRecord{
			Id:    record.ID,
			Event: record.Event,
			Actor: &frontierv1beta1.AuditRecordActor{
				Id:       record.Actor.ID,
				Type:     record.Actor.Type,
				Name:     record.Actor.Name,
				Metadata: actorMetaData,
			},
			Resource: &frontierv1beta1.AuditRecordResource{
				Id:       record.Resource.ID,
				Type:     record.Resource.Type,
				Name:     record.Resource.Name,
				Metadata: resourceMetaData,
			},
			Target:     target,
			OccurredAt: timestamppb.New(record.OccurredAt),
			OrgId:      record.OrgID,
			ReqId:      reqId,
			CreatedAt:  timestamppb.New(record.CreatedAt),
			Metadata:   metaData,
		},
	}, nil
}
