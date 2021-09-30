package handler

import (
	"context"
	"fmt"
	"github.com/odpf/salt/server"
	v1 "github.com/odpf/shield/api/handler/v1"
	"net/http"
)

type Deps struct {
	V1 v1.Dep
}

func Register(ctx context.Context, s *server.MuxServer, gw *server.GRPCGateway, deps Deps) {
	s.RegisterHandler("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}))

	// grpc gateway api will have version endpoints
	s.SetGateway("/", gw)

	v1.RegisterV1(ctx, s, gw, deps.V1)
}
