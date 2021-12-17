package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/odpf/shield/internal/bootstrap"
	"github.com/odpf/shield/internal/group"

	"github.com/odpf/shield/internal/relation"
	"github.com/odpf/shield/internal/resource"

	"github.com/odpf/shield/api/handler"
	v1 "github.com/odpf/shield/api/handler/v1beta1"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/hook"
	authz_hook "github.com/odpf/shield/hook/authz"
	"github.com/odpf/shield/internal/authz"
	"github.com/odpf/shield/internal/org"
	"github.com/odpf/shield/internal/project"
	"github.com/odpf/shield/internal/roles"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/internal/user"
	"github.com/odpf/shield/pkg/sql"
	"github.com/odpf/shield/proxy"
	blobstore "github.com/odpf/shield/store/blob"
	"github.com/odpf/shield/store/postgres"

	"github.com/odpf/salt/log"
	"github.com/odpf/salt/server"
	"github.com/pkg/errors"
	"github.com/pkg/profile"
	cli "github.com/spf13/cobra"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	proxyTermChan         = make(chan os.Signal, 1)
	ruleCacheRefreshDelay = time.Minute * 2
)

func serveCommand(logger log.Logger, appConfig *config.Shield) *cli.Command {
	c := &cli.Command{
		Use:     "serve",
		Short:   "Start server and proxy default on port 8080",
		Example: "shield serve",
		RunE: func(cmd *cli.Command, args []string) error {
			return serve(logger, appConfig)
		},
	}
	return c
}

func serve(logger log.Logger, appConfig *config.Shield) error {
	if profiling := os.Getenv("SHIELD_PROFILE"); profiling == "true" || profiling == "1" {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	}

	// @TODO: need to inject custom logger wrapper over zap into ctx to use it internally
	ctx, cancelFunc := context.WithCancel(server.HandleSignals(context.Background()))
	defer cancelFunc()

	db, dbShutdown := setupDB(appConfig.DB, logger)
	defer dbShutdown()

	var cleanUpFunc []func() error
	var cleanUpProxies []func(ctx context.Context) error
	cleanUpFunc, cleanUpProxies, err := startProxy(logger, appConfig, ctx, cleanUpFunc, cleanUpProxies)
	if err != nil {
		return err
	}

	muxServer := startServer(logger, appConfig, err, ctx, db)

	waitForTermSignal(ctx)
	cleanup(logger, ctx, cleanUpFunc, cleanUpProxies, muxServer)

	return nil
}

func cleanup(logger log.Logger, ctx context.Context, cleanUpFunc []func() error, cleanUpProxies []func(ctx context.Context) error, s *server.MuxServer) {
	for _, f := range cleanUpFunc {
		if err := f(); err != nil {
			logger.Warn("error occurred during shutdown", "err", err)
		}
	}
	for _, f := range cleanUpProxies {
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Second*20)
		if err := f(shutdownCtx); err != nil {
			shutdownCancel()
			logger.Warn("error occurred during shutdown", "err", err)
			continue
		}
		shutdownCancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer shutdownCancel()

	s.Shutdown(shutdownCtx)
}

func startServer(logger log.Logger, appConfig *config.Shield, err error, ctx context.Context, db *sql.SQL) *server.MuxServer {
	s, err := server.NewMux(server.Config{
		Port: appConfig.App.Port,
	}, server.WithMuxGRPCServerOptions(getGRPCMiddleware(appConfig, logger)))
	if err != nil {
		panic(err)
	}

	gw, err := server.NewGateway("", appConfig.App.Port, server.WithGatewayMuxOptions(
		runtime.WithIncomingHeaderMatcher(customHeaderMatcherFunc(map[string]bool{appConfig.App.IdentityProxyHeader: true}))),
	)
	if err != nil {
		panic(err)
	}

	handler.Register(ctx, s, gw, apiDependencies(ctx, db, appConfig, logger))

	go s.Serve()

	logger.Info("[shield] api is up", "port", appConfig.App.Port)

	// we'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	signal.Notify(proxyTermChan, os.Interrupt, os.Kill, syscall.SIGTERM)
	return s
}

