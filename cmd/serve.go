package cmd

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/raystack/frontier/core/aggregates/orgbilling"
	"github.com/raystack/frontier/core/aggregates/orginvoices"
	"github.com/raystack/frontier/core/aggregates/orgprojects"
	"github.com/raystack/frontier/core/aggregates/orgserviceuser"
	"github.com/raystack/frontier/core/aggregates/orgserviceusercredentials"
	"github.com/raystack/frontier/core/aggregates/orgtokens"
	"github.com/raystack/frontier/core/aggregates/orgusers"
	"github.com/raystack/frontier/core/aggregates/projectusers"
	"github.com/raystack/frontier/core/aggregates/userorgs"
	"github.com/raystack/frontier/core/aggregates/userprojects"
	"github.com/raystack/frontier/core/auditrecord"

	"github.com/raystack/frontier/core/kyc"
	"github.com/raystack/frontier/core/prospect"

	"golang.org/x/exp/slices"

	"github.com/jackc/pgx/v4"
	"github.com/stripe/stripe-go/v79"

	prometheusmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/raystack/frontier/internal/metrics"

	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/raystack/frontier/core/webhook"

	"github.com/raystack/frontier/core/event"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"github.com/raystack/frontier/billing/invoice"

	"github.com/raystack/frontier/billing/usage"

	"github.com/raystack/frontier/billing/credit"

	"github.com/raystack/frontier/billing/checkout"

	"github.com/raystack/frontier/billing/entitlement"

	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"

	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/stripe/stripe-go/v79/client"

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

	"github.com/go-webauthn/webauthn/webauthn"
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

var ruleCacheRefreshDelay = time.Minute * 2
var GetStripeClientFunc func(logger log.Logger, cfg *config.Frontier) *client.API

func StartServer(logger *log.Zap, cfg *config.Frontier) error {
	logger.Info("frontier starting", "version", config.Version)
	if profiling := os.Getenv("FRONTIER_PROFILE"); profiling == "true" || profiling == "1" {
		defer profile.Start(profile.CPUProfile, profile.ProfilePath("."), profile.NoShutdownHook).Stop()
	}

	ctx, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancelFunc()

	ctx = ctxzap.ToContext(ctx, logger.GetInternalZapLogger().Desugar())

	dbClient, err := setupDB(cfg.DB, logger)
	if err != nil {
		return err
	}
	defer func() {
		logger.Debug("cleaning up db")
		if err := dbClient.Close(); err != nil {
			logger.Warn("db cleanup failed", "err", err)
		}
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
		logger.Debug("cleaning up resource blob")
		if err := resourceBlobRepository.Close(); err != nil {
			logger.Warn("resource blob cleanup failed", "err", err)
		}
	}()

	// load billing plans
	billingBlobFS, err := blob.NewStore(ctx, cfg.Billing.PlansPath, "")
	if err != nil {
		return err
	}
	billingPlanRepository := blob.NewPlanRepository(billingBlobFS)

	// setup telemetry
	nrApp, err := setupNewRelic(cfg.NewRelic, logger)
	if err != nil {
		return err
	}
	promRegistry := prometheus.NewRegistry()
	promMetrics := prometheusmiddleware.NewClientMetrics(
		prometheusmiddleware.WithClientHandlingTimeHistogram(),
	)
	promRegistry.MustRegister(promMetrics)
	if cfg.App.MetricsPort > 0 {
		var dbName string
		if parsedUrl, err := pgx.ParseConfig(cfg.DB.URL); err == nil {
			dbName = parsedUrl.Database
		}
		dbPromCollector := collectors.NewDBStatsCollector(dbClient.DB.DB, dbName)
		promRegistry.MustRegister(
			dbPromCollector,
			collectors.NewGoCollector(),
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		)
		prometheus.DefaultRegisterer = promRegistry
		metrics.Init()
	}

	spiceDBClient, err := spicedb.New(cfg.SpiceDB, logger, promMetrics)
	if err != nil {
		return err
	}

	deps, err := buildAPIDependencies(logger, cfg, dbClient, spiceDBClient, resourceBlobRepository, billingPlanRepository)
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
	logger.Info("migrated authz schema")

	// apply billing plans
	if cfg.Billing.PlansPath != "" {
		if err = deps.BootstrapService.MigrateBillingPlans(ctx); err != nil {
			return err
		}
		logger.Info("migrated billing plans")
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
		logger.Warn("sessions initialization failed", "err", err)
	}
	defer func() {
		logger.Debug("cleaning up sessions")
		if err := deps.SessionService.Close(); err != nil {
			logger.Warn("sessions cleanup failed", "err", err)
		}
	}()

	if err := deps.DomainService.InitDomainVerification(ctx); err != nil {
		logger.Warn("domain initialization failed", "err", err)
	}
	defer func() {
		logger.Debug("cleaning up domains")
		if err := deps.DomainService.Close(); err != nil {
			logger.Warn("domain cleanup failed", "err", err)
		}
	}()

	if err := deps.AuthnService.InitFlows(ctx); err != nil {
		logger.Warn("Authn initialization failed", "err", err)
	}
	defer func() {
		logger.Debug("cleaning up authn")
		if err := deps.AuthnService.Close(); err != nil {
			logger.Warn("Authn cleanup failed", "err", err)
		}
	}()

	if cfg.Billing.StripeKey != "" {
		// billing services initialization and cleanup
		if err := deps.CustomerService.Init(ctx); err != nil {
			return err
		}
		defer func() {
			logger.Debug("cleaning up customers")
			if err := deps.CustomerService.Close(); err != nil {
				logger.Warn("customer service cleanup failed", "err", err)
			}
		}()

		if err := deps.CheckoutService.Init(ctx); err != nil {
			return err
		}
		defer func() {
			logger.Debug("cleaning up checkouts")
			if err := deps.CheckoutService.Close(); err != nil {
				logger.Warn("checkout service cleanup failed", "err", err)
			}
		}()

		if err := deps.SubscriptionService.Init(ctx); err != nil {
			return err
		}
		defer func() {
			logger.Debug("cleaning up subscriptions")
			if err := deps.SubscriptionService.Close(); err != nil {
				logger.Warn("subscription service cleanup failed", "err", err)
			}
		}()

		if err := deps.InvoiceService.Init(ctx); err != nil {
			return err
		}
		defer func() {
			logger.Debug("cleaning up invoices")
			if err := deps.InvoiceService.Close(); err != nil {
				logger.Warn("invoice service cleanup failed", "err", err)
			}
		}()
	}

	go func() {
		if err := deps.LogListener.Listen(ctx); err != nil {
			logger.Warn("log listener failed", "err", err)
		}
	}()

	go server.ServeUI(ctx, logger, cfg.UI, cfg.App)

	// start connect server
	go func() {
		if err := server.ServeConnect(ctx, logger, cfg.App, deps, promRegistry); err != nil {
			logger.Fatal("connect server failed", "err", err.Error())
		}
	}()

	// serving grpc server
	return server.Serve(ctx, logger, cfg.App, nrApp, deps, promRegistry)
}

