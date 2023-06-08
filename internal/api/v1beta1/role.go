package v1beta1

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/odpf/shield/core/permission"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/pkg/metadata"
	"github.com/odpf/shield/pkg/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

var grpcRoleNotFoundErr = status.Errorf(codes.NotFound, "role doesn't exist")

//go:generate mockery --name=RoleService -r --case underscore --with-expecter --structname RoleService --filename role_service.go --output=./mocks
type RoleService interface {
	Get(ctx context.Context, id string) (role.Role, error)
	Upsert(ctx context.Context, toCreate role.Role) (role.Role, error)
	List(ctx context.Context, f role.Filter) ([]role.Role, error)
	Update(ctx context.Context, toUpdate role.Role) (role.Role, error)
	Delete(ctx context.Context, id string) error
}

func (h Handler) ListOrganizationRoles(ctx context.Context, request *shieldv1beta1.ListOrganizationRolesRequest) (*shieldv1beta1.ListOrganizationRolesResponse, error) {
	logger := grpczap.Extract(ctx)
	var roles []*shieldv1beta1.Role

	roleList, err := h.roleService.List(ctx, role.Filter{
		OrgID: request.GetOrgId(),
	})
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

	return &shieldv1beta1.ListOrganizationRolesResponse{Roles: roles}, nil
}

func (h Handler) ListRoles(ctx context.Context, request *shieldv1beta1.ListRolesRequest) (*shieldv1beta1.ListRolesResponse, error) {
	logger := grpczap.Extract(ctx)
	var roles []*shieldv1beta1.Role

	roleList, err := h.roleService.List(ctx, role.Filter{
		OrgID: uuid.Nil.String(),
	})
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

	return &shieldv1beta1.ListRolesResponse{Roles: roles}, nil
}

func (h Handler) CreateOrganizationRole(ctx context.Context, request *shieldv1beta1.CreateOrganizationRoleRequest) (*shieldv1beta1.CreateOrganizationRoleResponse, error) {
	logger := grpczap.Extract(ctx)
	if utils.IsNullUUID(request.GetOrgId()) {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	if err := h.metaSchemaService.Validate(metaDataMap, roleMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	newRole, err := h.roleService.Upsert(ctx, role.Role{
		Name:        request.GetBody().GetName(),
		Permissions: request.GetBody().GetPermissions(),
		OrgID:       request.GetOrgId(),
		Metadata:    metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrNotExist),
			errors.Is(err, role.ErrInvalidID),
			errors.Is(err, role.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, role.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	rolePB, err := transformRoleToPB(newRole)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateOrganizationRoleResponse{Role: &rolePB}, nil
}

func (h Handler) GetOrganizationRole(ctx context.Context, request *shieldv1beta1.GetOrganizationRoleRequest) (*shieldv1beta1.GetOrganizationRoleResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedRole, err := h.roleService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, role.ErrNotExist), errors.Is(err, role.ErrInvalidID):
			return nil, grpcRoleNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	rolePB, err := transformRoleToPB(fetchedRole)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.GetOrganizationRoleResponse{Role: &rolePB}, nil
}

func (h Handler) UpdateOrganizationRole(ctx context.Context, request *shieldv1beta1.UpdateOrganizationRoleRequest) (*shieldv1beta1.UpdateOrganizationRoleResponse, error) {
	logger := grpczap.Extract(ctx)
	if utils.IsNullUUID(request.GetOrgId()) {
		return nil, grpcBadBodyError
	}
	if len(request.GetBody().GetPermissions()) == 0 {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	if err := h.metaSchemaService.Validate(metaDataMap, roleMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	updatedRole, err := h.roleService.Update(ctx, role.Role{
		ID:          request.GetId(),
		OrgID:       request.GetOrgId(),
		Name:        request.GetBody().GetName(),
		Permissions: request.GetBody().GetPermissions(),
		Metadata:    metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, role.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, role.ErrInvalidID),
			errors.Is(err, role.ErrNotExist):
			return nil, grpcRoleNotFoundErr
		case errors.Is(err, role.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	rolePB, err := transformRoleToPB(updatedRole)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdateOrganizationRoleResponse{Role: &rolePB}, nil
}

func (h Handler) DeleteOrganizationRole(ctx context.Context, request *shieldv1beta1.DeleteOrganizationRoleRequest) (*shieldv1beta1.DeleteOrganizationRoleResponse, error) {
	logger := grpczap.Extract(ctx)

	err := h.roleService.Delete(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, role.ErrNotExist), errors.Is(err, role.ErrInvalidID):
			return nil, grpcRoleNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	return &shieldv1beta1.DeleteOrganizationRoleResponse{}, nil
}

func transformRoleToPB(from role.Role) (shieldv1beta1.Role, error) {
	metaData, err := from.Metadata.ToStructPB()
	if err != nil {
		return shieldv1beta1.Role{}, err
	}

	return shieldv1beta1.Role{
		Id:   from.ID,
		Name: from.Name,

		Permissions: from.Permissions,
		OrgId:       from.OrgID,
		State:       from.State.String(),
		Metadata:    metaData,
		CreatedAt:   timestamppb.New(from.CreatedAt),
		UpdatedAt:   timestamppb.New(from.UpdatedAt),
	}, nil
}
