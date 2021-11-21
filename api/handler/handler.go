package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/odpf/shield/model"
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
		switch r.Method {
		case "GET":
			policies, err := deps.V1.PolicyService.ListPolicies(context.Background())
			if err != nil {
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(policies)
		case "POST":
			var payload model.Policy
			err := json.NewDecoder(r.Body).Decode(&payload)
			if err != nil {
				w.WriteHeader(400)
				fmt.Fprintf(w, "Decode error! please check your JSON formating.")
				return
			}
			policy, err := deps.V1.PolicyService.CreatePolicy(context.Background(), payload)
			if err != nil {
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(policy)
		default:
			fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
		}
	}))
	v1.RegisterV1(ctx, s, gw, deps.V1)
}
