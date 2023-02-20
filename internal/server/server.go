package server

import (
	"context"
	"errors"
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
	"github.com/odpf/salt/mux"
	"github.com/odpf/shield/internal/api"
	"github.com/odpf/shield/internal/api/v1beta1"
	"github.com/odpf/shield/internal/server/grpc_interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func registerHandler(ctx context.Context, sh *http.ServeMux, gw *runtime.ServeMux, sg *grpc.Server, deps api.Deps, address string) error {
	sh.Handle("/admin*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}))

	sh.Handle("/admin/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}))

	sh.Handle("/admin/", http.StripPrefix("/admin", gw))

	return v1beta1.Register(ctx, address, sg, gw, deps)
}

func Serve(
	ctx context.Context,
	logger log.Logger,
	cfg Config,
	nrApp newrelic.Application,
	deps api.Deps,
) error {
	httpMux := http.NewServeMux()

	grpcGateway := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(customHeaderMatcherFunc(map[string]bool{cfg.IdentityProxyHeader: true})))

	grpcServer := grpc.NewServer(getGRPCMiddleware(cfg, logger, nrApp))
	reflection.Register(grpcServer)

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.GRPCPort)
	err := registerHandler(ctx, httpMux, grpcGateway, grpcServer, deps, address)
	if err != nil {
		return err
	}

	logger.Info("[shield] api server starting", "http-port", cfg.HTTPPort, "grpc-port", cfg.GRPCPort)

	if err := mux.Serve(
		ctx,
		mux.WithHTTPTarget(fmt.Sprintf(":%d", cfg.HTTPPort), &http.Server{
			Handler:        httpMux,
			ReadTimeout:    120 * time.Second,
			WriteTimeout:   120 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}),
		mux.WithGRPCTarget(fmt.Sprintf(":%d", cfg.GRPCPort), grpcServer),
		mux.WithGracePeriod(5*time.Second),
	); !errors.Is(err, context.Canceled) {
		logger.Error("mux serve error", "err", err)
		return nil
	}

	logger.Info("server stopped gracefully")

	return nil
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
