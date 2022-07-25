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

	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/bootstrap"
	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/rule"
	"github.com/odpf/shield/core/user"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/odpf/shield/api/handler"
	v1 "github.com/odpf/shield/api/handler/v1beta1"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/hook"
	authz_hook "github.com/odpf/shield/hook/authz"
	"github.com/odpf/shield/pkg/sql"
	"github.com/odpf/shield/proxy"
	blobstore "github.com/odpf/shield/store/blob"
	"github.com/odpf/shield/store/postgres"
	"github.com/odpf/shield/store/spicedb"

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

	resourceConfig, err := loadResourceConfig(ctx, logger, appConfig)
	if err != nil {
		return err
	}

	serviceStore := postgres.NewStore(db)
	authzStore, err := spicedb.New(appConfig.SpiceDB, logger)
	if err != nil {
		return err
	}

	deps, err := apiDependencies(ctx, logger, db, appConfig, resourceConfig, serviceStore, authzStore)
	if err != nil {
		return err
	}

	cleanUpFunc, cleanUpProxies, err = startProxy(logger, appConfig, ctx, deps, cleanUpFunc, cleanUpProxies)
	if err != nil {
		return err
	}

	muxServer := startServer(logger, appConfig, err, ctx, deps)

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

func startServer(logger log.Logger, appConfig *config.Shield, err error, ctx context.Context, deps handler.Deps) *server.MuxServer {
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

	handler.Register(ctx, s, gw, deps)

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

func loadResourceConfig(ctx context.Context, logger log.Logger, appConfig *config.Shield) (*blobstore.ResourcesRepository, error) {
	// load resource config
	if appConfig.App.ResourcesConfigPath == "" {
		return nil, errors.New("resource config path cannot be left empty")
	}
	resourceBlobFS, err := (&blobFactory{}).New(ctx, appConfig.App.ResourcesConfigPath, appConfig.App.ResourcesConfigPathSecret)
	if err != nil {
		return nil, err
	}

	resourceRepo := blobstore.NewResourcesRepository(logger, resourceBlobFS)
	if err := resourceRepo.InitCache(ctx, ruleCacheRefreshDelay); err != nil {
		return nil, err
	}

	return resourceRepo, nil
}

func startProxy(logger log.Logger, appConfig *config.Shield, ctx context.Context, deps handler.Deps, cleanUpFunc []func() error, cleanUpProxies []func(ctx context.Context) error) ([]func() error, []func(ctx context.Context) error, error) {
	for _, service := range appConfig.Proxy.Services {
		h2cProxy := proxy.NewH2c(proxy.NewH2cRoundTripper(logger, buildHookPipeline(logger, deps)), proxy.NewDirector())

		// load rules sets
		if service.RulesPath == "" {
			return nil, nil, errors.New("ruleset field cannot be left empty")
		}
		blobFS, err := (&blobFactory{}).New(ctx, service.RulesPath, service.RulesPathSecret)
		if err != nil {
			return nil, nil, err
		}

		ruleRepo := blobstore.NewRuleRepository(logger, blobFS)
		if err := ruleRepo.InitCache(ctx, ruleCacheRefreshDelay); err != nil {
			return nil, nil, err
		}

		ruleService := rule.NewService(ruleRepo)
		deps.V1beta1.RuleService = ruleService

		cleanUpFunc = append(cleanUpFunc, ruleRepo.Close)
		middlewarePipeline := buildMiddlewarePipeline(logger, h2cProxy, appConfig.App.IdentityProxyHeader, deps)
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

func buildHookPipeline(log log.Logger, deps handler.Deps) hook.Service {
	rootHook := hook.New()
	return authz_hook.New(log, rootHook, rootHook, deps)
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

func apiDependencies(
	ctx context.Context,
	logger log.Logger,
	db *sql.SQL,
	appConfig *config.Shield,
	blobStore *blobstore.ResourcesRepository,
	serviceStore postgres.Store,
	authzStore *spicedb.SpiceDB) (handler.Deps, error) {
	actionService := action.NewService(serviceStore)
	namespaceService := namespace.NewService(serviceStore)
	userService := user.NewService(appConfig.App.IdentityProxyHeader, serviceStore)
	relationService := relation.NewService(serviceStore, authzStore)
	groupService := group.NewService(serviceStore, relationService, userService)
	organizationService := organization.NewService(serviceStore, relationService, userService)
	projectService := project.NewService(serviceStore, relationService, userService)
	policyService := policy.NewService(serviceStore, authzStore)
	roleService := role.NewService(serviceStore)
	bootstrapService := bootstrap.NewService(logger, policyService, actionService, namespaceService, roleService, blobStore)
	resourceService := resource.NewService(serviceStore, authzStore, blobStore, relationService, userService)

	bootstrapService.BootstrapDefaultDefinitions(ctx)
	err := bootstrapService.BootstrapResources(ctx)
	if err != nil {
		return handler.Deps{}, err
	}

	dependencies := handler.Deps{
		V1beta1: v1.Dep{
			OrgService:          organizationService,
			UserService:         userService,
			ProjectService:      projectService,
			GroupService:        groupService,
			RelationService:     relationService,
			ResourceService:     resourceService,
			RoleService:         roleService,
			PolicyService:       policyService,
			ActionService:       actionService,
			NamespaceService:    namespaceService,
			IdentityProxyHeader: appConfig.App.IdentityProxyHeader,
		},
	}
	return dependencies, nil
}
