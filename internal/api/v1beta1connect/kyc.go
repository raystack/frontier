package v1beta1connect

import (
	"context"
	"strconv"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/kyc"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type KycService interface {
	GetKyc(context.Context, string) (kyc.KYC, error)
	SetKyc(context.Context, kyc.KYC) (kyc.KYC, error)
	ListKycs(context.Context) ([]kyc.KYC, error)
}

func (h *ConnectHandler) SetOrganizationKyc(ctx context.Context, request *connect.Request[frontierv1beta1.SetOrganizationKycRequest]) (*connect.Response[frontierv1beta1.SetOrganizationKycResponse], error) {
	orgKyc, err := h.orgKycService.SetKyc(ctx, kyc.KYC{
		OrgID:  request.Msg.GetOrgId(),
		Status: request.Msg.GetStatus(),
		Link:   request.Msg.GetLink(),
	})
	if err != nil {
		switch {
		case errors.Is(err, kyc.ErrKycLinkNotSet):
			return nil, connect.NewError(connect.CodeInvalidArgument, kyc.ErrKycLinkNotSet)
		case errors.Is(err, kyc.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeInvalidArgument, kyc.ErrInvalidUUID)
		case errors.Is(err, kyc.ErrOrgDoesntExist):
			return nil, connect.NewError(connect.CodeInvalidArgument, kyc.ErrOrgDoesntExist)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	// Add audit log
	audit.GetAuditor(ctx, orgKyc.OrgID).
		LogWithAttrs(audit.OrgKycUpdatedEvent, audit.OrgTarget(orgKyc.OrgID), map[string]string{
			"status": strconv.FormatBool(orgKyc.Status),
			"link":   orgKyc.Link,
		})

	return connect.NewResponse(&frontierv1beta1.SetOrganizationKycResponse{OrganizationKyc: transformOrgKycToPB(orgKyc)}), nil
}

func (h *ConnectHandler) GetOrganizationKyc(ctx context.Context, request *connect.Request[frontierv1beta1.GetOrganizationKycRequest]) (*connect.Response[frontierv1beta1.GetOrganizationKycResponse], error) {
	orgKyc, err := h.orgKycService.GetKyc(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, kyc.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, kyc.ErrNotExist)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	return connect.NewResponse(&frontierv1beta1.GetOrganizationKycResponse{OrganizationKyc: transformOrgKycToPB(orgKyc)}), nil
}

func (h *ConnectHandler) ListOrganizationsKyc(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationsKycRequest]) (*connect.Response[frontierv1beta1.ListOrganizationsKycResponse], error) {
	orgKycs, err := h.orgKycService.ListKycs(ctx)
	if err != nil {
		switch {
		case errors.Is(err, kyc.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, kyc.ErrNotExist)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	resp := make([]*frontierv1beta1.OrganizationKyc, len(orgKycs))
	for i, orgKyc := range orgKycs {
		resp[i] = transformOrgKycToPB(orgKyc)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationsKycResponse{OrganizationsKyc: resp}), nil
}

func transformOrgKycToPB(k kyc.KYC) *frontierv1beta1.OrganizationKyc {
	return &frontierv1beta1.OrganizationKyc{
		OrgId:     k.OrgID,
		Status:    k.Status,
		Link:      k.Link,
		CreatedAt: timestamppb.New(k.CreatedAt),
		UpdatedAt: timestamppb.New(k.UpdatedAt),
	}
}
