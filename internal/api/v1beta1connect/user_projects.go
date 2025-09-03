package v1beta1connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/pkg/errors"
	"github.com/raystack/frontier/core/aggregates/userprojects"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserProjectsService interface {
	Search(ctx context.Context, userId string, orgId string, query *rql.Query) (userprojects.UserProjects, error)
}

func (h *ConnectHandler) SearchUserProjects(ctx context.Context, request *connect.Request[frontierv1beta1.SearchUserProjectsRequest]) (*connect.Response[frontierv1beta1.SearchUserProjectsResponse], error) {
	var userProjects []*frontierv1beta1.SearchUserProjectsResponse_UserProject

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), userprojects.AggregatedProject{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, userprojects.AggregatedProject{})
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

	userProjectsData, err := h.userProjectsService.Search(ctx, request.Msg.GetUserId(), request.Msg.GetOrgId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range userProjectsData.Projects {
		userProjects = append(userProjects, transformAggregatedUserProjectToPB(v))
	}

	return connect.NewResponse(&frontierv1beta1.SearchUserProjectsResponse{
		UserProjects: userProjects,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(userProjectsData.Pagination.Offset),
			Limit:  uint32(userProjectsData.Pagination.Limit),
		},
		Group: nil,
	}), nil
}

func transformAggregatedUserProjectToPB(v userprojects.AggregatedProject) *frontierv1beta1.SearchUserProjectsResponse_UserProject {
	return &frontierv1beta1.SearchUserProjectsResponse_UserProject{
		ProjectId:        v.ProjectID,
		ProjectTitle:     v.ProjectTitle,
		ProjectName:      v.ProjectName,
		ProjectCreatedOn: timestamppb.New(v.CreatedOn),
		UserNames:        v.UserNames,
		UserTitles:       v.UserTitles,
		UserIds:          v.UserIDs,
		UserAvatars:      v.UserAvatars,
		OrgId:            v.OrgID,
		UserId:           v.UserID,
	}
}
