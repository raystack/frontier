package v1

import (
	"context"
	"github.com/odpf/salt/server"

	shieldv1 "github.com/odpf/shield/api/protos/github.com/odpf/proton/shield/v1"
)

type Dep struct {
	shieldv1.UnimplementedShieldServiceServer
}

func RegisterV1(ctx context.Context, s *server.MuxServer, gw *server.GRPCGateway, dep Dep) {
	gw.RegisterHandler(ctx, shieldv1.RegisterShieldServiceHandlerFromEndpoint)

	s.RegisterService(
		&shieldv1.ShieldService_ServiceDesc,
		&dep,
	)
}
