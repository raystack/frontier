package v1beta1

import (
	"context"
	"errors"

	"github.com/odpf/shield/core/metaschema"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/pkg/metadata"

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
	Create(ctx context.Context, toCreate role.Role) (role.Role, error)
	List(ctx context.Context) ([]role.Role, error)
	Update(ctx context.Context, toUpdate role.Role) (role.Role, error)
}

func (h Handler) ListRoles(ctx context.Context, request *shieldv1beta1.ListRolesRequest) (*shieldv1beta1.ListRolesResponse, error) {
	logger := grpczap.Extract(ctx)
	var roles []*shieldv1beta1.Role

	roleList, err := h.roleService.List(ctx)
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

func (h Handler) CreateRole(ctx context.Context, request *shieldv1beta1.CreateRoleRequest) (*shieldv1beta1.CreateRoleResponse, error) {
	logger := grpczap.Extract(ctx)
	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	newRole, err := h.roleService.Create(ctx, role.Role{
		ID:          request.GetBody().GetId(),
		Name:        request.GetBody().GetName(),
		Types:       request.GetBody().GetTypes(),
		NamespaceID: request.GetBody().GetNamespaceId(),
		Metadata:    metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, namespace.ErrNotExist),
			errors.Is(err, role.ErrInvalidID),
			errors.Is(err, role.ErrInvalidDetail):
			return nil, grpcBadBodyError
		case errors.Is(err, role.ErrConflict):
			return nil, grpcConflictError
		case errors.Is(errors.Unwrap(err), metaschema.ErrInvalidMetaSchema):
			return nil, grpcBadBodyMetaSchemaError
		default:
			return nil, grpcInternalServerError
		}
	}

	rolePB, err := transformRoleToPB(newRole)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.CreateRoleResponse{Role: &rolePB}, nil
}

func (h Handler) GetRole(ctx context.Context, request *shieldv1beta1.GetRoleRequest) (*shieldv1beta1.GetRoleResponse, error) {
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

	return &shieldv1beta1.GetRoleResponse{Role: &rolePB}, nil
}

func (h Handler) UpdateRole(ctx context.Context, request *shieldv1beta1.UpdateRoleRequest) (*shieldv1beta1.UpdateRoleResponse, error) {
	logger := grpczap.Extract(ctx)

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	updatedRole, err := h.roleService.Update(ctx, role.Role{
		ID:          request.GetId(),
		Name:        request.GetBody().GetName(),
		Types:       request.GetBody().GetTypes(),
		NamespaceID: request.GetBody().GetNamespaceId(),
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
		case errors.Is(errors.Unwrap(err), metaschema.ErrInvalidMetaSchema):
			return nil, grpcBadBodyMetaSchemaError
		default:
			return nil, grpcInternalServerError
		}
	}

	rolePB, err := transformRoleToPB(updatedRole)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.UpdateRoleResponse{Role: &rolePB}, nil
}

func transformRoleToPB(from role.Role) (shieldv1beta1.Role, error) {
	metaData, err := from.Metadata.ToStructPB()
	if err != nil {
		return shieldv1beta1.Role{}, err
	}

	if err != nil {
		return shieldv1beta1.Role{}, err
	}

	return shieldv1beta1.Role{
		Id:    from.ID,
		Name:  from.Name,
		Types: from.Types,
		// TODO(krtkvrm): use namespace id here instead of namespace as object
		//Namespace: &namespace,
		//Tags:      nil,
		//Actions:   nil,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(from.CreatedAt),
		UpdatedAt: timestamppb.New(from.UpdatedAt),
	}, nil
}
