package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/odpf/salt/server"
	v1 "github.com/odpf/shield/api/handler/v1"
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
	s.RegisterHandler("/policies", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		policies, err := deps.V1.PolicyService.ListPolicies(context.Background())
		if err != nil {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(policies)
	}))
	v1.RegisterV1(ctx, s, gw, deps.V1)
}
