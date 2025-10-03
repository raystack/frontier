package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/str"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"go.uber.org/zap"
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

	// Auto-generate name from title if name is empty but title is provided
	requestBody := request.Msg.GetBody()
	name := requestBody.GetName()
	if name == "" && requestBody.GetTitle() != "" {
		name = str.GenerateSlug(requestBody.GetTitle())
	}

	newGroup, err := h.groupService.Create(ctx, group.Group{
		Name:           name,
		Title:          requestBody.GetTitle(),
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

func (h *ConnectHandler) UpdateGroup(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateGroupRequest]) (*connect.Response[frontierv1beta1.UpdateGroupResponse], error) {
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

	updatedGroup, err := h.groupService.Update(ctx, group.Group{
		ID:             request.Msg.GetId(),
		Name:           request.Msg.GetBody().GetName(),
		Title:          request.Msg.GetBody().GetTitle(),
		OrganizationID: orgResp.ID,
		Metadata:       metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist), errors.Is(err, group.ErrInvalidUUID), errors.Is(err, group.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrGroupNotFound)
		case errors.Is(err, group.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		case errors.Is(err, group.ErrInvalidDetail), errors.Is(err, organization.ErrInvalidUUID), errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	groupPB, err := transformGroupToPB(updatedGroup)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, orgResp.ID).Log(audit.GroupUpdatedEvent, audit.GroupTarget(updatedGroup.ID))
	return connect.NewResponse(&frontierv1beta1.UpdateGroupResponse{Group: &groupPB}), nil
}

func (h *ConnectHandler) ListGroupUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListGroupUsersRequest]) (*connect.Response[frontierv1beta1.ListGroupUsersResponse], error) {
	logger := grpczap.Extract(ctx)
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

	var userPBs []*frontierv1beta1.User
	var rolePairPBs []*frontierv1beta1.ListGroupUsersResponse_RolePair
	users, err := h.userService.ListByGroup(ctx, request.Msg.GetId(), "")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, user := range users {
		userPb, err := transformUserToPB(user)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		userPBs = append(userPBs, userPb)
	}

	if request.Msg.GetWithRoles() {
		for _, user := range users {
			roles, err := h.policyService.ListRoles(ctx, schema.UserPrincipal, user.ID, schema.GroupNamespace, request.Msg.GetId())
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
			}

			rolesPb := utils.Filter(utils.Map(roles, func(role role.Role) *frontierv1beta1.Role {
				pb, err := transformRoleToPB(role)
				if err != nil {
					logger.Error("failed to transform role for group", zap.Error(err))
					return nil
				}
				return &pb
			}), func(role *frontierv1beta1.Role) bool {
				return role != nil
			})
			rolePairPBs = append(rolePairPBs, &frontierv1beta1.ListGroupUsersResponse_RolePair{
				UserId: user.ID,
				Roles:  rolesPb,
			})
		}
	}

	return connect.NewResponse(&frontierv1beta1.ListGroupUsersResponse{
		Users:     userPBs,
		RolePairs: rolePairPBs,
	}), nil
}

func (h *ConnectHandler) AddGroupUsers(ctx context.Context, request *connect.Request[frontierv1beta1.AddGroupUsersRequest]) (*connect.Response[frontierv1beta1.AddGroupUsersResponse], error) {
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

	if err := h.groupService.AddUsers(ctx, request.Msg.GetId(), request.Msg.GetUserIds()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.AddGroupUsersResponse{}), nil
}

func (h *ConnectHandler) RemoveGroupUser(ctx context.Context, request *connect.Request[frontierv1beta1.RemoveGroupUserRequest]) (*connect.Response[frontierv1beta1.RemoveGroupUserResponse], error) {
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

	// before deleting the user, check if the user is the only owner of the group
	owners, err := h.userService.ListByGroup(ctx, request.Msg.GetId(), group.AdminRole)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	if len(owners) == 1 && owners[0].ID == request.Msg.GetUserId() {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrGroupMinOwnerCount)
	}

	// delete the user
	if err := h.groupService.RemoveUsers(ctx, request.Msg.GetId(), []string{request.Msg.GetUserId()}); err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.RemoveGroupUserResponse{}), nil
}

func (h *ConnectHandler) EnableGroup(ctx context.Context, request *connect.Request[frontierv1beta1.EnableGroupRequest]) (*connect.Response[frontierv1beta1.EnableGroupResponse], error) {
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
	if err := h.groupService.Enable(ctx, request.Msg.GetId()); err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrGroupNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}
	return connect.NewResponse(&frontierv1beta1.EnableGroupResponse{}), nil
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
