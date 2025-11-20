package v1beta1connect

import (
	"context"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/internal/api"
	frontierv1beta1connect "github.com/raystack/frontier/proto/v1beta1/frontierv1beta1connect"
	"go.uber.org/zap"
)

const loggerContextKey = "logger"

type ConnectHandler struct {
	frontierv1beta1connect.UnimplementedAdminServiceHandler
	frontierv1beta1connect.UnimplementedFrontierServiceHandler

	authConfig                       authenticate.Config
	orgService                       OrganizationService
	orgKycService                    KycService
	projectService                   ProjectService
	groupService                     GroupService
	roleService                      RoleService
	policyService                    PolicyService
	userService                      UserService
	namespaceService                 NamespaceService
	permissionService                PermissionService
	relationService                  RelationService
	resourceService                  ResourceService
	sessionService                   SessionService
	authnService                     AuthnService
	deleterService                   CascadeDeleter
	metaSchemaService                MetaSchemaService
	bootstrapService                 BootstrapService
	invitationService                InvitationService
	serviceUserService               ServiceUserService
	domainService                    DomainService
	preferenceService                PreferenceService
	customerService                  CustomerService
	planService                      PlanService
	subscriptionService              SubscriptionService
	productService                   ProductService
	entitlementService               EntitlementService
	checkoutService                  CheckoutService
	creditService                    CreditService
	usageService                     UsageService
	invoiceService                   InvoiceService
	webhookService                   WebhookService
	eventService                     EventService
	prospectService                  ProspectService
	orgBillingService                OrgBillingService
	orgInvoicesService               OrgInvoicesService
	orgTokensService                 OrgTokensService
	orgUsersService                  OrgUsersService
	orgProjectsService               OrgProjectsService
	projectUsersService              ProjectUsersService
	orgServiceUserCredentialsService OrgServiceUserCredentialsService
	orgServiceUserService            OrgServiceUserService
	userOrgsService                  UserOrgsService
	userProjectsService              UserProjectsService
	auditRecordService               AuditRecordService
}

func NewConnectHandler(deps api.Deps, authConf authenticate.Config) *ConnectHandler {
	return &ConnectHandler{
		authConfig:                       authConf,
		orgService:                       deps.OrgService,
		orgKycService:                    deps.OrgKycService,
		projectService:                   deps.ProjectService,
		groupService:                     deps.GroupService,
		roleService:                      deps.RoleService,
		policyService:                    deps.PolicyService,
		userService:                      deps.UserService,
		namespaceService:                 deps.NamespaceService,
		permissionService:                deps.PermissionService,
		relationService:                  deps.RelationService,
		resourceService:                  deps.ResourceService,
		sessionService:                   deps.SessionService,
		authnService:                     deps.AuthnService,
		deleterService:                   deps.DeleterService,
		metaSchemaService:                deps.MetaSchemaService,
		bootstrapService:                 deps.BootstrapService,
		invitationService:                deps.InvitationService,
		serviceUserService:               deps.ServiceUserService,
		domainService:                    deps.DomainService,
		preferenceService:                deps.PreferenceService,
		customerService:                  deps.CustomerService,
		planService:                      deps.PlanService,
		subscriptionService:              deps.SubscriptionService,
		productService:                   deps.ProductService,
		entitlementService:               deps.EntitlementService,
		checkoutService:                  deps.CheckoutService,
		creditService:                    deps.CreditService,
		usageService:                     deps.UsageService,
		invoiceService:                   deps.InvoiceService,
		webhookService:                   deps.WebhookService,
		eventService:                     deps.EventService,
		prospectService:                  deps.ProspectService,
		orgBillingService:                deps.OrgBillingService,
		orgInvoicesService:               deps.OrgInvoicesService,
		orgTokensService:                 deps.OrgTokensService,
		orgUsersService:                  deps.OrgUsersService,
		orgProjectsService:               deps.OrgProjectsService,
		projectUsersService:              deps.ProjectUsersService,
		orgServiceUserCredentialsService: deps.OrgServiceUserCredentialsService,
		orgServiceUserService:            deps.OrgServiceUserService,
		userOrgsService:                  deps.UserOrgsService,
		userProjectsService:              deps.UserProjectsService,
		auditRecordService:               deps.AuditRecordService,
	}
}

// GetBillingAccountFromOrgID returns the billing account ID for a given organization
func (h *ConnectHandler) GetBillingAccountFromOrgID(ctx context.Context, orgID string) (string, error) {
	customer, err := h.customerService.GetByOrgID(ctx, orgID)
	if err != nil {
		return "", err
	}
	return customer.ID, nil
}

// GetOrgIDFromSubscriptionID returns the organization ID for a given subscription
func (h *ConnectHandler) GetOrgIDFromSubscriptionID(ctx context.Context, subscriptionID string) (string, error) {
	sub, err := h.subscriptionService.GetByID(ctx, subscriptionID)
	if err != nil {
		return "", err
	}
	// Get the billing account (customer) to find the org
	customer, err := h.customerService.GetByID(ctx, sub.CustomerID)
	if err != nil {
		return "", err
	}
	return customer.OrgID, nil
}

func ExtractLogger(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerContextKey).(*zap.Logger); ok {
		return logger
	}
	return nil
}
