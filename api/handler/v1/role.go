package v1

import (
	"context"
	"errors"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/odpf/shield/internal/roles"
	"github.com/odpf/shield/model"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	shieldv1 "go.buf.build/odpf/gwv/odpf/proton/odpf/shield/v1"
)

type RoleService interface {
	Get(ctx context.Context, id string) (model.Role, error)
	Create(ctx context.Context, toCreate model.Role) (model.Role, error)
	List(ctx context.Context) ([]model.Role, error)
	Update(ctx context.Context, toUpdate model.Role) (model.Role, error)
}

func (v Dep) ListRoles(ctx context.Context, request *shieldv1.ListRolesRequest) (*shieldv1.ListRolesResponse, error) {
	logger := grpczap.Extract(ctx)
	var roles []*shieldv1.Role

	roleList, err := v.RoleService.List(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, v := range roleList {
		rolePB, err := transformRoleToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		roles = append(roles, &rolePB)
	}

	return &shieldv1.ListRolesResponse{Roles: roles}, nil
}

func (v Dep) CreateRole(ctx context.Context, request *shieldv1.CreateRoleRequest) (*shieldv1.CreateRoleResponse, error) {
	logger := grpczap.Extract(ctx)
	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	newRole, err := v.RoleService.Create(ctx, model.Role{
		Id:        request.GetBody().Id,
		Name:      request.GetBody().Name,
		Types:     request.GetBody().Types,
		Namespace: request.GetBody().Namespace,
		Metadata:  metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	rolePB, err := transformRoleToPB(newRole)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1.CreateRoleResponse{Role: &rolePB}, nil
}

func (v Dep) GetRole(ctx context.Context, request *shieldv1.GetRoleRequest) (*shieldv1.GetRoleResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedRole, err := v.RoleService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, roles.RoleDoesntExist):
			return nil, grpcProjectNotFoundErr
		case errors.Is(err, roles.InvalidUUID):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	rolePB, err := transformRoleToPB(fetchedRole)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1.GetRoleResponse{Role: &rolePB}, nil
}

func (v Dep) UpdateRole(ctx context.Context, request *shieldv1.UpdateRoleRequest) (*shieldv1.UpdateRoleResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap, err := mapOfStringValues(request.GetBody().Metadata.AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	updatedRole, err := v.RoleService.Update(ctx, model.Role{
		Id:        request.GetBody().Id,
		Name:      request.GetBody().Name,
		Types:     request.GetBody().Types,
		Namespace: request.GetBody().Namespace,
		Metadata:  metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	rolePB, err := transformRoleToPB(updatedRole)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1.UpdateRoleResponse{Role: &rolePB}, nil
}

func transformRoleToPB(from model.Role) (shieldv1.Role, error) {
	metaData, err := structpb.NewStruct(mapOfInterfaceValues(from.Metadata))
	if err != nil {
		return shieldv1.Role{}, err
	}

	return shieldv1.Role{
		Id:        from.Id,
		Name:      from.Name,
		Types:     from.Types,
		Namespace: from.NamespaceId,
		//Tags:      nil,
		//Actions:   nil,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(from.CreatedAt),
		UpdatedAt: timestamppb.New(from.UpdatedAt),
	}, nil
}
