package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	svc "github.com/raystack/frontier/core/aggregates/orgserviceusercredentials"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgServiceUserCredentialsService interface {
	Search(ctx context.Context, id string, query *rql.Query) (svc.OrganizationServiceUserCredentials, error)
}

func (h *ConnectHandler) SearchOrganizationServiceUserCredentials(ctx context.Context, request *connect.Request[frontierv1beta1.SearchOrganizationServiceUserCredentialsRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationServiceUserCredentialsResponse], error) {
	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), svc.AggregatedServiceUserCredential{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, svc.AggregatedServiceUserCredential{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	credentialsData, err := h.orgServiceUserCredentialsService.Search(ctx, request.Msg.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var orgCredentials []*frontierv1beta1.SearchOrganizationServiceUserCredentialsResponse_OrganizationServiceUserCredential
	for _, v := range credentialsData.Credentials {
		orgCredentials = append(orgCredentials, transformAggregatedServiceUserCredentialToPB(v))
	}

	return connect.NewResponse(&frontierv1beta1.SearchOrganizationServiceUserCredentialsResponse{
		OrganizationServiceuserCredentials: orgCredentials,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(credentialsData.Pagination.Offset),
			Limit:  uint32(credentialsData.Pagination.Limit),
		},
		Group: nil,
	}), nil
}

func transformAggregatedServiceUserCredentialToPB(v svc.AggregatedServiceUserCredential) *frontierv1beta1.SearchOrganizationServiceUserCredentialsResponse_OrganizationServiceUserCredential {
	return &frontierv1beta1.SearchOrganizationServiceUserCredentialsResponse_OrganizationServiceUserCredential{
		Title:            v.Title,
		ServiceuserTitle: v.ServiceUserTitle,
		CreatedAt:        timestamppb.New(v.CreatedAt),
		OrgId:            v.OrgID,
	}
}
