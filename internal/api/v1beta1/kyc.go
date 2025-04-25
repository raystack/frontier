package v1beta1

import (
	"context"
	"errors"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/kyc"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type KycService interface {
	GetKyc(context.Context, string) (kyc.KYC, error)
	SetKyc(context.Context, kyc.KYC) (kyc.KYC, error)
	ListKycs(context.Context) ([]kyc.KYC, error)
}

var grpcOrgKycNotFoundErr = status.Errorf(codes.NotFound, kyc.ErrNotExist.Error())

func (h Handler) SetOrganizationKyc(ctx context.Context, request *frontierv1beta1.SetOrganizationKycRequest) (*frontierv1beta1.SetOrganizationKycResponse, error) {
	orgKyc, err := h.orgKycService.SetKyc(ctx, kyc.KYC{
		OrgID:  request.GetOrgId(),
		Status: request.GetStatus(),
		Link:   request.GetLink(),
	})
	if err != nil {
		switch {
		case errors.Is(err, kyc.ErrKycLinkNotSet):
			return nil, ErrInvalidInput(kyc.ErrKycLinkNotSet.Error())
		case errors.Is(err, kyc.ErrInvalidUUID):
			return nil, ErrInvalidInput(kyc.ErrInvalidUUID.Error())
		case errors.Is(err, kyc.ErrOrgDoesntExist):
			return nil, ErrInvalidInput(kyc.ErrOrgDoesntExist.Error())
		default:
			return nil, err
		}
	}

	// Add audit log
	audit.GetAuditor(ctx, orgKyc.OrgID).
		LogWithAttrs(audit.OrgKycUpdatedEvent, audit.OrgTarget(orgKyc.OrgID), map[string]string{
			"status": boolToString(orgKyc.Status),
			"link":   orgKyc.Link,
		})

	return &frontierv1beta1.SetOrganizationKycResponse{OrganizationKyc: transformOrgKycToPB(orgKyc)}, nil
}

func (h Handler) GetOrganizationKyc(ctx context.Context, request *frontierv1beta1.GetOrganizationKycRequest) (*frontierv1beta1.GetOrganizationKycResponse, error) {
	orgKyc, err := h.orgKycService.GetKyc(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, kyc.ErrNotExist):
			return nil, grpcOrgKycNotFoundErr
		default:
			return nil, err
		}
	}
	return &frontierv1beta1.GetOrganizationKycResponse{OrganizationKyc: transformOrgKycToPB(orgKyc)}, nil
}

func (h Handler) ListOrganizationsKyc(ctx context.Context, request *frontierv1beta1.ListOrganizationsKycRequest) (*frontierv1beta1.ListOrganizationsKycResponse, error) {
	orgKycs, err := h.orgKycService.ListKycs(ctx)
	if err != nil {
		switch {
		case errors.Is(err, kyc.ErrNotExist):
			return nil, grpcOrgKycNotFoundErr
		default:
			return nil, err
		}
	}
	resp := make([]*frontierv1beta1.OrganizationKyc, len(orgKycs))
	for i, orgKyc := range orgKycs {
		resp[i] = transformOrgKycToPB(orgKyc)
	}

	return &frontierv1beta1.ListOrganizationsKycResponse{OrganizationsKyc: resp}, nil
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

// Helper function to convert boolean to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
