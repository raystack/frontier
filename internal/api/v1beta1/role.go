package v1beta1

import (
	"context"
	"errors"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/core/permission"

	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

var grpcRoleNotFoundErr = status.Errorf(codes.NotFound, "role doesn't exist")

type RoleService interface {
	Get(ctx context.Context, id string) (role.Role, error)
	Upsert(ctx context.Context, toCreate role.Role) (role.Role, error)
	List(ctx context.Context, f role.Filter) ([]role.Role, error)
	Update(ctx context.Context, toUpdate role.Role) (role.Role, error)
	Delete(ctx context.Context, id string) error
}

func (h Handler) ListOrganizationRoles(ctx context.Context, request *frontierv1beta1.ListOrganizationRolesRequest) (*frontierv1beta1.ListOrganizationRolesResponse, error) {
	logger := grpczap.Extract(ctx)
	var roles []*frontierv1beta1.Role

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

	return &frontierv1beta1.ListOrganizationRolesResponse{Roles: roles}, nil
}

func (h Handler) ListRoles(ctx context.Context, request *frontierv1beta1.ListRolesRequest) (*frontierv1beta1.ListRolesResponse, error) {
	logger := grpczap.Extract(ctx)
	var roles []*frontierv1beta1.Role

	roleList, err := h.roleService.List(ctx, role.Filter{
		OrgID: schema.PlatformOrgID.String(),
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

	return &frontierv1beta1.ListRolesResponse{Roles: roles}, nil
}

func (h Handler) CreateRole(ctx context.Context, request *frontierv1beta1.CreateRoleRequest) (*frontierv1beta1.CreateRoleResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	var metaDataMap metadata.Metadata
	var err error
	if request.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.GetBody().GetMetadata().AsMap())

		if err := h.metaSchemaService.Validate(metaDataMap, roleMetaSchema); err != nil {
			logger.Error(err.Error())
			return nil, grpcBadBodyMetaSchemaError
		}
	}

	newRole, err := h.roleService.Upsert(ctx, role.Role{
		Name:        request.GetBody().GetName(),
		Permissions: request.GetBody().GetPermissions(),
		Title:       request.GetBody().GetTitle(),
		OrgID:       schema.PlatformOrgID.String(), // to create a platform wide role
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

	audit.GetAuditor(ctx, schema.PlatformOrgID.String()).Log(audit.RoleCreatedEvent, audit.Target{
		ID:   newRole.ID,
		Type: schema.RoleNamespace,
		Name: newRole.Name,
	})
	return &frontierv1beta1.CreateRoleResponse{Role: &rolePB}, nil
}

func (h Handler) UpdateRole(ctx context.Context, request *frontierv1beta1.UpdateRoleRequest) (*frontierv1beta1.UpdateRoleResponse, error) {
	logger := grpczap.Extract(ctx)
	if len(request.GetBody().GetPermissions()) == 0 {
		return nil, grpcBadBodyError
	}

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, roleMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	updatedRole, err := h.roleService.Update(ctx, role.Role{
		ID:          request.GetId(),
		OrgID:       schema.PlatformOrgID.String(), // to create a platform wide role
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

	audit.GetAuditor(ctx, schema.PlatformOrgID.String()).Log(audit.RoleUpdatedEvent, audit.Target{
		ID:   updatedRole.ID,
		Type: schema.RoleNamespace,
		Name: updatedRole.Name,
	})
	return &frontierv1beta1.UpdateRoleResponse{Role: &rolePB}, nil
}

func (h Handler) CreateOrganizationRole(ctx context.Context, request *frontierv1beta1.CreateOrganizationRoleRequest) (*frontierv1beta1.CreateOrganizationRoleResponse, error) {
	logger := grpczap.Extract(ctx)
	if utils.IsNullUUID(request.GetOrgId()) {
		return nil, grpcBadBodyError
	}

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, roleMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	newRole, err := h.roleService.Upsert(ctx, role.Role{
		Name:        request.GetBody().GetName(),
		Title:       request.GetBody().GetTitle(),
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

	audit.GetAuditor(ctx, request.GetOrgId()).Log(audit.RoleCreatedEvent, audit.Target{
		ID:   newRole.ID,
		Type: schema.RoleNamespace,
		Name: newRole.Name,
	})
	return &frontierv1beta1.CreateOrganizationRoleResponse{Role: &rolePB}, nil
}

func (h Handler) GetOrganizationRole(ctx context.Context, request *frontierv1beta1.GetOrganizationRoleRequest) (*frontierv1beta1.GetOrganizationRoleResponse, error) {
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

	return &frontierv1beta1.GetOrganizationRoleResponse{Role: &rolePB}, nil
}

func (h Handler) UpdateOrganizationRole(ctx context.Context, request *frontierv1beta1.UpdateOrganizationRoleRequest) (*frontierv1beta1.UpdateOrganizationRoleResponse, error) {
	logger := grpczap.Extract(ctx)
	if utils.IsNullUUID(request.GetOrgId()) {
		return nil, grpcBadBodyError
	}
	if len(request.GetBody().GetPermissions()) == 0 {
		return nil, grpcBadBodyError
	}

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, roleMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	updatedRole, err := h.roleService.Update(ctx, role.Role{
		ID:          request.GetId(),
		OrgID:       request.GetOrgId(),
		Name:        request.GetBody().GetName(),
		Title:       request.GetBody().GetTitle(),
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

	audit.GetAuditor(ctx, request.GetOrgId()).Log(audit.RoleUpdatedEvent, audit.Target{
		ID:   updatedRole.ID,
		Type: schema.RoleNamespace,
		Name: updatedRole.Name,
	})
	return &frontierv1beta1.UpdateOrganizationRoleResponse{Role: &rolePB}, nil
}

func (h Handler) DeleteRole(ctx context.Context, request *frontierv1beta1.DeleteRoleRequest) (*frontierv1beta1.DeleteRoleResponse, error) {
	logger := grpczap.Extract(ctx)
	if utils.IsNullUUID(request.GetId()) {
		return nil, grpcBadBodyError
	}

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

	return &frontierv1beta1.DeleteRoleResponse{}, nil
}

func (h Handler) DeleteOrganizationRole(ctx context.Context, request *frontierv1beta1.DeleteOrganizationRoleRequest) (*frontierv1beta1.DeleteOrganizationRoleResponse, error) {
	logger := grpczap.Extract(ctx)
	if utils.IsNullUUID(request.GetOrgId()) || utils.IsNullUUID(request.GetId()) {
		return nil, grpcBadBodyError
	}

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

	audit.GetAuditor(ctx, request.GetOrgId()).Log(audit.RoleDeletedEvent, audit.Target{
		ID:   request.GetId(),
		Type: schema.RoleNamespace,
	})
	return &frontierv1beta1.DeleteOrganizationRoleResponse{}, nil
}

func transformRoleToPB(from role.Role) (frontierv1beta1.Role, error) {
	metaData, err := from.Metadata.ToStructPB()
	if err != nil {
		return frontierv1beta1.Role{}, err
	}

	return frontierv1beta1.Role{
		Id:    from.ID,
		Name:  from.Name,
		Title: from.Title,

		Permissions: from.Permissions,
		OrgId:       from.OrgID,
		State:       from.State.String(),
		Metadata:    metaData,
		CreatedAt:   timestamppb.New(from.CreatedAt),
		UpdatedAt:   timestamppb.New(from.UpdatedAt),
	}, nil
}
