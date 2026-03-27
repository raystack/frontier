package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	svc "github.com/raystack/frontier/core/aggregates/orgpats"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const orgPATsDefaultLimit = 30
const orgPATsMaxLimit = 30

func (h *ConnectHandler) SearchOrganizationPATs(ctx context.Context, request *connect.Request[frontierv1beta1.SearchOrganizationPATsRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationPATsResponse], error) {
	errorLogger := NewErrorLogger()

	if request.Msg.Validate() != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), svc.PATSearchFields{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	if err = rql.ValidateQuery(rqlQuery, svc.PATSearchFields{}); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	// Cap limit — override user-requested limit if it exceeds max
	if rqlQuery.Limit <= 0 || rqlQuery.Limit > orgPATsMaxLimit {
		grpczap.Extract(ctx).Warn("overriding requested limit to max allowed",
			zap.Int("requested_limit", rqlQuery.Limit),
			zap.Int("applied_limit", orgPATsDefaultLimit))
		rqlQuery.Limit = orgPATsDefaultLimit
	}

	result, err := h.orgPATsService.Search(ctx, request.Msg.GetOrgId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		errorLogger.LogServiceError(ctx, request, "SearchOrganizationPATs.Search", err,
			zap.String("org_id", request.Msg.GetOrgId()))
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	orgPATs := make([]*frontierv1beta1.SearchOrganizationPATsResponse_OrganizationPAT, 0, len(result.PATs))
	for _, pat := range result.PATs {
		orgPATs = append(orgPATs, transformAggregatedPATToPB(pat))
	}

	return connect.NewResponse(&frontierv1beta1.SearchOrganizationPATsResponse{
		OrganizationPats: orgPATs,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset:     uint32(result.Pagination.Offset),
			Limit:      uint32(result.Pagination.Limit),
			TotalCount: uint32(result.Pagination.TotalCount),
		},
	}), nil
}

func transformAggregatedPATToPB(pat svc.AggregatedPAT) *frontierv1beta1.SearchOrganizationPATsResponse_OrganizationPAT {
	pbPAT := &frontierv1beta1.SearchOrganizationPATsResponse_OrganizationPAT{
		Id:    pat.ID,
		Title: pat.Title,
		CreatedBy: &frontierv1beta1.SearchOrganizationPATsResponse_CreatedBy{
			Id:    pat.CreatedBy.ID,
			Title: pat.CreatedBy.Title,
			Email: pat.CreatedBy.Email,
		},
		Scopes:    transformScopesToPB(pat.Scopes),
		CreatedAt: timestamppb.New(pat.CreatedAt),
		ExpiresAt: timestamppb.New(pat.ExpiresAt),
	}
	if pat.LastUsedAt != nil {
		pbPAT.LastUsedAt = timestamppb.New(*pat.LastUsedAt)
	}
	return pbPAT
}

func transformScopesToPB(scopes []patmodels.PATScope) []*frontierv1beta1.PATScope {
	pbScopes := make([]*frontierv1beta1.PATScope, 0, len(scopes))
	for _, s := range scopes {
		pbScopes = append(pbScopes, &frontierv1beta1.PATScope{
			RoleId:       s.RoleID,
			ResourceType: s.ResourceType,
			ResourceIds:  s.ResourceIDs,
		})
	}
	return pbScopes
}
