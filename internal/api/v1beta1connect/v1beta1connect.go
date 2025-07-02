package v1beta1connect

import (
	"github.com/raystack/frontier/internal/api"
	apiv1beta1 "github.com/raystack/frontier/internal/api/v1beta1"
	frontierv1beta1connect "github.com/raystack/frontier/proto/v1beta1/frontierv1beta1connect"
)

// TODO: add a field "authConfig" in this struct which should be of type Config from core/authenticate
// removed to avoid lint errors
type ConnectHandler struct {
	frontierv1beta1connect.UnimplementedAdminServiceHandler
	frontierv1beta1connect.UnimplementedFrontierServiceHandler

	orgService                       apiv1beta1.OrganizationService
	orgKycService                    apiv1beta1.KycService
	projectService                   apiv1beta1.ProjectService
	groupService                     apiv1beta1.GroupService
	roleService                      apiv1beta1.RoleService
	policyService                    apiv1beta1.PolicyService
	userService                      apiv1beta1.UserService
	namespaceService                 apiv1beta1.NamespaceService
	permissionService                apiv1beta1.PermissionService
	relationService                  apiv1beta1.RelationService
	resourceService                  apiv1beta1.ResourceService
	sessionService                   apiv1beta1.SessionService
	authnService                     apiv1beta1.AuthnService
	deleterService                   apiv1beta1.CascadeDeleter
	metaSchemaService                apiv1beta1.MetaSchemaService
	bootstrapService                 apiv1beta1.BootstrapService
	invitationService                apiv1beta1.InvitationService
	serviceUserService               apiv1beta1.ServiceUserService
	auditService                     apiv1beta1.AuditService
	domainService                    apiv1beta1.DomainService
	preferenceService                apiv1beta1.PreferenceService
	customerService                  apiv1beta1.CustomerService
	planService                      apiv1beta1.PlanService
	subscriptionService              apiv1beta1.SubscriptionService
	productService                   apiv1beta1.ProductService
	entitlementService               apiv1beta1.EntitlementService
	checkoutService                  apiv1beta1.CheckoutService
	creditService                    apiv1beta1.CreditService
	usageService                     apiv1beta1.UsageService
	invoiceService                   apiv1beta1.InvoiceService
	webhookService                   apiv1beta1.WebhookService
	eventService                     apiv1beta1.EventService
	prospectService                  apiv1beta1.ProspectService
	orgBillingService                apiv1beta1.OrgBillingService
	orgInvoicesService               apiv1beta1.OrgInvoicesService
	orgTokensService                 apiv1beta1.OrgTokensService
	orgUsersService                  apiv1beta1.OrgUsersService
	orgProjectsService               apiv1beta1.OrgProjectsService
	projectUsersService              apiv1beta1.ProjectUsersService
	orgServiceUserCredentialsService apiv1beta1.OrgServiceUserCredentialsService
	orgServiceUserService            apiv1beta1.OrgServiceUserService
	userOrgsService                  apiv1beta1.UserOrgsService
	userProjectsService              apiv1beta1.UserProjectsService
}

func NewConnectHandler(deps api.Deps) *ConnectHandler {
	return &ConnectHandler{
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
		auditService:                     deps.AuditService,
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
	}
}
