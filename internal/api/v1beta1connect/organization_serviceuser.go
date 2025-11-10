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
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgServiceUserService interface {
	Search(ctx context.Context, orgID string, query *rql.Query) (orgserviceuser.OrganizationServiceUsers, error)
}

func (h *ConnectHandler) SearchOrganizationServiceUsers(ctx context.Context, request *connect.Request[frontierv1beta1.SearchOrganizationServiceUsersRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationServiceUsersResponse], error) {
	errorLogger := NewErrorLogger()

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), orgserviceuser.AggregatedServiceUser{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	serviceUsersData, err := h.orgServiceUserService.Search(ctx, request.Msg.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		}
		errorLogger.LogServiceError(ctx, request, "SearchOrganizationServiceUsers.Search", err,
			zap.String("org_id", request.Msg.GetId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
