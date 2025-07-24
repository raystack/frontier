package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"github.com/raystack/frontier/core/aggregates/orgusers"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgUsersService interface {
	Search(ctx context.Context, id string, query *rql.Query) (orgusers.OrgUsers, error)
	Export(ctx context.Context, orgID string) ([]byte, string, error)
}

func (h *ConnectHandler) SearchOrganizationUsers(ctx context.Context, request *connect.Request[frontierv1beta1.SearchOrganizationUsersRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationUsersResponse], error) {
	var orgUsers []*frontierv1beta1.SearchOrganizationUsersResponse_OrganizationUser

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), orgusers.AggregatedUser{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, orgusers.AggregatedUser{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	if len(rqlQuery.GroupBy) > 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("group by is not supported"))
	}

	orgUsersData, err := h.orgUsersService.Search(ctx, request.Msg.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
	return connect.NewResponse(&frontierv1beta1.SearchOrganizationUsersResponse{
		OrgUsers: orgUsers,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(orgUsersData.Pagination.Offset),
			Limit:  uint32(orgUsersData.Pagination.Limit),
		},
		Group: &frontierv1beta1.RQLQueryGroupResponse{
			Name: orgUsersData.Group.Name,
			Data: groupResponse,
		},
	}), nil
}

func transformAggregatedUserToPB(v orgusers.AggregatedUser) *frontierv1beta1.SearchOrganizationUsersResponse_OrganizationUser {
	return &frontierv1beta1.SearchOrganizationUsersResponse_OrganizationUser{
		Id:             v.ID,
		Name:           v.Name,
		Title:          v.Title,
		Email:          v.Email,
		State:          string(v.State),
		Avatar:         v.Avatar,
		RoleNames:      v.RoleNames,
		RoleTitles:     v.RoleTitles,
		RoleIds:        v.RoleIDs,
		OrganizationId: v.OrgID,
		OrgJoinedAt:    timestamppb.New(v.OrgJoinedAt),
	}
}
