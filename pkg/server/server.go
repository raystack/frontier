package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	frontierlogger "github.com/raystack/frontier/pkg/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/raystack/salt/spa"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/raystack/frontier/pkg/server/consts"
	"github.com/raystack/frontier/pkg/server/health"
	"github.com/raystack/frontier/ui"

	connectinterceptors "github.com/raystack/frontier/pkg/server/connect_interceptors"
	"github.com/raystack/frontier/pkg/server/interceptors"

	"connectrpc.com/connect"
	connecthealth "connectrpc.com/grpchealth"
	"github.com/gorilla/securecookie"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	newrelic "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrgrpc"
	"github.com/raystack/frontier/internal/api"
	"github.com/raystack/frontier/internal/api/v1beta1"
	"github.com/raystack/frontier/internal/api/v1beta1connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	frontierv1beta1connect "github.com/raystack/frontier/proto/v1beta1/frontierv1beta1connect"
	"github.com/raystack/salt/log"
	"github.com/raystack/salt/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	grpcDialTimeout = 5 * time.Second

	// keeping it in sync with https://github.com/raystack/salt/blob/v0.3.8/mux/mux.go#L15
	// which is being used in GRPC server shutdown
	connectServerGracePeriod = 10 * time.Second
)

type UIConfigApiResponse struct {
	Title             string   `json:"title"`
	Logo              string   `json:"logo"`
	AppUrl            string   `json:"app_url"`
	TokenProductId    string   `json:"token_product_id"`
	OrganizationTypes []string `json:"organization_types"`
}

func ServeUI(ctx context.Context, logger log.Logger, uiConfig UIConfig, apiServerConfig Config) {
	isUIPortNotExits := uiConfig.Port == 0
	if isUIPortNotExits {
		logger.Warn("ui server disabled: no port specified")
		return
	}

	spaHandler, err := spa.Handler(ui.Assets, "dist/ui", "index.html", false)
	if err != nil {
		logger.Warn("failed to load ui", "err", err)
		return
	} else {
		restRemoteHost := fmt.Sprintf("http://%s:%d", apiServerConfig.Host, apiServerConfig.Port)
		restRemote, err := url.Parse(restRemoteHost)
		if err != nil {
			logger.Error("ui server failed: unable to parse api server host")
			return
		}

		connectRemoteHost := fmt.Sprintf("http://%s:%d", apiServerConfig.Host, apiServerConfig.Connect.Port)
		connectRemote, err := url.Parse(connectRemoteHost)
		if err != nil {
			logger.Error("ui server failed: unable to parse connect server host")
			return
		}

		restProxyHandler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				r.URL.Path = strings.Replace(r.URL.Path, "/frontier-api", "", -1)
				r.Host = restRemoteHost
				p.ServeHTTP(w, r)
			}
		}

		connectProxyHandler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				r.URL.Path = strings.Replace(r.URL.Path, "/frontier-connect", "", -1)
				r.Host = connectRemoteHost
				p.ServeHTTP(w, r)
			}
		}

		restProxy := httputil.NewSingleHostReverseProxy(restRemote)
		connectProxy := httputil.NewSingleHostReverseProxy(connectRemote)

		http.HandleFunc("/configs", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			confResp := UIConfigApiResponse{
				Title:             uiConfig.Title,
				Logo:              uiConfig.Logo,
				AppUrl:            uiConfig.AppURL,
				TokenProductId:    uiConfig.TokenProductId,
				OrganizationTypes: uiConfig.OrganizationTypes,
			}
			json.NewEncoder(w).Encode(confResp)
		})

		http.HandleFunc("/frontier-api/", restProxyHandler(restProxy))
		http.HandleFunc("/frontier-connect/", connectProxyHandler(connectProxy))
		http.Handle("/", http.StripPrefix("/", spaHandler))
	}

	logger.Info("ui server starting", "http-port", uiConfig.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", uiConfig.Port), nil); err != nil {
		logger.Error("ui server failed", "err", err)
	}
}

