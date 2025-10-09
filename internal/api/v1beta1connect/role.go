package v1beta1connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RoleService interface {
	Get(ctx context.Context, id string) (role.Role, error)
	Upsert(ctx context.Context, toCreate role.Role) (role.Role, error)
	List(ctx context.Context, f role.Filter) ([]role.Role, error)
	Update(ctx context.Context, toUpdate role.Role) (role.Role, error)
	Delete(ctx context.Context, id string) error
}

func (h *ConnectHandler) CreateRole(ctx context.Context, request *connect.Request[frontierv1beta1.CreateRoleRequest]) (*connect.Response[frontierv1beta1.CreateRoleResponse], error) {
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	var metaDataMap metadata.Metadata
	if request.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

		if err := h.metaSchemaService.Validate(metaDataMap, roleMetaSchema); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidMetadata)
		}
	}

	newRole, err := h.roleService.Upsert(ctx, role.Role{
		Name:        request.Msg.GetBody().GetName(),
		Permissions: request.Msg.GetBody().GetPermissions(),
		Scopes:      request.Msg.GetBody().GetScopes(),
		Title:       request.Msg.GetBody().GetTitle(),
		OrgID:       schema.PlatformOrgID.String(), // to create a platform wide role
		Metadata:    metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrNotExist),
			errors.Is(err, role.ErrInvalidID),
			errors.Is(err, role.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, role.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	rolePB, err := transformRoleToPB(newRole)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, schema.PlatformOrgID.String()).Log(audit.RoleCreatedEvent, audit.Target{
		ID:   newRole.ID,
		Type: schema.RoleNamespace,
		Name: newRole.Name,
	})
	return connect.NewResponse(&frontierv1beta1.CreateRoleResponse{Role: &rolePB}), nil
}

