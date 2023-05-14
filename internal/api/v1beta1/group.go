package v1beta1

import (
	"context"
	"strings"

	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/pkg/metadata"
	"github.com/odpf/shield/pkg/str"
	"github.com/odpf/shield/pkg/uuid"
	"github.com/pkg/errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/metaschema"
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
	ListGroupRelations(ctx context.Context, objectId, subjectType, role string) ([]user.User, []group.Group, map[string][]string, map[string][]string, error)
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

var (
	grpcGroupNotFoundErr = status.Errorf(codes.NotFound, "group doesn't exist")
)

func (h Handler) ListGroups(ctx context.Context, request *shieldv1beta1.ListGroupsRequest) (*shieldv1beta1.ListGroupsResponse, error) {
	logger := grpczap.Extract(ctx)
	// TODO(kushsharma): apply admin level authz

	var groups []*shieldv1beta1.Group
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

	return &shieldv1beta1.ListGroupsResponse{Groups: groups}, nil
}

func (h Handler) ListOrganizationGroups(ctx context.Context, request *shieldv1beta1.ListOrganizationGroupsRequest) (*shieldv1beta1.ListOrganizationGroupsResponse, error) {
	logger := grpczap.Extract(ctx)

	var groups []*shieldv1beta1.Group
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

	return &shieldv1beta1.ListOrganizationGroupsResponse{Groups: groups}, nil
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

	if err := h.metaSchemaService.Validate(metaDataMap, groupMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	grp := group.Group{
		Name:           request.GetBody().GetName(),
		Slug:           request.GetBody().GetSlug(),
		OrganizationID: request.GetBody().GetOrgId(),
		Metadata:       metaDataMap,
	}

	if strings.TrimSpace(grp.Slug) == "" {
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
		case errors.Is(errors.Unwrap(err), metaschema.ErrInvalidMetaSchema):
			return nil, grpcBadBodyMetaSchemaError
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
		OrgId:     newGroup.OrganizationID,
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

func (h Handler) UpdateGroup(ctx context.Context, request *shieldv1beta1.UpdateGroupRequest) (*shieldv1beta1.UpdateGroupResponse, error) {
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

	var updatedGroup group.Group
	if uuid.IsValid(request.GetId()) {
		updatedGroup, err = h.groupService.Update(ctx, group.Group{
			ID:             request.GetId(),
			Name:           request.GetBody().GetName(),
			Slug:           request.GetBody().GetSlug(),
			OrganizationID: request.GetBody().GetOrgId(),
			Metadata:       metaDataMap,
		})
	} else {
		updatedGroup, err = h.groupService.Update(ctx, group.Group{
			Name:           request.GetBody().GetName(),
			Slug:           request.GetId(),
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
		case errors.Is(errors.Unwrap(err), metaschema.ErrInvalidMetaSchema):
			return nil, grpcBadBodyMetaSchemaError
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

func (h Handler) ListGroupRelations(ctx context.Context, request *shieldv1beta1.ListGroupRelationsRequest) (*shieldv1beta1.ListGroupRelationsResponse, error) {
	logger := grpczap.Extract(ctx)
	groupRelations := []*shieldv1beta1.GroupRelation{}

	users, groups, userIDRoleMap, groupIDRoleMap, err := h.groupService.ListGroupRelations(ctx, request.Id, request.SubjectType, request.Role)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, user := range users {
		userPb, err := transformUserToPB(user)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		for _, r := range userIDRoleMap[userPb.Id] {
			role := strings.Split(r, ":")

			grprel := &shieldv1beta1.GroupRelation{
				SubjectType: schema.UserPrincipal,
				Role:        role[1],
				Subject: &shieldv1beta1.GroupRelation_User{
					User: &userPb,
				},
			}
			groupRelations = append(groupRelations, grprel)
		}
	}

	for _, group := range groups {
		groupPb, err := transformGroupToPB(group)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		for _, r := range groupIDRoleMap[groupPb.Id] {
			role := strings.Split(r, ":")

			grprel := &shieldv1beta1.GroupRelation{
				SubjectType: schema.GroupPrincipal,
				Role:        role[1],
				Subject: &shieldv1beta1.GroupRelation_Group{
					Group: &groupPb,
				},
			}
			groupRelations = append(groupRelations, grprel)
		}
	}

	return &shieldv1beta1.ListGroupRelationsResponse{
		Relations: groupRelations,
	}, nil
}

func (h Handler) EnableGroup(ctx context.Context, request *shieldv1beta1.EnableGroupRequest) (*shieldv1beta1.EnableGroupResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.groupService.Enable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.EnableGroupResponse{}, nil
}

func (h Handler) DisableGroup(ctx context.Context, request *shieldv1beta1.DisableGroupRequest) (*shieldv1beta1.DisableGroupResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.groupService.Disable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.DisableGroupResponse{}, nil
}

func (h Handler) DeleteGroup(ctx context.Context, request *shieldv1beta1.DeleteGroupRequest) (*shieldv1beta1.DeleteGroupResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.groupService.Delete(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.DeleteGroupResponse{}, nil
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
		OrgId:     grp.OrganizationID,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(grp.CreatedAt),
		UpdatedAt: timestamppb.New(grp.UpdatedAt),
	}, nil
}
