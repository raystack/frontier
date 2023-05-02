package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/odpf/shield/core/deleter"

	"github.com/odpf/shield/core/authenticate/session"
	"github.com/odpf/shield/core/metaschema"
	"github.com/odpf/shield/internal/server/consts"

	_ "github.com/authzed/authzed-go/proto/authzed/api/v0"
	_ "github.com/jackc/pgx/v4/stdlib"
	newrelic "github.com/newrelic/go-agent"
	"github.com/odpf/shield/core/authenticate"

	"github.com/odpf/shield/config"
	"github.com/odpf/shield/core/action"
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
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/internal/server"
	"github.com/odpf/shield/internal/store/blob"
	"github.com/odpf/shield/internal/store/postgres"
	"github.com/odpf/shield/internal/store/spicedb"
	"github.com/odpf/shield/pkg/db"

	"github.com/odpf/salt/log"
	"github.com/pkg/profile"
	"google.golang.org/grpc/codes"
)

var (
	ruleCacheRefreshDelay = time.Minute * 2
)

func StartServer(logger *log.Zap, cfg *config.Shield) error {
	if profiling := os.Getenv("SHIELD_PROFILE"); profiling == "true" || profiling == "1" {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	}

	// @TODO: need to inject custom logger wrapper over zap into ctx to use it internally
	ctx, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancelFunc()

	dbClient, err := setupDB(cfg.DB, logger)
	if err != nil {
		return err
	}
	defer func() {
		logger.Info("cleaning up db")
		dbClient.Close()
	}()

	// load resource config
	if cfg.App.ResourcesConfigPath == "" {
		return errors.New("resource config path cannot be left empty")
	}

	resourceBlobFS, err := blob.NewStore(ctx, cfg.App.ResourcesConfigPath, cfg.App.ResourcesConfigPathSecret)
	if err != nil {
		return err
	}
	resourceBlobRepository := blob.NewResourcesRepository(logger, resourceBlobFS)
	if err := resourceBlobRepository.InitCache(ctx, ruleCacheRefreshDelay); err != nil {
		return err
	}
	defer func() {
		logger.Info("cleaning up resource blob")
		defer resourceBlobRepository.Close()
	}()

	spiceDBClient, err := spicedb.New(cfg.SpiceDB, logger)
	if err != nil {
		return err
	}

	nrApp, err := setupNewRelic(cfg.NewRelic, logger)
	if err != nil {
		return err
	}

	//
	actionRepository := postgres.NewActionRepository(dbClient)
	actionService := action.NewService(actionRepository)

	roleRepository := postgres.NewRoleRepository(dbClient)
	roleService := role.NewService(roleRepository)

	policyPGRepository := postgres.NewPolicyRepository(dbClient)
	policySpiceRepository := spicedb.NewPolicyRepository(logger, spiceDBClient)
	policyService := policy.NewService(policyPGRepository)

	namespaceRepository := postgres.NewNamespaceRepository(dbClient)
	namespaceService := namespace.NewService(namespaceRepository)

	deps, err := buildAPIDependencies(logger, cfg, resourceBlobRepository, dbClient, spiceDBClient)
	if err != nil {
		return err
	}
	// load metadata schema in memory from db
	if err := deps.MetaSchemaService.InitMetaSchemas(context.Background()); err != nil {
		logger.Warn("metaschemas initialization failed", "err", err)
	}

	// session service initialization and cleanup
	if err := deps.SessionService.InitSessions(context.Background()); err != nil {
		logger.Warn("sessions database cleanup failed", "err", err)
	}
	defer func() {
		logger.Debug("cleaning up cron jobs")
		deps.SessionService.Close()
	}()

	if err := deps.RegistrationService.InitFlows(context.Background()); err != nil {
		logger.Warn("flows database cleanup failed", "err", err)
	}
	defer func() {
		deps.RegistrationService.Close()
	}()

	s := schema.NewSchemaMigrationService(
		blob.NewSchemaConfigRepository(resourceBlobFS),
		namespaceService,
		roleService,
		actionService,
		policyService,
		policySpiceRepository,
	)

	err = s.RunMigrations(ctx)
	if err != nil {
		return err
	}

	// serving proxies
	cbs, cps, err := serveProxies(ctx, logger, cfg.App.IdentityProxyHeader, cfg.App.UserIDHeader, cfg.Proxy, deps.ResourceService, deps.RelationService, deps.UserService, deps.ProjectService)
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

	// serving server
	return server.Serve(ctx, logger, cfg.App, nrApp, deps)
}

