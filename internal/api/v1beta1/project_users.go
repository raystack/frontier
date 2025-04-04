package v1beta1

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/aggregates/projectusers"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProjectUsersService interface {
	Search(ctx context.Context, id string, query *rql.Query) (projectusers.ProjectUsers, error)
}

func (h Handler) SearchProjectUsers(ctx context.Context, request *frontierv1beta1.SearchProjectUsersRequest) (*frontierv1beta1.SearchProjectUsersResponse, error) {
	var projectUsers []*frontierv1beta1.SearchProjectUsersResponse_ProjectUser

	rqlQuery, err := utils.TransformProtoToRQL(request.GetQuery(), projectusers.AggregatedUser{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, projectusers.AggregatedUser{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to validate rql query: %v", err))
	}

	if len(rqlQuery.Filters) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "filters is not supported")
	}

	if len(rqlQuery.GroupBy) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "group by is not supported")
	}

	if len(rqlQuery.Sort) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "sorting is not supported")
	}

	projectUsersData, err := h.projectUsersService.Search(ctx, request.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, err
	}

	for _, v := range projectUsersData.Users {
		projectUsers = append(projectUsers, transformAggregatedProjectUserToPB(v))
	}

	return &frontierv1beta1.SearchProjectUsersResponse{
		ProjectUsers: projectUsers,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(projectUsersData.Pagination.Offset),
			Limit:  uint32(projectUsersData.Pagination.Limit),
		},
		Group: nil,
	}, nil
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
