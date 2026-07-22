package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/raystack/salt/server/spa"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/raystack/frontier/web/apps/admin"

	connectinterceptors "github.com/raystack/frontier/pkg/server/connect_interceptors"

	"connectrpc.com/connect"
	connecthealth "connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"github.com/gorilla/securecookie"
	"github.com/raystack/frontier/internal/api"
	"github.com/raystack/frontier/internal/api/v1beta1connect"
	frontierv1beta1connect "github.com/raystack/frontier/proto/v1beta1/frontierv1beta1connect"
)

type WebhooksConfigApiResponse struct {
	EnableDelete bool `json:"enable_delete"`
}

type UIConfigApiResponse struct {
	Title             string                    `json:"title"`
	Logo              string                    `json:"logo"`
	AppUrl            string                    `json:"app_url"`
	TokenProductId    string                    `json:"token_product_id"`
	OrganizationTypes []string                  `json:"organization_types"`
	Webhooks          WebhooksConfigApiResponse `json:"webhooks"`
	Terminology       TerminologyConfig         `json:"terminology"`
}

func ServeUI(ctx context.Context, logger *slog.Logger, uiConfig UIConfig, apiServerConfig Config) {
	isUIPortNotExits := uiConfig.Port == 0
	if isUIPortNotExits {
		logger.Warn("ui server disabled: no port specified")
		return
	}

	spaHandler, err := spa.Handler(admin.Assets, "dist/admin", "index.html", false)
	if err != nil {
		logger.Warn("failed to load ui", "err", err)
		return
	}

	connectRemoteHost := fmt.Sprintf("http://%s:%d", apiServerConfig.Host, apiServerConfig.Connect.Port)
	connectRemote, err := url.Parse(connectRemoteHost)
	if err != nil {
		logger.Error("ui server failed: unable to parse connect server host", "err", err)
		return
	}

	connectProxyHandler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = strings.Replace(r.URL.Path, "/frontier-connect", "", -1)
			r.Host = connectRemoteHost
			p.ServeHTTP(w, r)
		}
	}

	connectProxy := httputil.NewSingleHostReverseProxy(connectRemote)

	mux := http.NewServeMux()
	mux.HandleFunc("/configs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		confResp := UIConfigApiResponse{
			Title:             uiConfig.Title,
			Logo:              uiConfig.Logo,
			AppUrl:            uiConfig.AppURL,
			TokenProductId:    uiConfig.TokenProductId,
			OrganizationTypes: uiConfig.OrganizationTypes,
			Webhooks: WebhooksConfigApiResponse{
				EnableDelete: uiConfig.Webhooks.EnableDelete,
			},
			Terminology: uiConfig.Terminology,
		}
		_ = json.NewEncoder(w).Encode(confResp)
	})

	mux.HandleFunc("/frontier-connect/", connectProxyHandler(connectProxy))
	mux.Handle("/", http.StripPrefix("/", spaHandler))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", uiConfig.Port),
		Handler: mux,
	}

	logger.Info("ui server starting", "http-port", uiConfig.Port)
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		logger.Error("ui server failed", "err", err)
		return
	case <-ctx.Done():
	}

	ctxShutdown, cancel := context.WithTimeout(context.Background(), apiServerConfig.ShutdownGracePeriod)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		logger.ErrorContext(ctxShutdown, "ui server shutdown error", "err", err)
		return
	}

	logger.Info("Graceful shutdown of ui server complete")
}

