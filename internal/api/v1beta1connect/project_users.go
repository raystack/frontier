package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/aggregates/projectusers"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
)

type ProjectUsersService interface {
	Search(ctx context.Context, id string, query *rql.Query) (projectusers.ProjectUsers, error)
}

func (h *ConnectHandler) SearchProjectUsers(ctx context.Context, request *connect.Request[frontierv1beta1.SearchProjectUsersRequest]) (*connect.Response[frontierv1beta1.SearchProjectUsersResponse], error) {
	var projectUsers []*frontierv1beta1.SearchProjectUsersResponse_ProjectUser

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), projectusers.AggregatedUser{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, projectusers.AggregatedUser{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	if len(rqlQuery.Filters) > 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("filters is not supported"))
	}

	if len(rqlQuery.GroupBy) > 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("group by is not supported"))
	}

	if len(rqlQuery.Sort) > 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("sorting is not supported"))
	}

	projectUsersData, err := h.projectUsersService.Search(ctx, request.Msg.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range projectUsersData.Users {
		projectUsers = append(projectUsers, transformAggregatedProjectUserToPB(v))
	}

	return connect.NewResponse(&frontierv1beta1.SearchProjectUsersResponse{
		ProjectUsers: projectUsers,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(projectUsersData.Pagination.Offset),
			Limit:  uint32(projectUsersData.Pagination.Limit),
		},
	}), nil
}

func transformAggregatedProjectUserToPB(v projectusers.AggregatedUser) *frontierv1beta1.SearchProjectUsersResponse_ProjectUser {
	return &frontierv1beta1.SearchProjectUsersResponse_ProjectUser{
		Id:         v.ID,
		Name:       v.Name,
		Email:      v.Email,
		Title:      v.Title,
		Avatar:     v.Avatar,
		RoleNames:  v.RoleNames,
		RoleTitles: v.RoleTitles,
		RoleIds:    v.RoleIDs,
		ProjectId:  v.ProjectID,
	}
}