func (h *ConnectHandler) UpdateRole(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateRoleRequest]) (*connect.Response[frontierv1beta1.UpdateRoleResponse], error) {
	if len(request.Msg.GetBody().GetPermissions()) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, roleMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	updatedRole, err := h.roleService.Update(ctx, role.Role{
		ID:          request.Msg.GetId(),
		OrgID:       schema.PlatformOrgID.String(), // to create a platform wide role
		Title:       request.Msg.GetBody().GetTitle(),
		Name:        request.Msg.GetBody().GetName(),
		Scopes:      request.Msg.GetBody().GetScopes(),
		Permissions: request.Msg.GetBody().GetPermissions(),
		Metadata:    metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, role.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, role.ErrInvalidID),
			errors.Is(err, role.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, role.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	rolePB, err := transformRoleToPB(updatedRole)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, schema.PlatformOrgID.String()).Log(audit.RoleUpdatedEvent, audit.Target{
		ID:   updatedRole.ID,
		Type: schema.RoleNamespace,
		Name: updatedRole.Name,
	})
	return connect.NewResponse(&frontierv1beta1.UpdateRoleResponse{Role: &rolePB}), nil
}

func transformRoleToPB(from role.Role) (frontierv1beta1.Role, error) {
	metaData, err := from.Metadata.ToStructPB()
	if err != nil {
		return frontierv1beta1.Role{}, err
	}

	return frontierv1beta1.Role{
		Id:          from.ID,
		Name:        from.Name,
		Title:       from.Title,
		Scopes:      from.Scopes,
		Permissions: from.Permissions,
		OrgId:       from.OrgID,
		State:       from.State.String(),
		Metadata:    metaData,
		CreatedAt:   timestamppb.New(from.CreatedAt),
		UpdatedAt:   timestamppb.New(from.UpdatedAt),
	}, nil
}

func (h *ConnectHandler) ListRoles(ctx context.Context, request *connect.Request[frontierv1beta1.ListRolesRequest]) (*connect.Response[frontierv1beta1.ListRolesResponse], error) {
	var roles []*frontierv1beta1.Role

	roleList, err := h.roleService.List(ctx, role.Filter{
		OrgID:  schema.PlatformOrgID.String(),
		Scopes: request.Msg.GetScopes(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range roleList {
		rolePB, err := transformRoleToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		roles = append(roles, &rolePB)
	}

	return connect.NewResponse(&frontierv1beta1.ListRolesResponse{Roles: roles}), nil
}

func (h *ConnectHandler) ListOrganizationRoles(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationRolesRequest]) (*connect.Response[frontierv1beta1.ListOrganizationRolesResponse], error) {
	var roles []*frontierv1beta1.Role

	roleList, err := h.roleService.List(ctx, role.Filter{
		OrgID:  request.Msg.GetOrgId(),
		Scopes: request.Msg.GetScopes(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, v := range roleList {
		rolePB, err := transformRoleToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}

		roles = append(roles, &rolePB)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationRolesResponse{Roles: roles}), nil
}

func (h *ConnectHandler) CreateOrganizationRole(ctx context.Context, request *connect.Request[frontierv1beta1.CreateOrganizationRoleRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationRoleResponse], error) {
	if utils.IsNullUUID(request.Msg.GetOrgId()) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, roleMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	newRole, err := h.roleService.Upsert(ctx, role.Role{
		Name:        request.Msg.GetBody().GetName(),
		Title:       request.Msg.GetBody().GetTitle(),
		Scopes:      request.Msg.GetBody().GetScopes(),
		Permissions: request.Msg.GetBody().GetPermissions(),
		OrgID:       request.Msg.GetOrgId(),
		Metadata:    metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, permission.ErrNotExist),
			errors.Is(err, role.ErrInvalidID),
			errors.Is(err, role.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, role.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	rolePB, err := transformRoleToPB(newRole)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, request.Msg.GetOrgId()).Log(audit.RoleCreatedEvent, audit.Target{
		ID:   newRole.ID,
		Type: schema.RoleNamespace,
		Name: newRole.Name,
	})
	return connect.NewResponse(&frontierv1beta1.CreateOrganizationRoleResponse{Role: &rolePB}), nil
}

func (h *ConnectHandler) GetOrganizationRole(ctx context.Context, request *connect.Request[frontierv1beta1.GetOrganizationRoleRequest]) (*connect.Response[frontierv1beta1.GetOrganizationRoleResponse], error) {
	fetchedRole, err := h.roleService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, role.ErrNotExist), errors.Is(err, role.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	rolePB, err := transformRoleToPB(fetchedRole)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.GetOrganizationRoleResponse{Role: &rolePB}), nil
}

func (h *ConnectHandler) UpdateOrganizationRole(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateOrganizationRoleRequest]) (*connect.Response[frontierv1beta1.UpdateOrganizationRoleResponse], error) {
	if utils.IsNullUUID(request.Msg.GetOrgId()) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}
	if len(request.Msg.GetBody().GetPermissions()) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, roleMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	updatedRole, err := h.roleService.Update(ctx, role.Role{
		ID:          request.Msg.GetId(),
		OrgID:       request.Msg.GetOrgId(),
		Name:        request.Msg.GetBody().GetName(),
		Title:       request.Msg.GetBody().GetTitle(),
		Scopes:      request.Msg.GetBody().GetScopes(),
		Permissions: request.Msg.GetBody().GetPermissions(),
		Metadata:    metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, role.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, role.ErrInvalidID),
			errors.Is(err, role.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, role.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	rolePB, err := transformRoleToPB(updatedRole)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, request.Msg.GetOrgId()).Log(audit.RoleUpdatedEvent, audit.Target{
		ID:   updatedRole.ID,
		Type: schema.RoleNamespace,
		Name: updatedRole.Name,
	})
	return connect.NewResponse(&frontierv1beta1.UpdateOrganizationRoleResponse{Role: &rolePB}), nil
}

func (h *ConnectHandler) DeleteOrganizationRole(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteOrganizationRoleRequest]) (*connect.Response[frontierv1beta1.DeleteOrganizationRoleResponse], error) {
	if utils.IsNullUUID(request.Msg.GetOrgId()) || utils.IsNullUUID(request.Msg.GetId()) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	err := h.roleService.Delete(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, role.ErrNotExist), errors.Is(err, role.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrRoleNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	audit.GetAuditor(ctx, request.Msg.GetOrgId()).Log(audit.RoleDeletedEvent, audit.Target{
		ID:   request.Msg.GetId(),
		Type: schema.RoleNamespace,
	})
	return connect.NewResponse(&frontierv1beta1.DeleteOrganizationRoleResponse{}), nil
}

func (h *ConnectHandler) DeleteRole(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteRoleRequest]) (*connect.Response[frontierv1beta1.DeleteRoleResponse], error) {
	if utils.IsNullUUID(request.Msg.GetId()) {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	err := h.roleService.Delete(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, role.ErrNotExist), errors.Is(err, role.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.DeleteRoleResponse{}), nil
}
