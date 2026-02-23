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
	"github.com/gorilla/securecookie"
	"github.com/raystack/frontier/internal/api"
	"github.com/raystack/frontier/internal/api/v1beta1connect"
	frontierv1beta1connect "github.com/raystack/frontier/proto/v1beta1/frontierv1beta1connect"
	"github.com/raystack/salt/log"
	"go.uber.org/zap"
)

const (
	connectServerGracePeriod = 10 * time.Second
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
}

func ServeUI(ctx context.Context, logger log.Logger, uiConfig UIConfig, apiServerConfig Config) {
	isUIPortNotExits := uiConfig.Port == 0
	if isUIPortNotExits {
		logger.Warn("ui server disabled: no port specified")
		return
	}

	spaHandler, err := spa.Handler(admin.Assets, "dist/admin", "index.html", false)
	if err != nil {
		logger.Warn("failed to load ui", "err", err)
		return
	} else {
		connectRemoteHost := fmt.Sprintf("http://%s:%d", apiServerConfig.Host, apiServerConfig.Connect.Port)
		connectRemote, err := url.Parse(connectRemoteHost)
		if err != nil {
			logger.Error("ui server failed: unable to parse connect server host")
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

		http.HandleFunc("/configs", func(w http.ResponseWriter, r *http.Request) {
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
			}
			json.NewEncoder(w).Encode(confResp)
		})

		http.HandleFunc("/frontier-connect/", connectProxyHandler(connectProxy))
		http.Handle("/", http.StripPrefix("/", spaHandler))
	}

	logger.Info("ui server starting", "http-port", uiConfig.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", uiConfig.Port), nil); err != nil {
		logger.Error("ui server failed", "err", err)
	}
}

func ServeConnect(ctx context.Context, logger log.Logger, cfg Config, deps api.Deps, promRegistry *prometheus.Registry) error {
	frontierService := v1beta1connect.NewConnectHandler(deps, cfg.Authentication)

	sessionCookieCutter := getSessionCookieCutter(cfg.Authentication.Session.BlockSecretKey, cfg.Authentication.Session.HashSecretKey, logger)
	zapLogger := zap.NewExample().Sugar()
	loggerZap, ok := logger.(*log.Zap)
	if ok {
		zapLogger = loggerZap.GetInternalZapLogger()
	}
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
		logger.Fatal(err.Error())
	}

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(promExporter))
	otelInterceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithMeterProvider(provider),
		otelconnect.WithoutTracing(),
		otelconnect.WithoutServerPeerAttributes(),
	)
	if err != nil {
		logger.Fatal("OTEL ConnectRPC interceptor init error: %v", err)
		return err
	}

	authNInterceptor := connectinterceptors.NewAuthenticationInterceptor(frontierService, cfg.Authentication.Session.Headers)
	authZInterceptor := connectinterceptors.NewAuthorizationInterceptor(frontierService)
	sessionInterceptor := connectinterceptors.NewSessionInterceptor(sessionCookieCutter, cfg.Authentication.Session, frontierService)
	auditInterceptor := connectinterceptors.NewAuditInterceptor(deps.AuditService)

	interceptors := connect.WithInterceptors(
		otelInterceptor,
		connectinterceptors.UnaryConnectLoggerInterceptor(zapLogger.Desugar(), loggerOpts),
		connectinterceptors.UnaryConnectErrorResponseInterceptor(),
		sessionInterceptor,
		authNInterceptor,
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
		json.NewEncoder(w).Encode(map[string]string{"status": "SERVING"})
	})

	// Configure and create the server
	handler := h2c.NewHandler(mux, &http2.Server{})
	handler = connectinterceptors.WithConnectCORS(handler, cfg.ConnectCors)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Connect.Port),
		Handler: handler,
	}

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
		go func() {
			logger.Info("metrics server starting", "port", cfg.MetricsPort)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("metrics server failed", "err", err)
			}
		}()
		go func() {
			<-ctx.Done()
			metricsServer.Shutdown(context.Background())
		}()
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
