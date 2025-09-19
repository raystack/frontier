package v1beta1connect

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	IdempotencyReplyHeader = "X-Idempotency-Replayed"
	HttpChunkSize          = 204800 // 200KB
)

type AuditRecordService interface {
	Create(ctx context.Context, record auditrecord.AuditRecord) (auditrecord.AuditRecord, bool, error)
	List(ctx context.Context, query *rql.Query) (auditrecord.AuditRecordsList, error)
	Export(ctx context.Context, query *rql.Query) (io.Reader, string, error)
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
	if request.Msg.GetRequestId() != "" {
		reqID := request.Msg.GetRequestId()
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

func (h *ConnectHandler) ListAuditRecords(ctx context.Context, request *connect.Request[frontierv1beta1.ListAuditRecordsRequest]) (*connect.Response[frontierv1beta1.ListAuditRecordsResponse], error) {
	requestQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), auditrecord.AuditRecordRQLSchema{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to transform rql query: %v", err))
	}
	err = rql.ValidateQuery(requestQuery, auditrecord.AuditRecordRQLSchema{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}
	records, err := h.auditRecordService.List(ctx, requestQuery)
	if err != nil {
		switch {
		case errors.Is(err, auditrecord.ErrInvalidUUID) || errors.Is(err, auditrecord.ErrRepositoryBadInput):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case errors.Is(err, auditrecord.ErrNotFound):
			return nil, connect.NewError(connect.CodeNotFound, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	pbRecords := make([]*frontierv1beta1.AuditRecord, 0)
	for _, auditRecord := range records.AuditRecords {
		pbRecord, err := TransformAuditRecordToPB(auditRecord)
		if err != nil {
			return nil, err
		}
		pbRecords = append(pbRecords, pbRecord.GetAuditRecord())
	}

	var pbGroups *frontierv1beta1.RQLQueryGroupResponse

	if records.Group != nil && len(records.Group.Data) > 0 {
		groupResponse := make([]*frontierv1beta1.RQLQueryGroupData, 0)
		for _, groupItem := range records.Group.Data {
			groupResponse = append(groupResponse, &frontierv1beta1.RQLQueryGroupData{
				Name:  groupItem.Name,
				Count: uint32(groupItem.Count),
			})
		}
		pbGroups = &frontierv1beta1.RQLQueryGroupResponse{
			Name: records.Group.Name,
			Data: groupResponse,
		}
	}

	pagination := &frontierv1beta1.RQLQueryPaginationResponse{
		Offset:     uint32(records.Page.Offset),
		Limit:      uint32(records.Page.Limit),
		TotalCount: uint32(records.Page.TotalCount),
	}

	return connect.NewResponse(&frontierv1beta1.ListAuditRecordsResponse{AuditRecords: pbRecords, Group: pbGroups, Pagination: pagination}), nil
}

func (h *ConnectHandler) ExportAuditRecords(ctx context.Context, request *connect.Request[frontierv1beta1.ExportAuditRecordsRequest], stream *connect.ServerStream[httpbody.HttpBody]) error {
	requestQuery, err := utils.TransformExportProtoToRQL(request.Msg.GetQuery(), auditrecord.AuditRecordRQLSchema{})
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to transform rql query: %v", err))
	}
	err = rql.ValidateQuery(requestQuery, auditrecord.AuditRecordRQLSchema{})
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}
	reader, contentType, err := h.auditRecordService.Export(ctx, requestQuery)
	if err != nil {
		switch {
		case errors.Is(err, auditrecord.ErrInvalidUUID) || errors.Is(err, auditrecord.ErrRepositoryBadInput):
			return connect.NewError(connect.CodeInvalidArgument, err)
		case errors.Is(err, auditrecord.ErrNotFound):
			return connect.NewError(connect.CodeNotFound, err)
		default:
			return connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	// Stream the data using io.Reader
	return streamReaderInChunks(reader, contentType, stream)
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

	var requestID string
	if record.RequestID != nil {
		requestID = *record.RequestID
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
			RequestId:  requestID,
			CreatedAt:  timestamppb.New(record.CreatedAt),
			Metadata:   metaData,
		},
	}, nil
}

// streamReaderInChunks reads from io.Reader and streams data in HTTP-friendly chunks
func streamReaderInChunks(reader io.Reader, contentType string, stream *connect.ServerStream[httpbody.HttpBody]) error {
	buffer := make([]byte, HttpChunkSize)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		if n > 0 {
			msg := &httpbody.HttpBody{
				ContentType: contentType,
				Data:        buffer[:n],
			}
			if sendErr := stream.Send(msg); sendErr != nil {
				return connect.NewError(connect.CodeInternal, ErrInternalServerError)
			}
		}
	}

	return nil
}
