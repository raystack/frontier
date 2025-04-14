package v1beta1

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/aggregates/userorgs"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserOrgsService interface {
	Search(ctx context.Context, id string, query *rql.Query) (userorgs.UserOrgs, error)
}

func (h Handler) SearchUserOrganizations(ctx context.Context, request *frontierv1beta1.SearchUserOrganizationsRequest) (*frontierv1beta1.SearchUserOrganizationsResponse, error) {
	var userOrgs []*frontierv1beta1.SearchUserOrganizationsResponse_UserOrganization

	rqlQuery, err := utils.TransformProtoToRQL(request.GetQuery(), userorgs.AggregatedUserOrganization{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, userorgs.AggregatedUserOrganization{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to validate rql query: %v", err))
	}

	if len(rqlQuery.GroupBy) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "group by is not supported")
	}

	userOrgsData, err := h.userOrgsService.Search(ctx, request.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, err
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

	return &frontierv1beta1.SearchUserOrganizationsResponse{
		UserOrganizations: userOrgs,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(userOrgsData.Pagination.Offset),
			Limit:  uint32(userOrgsData.Pagination.Limit),
		},
		Group: &frontierv1beta1.RQLQueryGroupResponse{
			Name: userOrgsData.Group.Name,
			Data: groupResponse,
		},
	}, nil
}

// Helper function to transform aggregated organization to protobuf
func transformAggregatedUserOrganizationToPB(userOrg userorgs.AggregatedUserOrganization) *frontierv1beta1.SearchUserOrganizationsResponse_UserOrganization {
	return &frontierv1beta1.SearchUserOrganizationsResponse_UserOrganization{
		OrgId:    userOrg.OrgID,
		OrgTitle: userOrg.OrgTitle,
		OrgName:  userOrg.OrgName,
		// OrgAvatar:    userOrg.OrgAvatar,
		ProjectCount: userOrg.ProjectCount,
		RoleNames:    userOrg.RoleNames,
		RoleTitles:   userOrg.RoleTitles,
		RoleIds:      userOrg.RoleIDs,
		OrgJoinedOn:  timestamppb.New(userOrg.OrgJoinedOn),
		UserId:       userOrg.UserID,
	}
}
