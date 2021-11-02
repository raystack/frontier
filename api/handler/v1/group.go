package v1

import (
	"context"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/odpf/shield/internal/group"
	"github.com/odpf/shield/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

type GroupService interface {
	CreateGroup(ctx context.Context, grp group.Group) (group.Group, error)
}

func (v Dep) ListGroups(ctx context.Context, request *shieldv1.ListGroupsRequest) (*shieldv1.ListGroupsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) CreateGroup(ctx context.Context, request *shieldv1.CreateGroupRequest) (*shieldv1.CreateGroupResponse, error) {
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

	newGroup, err := v.GroupService.CreateGroup(ctx, group.Group{
		Name:         request.Body.Name,
		Slug:         slug,
		Organization: model.Organization{Id: "23cb5c8f-859f-43ad-ae2b-47fae181bd8a"},
		Metadata:     metaDataMap,
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

	return &shieldv1.CreateGroupResponse{Group: &shieldv1.Group{
		Id:        newGroup.Id,
		Name:      newGroup.Name,
		Slug:      newGroup.Slug,
		OrgId:     newGroup.Organization.Id,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(newGroup.CreatedAt),
		UpdatedAt: timestamppb.New(newGroup.UpdatedAt),
	}}, nil
}

func (v Dep) GetGroup(ctx context.Context, request *shieldv1.GetGroupRequest) (*shieldv1.GetGroupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) ListGroupUsers(ctx context.Context, request *shieldv1.ListGroupUsersRequest) (*shieldv1.ListGroupUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

func (v Dep) UpdateGroup(ctx context.Context, request *shieldv1.UpdateGroupRequest) (*shieldv1.UpdateGroupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}
