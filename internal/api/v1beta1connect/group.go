package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/group"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) ListGroups(ctx context.Context, request *connect.Request[frontierv1beta1.ListGroupsRequest]) (*connect.Response[frontierv1beta1.ListGroupsResponse], error) {
	var groups []*frontierv1beta1.Group
	groupList, err := h.groupService.List(ctx, group.Filter{
		SU:             true,
		OrganizationID: request.Msg.GetOrgId(),
		State:          group.State(request.Msg.GetState()),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range groupList {
		groupPB, err := transformGroupToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		groups = append(groups, &groupPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListGroupsResponse{Groups: groups}), nil
}

func transformGroupToPB(grp group.Group) (frontierv1beta1.Group, error) {
	metaData, err := grp.Metadata.ToStructPB()
	if err != nil {
		return frontierv1beta1.Group{}, err
	}

	return frontierv1beta1.Group{
		Id:           grp.ID,
		Name:         grp.Name,
		Title:        grp.Title,
		OrgId:        grp.OrganizationID,
		Metadata:     metaData,
		CreatedAt:    timestamppb.New(grp.CreatedAt),
		UpdatedAt:    timestamppb.New(grp.UpdatedAt),
		MembersCount: int32(grp.MemberCount),
	}, nil
}
