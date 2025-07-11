package v1beta1

import (
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/internal/api"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc"
)

type Handler struct {
	frontierv1beta1.UnimplementedFrontierServiceServer
	frontierv1beta1.UnimplementedAdminServiceServer

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
	auditService                     AuditService
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
}

func Register(s *grpc.Server, deps api.Deps, authConf authenticate.Config) {
	handler := &Handler{
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
	s.RegisterService(&frontierv1beta1.FrontierService_ServiceDesc, handler)
	s.RegisterService(&frontierv1beta1.AdminService_ServiceDesc, handler)
}
