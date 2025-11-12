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
	"go.uber.org/zap"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) SearchOrganizationUsers(ctx context.Context, request *connect.Request[frontierv1beta1.SearchOrganizationUsersRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationUsersResponse], error) {
	errorLogger := NewErrorLogger()

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
		errorLogger.LogServiceError(ctx, request, "SearchOrganizationUsers.Search", err,
			zap.String("org_id", request.Msg.GetId()))
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

func (h *ConnectHandler) ExportOrganizationUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ExportOrganizationUsersRequest], stream *connect.ServerStream[httpbody.HttpBody]) error {
	errorLogger := NewErrorLogger()

	orgUsersDataBytes, contentType, err := h.orgUsersService.Export(ctx, request.Msg.GetId())
	if err != nil {
		if errors.Is(err, orgusers.ErrNoContent) {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("no data to export: %v", err))
		}
		errorLogger.LogServiceError(ctx, request, "ExportOrganizationUsers.Export", err,
			zap.String("org_id", request.Msg.GetId()))
		return connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return streamBytesInChunks(orgUsersDataBytes, contentType, stream)
}

func streamBytesInChunks(data []byte, contentType string, stream *connect.ServerStream[httpbody.HttpBody]) error {
	chunkSize := 1024 * 200 // 200KB
	for i := 0; i < len(data); i += chunkSize {
		end := min(i+chunkSize, len(data))

		chunk := data[i:end]
		msg := &httpbody.HttpBody{
			ContentType: contentType,
			Data:        chunk,
		}

		if err := stream.Send(msg); err != nil {
			return connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	return nil
}
