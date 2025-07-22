package v1beta1

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/core/aggregates/orgserviceuser"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgServiceUserService interface {
	Search(ctx context.Context, orgID string, query *rql.Query) (orgserviceuser.OrganizationServiceUsers, error)
}

func (h Handler) SearchOrganizationServiceUsers(ctx context.Context, request *frontierv1beta1.SearchOrganizationServiceUsersRequest) (*frontierv1beta1.SearchOrganizationServiceUsersResponse, error) {
	rqlQuery, err := utils.TransformProtoToRQL(request.GetQuery(), orgserviceuser.AggregatedServiceUser{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to read rql query: %v", err))
	}

	serviceUsersData, err := h.orgServiceUserService.Search(ctx, request.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, err
	}

	var orgServiceUsers []*frontierv1beta1.SearchOrganizationServiceUsersResponse_OrganizationServiceUser
	for _, v := range serviceUsersData.ServiceUsers {
		orgServiceUsers = append(orgServiceUsers, transformAggregatedServiceUserToPB(v))
	}

	return &frontierv1beta1.SearchOrganizationServiceUsersResponse{
		OrganizationServiceUsers: orgServiceUsers,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(serviceUsersData.Pagination.Offset),
			Limit:  uint32(serviceUsersData.Pagination.Limit),
		},
		Group: nil,
	}, nil
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
