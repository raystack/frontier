package v1beta1

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/aggregates/orgusers"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgUsersService interface {
	Search(ctx context.Context, id string, query *rql.Query) (orgusers.OrgUsers, error)
	Export(ctx context.Context, orgID string) ([]byte, string, error)
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

	if len(rqlQuery.GroupBy) > 0 {
		return nil, status.Errorf(codes.InvalidArgument, "group by is not supported")
	}

	orgUsersData, err := h.orgUsersService.Search(ctx, request.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
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

func (h Handler) ExportOrganizationUsers(req *frontierv1beta1.ExportOrganizationUsersRequest, stream frontierv1beta1.AdminService_ExportOrganizationUsersServer) error {
	orgUsersDataBytes, contentType, err := h.orgUsersService.Export(stream.Context(), req.GetId())
	if err != nil {
		if errors.Is(err, orgusers.ErrNoContent) {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("no data to export: %v", err))
		}
		return err
	}

	chunkSize := 1024 * 200 // 200KB

	for i := 0; i < len(orgUsersDataBytes); i += chunkSize {
		end := min(i+chunkSize, len(orgUsersDataBytes))

		chunk := orgUsersDataBytes[i:end]
		msg := &httpbody.HttpBody{
			ContentType: contentType,
			Data:        chunk,
		}

		if err := stream.Send(msg); err != nil {
			return fmt.Errorf("failed to send chunk: %v", err)
		}
	}
	return nil
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
