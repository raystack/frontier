package v1beta1

import (
	"context"
	"errors"
	"fmt"

	svc "github.com/raystack/frontier/core/aggregates/orgserviceusercredentials"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgServiceUserCredentialsService interface {
	Search(ctx context.Context, id string, query *rql.Query) (svc.OrganizationServiceUserCredentials, error)
}

func (h Handler) SearchOrganizationServiceUserCredentials(ctx context.Context, request *frontierv1beta1.SearchOrganizationServiceUserCredentialsRequest) (*frontierv1beta1.SearchOrganizationServiceUserCredentialsResponse, error) {
	rqlQuery, err := utils.TransformProtoToRQL(request.GetQuery(), svc.AggregatedServiceUserCredential{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to read rql query: %v", err))
	}

	credentialsData, err := h.orgServiceUserCredentialsService.Search(ctx, request.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, err
	}

	var orgCredentials []*frontierv1beta1.SearchOrganizationServiceUserCredentialsResponse_OrganizationServiceUserCredential
	for _, v := range credentialsData.Credentials {
		orgCredentials = append(orgCredentials, transformAggregatedServiceUserCredentialToPB(v))
	}

	return &frontierv1beta1.SearchOrganizationServiceUserCredentialsResponse{
		OrganizationServiceuserCredentials: orgCredentials,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(credentialsData.Pagination.Offset),
			Limit:  uint32(credentialsData.Pagination.Limit),
		},
		Group: nil,
	}, nil
}

func transformAggregatedServiceUserCredentialToPB(v svc.AggregatedServiceUserCredential) *frontierv1beta1.SearchOrganizationServiceUserCredentialsResponse_OrganizationServiceUserCredential {
	return &frontierv1beta1.SearchOrganizationServiceUserCredentialsResponse_OrganizationServiceUserCredential{
		Title:            v.Title,
		ServiceuserTitle: v.ServiceUserTitle,
		CreatedAt:        timestamppb.New(v.CreatedAt),
		OrgId:            v.OrgID,
	}
}
