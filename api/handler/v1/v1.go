package v1

import (
	"context"
	"errors"

	"github.com/odpf/salt/server"

	shieldv1 "github.com/odpf/shield/proto/v1"
)

type Dep struct {
	shieldv1.UnimplementedShieldServiceServer
	OrgService          OrganizationService
	ProjectService      ProjectService
	GroupService        GroupService
	RoleService         RoleService
	PolicyService       PolicyService
	UserService         UserService
	NamespaceService    NamespaceService
	ActionService       ActionService
	IdentityProxyHeader string
}

var (
	internalServerError = errors.New("internal server error")
	badRequestError     = errors.New("invalid syntax in body")
)

func RegisterV1(ctx context.Context, s *server.MuxServer, gw *server.GRPCGateway, dep Dep) {
	gw.RegisterHandler(ctx, shieldv1.RegisterShieldServiceHandlerFromEndpoint)

	s.RegisterService(
		&shieldv1.ShieldService_ServiceDesc,
		&dep,
	)
}
