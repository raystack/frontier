package v1beta1

import (
	"context"
	"fmt"

	"github.com/raystack/frontier/core/aggregates/orgusers"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrgUsersService interface {
	Search(ctx context.Context, query *rql.Query) (orgusers.OrgUsers, error)
}

func (h Handler) SearchOrganizationUsers(ctx context.Context, request *frontierv1beta1.SearchOrganizationUsersRequest) (*frontierv1beta1.SearchOrganizationUsersResponse, error) {
	var orgUsers []*frontierv1beta1.SearchOrganizationUsersResponse_OrganizationUser

	rqlQuery, err := utils.TransformProtoToRQL(request.GetQuery(), orgusers.AggregatedUser{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, orgusers.AggregatedUser{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to validate rql query: %v", err))
	}

	orgUsersData, err := h.orgUsersService.Search(ctx, rqlQuery)
	if err != nil {
		return nil, err
	}

	for _, v := range orgUsersData.Users {
		orgUsers = append(orgUsers, transformAggregatedUserToPB(v))
	}

	groupResponse := make([]*frontierv1beta1.RQLQueryGroupData, 0)
	for _, groupItem := range orgUsersData.Group.Data {
		groupResponse = append(groupResponse, &frontierv1beta1.RQLQueryGroupData{
			Name:  groupItem.Name,
			Count: uint32(groupItem.Count),
		})
	}
	return &frontierv1beta1.SearchOrganizationUsersResponse{
		OrgUsers: orgUsers,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(orgUsersData.Pagination.Offset),
			Limit:  uint32(orgUsersData.Pagination.Limit),
		},
		Group: &frontierv1beta1.RQLQueryGroupResponse{
			Name: orgUsersData.Group.Name,
			Data: groupResponse,
		},
	}, nil
}

func transformAggregatedUserToPB(v orgusers.AggregatedUser) *frontierv1beta1.SearchOrganizationUsersResponse_OrganizationUser {
	return &frontierv1beta1.SearchOrganizationUsersResponse_OrganizationUser{
		Id:        v.ID,
		Name:      v.Name,
		Title:     v.Title,
		Email:     v.Email,
		State:     string(v.State),
		Avatar:    v.Avatar,
		RoleName:  v.RoleName,
		RoleTitle: v.RoleTitle,
		RoleId:    v.RoleID,
	}
}