func buildAPIDependencies(
	logger log.Logger,
	cfg *config.Frontier,
	dbc *db.Client,
	sdb *spicedb.SpiceDB,
	resourceBlobRepository *blob.ResourcesRepository,
	planBlobRepository *blob.PlanRepository,
) (api.Deps, error) {
	preferenceService := preference.NewService(postgres.NewPreferenceRepository(dbc))

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
	consistencyLevel := spicedb.ConsistencyLevel(cfg.SpiceDB.Consistency)
	if cfg.SpiceDB.FullyConsistent {
		consistencyLevel = spicedb.ConsistencyLevelFull
	}
	if !slices.Contains([]spicedb.ConsistencyLevel{
		spicedb.ConsistencyLevelFull,
		spicedb.ConsistencyLevelBestEffort,
		spicedb.ConsistencyLevelMinimizeLatency}, consistencyLevel) {
		return api.Deps{}, fmt.Errorf("invalid consistency level: %s", consistencyLevel)
	}
	authzRelationRepository := spicedb.NewRelationRepository(sdb, consistencyLevel, cfg.SpiceDB.CheckTrace)

	permissionRepository := postgres.NewPermissionRepository(dbc)
	permissionService := permission.NewService(permissionRepository)

	relationPGRepository := postgres.NewRelationRepository(dbc)
	relationService := relation.NewService(relationPGRepository, authzRelationRepository)

	roleRepository := postgres.NewRoleRepository(dbc)
	roleService := role.NewService(roleRepository, relationService, permissionService)

	policyPGRepository := postgres.NewPolicyRepository(dbc)
	policyService := policy.NewService(policyPGRepository, relationService, roleService)

	userRepository := postgres.NewUserRepository(dbc)
	userService := user.NewService(userRepository, relationService, policyService, roleService)

	prospectRepository := postgres.NewProspectRepository(dbc)
	prospectService := prospect.NewService(prospectRepository)

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
			cfg.App.Mailer.TLSPolicy(),
		)
		logger.Info("mailer enabled", "host", cfg.App.Mailer.SMTPHost, "port", cfg.App.Mailer.SMTPPort)
	}

	wconfig := &webauthn.Config{
		RPDisplayName: cfg.App.Authentication.PassKey.RPDisplayName,
		RPID:          cfg.App.Authentication.PassKey.RPID,
		RPOrigins:     cfg.App.Authentication.PassKey.RPOrigins,
	}
	webAuthConfig, err := webauthn.New(wconfig)
	if err != nil {
		if wconfig.RPDisplayName == "" && wconfig.RPID == "" && wconfig.RPOrigins == nil {
			webAuthConfig = nil
		} else {
			return api.Deps{}, fmt.Errorf("failed to parse passkey config: %w", err)
		}
	}
	authnService := authenticate.NewService(logger, cfg.App.Authentication,
		postgres.NewFlowRepository(logger, dbc), mailDialer, tokenService, sessionService, userService, serviceUserService, webAuthConfig)

	groupRepository := postgres.NewGroupRepository(dbc)
	groupService := group.NewService(groupRepository, relationService, authnService, policyService)

	organizationRepository := postgres.NewOrganizationRepository(dbc)
	organizationService := organization.NewService(organizationRepository, relationService, userService,
		authnService, policyService, preferenceService)

	orgKycRepository := postgres.NewOrgKycRepository(dbc)
	orgKycService := kyc.NewService(orgKycRepository)

	orgBillingRepository := postgres.NewOrgBillingRepository(dbc)
	orgBillingService := orgbilling.NewService(orgBillingRepository)

	orgInvoicesRepository := postgres.NewOrgInvoicesRepository(dbc)
	orgInvoicesService := orginvoices.NewService(orgInvoicesRepository)

	orgTokensRepository := postgres.NewOrgTokensRepository(dbc)
	orgTokensService := orgtokens.NewService(orgTokensRepository)

	orgUsersRepository := postgres.NewOrgUsersRepository(dbc)
	orgUserService := orgusers.NewService(orgUsersRepository)

	projectUsersRepository := postgres.NewProjectUsersRepository(dbc)
	projectUserService := projectusers.NewService(projectUsersRepository)

	orgProjectsRepository := postgres.NewOrgProjectsRepository(dbc)
	orgProjectsService := orgprojects.NewService(orgProjectsRepository)

	orgServiceUserCredentialsRepository := postgres.NewOrgServiceUserCredentialsRepository(dbc)
	orgServiceUserCredentialsService := orgserviceusercredentials.NewService(orgServiceUserCredentialsRepository)

	orgServiceUserRepository := postgres.NewOrgServiceUserRepository(dbc)
	orgServiceUserService := orgserviceuser.NewService(orgServiceUserRepository)

	userOrgsRepository := postgres.NewUserOrgsRepository(dbc)
	userOrgsService := userorgs.NewService(userOrgsRepository)

	userProjectsRepository := postgres.NewUserProjectsRepository(dbc)
	userProjectsService := userprojects.NewService(userProjectsRepository)

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
		organizationService, groupService, userService, relationService, policyService, preferenceService)

	if GetStripeClientFunc == nil {
		// allow to override the stripe client creation function in tests
		GetStripeClientFunc = getStripeClient
	}
	stripeClient := GetStripeClientFunc(logger, cfg)

	creditService := credit.NewService(postgres.NewBillingTransactionRepository(dbc))
	customerService := customer.NewService(
		stripeClient,
		postgres.NewBillingCustomerRepository(dbc), cfg.Billing, creditService)
	featureRepository := postgres.NewBillingFeatureRepository(dbc)
	priceRepository := postgres.NewBillingPriceRepository(dbc)
	productService := product.NewService(
		stripeClient,
		postgres.NewBillingProductRepository(dbc),
		priceRepository,
		featureRepository,
	)
	planService := plan.NewService(
		stripeClient,
		postgres.NewBillingPlanRepository(dbc),
		productService,
		featureRepository,
		priceRepository,
	)
	subscriptionService := subscription.NewService(
		stripeClient, cfg.Billing,
		postgres.NewBillingSubscriptionRepository(dbc),
		customerService, planService, organizationService,
		productService, creditService)
	entitlementService := entitlement.NewEntitlementService(subscriptionService, productService,
		planService, organizationService)
	checkoutService := checkout.NewService(stripeClient, cfg.Billing, postgres.NewBillingCheckoutRepository(dbc),
		customerService, planService, subscriptionService, productService, creditService, organizationService,
		authnService)

	invoiceService := invoice.NewService(stripeClient, postgres.NewBillingInvoiceRepository(dbc),
		customerService, creditService, productService, dbc, cfg.Billing)

	usageService := usage.NewService(creditService)

	resourceSchemaRepository := blob.NewSchemaConfigRepository(resourceBlobRepository.Bucket)
	bootstrapService := bootstrap.NewBootstrapService(
		cfg.App.Admin,
		resourceSchemaRepository,
		namespaceService,
		roleService,
		permissionService,
		userService,
		authzSchemaRepository,
		planService,
		planBlobRepository,
	)

	cascadeDeleter := deleter.NewCascadeDeleter(organizationService, projectService, resourceService,
		groupService, policyService, roleService, invitationService, userService, customerService,
		subscriptionService, invoiceService,
	)

	// we should default it with a stdout logger repository as postgres can start to bloat really fast
	var auditRepository audit.Repository
	switch cfg.Log.AuditEvents {
	case "db":
		auditRepository = postgres.NewAuditRepository(dbc)
	case "stdout":
		auditRepository = audit.NewWriteOnlyRepository(os.Stdout)
	default:
		auditRepository = audit.NewNoopRepository()
	}
	eventProcessor := event.NewService(cfg.Billing, organizationService, checkoutService, customerService,
		planService, userService, subscriptionService, creditService, invoiceService)
	eventChannel := make(chan audit.Log, 10) // buffered channel to avoid blocking the event processor
	logPublisher := event.NewChanPublisher(eventChannel)
	logListener := event.NewChanListener(eventChannel, eventProcessor)

	webhookService := webhook.NewService(postgres.NewWebhookEndpointRepository(dbc, []byte(cfg.App.Webhook.EncryptionKey)))
	auditService := audit.NewService("frontier",
		auditRepository, webhookService,
		audit.WithLogPublisher(logPublisher),
		audit.WithIgnoreList(cfg.Log.IgnoredAuditEvents),
	)

	auditRecordRepository := postgres.NewAuditRecordRepository(dbc)
	auditRecordService := auditrecord.NewService(auditRecordRepository, userService, serviceUserService)

	dependencies := api.Deps{
		OrgService:                       organizationService,
		OrgKycService:                    orgKycService,
		ProjectService:                   projectService,
		GroupService:                     groupService,
		RoleService:                      roleService,
		PolicyService:                    policyService,
		UserService:                      userService,
		NamespaceService:                 namespaceService,
		PermissionService:                permissionService,
		RelationService:                  relationService,
		ResourceService:                  resourceService,
		SessionService:                   sessionService,
		AuthnService:                     authnService,
		DeleterService:                   cascadeDeleter,
		MetaSchemaService:                metaschemaService,
		BootstrapService:                 bootstrapService,
		InvitationService:                invitationService,
		ServiceUserService:               serviceUserService,
		AuditService:                     auditService,
		DomainService:                    domainService,
		PreferenceService:                preferenceService,
		CustomerService:                  customerService,
		SubscriptionService:              subscriptionService,
		ProductService:                   productService,
		PlanService:                      planService,
		EntitlementService:               entitlementService,
		CheckoutService:                  checkoutService,
		CreditService:                    creditService,
		UsageService:                     usageService,
		InvoiceService:                   invoiceService,
		LogListener:                      logListener,
		WebhookService:                   webhookService,
		EventService:                     eventProcessor,
		ProspectService:                  prospectService,
		OrgBillingService:                orgBillingService,
		OrgInvoicesService:               orgInvoicesService,
		OrgTokensService:                 orgTokensService,
		OrgUsersService:                  orgUserService,
		OrgProjectsService:               orgProjectsService,
		OrgServiceUserCredentialsService: orgServiceUserCredentialsService,
		OrgServiceUserService:            orgServiceUserService,
		ProjectUsersService:              projectUserService,
		UserOrgsService:                  userOrgsService,
		UserProjectsService:              userProjectsService,
		AuditRecordService:               auditRecordService,
	}
	return dependencies, nil
}

