package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/odpf/salt/server"
	"github.com/odpf/shield/api/handler/v1beta1"
)

type Deps struct {
	V1beta1 v1beta1.Dep
}

func Register(ctx context.Context, s *server.MuxServer, gw *server.GRPCGateway, deps Deps) {
	s.RegisterHandler("/admin/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}))

	// grpc gateway api will have version endpoints
	s.SetGateway("/admin", gw)
	v1beta1.RegisterV1(ctx, s, gw, deps.V1beta1)
}
