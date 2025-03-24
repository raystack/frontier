package v1beta1

import (
	"context"
	"errors"
	"fmt"

	"github.com/raystack/frontier/core/aggregates/orgprojects"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/rql"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrgProjectsService interface {
	Search(ctx context.Context, id string, query *rql.Query) (orgprojects.OrgProjects, error)
	Export(ctx context.Context, orgID string) ([]byte, string, error)
}

func (h Handler) SearchOrganizationProjects(ctx context.Context, request *frontierv1beta1.SearchOrganizationProjectsRequest) (*frontierv1beta1.SearchOrganizationProjectsResponse, error) {
	var orgProjects []*frontierv1beta1.SearchOrganizationProjectsResponse_OrganizationProject

	rqlQuery, err := utils.TransformProtoToRQL(request.GetQuery(), orgprojects.AggregatedProject{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to read rql query: %v", err))
	}

	err = rql.ValidateQuery(rqlQuery, orgprojects.AggregatedProject{})
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("failed to validate rql query: %v", err))
	}

	orgProjectsData, err := h.orgProjectsService.Search(ctx, request.GetId(), rqlQuery)
	if err != nil {
		if errors.Is(err, postgres.ErrBadInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
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

	return &frontierv1beta1.SearchOrganizationProjectsResponse{
		OrgProjects: orgProjects,
		Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
			Offset: uint32(orgProjectsData.Pagination.Offset),
			Limit:  uint32(orgProjectsData.Pagination.Limit),
		},
		Group: &frontierv1beta1.RQLQueryGroupResponse{
			Name: orgProjectsData.Group.Name,
			Data: groupResponse,
		},
	}, nil
}

func (h Handler) ExportOrganizationProjects(req *frontierv1beta1.ExportOrganizationProjectsRequest, stream frontierv1beta1.AdminService_ExportOrganizationProjectsServer) error {
	orgProjectsDataBytes, contentType, err := h.orgProjectsService.Export(stream.Context(), req.GetId())
	if err != nil {
		if errors.Is(err, orgprojects.ErrNoContent) {
			return status.Errorf(codes.InvalidArgument, fmt.Sprintf("no data to export: %v", err))
		}
		return err
	}

	chunkSize := 1024 * 200 // 200KB chunks

	for i := 0; i < len(orgProjectsDataBytes); i += chunkSize {
		end := min(i+chunkSize, len(orgProjectsDataBytes))

		chunk := orgProjectsDataBytes[i:end]
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