func ServeConnect(ctx context.Context, logger log.Logger, cfg Config, deps api.Deps) error {
	// Create the server handler with both services
	frontierService := v1beta1connect.NewConnectHandler(deps, cfg.Authentication)

	sessionCookieCutter := getSessionCookieCutter(cfg.Authentication.Session.BlockSecretKey, cfg.Authentication.Session.HashSecretKey, logger)
	sessionMiddleware := connectinterceptors.NewSession(sessionCookieCutter, cfg.Authentication.Session)

	// grpcZapLogger := zap.Must(zap.NewProduction())
	grpcZapLogger := zap.NewExample().Sugar()
	loggerZap, ok := logger.(*log.Zap)
	if ok {
		grpcZapLogger = loggerZap.GetInternalZapLogger()
	}
	loggerOpts := connectinterceptors.NewLoggerOptions(connectinterceptors.LoggerOption{
		Decider: func(procedure string) bool {
			return procedure != "/grpc.health.v1.Health/Check"
		},
	})

	interceptors := connect.WithInterceptors(
		connectinterceptors.UnaryConnectLoggerInterceptor(grpcZapLogger.Desugar(), loggerOpts),
		sessionMiddleware.UnaryConnectRequestHeadersAnnotator(),
		connectinterceptors.UnaryAuthenticationCheck(frontierService),
		connectinterceptors.UnaryAuthorizationCheck(frontierService),
		sessionMiddleware.UnaryConnectResponseInterceptor())

	// Initialize connect handlers
	frontierPath, frontierHandler := frontierv1beta1connect.NewFrontierServiceHandler(frontierService, interceptors)
	adminPath, adminHandler := frontierv1beta1connect.NewAdminServiceHandler(frontierService, interceptors)

	// Create mux and register handlers
	mux := http.NewServeMux()
	mux.Handle(frontierPath, frontierHandler)
	mux.Handle(adminPath, adminHandler)

	// configure healthcheck
	// curl --header "Content-Type: application/json" \
	// --data '{"service":"raystack.frontier.v1beta1.AdminService"}' \
	// http://localhost:8002/grpc.health.v1.Health/Check
	checker := connecthealth.NewStaticChecker(
		"raystack.frontier.v1beta1.FrontierService",
		"raystack.frontier.v1beta1.AdminService",
	)

	mux.Handle(connecthealth.NewHandler(checker))

	// Configure and create the server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Connect.Port),
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	logger.Info("connect server starting", "port", cfg.Connect.Port)

	go func() {
		<-ctx.Done()

		ctxShutdown, cancel := context.WithTimeout(context.Background(), connectServerGracePeriod)
		defer cancel()

		if err := server.Shutdown(ctxShutdown); err != nil {
			logger.Fatal("HTTP shutdown error: %v", err)
		}

		logger.Info("Graceful shutdown of connect server complete")
	}()

	// Start server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("connect server failed: %w", err)
	}

	return nil
}

