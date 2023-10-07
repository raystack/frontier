package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raystack/frontier/core/preference"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/domain"

	"github.com/raystack/frontier/core/serviceuser"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/frontier/core/authenticate/token"

	"github.com/raystack/frontier/pkg/server"

	"github.com/raystack/frontier/core/invitation"

	"github.com/raystack/frontier/pkg/mailer"

	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/internal/bootstrap"

	"github.com/raystack/frontier/core/deleter"

	_ "github.com/authzed/authzed-go/proto/authzed/api/v0"
	_ "github.com/jackc/pgx/v4/stdlib"
	newrelic "github.com/newrelic/go-agent"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/metaschema"

	"github.com/raystack/frontier/config"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api"
	"github.com/raystack/frontier/internal/store/blob"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/internal/store/spicedb"
	"github.com/raystack/frontier/pkg/db"

	"github.com/pkg/profile"
	"github.com/raystack/salt/log"
	"google.golang.org/grpc/codes"
)

var (
	ruleCacheRefreshDelay = time.Minute * 2
)

func StartServer(logger *log.Zap, cfg *config.Frontier) error {
	logger.Info("frontier starting", "version", config.Version)
	if profiling := os.Getenv("FRONTIER_PROFILE"); profiling == "true" || profiling == "1" {
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

	nrApp, err := setupNewRelic(cfg.NewRelic, logger)
	if err != nil {
		return err
	}

	spiceDBClient, err := spicedb.New(cfg.SpiceDB, logger)
	if err != nil {
		return err
	}

	deps, err := buildAPIDependencies(ctx, logger, cfg, resourceBlobRepository, dbClient, spiceDBClient)
	if err != nil {
		return err
	}
	// load metadata schema in memory from db
	if schemas, err := deps.MetaSchemaService.List(context.Background()); err != nil {
		logger.Warn("metaschemas initialization failed", "err", err)
	} else {
		logger.Info("metaschemas loaded", "count", len(schemas))
	}

	// apply schema
	if err = deps.BootstrapService.MigrateSchema(ctx); err != nil {
		return err
	}

	// apply roles over nil org id
	// nil org is the default org of platform
	if err = deps.BootstrapService.MigrateRoles(ctx); err != nil {
		return err
	}
	// promote normal users to superusers
	if err = deps.BootstrapService.MakeSuperUsers(ctx); err != nil {
		return err
	}

	// session service initialization and cleanup
	if err := deps.SessionService.InitSessions(ctx); err != nil {
		logger.Warn("sessions database cleanup failed", "err", err)
	}
	defer func() {
		logger.Debug("cleaning up cron jobs")
		deps.SessionService.Close()
	}()

	if err := deps.DomainService.InitDomainVerification(ctx); err != nil {
		logger.Warn("domains database cleanup failed", "err", err)
	}
	defer func() {
		deps.DomainService.Close()
	}()

	if err := deps.AuthnService.InitFlows(ctx); err != nil {
		logger.Warn("flows database cleanup failed", "err", err)
	}
	defer func() {
		deps.AuthnService.Close()
	}()

	// serving server
	return server.Serve(ctx, logger, cfg.App, nrApp, deps)
}

func buildAPIDependencies(
	ctx context.Context,
	logger log.Logger,
	cfg *config.Frontier,
	resourceBlobRepository *blob.ResourcesRepository,
	dbc *db.Client,
	sdb *spicedb.SpiceDB,
) (api.Deps, error) {
	preferenceService := preference.NewService(postgres.NewPreferenceRepository(dbc))

	// load and apply config from preferences database or use default values for platform wide configs
	cfgMap, err := preferenceService.LoadPlatformPreferences(ctx)
	if err != nil {
		return api.Deps{}, fmt.Errorf("failed to load platform preferences: %w", err)
	}
	applyPlatformPreference(logger, cfg, cfgMap)

	var tokenKeySet jwk.Set
	if len(cfg.App.Authentication.Token.RSAPath) > 0 {
		if ks, err := jwk.ReadFile(cfg.App.Authentication.Token.RSAPath); err != nil {
			return api.Deps{}, fmt.Errorf("failed to parse rsa key: %w", err)
		} else {
			tokenKeySet = ks
		}
	}
	if len(cfg.App.Authentication.Token.RSABase64) > 0 {
		rawDecoded, err := base64.StdEncoding.DecodeString(cfg.App.Authentication.Token.RSABase64)
		if err != nil {
			return api.Deps{}, fmt.Errorf("failed to decode rsa key as std-base64: %w", err)
		}
		if ks, err := jwk.Parse(rawDecoded); err != nil {
			return api.Deps{}, fmt.Errorf("failed to parse rsa key: %w", err)
		} else {
			tokenKeySet = ks
		}
	}
	tokenService := token.NewService(tokenKeySet, cfg.App.Authentication.Token.Issuer,
		cfg.App.Authentication.Token.Validity)
	sessionService := session.NewService(logger, postgres.NewSessionRepository(logger, dbc), cfg.App.Authentication.Session.Validity)

	namespaceRepository := postgres.NewNamespaceRepository(dbc)
	namespaceService := namespace.NewService(namespaceRepository)

	authzSchemaRepository := spicedb.NewSchemaRepository(logger, sdb)
	authzRelationRepository := spicedb.NewRelationRepository(sdb, cfg.SpiceDB.FullyConsistent)

	permissionRepository := postgres.NewPermissionRepository(dbc)
	permissionService := permission.NewService(permissionRepository)

	relationPGRepository := postgres.NewRelationRepository(dbc)
	relationService := relation.NewService(relationPGRepository, authzRelationRepository)

	roleRepository := postgres.NewRoleRepository(dbc)
	roleService := role.NewService(roleRepository, relationService, permissionService)

	policyPGRepository := postgres.NewPolicyRepository(dbc)
	policyService := policy.NewService(policyPGRepository, relationService, roleService)

	userRepository := postgres.NewUserRepository(dbc)
	userService := user.NewService(userRepository, relationService)

	svUserRepo := postgres.NewServiceUserRepository(dbc)
	scUserCredRepo := postgres.NewServiceUserCredentialRepository(dbc)
	serviceUserService := serviceuser.NewService(svUserRepo, scUserCredRepo, relationService)

	var mailDialer mailer.Dialer = mailer.NewMockDialer()
	if cfg.App.Mailer.SMTPHost != "" && cfg.App.Mailer.SMTPHost != "smtp.example.com" {
		mailDialer = mailer.NewDialerImpl(cfg.App.Mailer.SMTPHost,
			cfg.App.Mailer.SMTPPort,
			cfg.App.Mailer.SMTPUsername,
			cfg.App.Mailer.SMTPPassword,
			cfg.App.Mailer.SMTPInsecure,
			cfg.App.Mailer.Headers,
		)
		logger.Info("mailer enabled", "host", cfg.App.Mailer.SMTPHost, "port", cfg.App.Mailer.SMTPPort)
	}
	authnService := authenticate.NewService(logger, cfg.App.Authentication,
		postgres.NewFlowRepository(logger, dbc), mailDialer, tokenService, sessionService, userService, serviceUserService)

	groupRepository := postgres.NewGroupRepository(dbc)
	groupService := group.NewService(groupRepository, relationService, authnService, policyService)

	resourceSchemaRepository := blob.NewSchemaConfigRepository(resourceBlobRepository.Bucket)
	bootstrapService := bootstrap.NewBootstrapService(
		cfg.App.Admin,
		resourceSchemaRepository,
		namespaceService,
		roleService,
		permissionService,
		userService,
		authzSchemaRepository,
	)

	organizationRepository := postgres.NewOrganizationRepository(dbc)
	organizationService := organization.NewService(organizationRepository, relationService, userService,
		authnService, policyService, cfg.App.DisableOrgsOnCreate)

	domainRepository := postgres.NewDomainRepository(logger, dbc)
	domainService := domain.NewService(logger, domainRepository, userService, organizationService)

	metaschemaRepository := postgres.NewMetaSchemaRepository(logger, dbc)
	metaschemaService := metaschema.NewService(metaschemaRepository)
	projectRepository := postgres.NewProjectRepository(dbc)
	projectService := project.NewService(projectRepository, relationService, userService, policyService,
		authnService, serviceUserService, groupService)

	resourcePGRepository := postgres.NewResourceRepository(dbc)
	resourceService := resource.NewService(
		resourcePGRepository,
		resourceBlobRepository,
		relationService,
		authnService,
		projectService,
		organizationService,
	)

	invitationService := invitation.NewService(mailDialer, postgres.NewInvitationRepository(logger, dbc),
		organizationService, groupService, userService, relationService, policyService, cfg.App.Invite)
	cascadeDeleter := deleter.NewCascadeDeleter(organizationService, projectService, resourceService,
		groupService, policyService, roleService, invitationService, userService)

	// we should default it with a stdout logger repository as postgres can start to bloat really fast
	var auditRepository audit.Repository
	switch cfg.Log.AuditEvents {
	case "db":
		auditRepository = postgres.NewAuditRepository(dbc)
	case "stdout":
		auditRepository = audit.NewWriteOnlyRepository(os.Stdout)
	default:
		auditRepository = audit.NewWriteOnlyRepository(io.Discard)
	}
	auditService := audit.NewService("frontier", auditRepository)

	dependencies := api.Deps{
		DisableOrgsListing:  cfg.App.DisableOrgsListing,
		DisableUsersListing: cfg.App.DisableUsersListing,
		DisableOrgOnCreate:  cfg.App.DisableOrgsOnCreate,
		OrgService:          organizationService,
		ProjectService:      projectService,
		GroupService:        groupService,
		RoleService:         roleService,
		PolicyService:       policyService,
		UserService:         userService,
		NamespaceService:    namespaceService,
		PermissionService:   permissionService,
		RelationService:     relationService,
		ResourceService:     resourceService,
		SessionService:      sessionService,
		AuthnService:        authnService,
		DeleterService:      cascadeDeleter,
		MetaSchemaService:   metaschemaService,
		BootstrapService:    bootstrapService,
		InvitationService:   invitationService,
		ServiceUserService:  serviceUserService,
		AuditService:        auditService,
		DomainService:       domainService,
		PreferenceService:   preferenceService,
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

// applyPlatformPreference applies platform wide preferences to server config
// if preference is not set in database it will use default value
func applyPlatformPreference(logger log.Logger, cfg *config.Frontier, cfgMap map[string]string) {
	cfg.App.DisableOrgsOnCreate = cfgMap[preference.PlatformDisableOrgsOnCreate] == "true"
	cfg.App.DisableOrgsListing = cfgMap[preference.PlatformDisableOrgsListing] == "true"
	cfg.App.DisableUsersListing = cfgMap[preference.PlatformDisableUsersListing] == "true"
	cfg.App.Invite.WithRoles = cfgMap[preference.PlatformInviteWithRoles] == "true"

	cfg.App.Invite.MailTemplate.Body = cfgMap[preference.PlatformInviteMailBody]
	cfg.App.Invite.MailTemplate.Subject = cfgMap[preference.PlatformInviteMailSubject]
}
