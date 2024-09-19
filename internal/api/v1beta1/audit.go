package v1beta1

import (
	"context"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuditService interface {
	List(ctx context.Context, filter audit.Filter) ([]audit.Log, error)
	GetByID(ctx context.Context, id string) (audit.Log, error)
	Create(ctx context.Context, log *audit.Log) error
}

func (h Handler) ListOrganizationAuditLogs(ctx context.Context, request *frontierv1beta1.ListOrganizationAuditLogsRequest) (*frontierv1beta1.ListOrganizationAuditLogsResponse, error) {
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	var logs []*frontierv1beta1.AuditLog
	logList, err := h.auditService.List(ctx, audit.Filter{
		OrgID:        orgResp.ID,
		Source:       request.GetSource(),
		Action:       request.GetAction(),
		StartTime:    request.GetStartTime().AsTime(),
		EndTime:      request.GetEndTime().AsTime(),
		IgnoreSystem: request.GetIgnoreSystem(),
	})
	if err != nil {
		return nil, err
	}
	for _, v := range logList {
		logs = append(logs, transformAuditLogToPB(v))
	}

	return &frontierv1beta1.ListOrganizationAuditLogsResponse{
		Logs: logs,
	}, nil
}

func (h Handler) CreateOrganizationAuditLogs(ctx context.Context, request *frontierv1beta1.CreateOrganizationAuditLogsRequest) (*frontierv1beta1.CreateOrganizationAuditLogsResponse, error) {
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	for _, log := range request.GetLogs() {
		if log.GetSource() == "" || log.GetAction() == "" {
			return nil, grpcBadBodyError
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
			return nil, err
		}
	}
	return &frontierv1beta1.CreateOrganizationAuditLogsResponse{}, nil
}

func (h Handler) GetOrganizationAuditLog(ctx context.Context, request *frontierv1beta1.GetOrganizationAuditLogRequest) (*frontierv1beta1.GetOrganizationAuditLogResponse, error) {
	_, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	log, err := h.auditService.GetByID(ctx, request.GetId())
	if err != nil {
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
