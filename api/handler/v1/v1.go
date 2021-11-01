package v1

import (
	"context"
	"errors"

	"github.com/odpf/salt/server"

	shieldv1 "go.buf.build/odpf/gw/odpf/proton/odpf/shield/v1"
)

type Dep struct {
	shieldv1.UnimplementedShieldServiceServer
	OrgService     OrganizationService
	ProjectService ProjectService
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
