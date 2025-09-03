package v1beta1connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/pkg/errors"
	"github.com/raystack/frontier/core/aggregates/userorgs"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserOrgsService interface {
	Search(ctx context.Context, id string, query *rql.Query) (userorgs.UserOrgs, error)
}

func (h *ConnectHandler) SearchUserOrganizations(ctx context.Context, request *connect.Request[frontierv1beta1.SearchUserOrganizationsRequest]) (*connect.Response[frontierv1beta1.SearchUserOrganizationsResponse], error) {
	var userOrgs []*frontierv1beta1.SearchUserOrganizationsResponse_UserOrganization

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), userorgs.AggregatedUserOrganization{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, userorgs.AggregatedUserOrganization{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	if len(rqlQuery.GroupBy) > 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("group by is not supported"))
	}

	userOrgsData, err := h.userOrgsService.Search(ctx, request.Msg.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range userOrgsData.Organizations {
		userOrgs = append(userOrgs, transformAggregatedUserOrganizationToPB(v))
	}

	groupResponse := make([]*frontierv1beta1.RQLQueryGroupData, 0)
	for _, groupItem := range userOrgsData.Group.Data {
		groupResponse = append(groupResponse, &frontierv1beta1.RQLQueryGroupData{
			Name:  groupItem.Name,
			Count: uint32(groupItem.Count),
		})
	}

	return connect.NewResponse(&frontierv1beta1.SearchUserOrganizationsResponse{
		UserOrganizations: userOrgs,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(userOrgsData.Pagination.Offset),
			Limit:  uint32(userOrgsData.Pagination.Limit),
		},
		Group: &frontierv1beta1.RQLQueryGroupResponse{
			Name: userOrgsData.Group.Name,
			Data: groupResponse,
		},
	}), nil
}

// Helper function to transform aggregated organization to protobuf
func transformAggregatedUserOrganizationToPB(userOrg userorgs.AggregatedUserOrganization) *frontierv1beta1.SearchUserOrganizationsResponse_UserOrganization {
	return &frontierv1beta1.SearchUserOrganizationsResponse_UserOrganization{
		OrgId:        userOrg.OrgID,
		OrgTitle:     userOrg.OrgTitle,
		OrgName:      userOrg.OrgName,
		OrgAvatar:    userOrg.OrgAvatar,
		ProjectCount: userOrg.ProjectCount,
		RoleNames:    userOrg.RoleNames,
		RoleTitles:   userOrg.RoleTitles,
		RoleIds:      userOrg.RoleIDs,
		OrgJoinedOn:  timestamppb.New(userOrg.OrgJoinedOn),
		UserId:       userOrg.UserID,
	}
}
