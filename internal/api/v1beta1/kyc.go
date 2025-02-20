package v1beta1

import (
	"context"
	"github.com/raystack/frontier/core/kyc"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type KycService interface {
	GetKyc(context.Context, string) (kyc.KYC, error)
	SetKyc(context.Context, kyc.KYC) (kyc.KYC, error)
}

func (h Handler) SetOrganizationKyc(ctx context.Context, request *frontierv1beta1.SetOrganizationKycRequest) (*frontierv1beta1.SetOrganizationKycResponse, error) {
	orgKyc, err := h.orgKycService.SetKyc(ctx, kyc.KYC{
		OrgID:  request.GetOrgId(),
		Status: request.GetStatus(),
		Link:   request.GetLink(),
	})
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.SetOrganizationKycResponse{OrganizationKyc: transformOrgKycToPB(orgKyc)}, nil
}

func (h Handler) GetOrganizationKyc(ctx context.Context, request *frontierv1beta1.GetOrganizationKycRequest) (*frontierv1beta1.GetOrganizationKycResponse, error) {
	orgKyc, err := h.orgKycService.GetKyc(ctx, request.GetOrgId())
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.GetOrganizationKycResponse{OrganizationKyc: transformOrgKycToPB(orgKyc)}, nil
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
