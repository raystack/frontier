package v1beta1connect

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/str"
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
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListGroups.List: org_id=%s state=%s: %w", request.Msg.GetOrgId(), request.Msg.GetState(), err))
	}

	for _, v := range groupList {
		groupPB, err := transformGroupToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListGroups: entity_id=%s: %w", v.ID, err))
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
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationGroups.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
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
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationGroups.List: org_id=%s state=%s: %w", request.Msg.GetOrgId(), request.Msg.GetState(), err))
	}

	for _, v := range groupList {
		groupPB, err := transformGroupToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationGroups: entity_id=%s: %w", v.ID, err))
		}

		if request.Msg.GetWithMembers() {
			groupUsers, err := h.listGroupUsers(ctx, v.ID, "")
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationGroups.listGroupUsers: group_id=%s: %w", v.ID, err))
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
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationGroups.transformUserToPB: group_id=%s: %w", v.ID, groupUsersErr))
			}
		}

		groups = append(groups, &groupPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationGroupsResponse{Groups: groups}), nil
}

// listGroupUsers returns the users that are members of the given group.
// When roleID is non-empty, only users with that role are returned.
func (h *ConnectHandler) listGroupUsers(ctx context.Context, groupID, roleID string) ([]user.User, error) {
	filter := membership.MemberFilter{PrincipalType: schema.UserPrincipal}
	if roleID != "" {
		filter.RoleIDs = []string{roleID}
	}
	members, err := h.membershipService.ListPrincipalsByResource(ctx, groupID, schema.GroupNamespace, filter)
	if err != nil {
		return nil, err
	}
	userIDs := utils.Map(members, func(m membership.Member) string { return m.PrincipalID })
	return h.userService.GetByIDs(ctx, userIDs)
}

func (h *ConnectHandler) CreateGroup(ctx context.Context, request *connect.Request[frontierv1beta1.CreateGroupRequest]) (*connect.Response[frontierv1beta1.CreateGroupResponse], error) {
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateGroup.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
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
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateGroup.Create: org_id=%s group_name=%s: %w", request.Msg.GetOrgId(), name, err))
		}
	}

	groupPB, err := transformGroupToPB(newGroup)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateGroup: entity_id=%s: %w", newGroup.ID, err))
	}

	audit.GetAuditor(ctx, request.Msg.GetOrgId()).Log(audit.GroupCreatedEvent, audit.GroupTarget(newGroup.ID))
	return connect.NewResponse(&frontierv1beta1.CreateGroupResponse{Group: &groupPB}), nil
}

func (h *ConnectHandler) GetGroup(ctx context.Context, request *connect.Request[frontierv1beta1.GetGroupRequest]) (*connect.Response[frontierv1beta1.GetGroupResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetGroup.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
		}
	}

	fetchedGroup, err := h.groupService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist), errors.Is(err, group.ErrInvalidID), errors.Is(err, group.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeNotFound, ErrGroupNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetGroup.Get: group_id=%s: %w", request.Msg.GetId(), err))
		}
	}

	groupPB, err := transformGroupToPB(fetchedGroup)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetGroup: entity_id=%s: %w", fetchedGroup.ID, err))
	}

	if request.Msg.GetWithMembers() {
		groupUsers, err := h.listGroupUsers(ctx, fetchedGroup.ID, "")
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetGroup.listGroupUsers: group_id=%s: %w", fetchedGroup.ID, err))
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
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetGroup.transformUserToPB: group_id=%s: %w", fetchedGroup.ID, groupUsersErr))
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
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateGroup.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
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
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateGroup.Update: group_id=%s group_name=%s: %w", request.Msg.GetId(), request.Msg.GetBody().GetName(), err))
		}
	}

	groupPB, err := transformGroupToPB(updatedGroup)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateGroup: entity_id=%s: %w", updatedGroup.ID, err))
	}

	audit.GetAuditor(ctx, orgResp.ID).Log(audit.GroupUpdatedEvent, audit.GroupTarget(updatedGroup.ID))
	return connect.NewResponse(&frontierv1beta1.UpdateGroupResponse{Group: &groupPB}), nil
}

func (h *ConnectHandler) ListGroupUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListGroupUsersRequest]) (*connect.Response[frontierv1beta1.ListGroupUsersResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListGroupUsers.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
		}
	}

	members, err := h.membershipService.ListPrincipalsByResource(ctx, request.Msg.GetId(), schema.GroupNamespace, membership.MemberFilter{
		PrincipalType: schema.UserPrincipal,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListGroupUsers.ListPrincipalsByResource: group_id=%s: %w", request.Msg.GetId(), err))
	}

	userIDs := utils.Map(members, func(m membership.Member) string { return m.PrincipalID })
	users, err := h.userService.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListGroupUsers.GetByIDs: group_id=%s: %w", request.Msg.GetId(), err))
	}

	var userPBs []*frontierv1beta1.User
	var rolePairPBs []*frontierv1beta1.ListGroupUsersResponse_RolePair
	for _, user := range users {
		userPb, err := transformUserToPB(user)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListGroupUsers: entity_id=%s: %w", user.ID, err))
		}
		userPBs = append(userPBs, userPb)
	}

	for _, m := range members {
		rolesPb := utils.Filter(utils.Map(m.Roles, func(r role.Role) *frontierv1beta1.Role {
			pb, err := transformRoleToPB(r)
			if err != nil {
				return nil
			}
			return &pb
		}), func(r *frontierv1beta1.Role) bool {
			return r != nil
		})
		rolePairPBs = append(rolePairPBs, &frontierv1beta1.ListGroupUsersResponse_RolePair{
			UserId: m.PrincipalID,
			Roles:  rolesPb,
		})
	}

	return connect.NewResponse(&frontierv1beta1.ListGroupUsersResponse{
		Users:     userPBs,
		RolePairs: rolePairPBs,
	}), nil
}