func buildAPIDependencies(
	logger log.Logger,
	cfg *config.Shield,
	resourceBlobRepository *blob.ResourcesRepository,
	dbc *db.Client,
	sdb *spicedb.SpiceDB,
) (api.Deps, error) {
	actionRepository := postgres.NewActionRepository(dbc)
	actionService := action.NewService(actionRepository)

	namespaceRepository := postgres.NewNamespaceRepository(dbc)
	namespaceService := namespace.NewService(namespaceRepository)

	sessionService := session.NewService(logger, postgres.NewSessionRepository(logger, dbc), consts.SessionValidity)

	roleRepository := postgres.NewRoleRepository(dbc)
	roleService := role.NewService(roleRepository)

	relationPGRepository := postgres.NewRelationRepository(dbc)
	relationSpiceRepository := spicedb.NewRelationRepository(sdb, cfg.SpiceDB.FullyConsistent)
	relationService := relation.NewService(relationPGRepository, relationSpiceRepository, roleService)

	userRepository := postgres.NewUserRepository(dbc)
	userService := user.NewService(userRepository, sessionService, relationService)

	groupRepository := postgres.NewGroupRepository(dbc)
	groupService := group.NewService(groupRepository, relationService, userService)

	organizationRepository := postgres.NewOrganizationRepository(dbc)
	organizationService := organization.NewService(organizationRepository, relationService, userService)

	projectRepository := postgres.NewProjectRepository(dbc)
	projectService := project.NewService(projectRepository, relationService, userService)

	policyPGRepository := postgres.NewPolicyRepository(dbc)
	policyService := policy.NewService(policyPGRepository)

	metaschemaRepository := postgres.NewMetaSchemaRepository(dbc)
	metaschemaService := metaschema.NewService(metaschemaRepository)

	resourcePGRepository := postgres.NewResourceRepository(dbc)
	resourceService := resource.NewService(
		resourcePGRepository,
		resourceBlobRepository,
		relationService,
		userService,
		projectService)

	registrationService := authenticate.NewRegistrationService(logger, cfg.App.Authentication, postgres.NewFlowRepository(logger, dbc), userService)

	cascadeDeleter := deleter.NewCascadeDeleter(organizationService, projectService, resourceService, groupService)

	dependencies := api.Deps{
		DisableOrgsListing:  cfg.App.DisableOrgsListing,
		DisableUsersListing: cfg.App.DisableUsersListing,
		OrgService:          organizationService,
		ProjectService:      projectService,
		GroupService:        groupService,
		RoleService:         roleService,
		PolicyService:       policyService,
		UserService:         userService,
		NamespaceService:    namespaceService,
		ActionService:       actionService,
		RelationService:     relationService,
		ResourceService:     resourceService,
		SessionService:      sessionService,
		RegistrationService: registrationService,
		DeleterService:      cascadeDeleter,
		MetaSchemaService:   metaschemaService,
	}
	return dependencies, nil
}

func setupNewRelic(cfg config.NewRelic, logger log.Logger) (newrelic.Application, error) {
	nrCfg := newrelic.NewConfig(cfg.AppName, cfg.License)
	nrCfg.Enabled = cfg.Enabled
	nrCfg.ErrorCollector.IgnoreStatusCodes = []int{
		http.StatusNotFound,
		http.StatusUnauthorized,
		int(codes.Unauthenticated),
		int(codes.PermissionDenied),
		int(codes.InvalidArgument),
		int(codes.AlreadyExists),
	}

	if nrCfg.Enabled {
		nrApp, err := newrelic.NewApplication(nrCfg)
		if err != nil {
			return nil, errors.New("failed to load Newrelic Application")
		}
		return nrApp, nil
	}
	return nil, nil
}

func setupDB(cfg db.Config, logger log.Logger) (dbc *db.Client, err error) {
	// prefer use pgx instead of lib/pq for postgres to catch pg error
	if cfg.Driver == "postgres" {
		cfg.Driver = "pgx"
	}
	dbc, err = db.New(cfg)
	if err != nil {
		err = fmt.Errorf("failed to setup db: %w", err)
		return
	}

	return
}
