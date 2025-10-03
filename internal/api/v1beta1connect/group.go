package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
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

func (h *ConnectHandler) ListOrganizationGroups(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationGroupsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationGroupsResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	var groups []*frontierv1beta1.Group
	groupList, err := h.groupService.List(ctx, group.Filter{
		OrganizationID:  orgResp.ID,
		State:           group.State(request.Msg.GetState()),
		GroupIDs:        request.Msg.GetGroupIds(),
		WithMemberCount: request.Msg.GetWithMemberCount(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range groupList {
		groupPB, err := transformGroupToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		if request.Msg.GetWithMembers() {
			groupUsers, err := h.userService.ListByGroup(ctx, v.ID, "")
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
			}
			var groupUsersErr error
			groupPB.Users = utils.Filter(utils.Map(groupUsers, func(user user.User) *frontierv1beta1.User {
				pb, err := transformUserToPB(user)
				if err != nil {
					groupUsersErr = errors.Join(groupUsersErr, err)
					return nil
				}
				return pb
			}), func(user *frontierv1beta1.User) bool {
				return user != nil
			})
			if groupUsersErr != nil {
				return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
			}
		}

		groups = append(groups, &groupPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationGroupsResponse{Groups: groups}), nil
}

func (h *ConnectHandler) CreateGroup(ctx context.Context, request *connect.Request[frontierv1beta1.CreateGroupRequest]) (*connect.Response[frontierv1beta1.CreateGroupResponse], error) {
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, groupMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	newGroup, err := h.groupService.Create(ctx, group.Group{
		Name:           request.Msg.GetBody().GetName(),
		Title:          request.Msg.GetBody().GetTitle(),
		OrganizationID: orgResp.ID,
		Metadata:       metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, group.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		case errors.Is(err, group.ErrInvalidDetail), errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	groupPB, err := transformGroupToPB(newGroup)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, request.Msg.GetOrgId()).Log(audit.GroupCreatedEvent, audit.GroupTarget(newGroup.ID))
	return connect.NewResponse(&frontierv1beta1.CreateGroupResponse{Group: &groupPB}), nil
}

func (h *ConnectHandler) GetGroup(ctx context.Context, request *connect.Request[frontierv1beta1.GetGroupRequest]) (*connect.Response[frontierv1beta1.GetGroupResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	fetchedGroup, err := h.groupService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist), errors.Is(err, group.ErrInvalidID), errors.Is(err, group.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeNotFound, ErrGroupNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	groupPB, err := transformGroupToPB(fetchedGroup)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	if request.Msg.GetWithMembers() {
		groupUsers, err := h.userService.ListByGroup(ctx, fetchedGroup.ID, "")
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		var groupUsersErr error
		groupPB.Users = utils.Filter(utils.Map(groupUsers, func(user user.User) *frontierv1beta1.User {
			pb, err := transformUserToPB(user)
			if err != nil {
				groupUsersErr = errors.Join(groupUsersErr, err)
				return nil
			}
			return pb
		}), func(user *frontierv1beta1.User) bool {
			return user != nil
		})
		if groupUsersErr != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.GetGroupResponse{Group: &groupPB}), nil
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