func Serve(
	ctx context.Context,
	logger log.Logger,
	cfg Config,
	nrApp newrelic.Application,
	deps api.Deps,
	promRegistry *prometheus.Registry,
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

	sessionCookieCutter := getSessionCookieCutter(cfg.Authentication.Session.BlockSecretKey, cfg.Authentication.Session.HashSecretKey, logger)
	sessionMiddleware := interceptors.NewSession(sessionCookieCutter, cfg.Authentication.Session)

	defaultMimeMarshaler := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames:   true,
			EmitUnpopulated: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}
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
					consts.ProjectRequestKey:                 true,
					consts.StripeTestClockRequestKey:         true,
					consts.StripeWebhookSignature:            true,
					consts.RequestIDHeader:                   true,
				},
			),
		),
		runtime.WithForwardResponseOption(sessionMiddleware.GatewayResponseModifier),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: defaultMimeMarshaler,
		}),
		runtime.WithMarshalerOption(interceptors.RawBytesMIME, &interceptors.RawJSONPb{
			JSONPb: defaultMimeMarshaler,
		}),
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
	)
	grpcGateway := runtime.NewServeMux(grpcGatewayServerInterceptors...)
	var rootHandler http.Handler = grpcGateway
	if len(cfg.Cors.AllowedOrigins) > 0 {
		rootHandler = interceptors.WithCors(rootHandler, cfg.Cors)
	}
	// add custom mimetype to use byte serializer for few endpoints
	rootHandler = interceptors.ByteMimeWrapper(rootHandler)

	httpMux.Handle("/", rootHandler)

	if err := frontierv1beta1.RegisterAdminServiceHandler(ctx, grpcGateway, grpcConn); err != nil {
		return err
	}
	if err := frontierv1beta1.RegisterFrontierServiceHandler(ctx, grpcGateway, grpcConn); err != nil {
		return err
	}

	// setup metrics
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120, 180, 240, 300}),
		),
	)

	grpcMiddleware := getGRPCMiddleware(logger, cfg.IdentityProxyHeader, nrApp, sessionMiddleware, srvMetrics, deps)
	grpcServerOpts := []grpc.ServerOption{
		grpcMiddleware,
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	}
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

	v1beta1.Register(grpcServer, deps, cfg.Authentication)

	var metricsOps = []mux.Option{
		mux.WithHTTPTarget(fmt.Sprintf(":%d", cfg.Port), &http.Server{
			Handler:        httpMux,
			ReadTimeout:    120 * time.Second,
			WriteTimeout:   120 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}),
		mux.WithGRPCTarget(fmt.Sprintf(":%d", cfg.GRPC.Port), grpcServer),
		mux.WithGracePeriod(10 * time.Second),
	}
	if cfg.MetricsPort > 0 {
		srvMetrics.InitializeMetrics(grpcServer)
		httpMuxMetrics := http.NewServeMux()
		httpMuxMetrics.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		}))
		if cfg.Profiler {
			// add debug handlers
			httpMuxMetrics.Handle("/debug/pprof/", http.DefaultServeMux)
		}
		promRegistry.MustRegister(srvMetrics)
		metricsOps = append(metricsOps, mux.WithHTTPTarget(fmt.Sprintf(":%d", cfg.MetricsPort), &http.Server{
			Handler:        httpMuxMetrics,
			ReadTimeout:    120 * time.Second,
			WriteTimeout:   120 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}))
	}

	logger.Info("api server starting", "http-port", cfg.Port, "connect-port", cfg.Connect.Port, "grpc-port", cfg.GRPC.Port, "metrics-port", cfg.MetricsPort)
	if err := mux.Serve(
		ctx,
		metricsOps...,
	); !errors.Is(err, context.Canceled) {
		logger.Error("mux serve error", "err", err)
		return nil
	}

	logger.Info("stopping server gracefully")
	return nil
}

func getSessionCookieCutter(blockSecretKey string, hashSecretKey string, logger log.Logger) securecookie.Codec {
	var sessionCookieCutter securecookie.Codec
	if len(hashSecretKey) != 32 || len(blockSecretKey) != 32 {
		// hash and block keys should be 32 bytes long
		logger.Warn("session management disabled", errors.New("authentication.session keys should be 32 chars long"))
	} else {
		sessionCookieCutter = securecookie.New(
			[]byte(hashSecretKey),
			[]byte(blockSecretKey),
		)
	}
	return sessionCookieCutter
}

// REVISIT: passing config.Frontier as reference
func getGRPCMiddleware(logger log.Logger, identityProxyHeader string, nrApp newrelic.Application,
	sessionMiddleware *interceptors.Session, srvMetrics *grpcprom.ServerMetrics, deps api.Deps,
) grpc.ServerOption {
	grpcZapLogger := zap.NewExample().Sugar()
	loggerZap, ok := logger.(*log.Zap)
	if ok {
		grpcZapLogger = loggerZap.GetInternalZapLogger()
	}
	recoveryFunc := func(p interface{}) (err error) {
		grpcZapLogger.Error(p)
		return status.Errorf(codes.Internal, "internal server error")
	}

	grpcRecoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recoveryFunc),
	}

	return grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(grpcRecoveryOpts...),
			grpc_zap.UnaryServerInterceptor(grpcZapLogger.Desugar(),
				grpc_zap.WithDecider(func(fullMethodName string, err error) bool {
					// don't log health check
					if fullMethodName == "/grpc.health.v1.Health/Check" {
						return false
					}
					return true
				}),
				grpc_zap.WithMessageProducer(frontierlogger.RequestLogFunc),
			),
			interceptors.UnaryRequestIDLoggerEnrich(),
			srvMetrics.UnaryServerInterceptor(),
			nrgrpc.UnaryServerInterceptor(nrApp),
			interceptors.EnrichCtxWithPassthroughEmail(identityProxyHeader),
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_validator.UnaryServerInterceptor(),
			sessionMiddleware.UnaryGRPCRequestHeadersAnnotator(),
			interceptors.UnaryErrorHandler(),
			interceptors.UnaryAuthenticationCheck(),
			interceptors.UnaryAPIRequestEnrich(),
			interceptors.UnaryAuthorizationCheck(),
			interceptors.UnaryCtxWithAudit(deps.AuditService),
		),
	)
}
