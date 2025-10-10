package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/organization"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuditService interface {
	List(ctx context.Context, filter audit.Filter) ([]audit.Log, error)
	GetByID(ctx context.Context, id string) (audit.Log, error)
	Create(ctx context.Context, log *audit.Log) error
}

func (h *ConnectHandler) ListOrganizationAuditLogs(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationAuditLogsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationAuditLogsResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	var logs []*frontierv1beta1.AuditLog
	logList, err := h.auditService.List(ctx, audit.Filter{
		OrgID:        orgResp.ID,
		Source:       request.Msg.GetSource(),
		Action:       request.Msg.GetAction(),
		StartTime:    request.Msg.GetStartTime().AsTime(),
		EndTime:      request.Msg.GetEndTime().AsTime(),
		IgnoreSystem: request.Msg.GetIgnoreSystem(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	for _, v := range logList {
		logs = append(logs, transformAuditLogToPB(v))
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationAuditLogsResponse{
		Logs: logs,
	}), nil
}

func (h *ConnectHandler) CreateOrganizationAuditLogs(ctx context.Context, request *connect.Request[frontierv1beta1.CreateOrganizationAuditLogsRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationAuditLogsResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	for _, log := range request.Msg.GetLogs() {
		if log.GetSource() == "" || log.GetAction() == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		if err := h.auditService.Create(ctx, &audit.Log{
			ID:    log.GetId(),
			OrgID: orgResp.ID,

			Source:    log.GetSource(),
			Action:    log.GetAction(),
			CreatedAt: log.GetCreatedAt().AsTime(),
			Actor: audit.Actor{
				ID:   log.GetActor().GetId(),
				Type: log.GetActor().GetType(),
				Name: log.GetActor().GetName(),
			},
			Target: audit.Target{
				ID:   log.GetTarget().GetId(),
				Type: log.GetTarget().GetType(),
				Name: log.GetTarget().GetName(),
			},
			Metadata: log.GetContext(),
		}); err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	return connect.NewResponse(&frontierv1beta1.CreateOrganizationAuditLogsResponse{}), nil
}

func (h *ConnectHandler) GetOrganizationAuditLog(ctx context.Context, request *connect.Request[frontierv1beta1.GetOrganizationAuditLogRequest]) (*connect.Response[frontierv1beta1.GetOrganizationAuditLogResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	log, err := h.auditService.GetByID(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetOrganizationAuditLogResponse{
		Log: transformAuditLogToPB(log),
	}), nil
}

func transformAuditLogToPB(log audit.Log) *frontierv1beta1.AuditLog {
	return &frontierv1beta1.AuditLog{
		Id:        log.ID,
		Source:    log.Source,
		Action:    log.Action,
		CreatedAt: timestamppb.New(log.CreatedAt),
		Actor: &frontierv1beta1.AuditLogActor{
			Id:   log.Actor.ID,
			Name: log.Actor.Name,
			Type: log.Actor.Type,
		},
		Target: &frontierv1beta1.AuditLogTarget{
			Id:   log.Target.ID,
			Name: log.Target.Name,
			Type: log.Target.Type,
		},
		Context: log.Metadata,
	}
}
