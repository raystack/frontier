package v1beta1

import (
	"context"

	"github.com/raystack/frontier/core/audit"

	"github.com/raystack/frontier/pkg/str"

	"github.com/pkg/errors"
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
	ListByUser(ctx context.Context, userId string, flt group.Filter) ([]group.Group, error)
	AddUsers(ctx context.Context, groupID string, userID []string) error
	RemoveUsers(ctx context.Context, groupID string, userID []string) error
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

var (
	grpcGroupNotFoundErr = status.Errorf(codes.NotFound, "group doesn't exist")
)

func (h Handler) ListGroups(ctx context.Context, request *frontierv1beta1.ListGroupsRequest) (*frontierv1beta1.ListGroupsResponse, error) {
	logger := grpczap.Extract(ctx)

	var groups []*frontierv1beta1.Group
	groupList, err := h.groupService.List(ctx, group.Filter{
		OrganizationID: request.GetOrgId(),
		State:          group.State(request.GetState()),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, v := range groupList {
		groupPB, err := transformGroupToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		groups = append(groups, &groupPB)
	}

	return &frontierv1beta1.ListGroupsResponse{Groups: groups}, nil
}

func (h Handler) ListOrganizationGroups(ctx context.Context, request *frontierv1beta1.ListOrganizationGroupsRequest) (*frontierv1beta1.ListOrganizationGroupsResponse, error) {
	logger := grpczap.Extract(ctx)

	var groups []*frontierv1beta1.Group
	groupList, err := h.groupService.List(ctx, group.Filter{
		OrganizationID: request.GetOrgId(),
		State:          group.State(request.GetState()),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, v := range groupList {
		groupPB, err := transformGroupToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		groups = append(groups, &groupPB)
	}

	return &frontierv1beta1.ListOrganizationGroupsResponse{Groups: groups}, nil
}

func (h Handler) CreateGroup(ctx context.Context, request *frontierv1beta1.CreateGroupRequest) (*frontierv1beta1.CreateGroupResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	if err := h.metaSchemaService.Validate(metaDataMap, groupMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	if request.GetBody().GetName() == "" && request.GetBody().GetTitle() != "" {
		request.GetBody().Name = str.GenerateSlug(request.GetBody().GetTitle())
	}

	newGroup, err := h.groupService.Create(ctx, group.Group{
		Name:           request.GetBody().GetName(),
		Title:          request.GetBody().GetTitle(),
		OrganizationID: request.GetOrgId(),
		Metadata:       metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrConflict):
			return nil, grpcConflictError
		case errors.Is(err, group.ErrInvalidDetail), errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidUUID):
			return nil, grpcBadBodyError
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		default:
			return nil, grpcInternalServerError
		}
	}

	metaData, err := newGroup.Metadata.ToStructPB()
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
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
	logger := grpczap.Extract(ctx)

	fetchedGroup, err := h.groupService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist), errors.Is(err, group.ErrInvalidID), errors.Is(err, group.ErrInvalidUUID):
			return nil, grpcGroupNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	groupPB, err := transformGroupToPB(fetchedGroup)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &frontierv1beta1.GetGroupResponse{Group: &groupPB}, nil
}

func (h Handler) UpdateGroup(ctx context.Context, request *frontierv1beta1.UpdateGroupRequest) (*frontierv1beta1.UpdateGroupResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	if err := h.metaSchemaService.Validate(metaDataMap, groupMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	updatedGroup, err := h.groupService.Update(ctx, group.Group{
		ID:             request.GetId(),
		Name:           request.GetBody().GetName(),
		Title:          request.GetBody().GetTitle(),
		OrganizationID: request.GetOrgId(),
		Metadata:       metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
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
			return nil, grpcInternalServerError
		}
	}

	groupPB, err := transformGroupToPB(updatedGroup)
	if err != nil {
		return nil, grpcInternalServerError
	}

	audit.GetAuditor(ctx, request.GetOrgId()).Log(audit.GroupUpdatedEvent, audit.GroupTarget(updatedGroup.ID))
	return &frontierv1beta1.UpdateGroupResponse{Group: &groupPB}, nil
}

func (h Handler) ListGroupUsers(ctx context.Context, request *frontierv1beta1.ListGroupUsersRequest) (*frontierv1beta1.ListGroupUsersResponse, error) {
	logger := grpczap.Extract(ctx)

	users, err := h.userService.ListByGroup(ctx, request.Id, group.MemberPermission)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var userPBs []*frontierv1beta1.User
	for _, user := range users {
		userPb, err := transformUserToPB(user)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		userPBs = append(userPBs, userPb)
	}
	return &frontierv1beta1.ListGroupUsersResponse{
		Users: userPBs,
	}, nil
}

func (h Handler) AddGroupUsers(ctx context.Context, request *frontierv1beta1.AddGroupUsersRequest) (*frontierv1beta1.AddGroupUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.groupService.AddUsers(ctx, request.GetId(), request.GetUserIds()); err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.AddGroupUsersResponse{}, nil
}

func (h Handler) RemoveGroupUser(ctx context.Context, request *frontierv1beta1.RemoveGroupUserRequest) (*frontierv1beta1.RemoveGroupUserResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.groupService.RemoveUsers(ctx, request.GetId(), []string{request.GetUserId()}); err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.RemoveGroupUserResponse{}, nil
}

func (h Handler) EnableGroup(ctx context.Context, request *frontierv1beta1.EnableGroupRequest) (*frontierv1beta1.EnableGroupResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.groupService.Enable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}
	return &frontierv1beta1.EnableGroupResponse{}, nil
}

func (h Handler) DisableGroup(ctx context.Context, request *frontierv1beta1.DisableGroupRequest) (*frontierv1beta1.DisableGroupResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.groupService.Disable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}
	return &frontierv1beta1.DisableGroupResponse{}, nil
}

func (h Handler) DeleteGroup(ctx context.Context, request *frontierv1beta1.DeleteGroupRequest) (*frontierv1beta1.DeleteGroupResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.groupService.Delete(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		default:
			return nil, grpcInternalServerError
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
		Id:        grp.ID,
		Name:      grp.Name,
		Title:     grp.Title,
		OrgId:     grp.OrganizationID,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(grp.CreatedAt),
		UpdatedAt: timestamppb.New(grp.UpdatedAt),
	}, nil
}
