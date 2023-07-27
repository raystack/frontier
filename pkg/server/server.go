package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/raystack/frontier/pkg/server/consts"
	"github.com/raystack/frontier/pkg/server/health"

	"github.com/raystack/frontier/pkg/server/interceptors"

	"github.com/gorilla/securecookie"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	newrelic "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrgrpc"
	"github.com/raystack/frontier/internal/api"
	"github.com/raystack/frontier/internal/api/v1beta1"
	"github.com/raystack/frontier/pkg/telemetry"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/ui"
	"github.com/raystack/salt/log"
	"github.com/raystack/salt/mux"
	"github.com/raystack/salt/spa"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

	grpcGatewayClientCreds := insecure.NewCredentials()
	if cfg.GRPC.TLSClientCAFile != "" {
		tlsCreds, err := credentials.NewClientTLSFromFile(cfg.GRPC.TLSClientCAFile, "")
		if err != nil {
			return err
		}
		grpcGatewayClientCreds = tlsCreds
	}
	// initialize grpc gateway client
	grpcConn, err := grpc.DialContext(
		grpcDialCtx,
		cfg.grpcAddr(),
		grpc.WithTransportCredentials(grpcGatewayClientCreds),
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
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)
	grpcGateway := runtime.NewServeMux(grpcGatewayServerInterceptors...)

	var rootHandler http.Handler = grpcGateway
	if cfg.CorsOrigin != "" {
		rootHandler = interceptors.WithCors(rootHandler, cfg.CorsOrigin)
	}

	httpMux.Handle("/", rootHandler)
	if err := frontierv1beta1.RegisterAdminServiceHandler(ctx, grpcGateway, grpcConn); err != nil {
		return err
	}
	if err := frontierv1beta1.RegisterFrontierServiceHandler(ctx, grpcGateway, grpcConn); err != nil {
		return err
	}

	spaHandler, err := spa.Handler(ui.Assets, "dist/ui", "index.html", false)
	if err != nil {
		logger.Warn("failed to load spa", "err", err)
	} else {
		httpMux.Handle("/console/", http.StripPrefix("/console/", spaHandler))
	}

	grpcMiddleware := getGRPCMiddleware(logger, cfg.IdentityProxyHeader, nrApp, sessionMiddleware, deps)
	grpcServerOpts := []grpc.ServerOption{grpcMiddleware}
	if cfg.GRPC.TLSCertFile != "" && cfg.GRPC.TLSKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.GRPC.TLSCertFile, cfg.GRPC.TLSKeyFile)
		if err != nil {
			return err
		}
		grpcServerOpts = append(grpcServerOpts, grpc.Creds(creds))
	}
	grpcServer := grpc.NewServer(grpcServerOpts...)
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

// REVISIT: passing config.Frontier as reference
func getGRPCMiddleware(logger log.Logger, identityProxyHeader string, nrApp newrelic.Application,
	sessionMiddleware *interceptors.Session, deps api.Deps) grpc.ServerOption {
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
			interceptors.EnrichCtxWithPassthroughEmail(identityProxyHeader),
			grpc_zap.UnaryServerInterceptor(grpcZapLogger.Desugar()),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_validator.UnaryServerInterceptor(),
			sessionMiddleware.UnaryGRPCRequestHeadersAnnotator(),
			interceptors.UnaryAuthenticationCheck(),
			interceptors.UnaryAuthorizationCheck(identityProxyHeader),
			interceptors.UnaryCtxWithAudit(deps.AuditService),
		),
	)
}