// StripeTransport wraps the default http.RoundTripper to add metrics.
type StripeTransport struct {
	Base http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *StripeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// extract the operation from path
	operationName := "unknown"
	pathParams := strings.Split(req.URL.Path, "/")
	if len(pathParams) >= 3 {
		operationName = pathParams[2]
	}
	if metrics.StripeAPILatency != nil {
		record := metrics.StripeAPILatency(operationName, req.Method)
		defer record()
	}

	// perform the request
	resp, err := t.Base.RoundTrip(req)

	return resp, err
}

func getStripeClient(logger log.Logger, cfg *config.Frontier) *client.API {
	stripeLogLevel := stripe.LevelError
	stripeBackends := &stripe.Backends{
		API: stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
			HTTPClient: &http.Client{
				Timeout: time.Second * 80,
				// custom transport to wrap calls
				Transport: &StripeTransport{
					Base: http.DefaultTransport,
				},
			},
			LeveledLogger: &stripe.LeveledLogger{
				Level: stripeLogLevel,
			},
		}),
		Connect: stripe.GetBackend(stripe.ConnectBackend),
		Uploads: stripe.GetBackend(stripe.UploadsBackend),
	}
	stripeClient := client.New(cfg.Billing.StripeKey, stripeBackends)
	if cfg.Billing.StripeKey == "" {
		logger.Warn("stripe key is empty, billing services will be non-functional")
	} else {
		logger.Info("stripe client created")
	}
	return stripeClient
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
