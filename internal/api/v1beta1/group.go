package v1beta1

import (
	"context"
	"strings"

	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/organization"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GroupService interface {
	CreateGroup(ctx context.Context, grp group.Group) (group.Group, error)
	GetGroup(ctx context.Context, id string) (group.Group, error)
	ListGroups(ctx context.Context, org organization.Organization) ([]group.Group, error)
	UpdateGroup(ctx context.Context, grp group.Group) (group.Group, error)
	AddUsersToGroup(ctx context.Context, groupId string, userIds []string) ([]user.User, error)
	ListUserGroups(ctx context.Context, userId string, roleId string) ([]group.Group, error)
	ListGroupUsers(ctx context.Context, groupId string) ([]user.User, error)
	ListGroupAdmins(ctx context.Context, groupId string) ([]user.User, error)
	RemoveUserFromGroup(ctx context.Context, groupId string, userId string) ([]user.User, error)
	AddAdminsToGroup(ctx context.Context, groupId string, userIds []string) ([]user.User, error)
	RemoveAdminFromGroup(ctx context.Context, groupId string, userId string) ([]user.User, error)
}

var (
	grpcGroupNotFoundErr = status.Errorf(codes.NotFound, "group doesn't exist")
)

func (h Handler) ListGroups(ctx context.Context, request *shieldv1beta1.ListGroupsRequest) (*shieldv1beta1.ListGroupsResponse, error) {
	logger := grpczap.Extract(ctx)

	var groups []*shieldv1beta1.Group

	groupList, err := h.groupService.ListGroups(ctx, organization.Organization{ID: request.OrgId})
	if errors.Is(err, group.ErrNotExist) {
		return nil, nil
	} else if err != nil {
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

	return &shieldv1beta1.ListGroupsResponse{Groups: groups}, nil
}

func (h Handler) CreateGroup(ctx context.Context, request *shieldv1beta1.CreateGroupRequest) (*shieldv1beta1.CreateGroupResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	slug := request.GetBody().Slug
	if strings.TrimSpace(slug) == "" {
		slug = generateSlug(request.GetBody().Name)
	}

	newGroup, err := h.groupService.CreateGroup(ctx, group.Group{
		Name:           request.Body.Name,
		Slug:           slug,
		OrganizationID: request.Body.OrgId,
		Metadata:       metaDataMap,
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	metaData, err := structpb.NewStruct(mapOfInterfaceValues(newGroup.Metadata))
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateGroupResponse{Group: &shieldv1beta1.Group{
		Id:        newGroup.ID,
		Name:      newGroup.Name,
		Slug:      newGroup.Slug,
		OrgId:     newGroup.Organization.ID,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newGroup.CreatedAt),
		UpdatedAt: timestamppb.New(newGroup.UpdatedAt),
	}}, nil
}

func (h Handler) GetGroup(ctx context.Context, request *shieldv1beta1.GetGroupRequest) (*shieldv1beta1.GetGroupResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedGroup, err := h.groupService.GetGroup(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		case errors.Is(err, group.ErrInvalidUUID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	groupPB, err := transformGroupToPB(fetchedGroup)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetGroupResponse{Group: &groupPB}, nil
}

func (h Handler) ListGroupUsers(ctx context.Context, request *shieldv1beta1.ListGroupUsersRequest) (*shieldv1beta1.ListGroupUsersResponse, error) {
	logger := grpczap.Extract(ctx)

	usersList, err := h.groupService.ListGroupUsers(ctx, request.GetId())

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var users []*shieldv1beta1.User

	for _, u := range usersList {
		userPB, err := transformUserToPB(u)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		users = append(users, &userPB)
	}

	return &shieldv1beta1.ListGroupUsersResponse{
		Users: users,
	}, nil
}

func (h Handler) AddGroupUser(ctx context.Context, request *shieldv1beta1.AddGroupUserRequest) (*shieldv1beta1.AddGroupUserResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.Body == nil {
		return nil, grpcBadBodyError
	}
	updatedUsers, err := h.groupService.AddUsersToGroup(ctx, request.GetId(), request.GetBody().UserIds)

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "group to be updated not found")
		case errors.Is(err, errors.Unauthorized):
			return nil, grpcPermissionDenied
		default:
			return nil, grpcInternalServerError
		}
	}

	var users []*shieldv1beta1.User

	for _, u := range updatedUsers {
		userPB, err := transformUserToPB(u)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		users = append(users, &userPB)
	}

	return &shieldv1beta1.AddGroupUserResponse{
		Users: users,
	}, nil
}

func (h Handler) RemoveGroupUser(ctx context.Context, request *shieldv1beta1.RemoveGroupUserRequest) (*shieldv1beta1.RemoveGroupUserResponse, error) {
	logger := grpczap.Extract(ctx)
	_, err := h.groupService.RemoveUserFromGroup(ctx, request.GetId(), request.GetUserId())

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "group to be updated not found")
		case errors.Is(err, errors.Unauthorized):
			return nil, grpcPermissionDenied
		default:
			return nil, grpcInternalServerError
		}
	}

	return &shieldv1beta1.RemoveGroupUserResponse{
		Message: "Removed User from group",
	}, nil
}

