package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/project"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) ListProjects(ctx context.Context, request *connect.Request[frontierv1beta1.ListProjectsRequest]) (*connect.Response[frontierv1beta1.ListProjectsResponse], error) {
	var projects []*frontierv1beta1.Project
	projectList, err := h.projectService.List(ctx, project.Filter{
		State: project.State(request.Msg.GetState()),
		OrgID: request.Msg.GetOrgId(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range projectList {
		projectPB, err := transformProjectToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		projects = append(projects, projectPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListProjectsResponse{Projects: projects}), nil
}

func transformProjectToPB(prj project.Project) (*frontierv1beta1.Project, error) {
	metaData, err := prj.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.Project{
		Id:           prj.ID,
		Name:         prj.Name,
		Title:        prj.Title,
		OrgId:        prj.Organization.ID,
		Metadata:     metaData,
		CreatedAt:    timestamppb.New(prj.CreatedAt),
		UpdatedAt:    timestamppb.New(prj.UpdatedAt),
		MembersCount: int32(prj.MemberCount),
	}, nil
}
