package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	newrelic "github.com/newrelic/go-agent"
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
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/api"
	"github.com/odpf/shield/internal/server"
	"github.com/odpf/shield/internal/store/blob"

	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/store/postgres"
	"github.com/odpf/shield/internal/store/spicedb"
	"github.com/odpf/shield/pkg/sql"

	"github.com/odpf/salt/log"
	salt_server "github.com/odpf/salt/server"
	"github.com/pkg/profile"
	cli "github.com/spf13/cobra"
)

var (
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

func serve(logger log.Logger, cfg *config.Shield) error {
	if profiling := os.Getenv("SHIELD_PROFILE"); profiling == "true" || profiling == "1" {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	}

	// @TODO: need to inject custom logger wrapper over zap into ctx to use it internally
	ctx, cancelFunc := context.WithCancel(salt_server.HandleSignals(context.Background()))
	defer cancelFunc()

	db, err := setupDB(cfg.DB, logger)
	if err != nil {
		return err
	}
	defer func() {
		logger.Info("cleaning up db")
		db.Close()
	}()

	// load resource config
	if cfg.App.ResourcesConfigPath == "" {
		return errors.New("resource config path cannot be left empty")
	}

	resourceBlobFS, err := blob.NewStore(ctx, cfg.App.ResourcesConfigPath, cfg.App.ResourcesConfigPathSecret)
	if err != nil {
		return err
	}
	resourceRepository := blob.NewResourcesRepository(logger, resourceBlobFS)
	if err := resourceRepository.InitCache(ctx, ruleCacheRefreshDelay); err != nil {
		return err
	}
	defer func() {
		logger.Info("cleaning up resource blob")
		defer resourceRepository.Close()
	}()

	serviceStore := postgres.NewStore(db)

	authzStore, err := spicedb.New(cfg.SpiceDB, logger)
	if err != nil {
		return err
	}

	nrApp, err := setupNewRelic(cfg.NewRelic, logger)
	if err != nil {
		return err
	}

	deps, err := buildAPIDependencies(ctx, logger, cfg.App.IdentityProxyHeader, resourceRepository, serviceStore, authzStore)
	if err != nil {
		return err
	}

	// serving proxies
	cbs, cps, err := serveProxies(ctx, logger, cfg.App.IdentityProxyHeader, cfg.Proxy, deps.ResourceService, deps.UserService)
	if err != nil {
		return err
	}
	defer func() {
		// clean up stage
		logger.Info("cleaning up rules proxy blob")
		for _, f := range cbs {
			if err := f(); err != nil {
				logger.Warn("error occurred during shutdown rules proxy blob storages", "err", err)
			}
		}

		logger.Info("cleaning up proxies")
		for _, f := range cps {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*20)
			if err := f(shutdownCtx); err != nil {
				shutdownCancel()
				logger.Warn("error occurred during shutdown proxies", "err", err)
				continue
			}
			shutdownCancel()
		}
	}()

	keystrokeTermChan := make(chan os.Signal, 1)
	// we'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	signal.Notify(keystrokeTermChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	// serving server
	muxServer, err := server.Serve(ctx, logger, cfg.App, nrApp, deps)
	if err != nil {
		return err
	}
	defer func() {
		logger.Info("cleaning up server")
		server.Cleanup(ctx, muxServer)
	}()

	// wait for termination
	select {
	case <-ctx.Done():
		fmt.Printf("process: ctx done bye\n")
		break
	case <-keystrokeTermChan:
		fmt.Printf("process: kill signal received. bye \n")
		break
	}

	return nil
}

func buildAPIDependencies(
	ctx context.Context,
	logger log.Logger,
	identityProxyHeader string,
	resourceRepository *blob.ResourcesRepository,
	serviceStore postgres.Store,
	authzStore *spicedb.SpiceDB,
) (api.Deps, error) {
	actionService := action.NewService(serviceStore)
	namespaceService := namespace.NewService(serviceStore)
	userService := user.NewService(identityProxyHeader, serviceStore)
	relationService := relation.NewService(serviceStore, authzStore)
	groupService := group.NewService(serviceStore, relationService, userService)
	organizationService := organization.NewService(serviceStore, relationService, userService)
	projectService := project.NewService(serviceStore, relationService, userService)
	policyService := policy.NewService(serviceStore, authzStore)
	roleService := role.NewService(serviceStore)
	bootstrapService := bootstrap.NewService(logger, policyService, actionService, namespaceService, roleService, resourceRepository)
	resourceService := resource.NewService(serviceStore, authzStore, resourceRepository, relationService, userService)

	bootstrapService.BootstrapDefaultDefinitions(ctx)
	err := bootstrapService.BootstrapResources(ctx)
	if err != nil {
		return api.Deps{}, err
	}

	dependencies := api.Deps{
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
		IdentityProxyHeader: identityProxyHeader,
	}
	return dependencies, nil
}

func setupNewRelic(cfg config.NewRelic, logger log.Logger) (newrelic.Application, error) {
	nrCfg := newrelic.NewConfig(cfg.AppName, cfg.License)
	nrCfg.Enabled = cfg.Enabled

	if nrCfg.Enabled {
		nrApp, err := newrelic.NewApplication(nrCfg)
		if err != nil {
			return nil, errors.New("failed to load Newrelic Application")
		}
		return nrApp, nil
	}
	return nil, nil
}

func setupDB(cfg postgres.Config, logger log.Logger) (db *sql.SQL, err error) {
	db, err = sql.New(sql.Config{
		Driver:              cfg.Driver,
		URL:                 cfg.URL,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxOpenConns:        cfg.MaxOpenConns,
		ConnMaxLifeTime:     cfg.ConnMaxLifeTime,
		MaxQueryTimeoutInMS: cfg.MaxQueryTimeout,
	})
	if err != nil {
		err = fmt.Errorf("failed to setup db: %w", err)
		return
	}

	return
}
