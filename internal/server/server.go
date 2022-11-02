package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	newrelic "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrgrpc"
	"github.com/odpf/salt/log"
	"github.com/odpf/salt/server"
	"github.com/odpf/shield/internal/api"
	"github.com/odpf/shield/internal/api/v1beta1"
	"github.com/odpf/shield/internal/server/grpc_interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func registerHandler(ctx context.Context, s *server.MuxServer, gw *server.GRPCGateway, deps api.Deps) {
	s.RegisterHandler("/admin*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}))

	s.RegisterHandler("/admin/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}))

	// grpc gateway api will have version endpoints
	s.SetGateway("/admin", gw)
	v1beta1.Register(ctx, s, gw, deps)
}

func Serve(
	ctx context.Context,
	logger log.Logger,
	cfg Config,
	nrApp newrelic.Application,
	deps api.Deps,
) (*server.MuxServer, error) {
	s, err := server.NewMux(server.Config{
		Port: cfg.Port,
	}, server.WithMuxGRPCServerOptions(getGRPCMiddleware(cfg, logger, nrApp)))
	if err != nil {
		return nil, err
	}

	gw, err := server.NewGateway("", cfg.Port, server.WithGatewayMuxOptions(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcherFunc(map[string]bool{cfg.IdentityProxyHeader: true}))),
	)
	if err != nil {
		return nil, err
	}

	registerHandler(ctx, s, gw, deps)

	go s.Serve()

	logger.Info("[shield] api is up", "port", cfg.Port)

	return s, nil
}

func Cleanup(ctx context.Context, s *server.MuxServer) {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer shutdownCancel()

	s.Shutdown(shutdownCtx)
}

// REVISIT: passing config.Shield as reference
func getGRPCMiddleware(cfg Config, logger log.Logger, nrApp newrelic.Application) grpc.ServerOption {
	recoveryFunc := func(p interface{}) (err error) {
		fmt.Println("-----------------------------")
		return status.Errorf(codes.Internal, "internal server error")
	}

	grpcRecoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recoveryFunc),
	}

	grpcZapLogger := zap.NewExample().Sugar()
	loggerZap, ok := logger.(*log.Zap)
	if ok {
		grpcZapLogger = loggerZap.GetInternalZapLogger()
	}
	return grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			grpc_interceptors.EnrichCtxWithIdentity(cfg.IdentityProxyHeader),
			grpc_zap.UnaryServerInterceptor(grpcZapLogger.Desugar()),
			grpc_recovery.UnaryServerInterceptor(grpcRecoveryOpts...),
			grpc_ctxtags.UnaryServerInterceptor(),
			nrgrpc.UnaryServerInterceptor(nrApp),
		))
}

func customHeaderMatcherFunc(headerKeys map[string]bool) func(key string) (string, bool) {
	return func(key string) (string, bool) {
		if _, ok := headerKeys[key]; ok {
			return key, true
		}
		return runtime.DefaultHeaderMatcher(key)
	}
}
