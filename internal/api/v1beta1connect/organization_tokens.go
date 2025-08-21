package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	svc "github.com/raystack/frontier/core/aggregates/orgtokens"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgTokensService interface {
	Search(ctx context.Context, id string, query *rql.Query) (svc.OrganizationTokens, error)
	Export(ctx context.Context, orgID string) ([]byte, string, error)
}

func (h *ConnectHandler) SearchOrganizationTokens(ctx context.Context, request *connect.Request[frontierv1beta1.SearchOrganizationTokensRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationTokensResponse], error) {
	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), svc.AggregatedToken{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, svc.AggregatedToken{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	tokensData, err := h.orgTokensService.Search(ctx, request.Msg.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var orgTokens []*frontierv1beta1.SearchOrganizationTokensResponse_OrganizationToken
	for _, v := range tokensData.Tokens {
		orgTokens = append(orgTokens, transformAggregatedTokenToPB(v))
	}

	return connect.NewResponse(&frontierv1beta1.SearchOrganizationTokensResponse{
		OrganizationTokens: orgTokens,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(tokensData.Pagination.Offset),
			Limit:  uint32(tokensData.Pagination.Limit),
		},
		Group: nil,
	}), nil
}

func transformAggregatedTokenToPB(v svc.AggregatedToken) *frontierv1beta1.SearchOrganizationTokensResponse_OrganizationToken {
	return &frontierv1beta1.SearchOrganizationTokensResponse_OrganizationToken{
		Amount:      v.Amount,
		Type:        v.Type,
		Description: v.Description,
		UserId:      v.UserID,
		UserTitle:   v.UserTitle,
		UserAvatar:  v.UserAvatar,
		CreatedAt:   timestamppb.New(v.CreatedAt),
		OrgId:       v.OrgID,
	}
}

func (h *ConnectHandler) ExportOrganizationTokens(ctx context.Context, request *connect.Request[frontierv1beta1.ExportOrganizationTokensRequest], stream *connect.ServerStream[httpbody.HttpBody]) error {
	orgTokensDataBytes, contentType, err := h.orgTokensService.Export(ctx, request.Msg.GetId())
	if err != nil {
		if errors.Is(err, svc.ErrNoContent) {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("no data to export: %v", err))
		}
		return connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return streamBytesInChunks(orgTokensDataBytes, contentType, stream)
}
