package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/odpf/shield/pkg/server/consts"
	"github.com/odpf/shield/pkg/server/health"

	"github.com/odpf/shield/pkg/server/interceptors"

	"github.com/gorilla/securecookie"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	newrelic "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrgrpc"
	"github.com/odpf/salt/log"
	"github.com/odpf/salt/mux"
	"github.com/odpf/salt/spa"
	"github.com/odpf/shield/internal/api"
	"github.com/odpf/shield/internal/api/v1beta1"
	"github.com/odpf/shield/pkg/telemetry"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"github.com/odpf/shield/ui"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	grpcDialTimeout = 5 * time.Second
)

func Serve(
	ctx context.Context,
	logger log.Logger,
	cfg Config,
	nrApp newrelic.Application,
	deps api.Deps,
) error {
	httpMux := http.NewServeMux()

	grpcDialCtx, grpcDialCancel := context.WithTimeout(ctx, grpcDialTimeout)
	defer grpcDialCancel()
	grpcConn, err := grpc.DialContext(
		grpcDialCtx,
		cfg.grpcAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(cfg.GRPC.MaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(cfg.GRPC.MaxSendMsgSize),
		))
	if err != nil {
		return err
	}

	var sessionCookieCutter securecookie.Codec
	if len(cfg.Authentication.Session.HashSecretKey) != 32 || len(cfg.Authentication.Session.BlockSecretKey) != 32 {
		// hash and block keys should be 32 bytes long
		logger.Warn("session management disabled", errors.New("authentication.session keys should be 32 chars long"))
	} else {
		sessionCookieCutter = securecookie.New(
			[]byte(cfg.Authentication.Session.HashSecretKey),
			[]byte(cfg.Authentication.Session.BlockSecretKey),
		)
	}
	sessionMiddleware := interceptors.NewSession(sessionCookieCutter)

	var grpcGatewayServerInterceptors []runtime.ServeMuxOption
	grpcGatewayServerInterceptors = append(grpcGatewayServerInterceptors,
		runtime.WithHealthEndpointAt(grpc_health_v1.NewHealthClient(grpcConn), "/ping"),
		runtime.WithIncomingHeaderMatcher(
			interceptors.GatewayHeaderMatcherFunc(
				map[string]bool{
					strings.ToLower(cfg.IdentityProxyHeader): true,
					consts.UserTokenRequestKey:               true,
					"cookie":                                 true,
					"authorization":                          true,
				},
			),
		),
		runtime.WithForwardResponseOption(sessionMiddleware.GatewayResponseModifier),
	)
	grpcGateway := runtime.NewServeMux(grpcGatewayServerInterceptors...)

	var rootHandler http.Handler = grpcGateway
	if cfg.CorsOrigin != "" {
		rootHandler = interceptors.WithCors(rootHandler, cfg.CorsOrigin)
	}

	httpMux.Handle("/", rootHandler)
	if err := shieldv1beta1.RegisterAdminServiceHandler(ctx, grpcGateway, grpcConn); err != nil {
		return err
	}
	if err := shieldv1beta1.RegisterShieldServiceHandler(ctx, grpcGateway, grpcConn); err != nil {
		return err
	}

	// json web key set handler
	if jwksHandler, err := NewTokenJWKSHandler(cfg.Authentication.Token.RSAPath); err != nil {
		return err
	} else {
		httpMux.Handle(fmt.Sprintf("/%s", consts.JWKSHandlerPath), jwksHandler)
	}

	spaHandler, err := spa.Handler(ui.Assets, "dist/ui", "index.html", false)
	if err != nil {
		logger.Warn("failed to load spa", "err", err)
	} else {
		httpMux.Handle("/console/", http.StripPrefix("/console/", spaHandler))
	}

	grpcMiddlewares := getGRPCMiddleware(logger, cfg.IdentityProxyHeader, nrApp, sessionMiddleware)
	grpcServer := grpc.NewServer(grpcMiddlewares)
	reflection.Register(grpcServer)
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewHandler())

	if err = v1beta1.Register(grpcServer, deps); err != nil {
		return err
	}

	pe, err := telemetry.SetupOpenCensus(ctx, cfg.TelemetryConfig)
	if err != nil {
		logger.Error("failed to setup OpenCensus", "err", err)
	}
	httpMuxMetrics := http.NewServeMux()
	httpMuxMetrics.Handle("/metrics", pe)

	logger.Info("api server starting", "http-port", cfg.Port, "grpc-port", cfg.GRPC.Port, "metrics-port", cfg.MetricsPort)
	if err := mux.Serve(
		ctx,
		mux.WithHTTPTarget(fmt.Sprintf(":%d", cfg.Port), &http.Server{
			Handler:        httpMux,
			ReadTimeout:    120 * time.Second,
			WriteTimeout:   120 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}),
		mux.WithHTTPTarget(fmt.Sprintf(":%d", cfg.MetricsPort), &http.Server{
			Handler:        httpMuxMetrics,
			ReadTimeout:    120 * time.Second,
			WriteTimeout:   120 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}),
		mux.WithGRPCTarget(fmt.Sprintf(":%d", cfg.GRPC.Port), grpcServer),
		mux.WithGracePeriod(5*time.Second),
	); !errors.Is(err, context.Canceled) {
		logger.Error("mux serve error", "err", err)
		return nil
	}

	logger.Info("server stopped gracefully")
	return nil
}

// REVISIT: passing config.Shield as reference
func getGRPCMiddleware(logger log.Logger, identityProxyHeader string, nrApp newrelic.Application,
	sessionMiddleware *interceptors.Session) grpc.ServerOption {
	recoveryFunc := func(p interface{}) (err error) {
		fmt.Println(p)
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
			grpc_recovery.UnaryServerInterceptor(grpcRecoveryOpts...),
			nrgrpc.UnaryServerInterceptor(nrApp),
			interceptors.EnrichCtxWithIdentity(identityProxyHeader),
			grpc_zap.UnaryServerInterceptor(grpcZapLogger.Desugar()),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_validator.UnaryServerInterceptor(),
			sessionMiddleware.UnaryGRPCRequestHeadersAnnotator(),
			interceptors.UnaryAuthenticationCheck(identityProxyHeader),
			interceptors.UnaryAuthorizationCheck(identityProxyHeader),
		),
	)
}
