package v1beta1

import (
	"context"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/audit"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuditService interface {
	List(ctx context.Context, filter audit.Filter) ([]audit.Log, error)
	GetByID(ctx context.Context, id string) (audit.Log, error)
	Create(ctx context.Context, log *audit.Log) error
}

func (h Handler) ListOrganizationAuditLogs(ctx context.Context, request *frontierv1beta1.ListOrganizationAuditLogsRequest) (*frontierv1beta1.ListOrganizationAuditLogsResponse, error) {
	logger := grpczap.Extract(ctx)

	var logs []*frontierv1beta1.AuditLog
	logList, err := h.auditService.List(ctx, audit.Filter{
		OrgID:     request.GetOrgId(),
		Source:    request.GetSource(),
		Action:    request.GetAction(),
		StartTime: request.GetStartTime().AsTime(),
		EndTime:   request.GetEndTime().AsTime(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	for _, v := range logList {
		logs = append(logs, transformAuditLogToPB(v))
	}

	return &frontierv1beta1.ListOrganizationAuditLogsResponse{
		Logs: logs,
	}, nil
}

func (h Handler) CreateOrganizationAuditLogs(ctx context.Context, request *frontierv1beta1.CreateOrganizationAuditLogsRequest) (*frontierv1beta1.CreateOrganizationAuditLogsResponse, error) {
	logger := grpczap.Extract(ctx)

	for _, log := range request.GetLogs() {
		if err := h.auditService.Create(ctx, &audit.Log{
			ID:    log.GetId(),
			OrgID: request.GetOrgId(),

			Source:    log.Source,
			Action:    log.Action,
			CreatedAt: log.CreatedAt.AsTime(),
			Actor: audit.Actor{
				ID:   log.GetActor().GetId(),
				Name: log.GetActor().GetName(),
			},
			Target: audit.Target{
				ID:   log.GetTarget().GetId(),
				Name: log.GetTarget().GetName(),
			},
			Metadata: log.Context,
		}); err != nil {
			logger.Error(err.Error())
			return nil, err
		}
	}
	return &frontierv1beta1.CreateOrganizationAuditLogsResponse{}, nil
}

func (h Handler) GetOrganizationAuditLog(ctx context.Context, request *frontierv1beta1.GetOrganizationAuditLogRequest) (*frontierv1beta1.GetOrganizationAuditLogResponse, error) {
	logger := grpczap.Extract(ctx)

	log, err := h.auditService.GetByID(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	return &frontierv1beta1.GetOrganizationAuditLogResponse{
		Log: transformAuditLogToPB(log),
	}, nil
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
		},
		Target: &frontierv1beta1.AuditLogTarget{
			Id:   log.Target.ID,
			Name: log.Target.Name,
		},
		Context: log.Metadata,
	}
}
