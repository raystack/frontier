package v1beta1

import (
	"context"

	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/pkg/utils"
	"go.uber.org/zap"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/pkg/str"

	"errors"

	"github.com/raystack/frontier/pkg/metadata"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GroupService interface {
	Create(ctx context.Context, grp group.Group) (group.Group, error)
	Get(ctx context.Context, id string) (group.Group, error)
	List(ctx context.Context, flt group.Filter) ([]group.Group, error)
	Update(ctx context.Context, grp group.Group) (group.Group, error)
	ListByUser(ctx context.Context, principalId, principalType string, flt group.Filter) ([]group.Group, error)
	AddUsers(ctx context.Context, groupID string, userID []string) error
	RemoveUsers(ctx context.Context, groupID string, userID []string) error
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

var (
	grpcGroupNotFoundErr = status.Errorf(codes.NotFound, "group doesn't exist")
	grpcMinOwnerCounrErr = status.Errorf(codes.InvalidArgument, "group must have at least one owner, consider adding another owner before removing")
)

func (h Handler) ListGroups(ctx context.Context, request *frontierv1beta1.ListGroupsRequest) (*frontierv1beta1.ListGroupsResponse, error) {
	var groups []*frontierv1beta1.Group
	groupList, err := h.groupService.List(ctx, group.Filter{
		SU:             true,
		OrganizationID: request.GetOrgId(),
		State:          group.State(request.GetState()),
	})
	if err != nil {
		return nil, err
	}

	for _, v := range groupList {
		groupPB, err := transformGroupToPB(v)
		if err != nil {
			return nil, err
		}

		groups = append(groups, &groupPB)
	}

	return &frontierv1beta1.ListGroupsResponse{Groups: groups}, nil
}

func (h Handler) ListOrganizationGroups(ctx context.Context, request *frontierv1beta1.ListOrganizationGroupsRequest) (*frontierv1beta1.ListOrganizationGroupsResponse, error) {
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	var groups []*frontierv1beta1.Group
	groupList, err := h.groupService.List(ctx, group.Filter{
		OrganizationID:  orgResp.ID,
		State:           group.State(request.GetState()),
		GroupIDs:        request.GetGroupIds(),
		WithMemberCount: request.GetWithMemberCount(),
	})
	if err != nil {
		return nil, err
	}

	for _, v := range groupList {
		groupPB, err := transformGroupToPB(v)
		if err != nil {
			return nil, err
		}

		if request.GetWithMembers() {
			groupUsers, err := h.userService.ListByGroup(ctx, v.ID, "")
			if err != nil {
				return nil, err
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
				return nil, groupUsersErr
			}
		}

		groups = append(groups, &groupPB)
	}

	return &frontierv1beta1.ListOrganizationGroupsResponse{Groups: groups}, nil
}

func (h Handler) CreateGroup(ctx context.Context, request *frontierv1beta1.CreateGroupRequest) (*frontierv1beta1.CreateGroupResponse, error) {
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, groupMetaSchema); err != nil {
		return nil, grpcBadBodyMetaSchemaError
	}

	if request.GetBody().GetName() == "" && request.GetBody().GetTitle() != "" {
		request.GetBody().Name = str.GenerateSlug(request.GetBody().GetTitle())
	}

	newGroup, err := h.groupService.Create(ctx, group.Group{
		Name:           request.GetBody().GetName(),
		Title:          request.GetBody().GetTitle(),
		OrganizationID: orgResp.ID,
		Metadata:       metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, group.ErrConflict):
			return nil, grpcConflictError
		case errors.Is(err, group.ErrInvalidDetail), errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidUUID):
			return nil, grpcBadBodyError
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		default:
			return nil, err
		}
	}

	metaData, err := newGroup.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	audit.GetAuditor(ctx, request.GetOrgId()).Log(audit.GroupCreatedEvent, audit.GroupTarget(newGroup.ID))
	return &frontierv1beta1.CreateGroupResponse{Group: &frontierv1beta1.Group{
		Id:        newGroup.ID,
		Name:      newGroup.Name,
		OrgId:     newGroup.OrganizationID,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newGroup.CreatedAt),
		UpdatedAt: timestamppb.New(newGroup.UpdatedAt),
	}}, nil
}

func (h Handler) GetGroup(ctx context.Context, request *frontierv1beta1.GetGroupRequest) (*frontierv1beta1.GetGroupResponse, error) {
	_, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	fetchedGroup, err := h.groupService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist), errors.Is(err, group.ErrInvalidID), errors.Is(err, group.ErrInvalidUUID):
			return nil, grpcGroupNotFoundErr
		default:
			return nil, err
		}
	}

	groupPB, err := transformGroupToPB(fetchedGroup)
	if err != nil {
		return nil, err
	}
	if request.GetWithMembers() {
		groupUsers, err := h.userService.ListByGroup(ctx, fetchedGroup.ID, "")
		if err != nil {
			return nil, err
		}
		var groupUsersErr error
		groupPB.Users = utils.Map(groupUsers, func(user user.User) *frontierv1beta1.User {
			pb, err := transformUserToPB(user)
			if err != nil {
				groupUsersErr = errors.Join(groupUsersErr, err)
				return nil
			}
			return pb
		})
		if groupUsersErr != nil {
			return nil, groupUsersErr
		}
	}

	return &frontierv1beta1.GetGroupResponse{Group: &groupPB}, nil
}