func (h *ConnectHandler) RemoveGroupUser(ctx context.Context, request *connect.Request[frontierv1beta1.RemoveGroupUserRequest]) (*connect.Response[frontierv1beta1.RemoveGroupUserResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("RemoveGroupUser.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
		}
	}

	if err := h.membershipService.RemoveGroupMember(ctx, request.Msg.GetId(), request.Msg.GetUserId(), schema.UserPrincipal); err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist), errors.Is(err, group.ErrInvalidID), errors.Is(err, group.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeNotFound, ErrGroupNotFound)
		case errors.Is(err, user.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
		case errors.Is(err, user.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, user.ErrDisabled)
		case errors.Is(err, membership.ErrInvalidPrincipalType), errors.Is(err, membership.ErrInvalidPrincipal):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case errors.Is(err, membership.ErrNotMember):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrNotMember)
		case errors.Is(err, membership.ErrLastGroupOwnerRole):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrGroupMinOwnerCount)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("RemoveGroupUser.RemoveGroupMember: group_id=%s user_id=%s: %w", request.Msg.GetId(), request.Msg.GetUserId(), err))
		}
	}
	return connect.NewResponse(&frontierv1beta1.RemoveGroupUserResponse{}), nil
}

func (h *ConnectHandler) SetGroupMemberRole(ctx context.Context, request *connect.Request[frontierv1beta1.SetGroupMemberRoleRequest]) (*connect.Response[frontierv1beta1.SetGroupMemberRoleResponse], error) {
	orgID := request.Msg.GetOrgId()
	groupID := request.Msg.GetGroupId()
	principalID := request.Msg.GetPrincipalId()
	principalType := request.Msg.GetPrincipalType()
	roleID := request.Msg.GetRoleId()

	if _, err := h.orgService.Get(ctx, orgID); err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("SetGroupMemberRole.GetOrg: org_id=%s: %w", orgID, err))
		}
	}

	if err := h.membershipService.SetGroupMemberRole(ctx, groupID, principalID, principalType, roleID); err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist), errors.Is(err, group.ErrInvalidID), errors.Is(err, group.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeNotFound, ErrGroupNotFound)
		case errors.Is(err, user.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
		case errors.Is(err, user.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, role.ErrNotExist), errors.Is(err, role.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrInvalidRoleID)
		case errors.Is(err, membership.ErrInvalidPrincipalType), errors.Is(err, membership.ErrInvalidPrincipal):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case errors.Is(err, membership.ErrInvalidGroupRole):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidGroupRole)
		case errors.Is(err, membership.ErrNotOrgMember):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrNotOrgMember)
		case errors.Is(err, membership.ErrLastGroupOwnerRole):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrLastGroupOwnerRole)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("SetGroupMemberRole: group_id=%s principal_id=%s: %w", groupID, principalID, err))
		}
	}

	return connect.NewResponse(&frontierv1beta1.SetGroupMemberRoleResponse{}), nil
}

func (h *ConnectHandler) EnableGroup(ctx context.Context, request *connect.Request[frontierv1beta1.EnableGroupRequest]) (*connect.Response[frontierv1beta1.EnableGroupResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("EnableGroup.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
		}
	}
	if err := h.groupService.Enable(ctx, request.Msg.GetId()); err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrGroupNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("EnableGroup.Enable: group_id=%s: %w", request.Msg.GetId(), err))
		}
	}
	return connect.NewResponse(&frontierv1beta1.EnableGroupResponse{}), nil
}

func (h *ConnectHandler) DisableGroup(ctx context.Context, request *connect.Request[frontierv1beta1.DisableGroupRequest]) (*connect.Response[frontierv1beta1.DisableGroupResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DisableGroup.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
		}
	}
	if err := h.groupService.Disable(ctx, request.Msg.GetId()); err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrGroupNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DisableGroup.Disable: group_id=%s: %w", request.Msg.GetId(), err))
		}
	}
	return connect.NewResponse(&frontierv1beta1.DisableGroupResponse{}), nil
}

func (h *ConnectHandler) DeleteGroup(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteGroupRequest]) (*connect.Response[frontierv1beta1.DeleteGroupResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteGroup.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
		}
	}
	if err := h.deleterService.DeleteGroup(ctx, request.Msg.GetId()); err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrGroupNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteGroup.DeleteGroup: group_id=%s: %w", request.Msg.GetId(), err))
		}
	}
	return connect.NewResponse(&frontierv1beta1.DeleteGroupResponse{}), nil
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
