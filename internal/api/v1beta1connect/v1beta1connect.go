package v1beta1connect

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/internal/api"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1connect "github.com/raystack/frontier/proto/v1beta1/frontierv1beta1connect"
)

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
	orgPATsService                   OrgPATsService
	orgUsersService                  OrgUsersService
	orgProjectsService               OrgProjectsService
	projectUsersService              ProjectUsersService
	orgServiceUserCredentialsService OrgServiceUserCredentialsService
	orgServiceUserService            OrgServiceUserService
	userOrgsService                  UserOrgsService
	userProjectsService              UserProjectsService
	auditRecordService               AuditRecordService
	userPATService                   UserPATService
	membershipService                MembershipService
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
		orgPATsService:                   deps.OrgPATsService,
		orgUsersService:                  deps.OrgUsersService,
		orgProjectsService:               deps.OrgProjectsService,
		projectUsersService:              deps.ProjectUsersService,
		orgServiceUserCredentialsService: deps.OrgServiceUserCredentialsService,
		orgServiceUserService:            deps.OrgServiceUserService,
		userOrgsService:                  deps.UserOrgsService,
		userProjectsService:              deps.UserProjectsService,
		auditRecordService:               deps.AuditRecordService,
		userPATService:                   deps.UserPATService,
		membershipService:                deps.MembershipService,
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
		if errors.Is(err, subscription.ErrNotFound) {
			return "", connect.NewError(connect.CodeNotFound, err)
		}
		slog.ErrorContext(ctx, "failed to get subscription for org lookup",
			"subscription_id", subscriptionID, "error", err)
		return "", connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	cust, err := h.customerService.GetByID(ctx, sub.CustomerID)
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return "", connect.NewError(connect.CodeNotFound, ErrNotFound)
		}
		slog.ErrorContext(ctx, "failed to get billing account for org lookup",
			"customer_id", sub.CustomerID, "subscription_id", subscriptionID, "error", err)
		return "", connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return cust.OrgID, nil
}

// GetOrgIDFromCheckoutID returns the organization ID for a given checkout
func (h *ConnectHandler) GetOrgIDFromCheckoutID(ctx context.Context, checkoutID string) (string, error) {
	co, err := h.checkoutService.GetByID(ctx, checkoutID)
	if err != nil {
		if errors.Is(err, checkout.ErrNotFound) {
			return "", connect.NewError(connect.CodeNotFound, err)
		}
		slog.ErrorContext(ctx, "failed to get checkout for org lookup",
			"checkout_id", checkoutID, "error", err)
		return "", connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	cust, err := h.customerService.GetByID(ctx, co.CustomerID)
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return "", connect.NewError(connect.CodeNotFound, ErrNotFound)
		}
		slog.ErrorContext(ctx, "failed to get billing account for org lookup",
			"customer_id", co.CustomerID, "checkout_id", checkoutID, "error", err)
		return "", connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return cust.OrgID, nil
}

// GetOrgIDFromBillingAccountID returns the organization ID for a given billing account
func (h *ConnectHandler) GetOrgIDFromBillingAccountID(ctx context.Context, billingAccountID string) (string, error) {
	cust, err := h.customerService.GetByID(ctx, billingAccountID)
	if err != nil {
		if errors.Is(err, customer.ErrNotFound) {
			return "", connect.NewError(connect.CodeNotFound, ErrNotFound)
		}
		if errors.Is(err, customer.ErrInvalidUUID) || errors.Is(err, customer.ErrInvalidID) {
			return "", connect.NewError(connect.CodeInvalidArgument, err)
		}
		slog.ErrorContext(ctx, "failed to get billing account for org lookup",
			"billing_account_id", billingAccountID, "error", err)
		return "", connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return cust.OrgID, nil
}