func (h Handler) UpdateGroup(ctx context.Context, request *frontierv1beta1.UpdateGroupRequest) (*frontierv1beta1.UpdateGroupResponse, error) {
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, groupMetaSchema); err != nil {
		return nil, grpcBadBodyMetaSchemaError
	}

	updatedGroup, err := h.groupService.Update(ctx, group.Group{
		ID:             request.GetId(),
		Name:           request.GetBody().GetName(),
		Title:          request.GetBody().GetTitle(),
		OrganizationID: orgResp.ID,
		Metadata:       metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist),
			errors.Is(err, group.ErrInvalidUUID),
			errors.Is(err, group.ErrInvalidID):
			return nil, grpcGroupNotFoundErr
		case errors.Is(err, group.ErrConflict):
			return nil, grpcConflictError
		case errors.Is(err, group.ErrInvalidDetail),
			errors.Is(err, organization.ErrInvalidUUID),
			errors.Is(err, organization.ErrNotExist):
			return nil, grpcBadBodyError
		default:
			return nil, err
		}
	}

	groupPB, err := transformGroupToPB(updatedGroup)
	if err != nil {
		return nil, err
	}

	audit.GetAuditor(ctx, orgResp.ID).Log(audit.GroupUpdatedEvent, audit.GroupTarget(updatedGroup.ID))
	return &frontierv1beta1.UpdateGroupResponse{Group: &groupPB}, nil
}

func (h Handler) ListGroupUsers(ctx context.Context, request *frontierv1beta1.ListGroupUsersRequest) (*frontierv1beta1.ListGroupUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	_, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	var userPBs []*frontierv1beta1.User
	var rolePairPBs []*frontierv1beta1.ListGroupUsersResponse_RolePair
	users, err := h.userService.ListByGroup(ctx, request.GetId(), "")
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		userPb, err := transformUserToPB(user)
		if err != nil {
			return nil, err
		}
		userPBs = append(userPBs, userPb)
	}

	if request.GetWithRoles() {
		for _, user := range users {
			roles, err := h.policyService.ListRoles(ctx, schema.UserPrincipal, user.ID, schema.GroupNamespace, request.GetId())
			if err != nil {
				return nil, err
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

	return &frontierv1beta1.ListGroupUsersResponse{
		Users:     userPBs,
		RolePairs: rolePairPBs,
	}, nil
}

func (h Handler) AddGroupUsers(ctx context.Context, request *frontierv1beta1.AddGroupUsersRequest) (*frontierv1beta1.AddGroupUsersResponse, error) {
	_, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	if err := h.groupService.AddUsers(ctx, request.GetId(), request.GetUserIds()); err != nil {
		return nil, err
	}
	return &frontierv1beta1.AddGroupUsersResponse{}, nil
}

func (h Handler) RemoveGroupUser(ctx context.Context, request *frontierv1beta1.RemoveGroupUserRequest) (*frontierv1beta1.RemoveGroupUserResponse, error) {
	_, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	// before deleting the user, check if the user is the only owner of the group
	owners, err := h.userService.ListByGroup(ctx, request.GetId(), group.AdminRole)
	if err != nil {
		return nil, err
	}
	if len(owners) == 1 && owners[0].ID == request.GetUserId() {
		return nil, grpcMinOwnerCounrErr
	}

	// delete the user
	if err := h.groupService.RemoveUsers(ctx, request.GetId(), []string{request.GetUserId()}); err != nil {
		return nil, err
	}
	return &frontierv1beta1.RemoveGroupUserResponse{}, nil
}

func (h Handler) EnableGroup(ctx context.Context, request *frontierv1beta1.EnableGroupRequest) (*frontierv1beta1.EnableGroupResponse, error) {
	_, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	if err := h.groupService.Enable(ctx, request.GetId()); err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		default:
			return nil, err
		}
	}
	return &frontierv1beta1.EnableGroupResponse{}, nil
}

func (h Handler) DisableGroup(ctx context.Context, request *frontierv1beta1.DisableGroupRequest) (*frontierv1beta1.DisableGroupResponse, error) {
	_, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	if err := h.groupService.Disable(ctx, request.GetId()); err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		default:
			return nil, err
		}
	}
	return &frontierv1beta1.DisableGroupResponse{}, nil
}

func (h Handler) DeleteGroup(ctx context.Context, request *frontierv1beta1.DeleteGroupRequest) (*frontierv1beta1.DeleteGroupResponse, error) {
	_, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, grpcOrgDisabledErr
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, err
		}
	}

	if err := h.groupService.Delete(ctx, request.GetId()); err != nil {
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		default:
			return nil, err
		}
	}
	return &frontierv1beta1.DeleteGroupResponse{}, nil
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
