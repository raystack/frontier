package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/aggregates/orgprojects"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgProjectsService interface {
	Search(ctx context.Context, id string, query *rql.Query) (orgprojects.OrgProjects, error)
	Export(ctx context.Context, orgID string) ([]byte, string, error)
}

func (h *ConnectHandler) SearchOrganizationProjects(ctx context.Context, request *connect.Request[frontierv1beta1.SearchOrganizationProjectsRequest]) (*connect.Response[frontierv1beta1.SearchOrganizationProjectsResponse], error) {
	var orgProjects []*frontierv1beta1.SearchOrganizationProjectsResponse_OrganizationProject

	rqlQuery, err := utils.TransformProtoToRQL(request.Msg.GetQuery(), orgprojects.AggregatedProject{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, orgprojects.AggregatedProject{})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to validate rql query: %v", err))
	}

	orgProjectsData, err := h.orgProjectsService.Search(ctx, request.Msg.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInternalServerError)
		}
		return nil, err
	}

	for _, v := range orgProjectsData.Projects {
		orgProjects = append(orgProjects, transformAggregatedProjectToPB(v))
	}

	groupResponse := make([]*frontierv1beta1.RQLQueryGroupData, 0)
	for _, groupItem := range orgProjectsData.Group.Data {
		groupResponse = append(groupResponse, &frontierv1beta1.RQLQueryGroupData{
			Name:  groupItem.Name,
			Count: uint32(groupItem.Count),
		})
	}

	return connect.NewResponse(&frontierv1beta1.SearchOrganizationProjectsResponse{
		OrgProjects: orgProjects,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(orgProjectsData.Pagination.Offset),
			Limit:  uint32(orgProjectsData.Pagination.Limit),
		},
		Group: &frontierv1beta1.RQLQueryGroupResponse{
			Name: orgProjectsData.Group.Name,
			Data: groupResponse,
		},
	}), nil
}

// Helper function to transform domain model to protobuf
func transformAggregatedProjectToPB(p orgprojects.AggregatedProject) *frontierv1beta1.SearchOrganizationProjectsResponse_OrganizationProject {
	return &frontierv1beta1.SearchOrganizationProjectsResponse_OrganizationProject{
		Id:             p.ID,
		Name:           p.Name,
		Title:          p.Title,
		State:          string(p.State),
		MemberCount:    p.MemberCount,
		UserIds:        p.UserIDs,
		CreatedAt:      timestamppb.New(p.CreatedAt),
		OrganizationId: p.OrganizationID,
	}
}
