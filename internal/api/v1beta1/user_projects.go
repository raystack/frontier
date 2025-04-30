package v1beta1

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/aggregates/userprojects"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserProjectsService interface {
	Search(ctx context.Context, userId string, orgId string, query *rql.Query) (userprojects.UserProjects, error)
}

func (h Handler) SearchUserProjects(ctx context.Context, request *frontierv1beta1.SearchUserProjectsRequest) (*frontierv1beta1.SearchUserProjectsResponse, error) {
	var userProjects []*frontierv1beta1.SearchUserProjectsResponse_UserProject

	rqlQuery, err := utils.TransformProtoToRQL(request.GetQuery(), userprojects.AggregatedProject{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, userprojects.AggregatedProject{})
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

	userProjectsData, err := h.userProjectsService.Search(ctx, request.GetUserId(), request.GetOrgId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, err
	}

	for _, v := range userProjectsData.Projects {
		userProjects = append(userProjects, transformAggregatedUserProjectToPB(v))
	}

	return &frontierv1beta1.SearchUserProjectsResponse{
		UserProjects: userProjects,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(userProjectsData.Pagination.Offset),
			Limit:  uint32(userProjectsData.Pagination.Limit),
		},
		Group: nil,
	}, nil
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
