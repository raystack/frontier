package v1beta1

import (
	"context"

	"github.com/odpf/shield/pkg/errors"
	"github.com/odpf/shield/pkg/metadata"
	"github.com/odpf/shield/pkg/str"
	"github.com/odpf/shield/pkg/uuid"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/user"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockery --name=GroupService -r --case underscore --with-expecter --structname GroupService --filename group_service.go --output=./mocks
type GroupService interface {
	Create(ctx context.Context, grp group.Group) (group.Group, error)
	Get(ctx context.Context, id string) (group.Group, error)
	List(ctx context.Context, flt group.Filter) ([]group.Group, error)
	Update(ctx context.Context, grp group.Group) (group.Group, error)
	ListUserGroups(ctx context.Context, userId string, roleId string) ([]group.Group, error)
	ListUsers(ctx context.Context, groupId string) ([]user.User, error)
	ListAdmins(ctx context.Context, groupId string) ([]user.User, error)
	AddUsers(ctx context.Context, groupId string, userIds []string) ([]user.User, error)
	RemoveUser(ctx context.Context, groupId string, userId string) ([]user.User, error)
	AddAdmins(ctx context.Context, groupId string, userIds []string) ([]user.User, error)
	RemoveAdmin(ctx context.Context, groupId string, userId string) ([]user.User, error)
}

var (
	grpcGroupNotFoundErr = status.Errorf(codes.NotFound, "group doesn't exist")
)

func (h Handler) ListGroups(ctx context.Context, request *shieldv1beta1.ListGroupsRequest) (*shieldv1beta1.ListGroupsResponse, error) {
	logger := grpczap.Extract(ctx)

	var groups []*shieldv1beta1.Group

	groupList, err := h.groupService.List(ctx, group.Filter{
		OrganizationID: request.GetOrgId(),
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

	return &shieldv1beta1.ListGroupsResponse{Groups: groups}, nil
}

func (h Handler) CreateGroup(ctx context.Context, request *shieldv1beta1.CreateGroupRequest) (*shieldv1beta1.CreateGroupResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	grp := group.Group{
		Name:           request.GetBody().GetName(),
		Slug:           request.GetBody().GetSlug(),
		Organization:   organization.Organization{ID: request.GetBody().GetOrgId()},
		OrganizationID: request.GetBody().GetOrgId(),
		Metadata:       metaDataMap,
	}

	if str.IsStringEmpty(grp.Slug) {
		grp.Slug = str.GenerateSlug(grp.Name)
	}

	newGroup, err := h.groupService.Create(ctx, grp)
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

	return &shieldv1beta1.GetGroupResponse{Group: &groupPB}, nil
}

func (h Handler) AddGroupUser(ctx context.Context, request *shieldv1beta1.AddGroupUserRequest) (*shieldv1beta1.AddGroupUserResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	//TODO might need to check userids are uuid here

	updatedUsers, err := h.groupService.AddUsers(ctx, request.GetId(), request.GetBody().GetUserIds())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		case errors.Is(err, errors.ErrForbidden):
			return nil, grpcPermissionDenied
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		case errors.Is(err, user.ErrInvalidUUID), errors.Is(err, user.ErrInvalidID):
			return nil, grpcBadBodyError
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

	if _, err := h.groupService.RemoveUser(ctx, request.GetId(), request.GetUserId()); err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		case errors.Is(err, errors.ErrForbidden):
			return nil, grpcPermissionDenied
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		case errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidUUID), errors.Is(err, user.ErrNotExist):
			return nil, grpcUserNotFoundError
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

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	var updatedGroup group.Group
	if uuid.IsValid(request.GetId()) {
		updatedGroup, err = h.groupService.Update(ctx, group.Group{
			ID:             request.GetId(),
			Name:           request.GetBody().GetName(),
			Slug:           request.GetBody().GetSlug(),
			Organization:   organization.Organization{ID: request.GetBody().GetOrgId()},
			OrganizationID: request.GetBody().GetOrgId(),
			Metadata:       metaDataMap,
		})
	} else {
		updatedGroup, err = h.groupService.Update(ctx, group.Group{
			Name:           request.GetBody().GetName(),
			Slug:           request.GetId(),
			Organization:   organization.Organization{ID: request.GetBody().GetOrgId()},
			OrganizationID: request.GetBody().GetOrgId(),
			Metadata:       metaDataMap,
		})
	}
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

	return &shieldv1beta1.UpdateGroupResponse{Group: &groupPB}, nil
}

func (h Handler) ListGroupAdmins(ctx context.Context, request *shieldv1beta1.ListGroupAdminsRequest) (*shieldv1beta1.ListGroupAdminsResponse, error) {
	logger := grpczap.Extract(ctx)

	usersList, err := h.groupService.ListAdmins(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
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

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	updatedUsers, err := h.groupService.AddAdmins(ctx, request.GetId(), request.GetBody().GetUserIds())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		case errors.Is(err, errors.ErrForbidden):
			return nil, grpcPermissionDenied
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		case errors.Is(err, user.ErrInvalidUUID), errors.Is(err, user.ErrInvalidID):
			return nil, grpcBadBodyError
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

	if _, err := h.groupService.RemoveAdmin(ctx, request.GetId(), request.GetUserId()); err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, grpcUnauthenticated
		case errors.Is(err, errors.ErrForbidden):
			return nil, grpcPermissionDenied
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		case errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidUUID), errors.Is(err, user.ErrNotExist):
			return nil, grpcUserNotFoundError
		default:
			return nil, grpcInternalServerError
		}
	}

	return &shieldv1beta1.RemoveGroupAdminResponse{
		Message: "Removed Admin from group",
	}, nil
}

func (h Handler) ListGroupUsers(ctx context.Context, request *shieldv1beta1.ListGroupUsersRequest) (*shieldv1beta1.ListGroupUsersResponse, error) {
	logger := grpczap.Extract(ctx)

	usersList, err := h.groupService.ListUsers(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrNotExist):
			return nil, grpcGroupNotFoundErr
		}
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

func transformGroupToPB(grp group.Group) (shieldv1beta1.Group, error) {
	metaData, err := grp.Metadata.ToStructPB()
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
