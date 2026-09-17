package v1beta1connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/aggregates/orgserviceuser"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) SearchOrganizationServiceUsers(ctx context.Context, request *connect.Request[frontierv1beta1.SearchOrganizationServiceUsersRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationServiceUsersResponse], error) {
	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), orgserviceuser.AggregatedServiceUser{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	// without this a bad filter value, say a datetime that cannot be parsed, reaches
	// postgres and comes back as an internal error instead of a bad request
	if err := rql.ValidateQuery(rqlQuery, orgserviceuser.AggregatedServiceUser{}); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	serviceUsersData, err := h.orgServiceUserService.Search(ctx, request.Msg.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("SearchOrganizationServiceUsers.Search: org_id=%s: %w", request.Msg.GetId(), err))
	}

	var orgServiceUsers []*frontierv1beta1.SearchOrganizationServiceUsersResponse_OrganizationServiceUser
	for _, v := range serviceUsersData.ServiceUsers {
		orgServiceUsers = append(orgServiceUsers, transformAggregatedServiceUserToPB(v))
	}

	return connect.NewResponse(&frontierv1beta1.SearchOrganizationServiceUsersResponse{
		OrganizationServiceUsers: orgServiceUsers,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(serviceUsersData.Pagination.Offset),
			Limit:  uint32(serviceUsersData.Pagination.Limit),
		},
		Group: nil,
	}), nil
}

func transformAggregatedServiceUserToPB(v orgserviceuser.AggregatedServiceUser) *frontierv1beta1.SearchOrganizationServiceUsersResponse_OrganizationServiceUser {
	var projects []*frontierv1beta1.SearchOrganizationServiceUsersResponse_Project
	for _, project := range v.Projects {
		projects = append(projects, &frontierv1beta1.SearchOrganizationServiceUsersResponse_Project{
			Id:    project.ID,
			Title: project.Title,
			Name:  project.Name,
		})
	}

	return &frontierv1beta1.SearchOrganizationServiceUsersResponse_OrganizationServiceUser{
		Id:        v.ID,
		Title:     v.Title,
		CreatedAt: timestamppb.New(v.CreatedAt),
		OrgId:     v.OrgID,
		Projects:  projects,
	}
}