func (h Handler) UpdateGroup(ctx context.Context, request *shieldv1beta1.UpdateGroupRequest) (*shieldv1beta1.UpdateGroupResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.Body == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	updatedGroup, err := h.groupService.UpdateGroup(ctx, group.Group{
		ID:           request.GetId(),
		Name:         request.GetBody().GetName(),
		Slug:         request.GetBody().GetSlug(),
		Organization: organization.Organization{ID: request.GetBody().OrgId},
		Metadata:     metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "group to be updated not found")
		default:
			return nil, grpcInternalServerError
		}
	}

	groupPB, err := transformGroupToPB(updatedGroup)
	if err != nil {
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdateGroupResponse{Group: &groupPB}, nil
}

func (h Handler) ListGroupAdmins(ctx context.Context, request *shieldv1beta1.ListGroupAdminsRequest) (*shieldv1beta1.ListGroupAdminsResponse, error) {
	logger := grpczap.Extract(ctx)
	usersList, err := h.groupService.ListGroupAdmins(ctx, request.GetId())

	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var users []*shieldv1beta1.User

	for _, u := range usersList {
		userPB, err := transformUserToPB(u)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		users = append(users, &userPB)
	}

	return &shieldv1beta1.ListGroupAdminsResponse{
		Users: users,
	}, nil
}

func (h Handler) AddGroupAdmin(ctx context.Context, request *shieldv1beta1.AddGroupAdminRequest) (*shieldv1beta1.AddGroupAdminResponse, error) {
	logger := grpczap.Extract(ctx)
	if request.Body == nil {
		return nil, grpcBadBodyError
	}
	updatedUsers, err := h.groupService.AddAdminsToGroup(ctx, request.GetId(), request.GetBody().UserIds)

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "group to be updated not found")
		case errors.Is(err, errors.Unauthorized):
			return nil, grpcPermissionDenied
		default:
			return nil, grpcInternalServerError
		}
	}

	var users []*shieldv1beta1.User

	for _, u := range updatedUsers {
		userPB, err := transformUserToPB(u)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		users = append(users, &userPB)
	}

	return &shieldv1beta1.AddGroupAdminResponse{
		Users: users,
	}, nil
}

func (h Handler) RemoveGroupAdmin(ctx context.Context, request *shieldv1beta1.RemoveGroupAdminRequest) (*shieldv1beta1.RemoveGroupAdminResponse, error) {
	logger := grpczap.Extract(ctx)
	_, err := h.groupService.RemoveAdminFromGroup(ctx, request.GetId(), request.GetUserId())

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, status.Errorf(codes.NotFound, "group to be updated not found")
		case errors.Is(err, errors.Unauthorized):
			return nil, grpcPermissionDenied
		default:
			return nil, grpcInternalServerError
		}
	}

	return &shieldv1beta1.RemoveGroupAdminResponse{
		Message: "Removed Admin from group",
	}, nil
}

func transformGroupToPB(grp group.Group) (shieldv1beta1.Group, error) {
	metaData, err := structpb.NewStruct(mapOfInterfaceValues(grp.Metadata))
	if err != nil {
		return shieldv1beta1.Group{}, err
	}

	return shieldv1beta1.Group{
		Id:        grp.ID,
		Name:      grp.Name,
		Slug:      grp.Slug,
		OrgId:     grp.Organization.ID,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(grp.CreatedAt),
		UpdatedAt: timestamppb.New(grp.UpdatedAt),
	}, nil
}