func ServeConnect(ctx context.Context, logger *slog.Logger, cfg Config, deps api.Deps, promRegistry *prometheus.Registry) error {
	frontierService := v1beta1connect.NewConnectHandler(deps, cfg.Authentication)

	sessionCookieCutter := getSessionCookieCutter(cfg.Authentication.Session.BlockSecretKey, cfg.Authentication.Session.HashSecretKey, logger)
	loggerOpts := connectinterceptors.NewLoggerOptions(connectinterceptors.LoggerOption{
		Decider: func(procedure string) bool {
			return procedure != "/grpc.health.v1.Health/Check"
		},
	})

	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	promExporter, err := promexporter.New(promexporter.WithNamespace("connect"),
		promexporter.WithRegisterer(promRegistry),
		promexporter.WithoutScopeInfo())
	if err != nil {
		return fmt.Errorf("prometheus exporter: %w", err)
	}

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(promExporter))
	otelInterceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithMeterProvider(provider),
		otelconnect.WithoutTracing(),
		otelconnect.WithoutServerPeerAttributes(),
	)
	if err != nil {
		return fmt.Errorf("OTEL ConnectRPC interceptor: %w", err)
	}

	authNInterceptor := connectinterceptors.NewAuthenticationInterceptor(frontierService, cfg.Authentication.Session.Headers)
	authZInterceptor := connectinterceptors.NewAuthorizationInterceptor(frontierService)
	sessionInterceptor := connectinterceptors.NewSessionInterceptor(sessionCookieCutter, cfg.Authentication.Session, frontierService, cfg.PAT)
	auditInterceptor := connectinterceptors.NewAuditInterceptor(deps.AuditService)

	validateInterceptor := validate.NewInterceptor()

	interceptors := connect.WithInterceptors(
		otelInterceptor,
		connectinterceptors.UnaryConnectErrorSanitizerInterceptor(),
		connectinterceptors.UnaryConnectLoggerInterceptor(logger, loggerOpts),
		connectinterceptors.UnaryConnectErrorResponseInterceptor(),
		sessionInterceptor,
		authNInterceptor,
		validateInterceptor,
		authZInterceptor,
		auditInterceptor,
		sessionInterceptor.UnaryConnectResponseInterceptor())

	frontierPath, frontierHandler := frontierv1beta1connect.NewFrontierServiceHandler(frontierService, interceptors, connect.WithCodec(connectCodec{}))
	adminPath, adminHandler := frontierv1beta1connect.NewAdminServiceHandler(frontierService, interceptors, connect.WithCodec(connectCodec{}))

	// Create mux and register handlers
	mux := http.NewServeMux()
	mux.Handle(frontierPath, frontierHandler)
	mux.Handle(adminPath, adminHandler)

	// Register webhook bridge handler to allow Stripe to call with provider in path
	// This uses frontierHandler which has all interceptors (auth, logging, audit, etc.) applied
	mux.HandleFunc("/billing/webhooks/callback/", WebhookBridgeHandler(frontierHandler))
	reflector := grpcreflect.NewStaticReflector(
		"raystack.frontier.v1beta1.FrontierService",
		"raystack.frontier.v1beta1.AdminService") // protoc-gen-connect-go generates package-level constants
	// for these fully-qualified protobuf service names, such as
	// frontierv1beta1.FrontierServiceName and frontierv1beta1.AdminServiceName

	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	// Many tools still expect the older version of the server reflection API, so
	// most servers should mount both handlers.
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// configure healthcheck
	// curl --header "Content-Type: application/json" \
	// --data '{"service":"raystack.frontier.v1beta1.AdminService"}' \
	// http://localhost:8002/grpc.health.v1.Health/Check
	checker := connecthealth.NewStaticChecker(
		"raystack.frontier.v1beta1.FrontierService",
		"raystack.frontier.v1beta1.AdminService",
	)

	mux.Handle(connecthealth.NewHandler(checker))

	// simple ping endpoint for liveness checks
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "SERVING"})
	})

	// Configure and create the server
	h2s := &http2.Server{}
	handler := h2c.NewHandler(mux, h2s)
	handler = connectinterceptors.WithConnectCORS(handler, cfg.ConnectCors)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Connect.Port),
		Handler: handler,
	}

	// Shutdown cannot reach h2c connections on its own: h2c hijacks them out
	// of the server's connection tracking. This hook makes Shutdown send
	// GOAWAY on them so clients finish up and reconnect elsewhere
	// (golang/go#26682).
	if err := http2.ConfigureServer(server, h2s); err != nil {
		return fmt.Errorf("configure http2 server: %w", err)
	}

	// counts shutdown goroutines still draining their servers
	var shutdownWG sync.WaitGroup

	// start dedicated metrics server if configured
	if cfg.MetricsPort > 0 {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		}))
		if cfg.Profiler {
			metricsMux.Handle("/debug/pprof/", http.DefaultServeMux)
		}
		metricsServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
			Handler: metricsMux,
		}
		metricsFailed := make(chan struct{})
		go func() {
			logger.Info("metrics server starting", "port", cfg.MetricsPort)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("metrics server failed", "err", err)
				close(metricsFailed)
			}
		}()
		shutdownWG.Add(1)
		go gracefulShutdown(ctx, logger, &shutdownWG, metricsServer, "metrics server", cfg.ShutdownGracePeriod, metricsFailed)
	}

	logger.Info("connect server starting", "port", cfg.Connect.Port)

	serveFailed := make(chan struct{})
	shutdownWG.Add(1)
	go gracefulShutdown(ctx, logger, &shutdownWG, server, "connect server", cfg.ShutdownGracePeriod, serveFailed)

	// Start server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		close(serveFailed)
		// No shutdownWG.Wait here: the metrics watcher only unblocks when ctx
		// is cancelled, which never happens on this path; it dies with the
		// process right after.
		return fmt.Errorf("connect server failed: %w", err)
	}

	// Wait for Shutdown to finish draining HTTP/1.1 requests before returning,
	// so callers don't tear down the database under them. Hijacked h2c
	// connections are not tracked by Shutdown and are not waited on; they get
	// GOAWAY through the http2.ConfigureServer hook instead.
	shutdownWG.Wait()
	return nil
}

// gracefulShutdown drains srv within the grace period once ctx is cancelled.
// It returns without logging when the server already failed, so a server
// that never started is not reported as gracefully shut down.
func gracefulShutdown(ctx context.Context, logger *slog.Logger, wg *sync.WaitGroup, srv *http.Server, name string, gracePeriod time.Duration, failed <-chan struct{}) {
	defer wg.Done()

	select {
	case <-ctx.Done():
	case <-failed:
		return
	}

	ctxShutdown, cancel := context.WithTimeout(context.Background(), gracePeriod)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.ErrorContext(ctxShutdown, name+" shutdown error", "err", err)
		return
	}

	logger.Info("Graceful shutdown of " + name + " complete")
}

func getSessionCookieCutter(blockSecretKey string, hashSecretKey string, logger *slog.Logger) securecookie.Codec {
	var sessionCookieCutter securecookie.Codec
	if len(hashSecretKey) != 32 || len(blockSecretKey) != 32 {
		// hash and block keys should be 32 bytes long
		logger.Warn("session management disabled", "reason", "authentication.session keys should be 32 chars long")
	} else {
		sessionCookieCutter = securecookie.New(
			[]byte(hashSecretKey),
			[]byte(blockSecretKey),
		)
	}
	return sessionCookieCutter
}
