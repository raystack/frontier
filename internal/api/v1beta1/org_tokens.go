package v1beta1

import (
	"context"
	"errors"
	"fmt"

	svc "github.com/raystack/frontier/core/aggregates/orgtokens"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	"github.com/raystack/salt/rql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgTokensService interface {
	Search(ctx context.Context, id string, query *rql.Query) (svc.OrganizationTokens, error)
	Export(ctx context.Context, orgID string) ([]byte, string, error)
}

func (h Handler) SearchOrganizationTokens(ctx context.Context, request *frontierv1beta1.SearchOrganizationTokensRequest) (*frontierv1beta1.SearchOrganizationTokensResponse, error) {
	rqlQuery, err := utils.TransformProtoToRQL(request.GetQuery(), svc.AggregatedToken{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, svc.AggregatedToken{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to validate rql query: %v", err))
	}

	tokensData, err := h.orgTokensService.Search(ctx, request.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, err
	}

	var orgTokens []*frontierv1beta1.SearchOrganizationTokensResponse_OrganizationToken
	for _, v := range tokensData.Tokens {
		orgTokens = append(orgTokens, transformAggregatedTokenToPB(v))
	}

	return &frontierv1beta1.SearchOrganizationTokensResponse{
		OrganizationTokens: orgTokens,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(tokensData.Pagination.Offset),
			Limit:  uint32(tokensData.Pagination.Limit),
		},
		Group: nil,
	}, nil
}

func (h Handler) ExportOrganizationTokens(req *frontierv1beta1.ExportOrganizationTokensRequest, stream frontierv1beta1.AdminService_ExportOrganizationTokensServer) error {
	orgTokensDataBytes, contentType, err := h.orgTokensService.Export(stream.Context(), req.GetId())
	if err != nil {
		if errors.Is(err, svc.ErrNoContent) {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("no data to export: %v", err))
		}
		return err
	}

	chunkSize := 1024 * 200 // 200KB

	for i := 0; i < len(orgTokensDataBytes); i += chunkSize {
		end := min(i+chunkSize, len(orgTokensDataBytes))

		chunk := orgTokensDataBytes[i:end]
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
