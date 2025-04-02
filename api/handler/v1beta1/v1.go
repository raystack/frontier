package v1beta1

import (
	"context"
	"errors"

	"github.com/goto/salt/server"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

type Dep struct {
	shieldv1beta1.UnimplementedShieldServiceServer
	OrgService             OrganizationService
	ProjectService         ProjectService
	GroupService           GroupService
	RoleService            RoleService
	PolicyService          PolicyService
	UserService            UserService
	NamespaceService       NamespaceService
	ActionService          ActionService
	RelationService        RelationService
	ResourceService        ResourceService
	IdentityProxyHeader    string
	PermissionCheckService PermissionCheckService
}

var (
	internalServerError   = errors.New("internal server error")
	badRequestError       = errors.New("invalid syntax in body")
	permissionDeniedError = errors.New("permission denied")
)

func RegisterV1(ctx context.Context, s *server.MuxServer, gw *server.GRPCGateway, dep Dep) {
	gw.RegisterHandler(ctx, shieldv1beta1.RegisterShieldServiceHandlerFromEndpoint)

	s.RegisterService(
		&shieldv1beta1.ShieldService_ServiceDesc,
		&dep,
	)
}