func customHeaderMatcherFunc(headerKeys map[string]bool) func(key string) (string, bool) {
	return func(key string) (string, bool) {
		if _, ok := headerKeys[key]; ok {
			return key, true
		}
		return runtime.DefaultHeaderMatcher(key)
	}
}

func startProxy(logger log.Logger, appConfig *config.Shield, ctx context.Context, cleanUpFunc []func() error, cleanUpProxies []func(ctx context.Context) error) ([]func() error, []func(ctx context.Context) error, error) {
	for _, service := range appConfig.Proxy.Services {
		if service.RulesPath == "" {
			return nil, nil, errors.New("ruleset field cannot be left empty")
		}
		blobFS, err := (&blobFactory{}).New(ctx, service.RulesPath, service.RulesPathSecret)
		if err != nil {
			return nil, nil, err
		}

		h2cProxy := proxy.NewH2c(proxy.NewH2cRoundTripper(logger, buildHookPipeline(logger)), proxy.NewDirector())

		ruleRepo := blobstore.NewRuleRepository(logger, blobFS)
		if err := ruleRepo.InitCache(ctx, ruleCacheRefreshDelay); err != nil {
			return nil, nil, err
		}
		cleanUpFunc = append(cleanUpFunc, ruleRepo.Close)
		middlewarePipeline := buildMiddlewarePipeline(logger, h2cProxy, ruleRepo, appConfig.App.IdentityProxyHeader)
		go func(thisService config.Service, handler http.Handler) {
			proxyURL := fmt.Sprintf("%s:%d", thisService.Host, thisService.Port)
			logger.Info("starting h2c proxy", "url", proxyURL)

			mux := http.NewServeMux()
			mux.Handle("/ping", healthCheck())
			mux.Handle("/", handler)

			//create a tcp listener
			proxyListener, err := net.Listen("tcp", proxyURL)
			if err != nil {
				logger.Fatal("failed to listen", "err", err)
			}

			proxySrv := http.Server{
				Addr:    proxyURL,
				Handler: h2c.NewHandler(mux, &http2.Server{}),
			}
			if err := proxySrv.Serve(proxyListener); err != nil && err != http.ErrServerClosed {
				logger.Fatal("failed to serve", "err", err)
			}
			cleanUpProxies = append(cleanUpProxies, proxySrv.Shutdown)
		}(service, middlewarePipeline)
	}
	time.Sleep(100 * time.Millisecond)
	logger.Info("[shield] proxy is up")
	return cleanUpFunc, cleanUpProxies, nil
}

func buildHookPipeline(log log.Logger) hook.Service {
	rootHook := hook.New()
	return authz_hook.New(log, rootHook, rootHook)
}

func waitForTermSignal(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("process: ctx done bye\n")
			return
		case <-proxyTermChan:
			fmt.Printf("process: kill signal received. bye \n")
			return
		}
	}
}

func healthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "pong")
	}
}

func apiDependencies(ctx context.Context, db *sql.SQL, appConfig *config.Shield, logger log.Logger) handler.Deps {
	serviceStore := postgres.NewStore(db)
	authzService := authz.New(appConfig, logger)

	schemaService := schema.Service{
		Store: serviceStore,
		Authz: authzService,
	}

	roleService := roles.Service{
		Store: serviceStore,
	}

	bootstrapService := bootstrap.Service{
		SchemaService: schemaService,
		RoleService:   roleService,
		Logger:        logger,
	}

	bootstrapService.BootstrapDefinitions(ctx)

	dependencies := handler.Deps{
		V1beta1: v1.Dep{
			OrgService: org.Service{
				Store: serviceStore,
			},
			UserService: user.Service{
				Store: serviceStore,
			},
			ProjectService: project.Service{
				Store: serviceStore,
			},
			GroupService: group.Service{
				Store: serviceStore,
			},
			RelationService: relation.Service{
				Store: serviceStore,
				Authz: authzService,
			},
			ResourceService: resource.Service{
				Store: serviceStore,
			},
			RoleService:         roleService,
			PolicyService:       schemaService,
			ActionService:       schemaService,
			NamespaceService:    schemaService,
			IdentityProxyHeader: appConfig.App.IdentityProxyHeader,
		},
	}
	return dependencies
}
