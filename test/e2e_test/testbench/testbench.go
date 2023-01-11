package testbench

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/config"
	"github.com/odpf/shield/internal/api"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/internal/store/spicedb"
	"github.com/odpf/shield/pkg/db"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"

	"context"
	"errors"
	"net/http"

	"github.com/odpf/shield/core/action"
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
	"github.com/odpf/shield/internal/api/v1beta1"
	"github.com/odpf/shield/internal/proxy"
	"github.com/odpf/shield/internal/proxy/hook"
	authz_hook "github.com/odpf/shield/internal/proxy/hook/authz"
	"github.com/odpf/shield/internal/proxy/middleware/attributes"
	"github.com/odpf/shield/internal/proxy/middleware/authz"
	"github.com/odpf/shield/internal/proxy/middleware/basic_auth"
	"github.com/odpf/shield/internal/proxy/middleware/prefix"
	"github.com/odpf/shield/internal/proxy/middleware/rulematch"
	"github.com/odpf/shield/internal/store/blob"
	"github.com/odpf/shield/internal/store/postgres"
)

const (
	preSharedKey         = "shield"
	waitContainerTimeout = 60 * time.Second
)

var (
	RuleCacheRefreshDelay = time.Minute * 2
)

type TestBench struct {
	PGConfig          db.Config
	SpiceDBConfig     spicedb.Config
	bridgeNetworkName string
	pool              *dockertest.Pool
	network           *docker.Network
	resources         []*dockertest.Resource
}

func Init(appConfig *config.Shield) (*TestBench, *config.Shield, error) {
	var (
		err    error
		logger = log.NewZap()
	)

	te := &TestBench{
		bridgeNetworkName: fmt.Sprintf("bridge-%s", uuid.New().String()),
		resources:         []*dockertest.Resource{},
	}

	te.pool, err = dockertest.NewPool("")
	if err != nil {
		return nil, nil, err
	}

	// Create a bridge network for testing.
	te.network, err = te.pool.Client.CreateNetwork(docker.CreateNetworkOptions{
		Name: te.bridgeNetworkName,
	})
	if err != nil {
		return nil, nil, err
	}

	// pg 1
	logger.Info("creating main postgres...")
	_, connMainPGExternal, res, err := initPG(logger, te.network, te.pool, "test_db")
	if err != nil {
		return nil, nil, err
	}
	te.resources = append(te.resources, res)
	logger.Info("main postgres is created")

	// pg 2
	logger.Info("creating spicedb postgres...")
	connSpicePGInternal, _, res, err := initPG(logger, te.network, te.pool, "spicedb")
	if err != nil {
		return nil, nil, err
	}
	te.resources = append(te.resources, res)
	logger.Info("spicedb postgres is created")

	logger.Info("migrating spicedb...")
	if err = migrateSpiceDB(logger, te.network, te.pool, connSpicePGInternal); err != nil {
		return nil, nil, err
	}
	logger.Info("spicedb is migrated")

	logger.Info("starting up spicedb...")
	spiceDBPort, res, err := startSpiceDB(logger, te.network, te.pool, connSpicePGInternal, preSharedKey)
	if err != nil {
		return nil, nil, err
	}
	te.resources = append(te.resources, res)
	logger.Info("spicedb is up")

	te.PGConfig = db.Config{
		Driver:              "postgres",
		URL:                 connMainPGExternal,
		MaxIdleConns:        10,
		MaxOpenConns:        10,
		ConnMaxLifeTime:     time.Millisecond * 100,
		MaxQueryTimeoutInMS: time.Millisecond * 100,
	}

	te.SpiceDBConfig = spicedb.Config{
		Host:         "localhost",
		Port:         spiceDBPort,
		PreSharedKey: preSharedKey,
	}

	appConfig.DB = te.PGConfig
	appConfig.SpiceDB = te.SpiceDBConfig

	logger.Info("migrating shield...")
	if err = migrateShield(appConfig); err != nil {
		return nil, nil, err
	}
	logger.Info("shield is migrated")

	logger.Info("starting up shield...")
	startShield(appConfig)
	logger.Info("shield is up")

	return te, appConfig, nil
}

func (te *TestBench) CleanUp() error {
	return nil
}

