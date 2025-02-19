package v1beta1

import (
	"context"
	"fmt"
	"github.com/raystack/frontier/core/kyc"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

type KycService interface {
	GetKyc(ctx context.Context, idOrSlug string) (kyc.KYC, error)
	SetKyc(ctx context.Context, idOrSlug string) (kyc.KYC, error)
}

func (h Handler) SetOrganizationKyc(ctx context.Context, request *frontierv1beta1.SetOrganizationKycRequest) (*frontierv1beta1.SetOrganizationKycResponse, error) {
	fmt.Println("meh")
	return &frontierv1beta1.SetOrganizationKycResponse{
		OrganizationKyc: &frontierv1beta1.OrganizationKyc{
			OrgId:  "blah",
			Status: true,
			Link:   "abcd",
		},
	}, nil
}

func (h Handler) GetOrganizationKyc(ctx context.Context, request *frontierv1beta1.GetOrganizationKycRequest) (*frontierv1beta1.GetOrganizationKycResponse, error) {
	fmt.Println("lol")
	return &frontierv1beta1.GetOrganizationKycResponse{
		OrganizationKyc: &frontierv1beta1.OrganizationKyc{
			OrgId:  "blah",
			Status: true,
			Link:   "abcd",
		},
	}, nil
}