func ServeProxies(
	ctx context.Context,
	logger log.Logger,
	identityProxyHeaderKey,
	userIDHeaderKey string,
	cfg proxy.ServicesConfig,
	resourceService *resource.Service,
	relationService *relation.Service,
	userService *user.Service,
	projectService *project.Service,
) ([]func() error, []func(ctx context.Context) error, error) {
	var cleanUpBlobs []func() error
	var cleanUpProxies []func(ctx context.Context) error

	for _, svcConfig := range cfg.Services {
		hookPipeline := buildHookPipeline(logger, resourceService, relationService, identityProxyHeaderKey)

		h2cProxy := proxy.NewH2c(
			proxy.NewH2cRoundTripper(logger, hookPipeline),
			proxy.NewDirector(),
		)

		// load rules sets
		if svcConfig.RulesPath == "" {
			return nil, nil, errors.New("ruleset field cannot be left empty")
		}

		ruleBlobFS, err := blob.NewStore(ctx, svcConfig.RulesPath, svcConfig.RulesPathSecret)
		if err != nil {
			return nil, nil, err
		}

		ruleBlobRepository := blob.NewRuleRepository(logger, ruleBlobFS)
		if err := ruleBlobRepository.InitCache(ctx, RuleCacheRefreshDelay); err != nil {
			return nil, nil, err
		}
		cleanUpBlobs = append(cleanUpBlobs, ruleBlobRepository.Close)

		ruleService := rule.NewService(ruleBlobRepository)

		middlewarePipeline := buildMiddlewarePipeline(logger, h2cProxy, identityProxyHeaderKey, userIDHeaderKey, resourceService, userService, ruleService, projectService)

		cps := proxy.Serve(ctx, logger, svcConfig, middlewarePipeline)
		cleanUpProxies = append(cleanUpProxies, cps)
	}

	logger.Info("[shield] proxy is up")
	return cleanUpBlobs, cleanUpProxies, nil
}

func buildHookPipeline(log log.Logger, resourceService v1beta1.ResourceService, relationService v1beta1.RelationService, identityProxyHeaderKey string) hook.Service {
	rootHook := hook.New()
	return authz_hook.New(log, rootHook, rootHook, resourceService, relationService, identityProxyHeaderKey)
}

// buildPipeline builds middleware sequence
func buildMiddlewarePipeline(
	logger *log.Zap,
	proxy http.Handler,
	identityProxyHeaderKey, userIDHeaderKey string,
	resourceService *resource.Service,
	userService *user.Service,
	ruleService *rule.Service,
	projectService *project.Service,
) http.Handler {
	// Note: execution order is bottom up
	prefixWare := prefix.New(logger, proxy)
	casbinAuthz := authz.New(logger, prefixWare, userIDHeaderKey, resourceService, userService)
	basicAuthn := basic_auth.New(logger, casbinAuthz)
	attributeExtractor := attributes.New(logger, basicAuthn, identityProxyHeaderKey, projectService)
	matchWare := rulematch.New(logger, attributeExtractor, rulematch.NewRouteMatcher(ruleService))
	return matchWare
}

func BuildAPIDependenciesAndMigrate(
	ctx context.Context,
	logger log.Logger,
	resourceBlobRepository *blob.ResourcesRepository,
	dbc *db.Client,
	sdb *spicedb.SpiceDB,
	rbfs blob.Bucket,
) (api.Deps, error) {
	actionRepository := postgres.NewActionRepository(dbc)
	actionService := action.NewService(actionRepository)

	namespaceRepository := postgres.NewNamespaceRepository(dbc)
	namespaceService := namespace.NewService(namespaceRepository)

	userRepository := postgres.NewUserRepository(dbc)
	userService := user.NewService(userRepository)

	roleRepository := postgres.NewRoleRepository(dbc)
	roleService := role.NewService(roleRepository)

	relationPGRepository := postgres.NewRelationRepository(dbc)
	relationSpiceRepository := spicedb.NewRelationRepository(sdb)
	relationService := relation.NewService(relationPGRepository, relationSpiceRepository, roleService, userService)

	groupRepository := postgres.NewGroupRepository(dbc)
	groupService := group.NewService(groupRepository, relationService, userService)

	organizationRepository := postgres.NewOrganizationRepository(dbc)
	organizationService := organization.NewService(organizationRepository, relationService, userService)

	projectRepository := postgres.NewProjectRepository(dbc)
	projectService := project.NewService(projectRepository, relationService, userService)

	policyPGRepository := postgres.NewPolicyRepository(dbc)
	policyService := policy.NewService(policyPGRepository)

	policySpiceRepository := spicedb.NewPolicyRepository(sdb)

	resourcePGRepository := postgres.NewResourceRepository(dbc)
	resourceService := resource.NewService(
		resourcePGRepository,
		resourceBlobRepository,
		relationService,
		userService)

	dependencies := api.Deps{
		OrgService:       organizationService,
		UserService:      userService,
		ProjectService:   projectService,
		GroupService:     groupService,
		RelationService:  relationService,
		ResourceService:  resourceService,
		RoleService:      roleService,
		PolicyService:    policyService,
		ActionService:    actionService,
		NamespaceService: namespaceService,
	}

	s := schema.NewSchemaMigrationService(
		blob.NewSchemaConfigRepository(rbfs),
		namespaceService,
		roleService,
		actionService,
		policyService,
		policySpiceRepository,
	)

	err := s.RunMigrations(ctx)
	if err != nil {
		return api.Deps{}, err
	}

	return dependencies, nil
}
