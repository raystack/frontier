package v1beta1connect

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/frontier/billing/checkout"
	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/invoice"
	"github.com/raystack/frontier/billing/plan"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/billing/subscription"
	"github.com/raystack/frontier/billing/usage"
	"github.com/raystack/frontier/core/aggregates/orgbilling"
	"github.com/raystack/frontier/core/aggregates/orginvoices"
	"github.com/raystack/frontier/core/aggregates/orgprojects"
	"github.com/raystack/frontier/core/aggregates/orgserviceuser"
	svc "github.com/raystack/frontier/core/aggregates/orgserviceusercredentials"
	"github.com/raystack/frontier/core/aggregates/orgtokens"
	"github.com/raystack/frontier/core/aggregates/orgusers"
	"github.com/raystack/frontier/core/aggregates/projectusers"
	"github.com/raystack/frontier/core/aggregates/userorgs"
	"github.com/raystack/frontier/core/aggregates/userprojects"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/authenticate"
	frontiersession "github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/core/event"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/kyc"
	"github.com/raystack/frontier/core/metaschema"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/prospect"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/core/webhook"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/salt/rql"
)

type PermissionService interface {
	Get(ctx context.Context, id string) (permission.Permission, error)
	List(ctx context.Context, filter permission.Filter) ([]permission.Permission, error)
	Upsert(ctx context.Context, perm permission.Permission) (permission.Permission, error)
	Update(ctx context.Context, perm permission.Permission) (permission.Permission, error)
}

type BootstrapService interface {
	AppendSchema(ctx context.Context, definition schema.ServiceDefinition) error
}

type WebhookService interface {
	CreateEndpoint(ctx context.Context, endpoint webhook.Endpoint) (webhook.Endpoint, error)
	UpdateEndpoint(ctx context.Context, endpoint webhook.Endpoint) (webhook.Endpoint, error)
	DeleteEndpoint(ctx context.Context, id string) error
	ListEndpoints(ctx context.Context, filter webhook.EndpointFilter) ([]webhook.Endpoint, error)
}

type UserOrgsService interface {
	Search(ctx context.Context, id string, query *rql.Query) (userorgs.UserOrgs, error)
}

type DomainService interface {
	Get(ctx context.Context, id string) (domain.Domain, error)
	List(ctx context.Context, flt domain.Filter) ([]domain.Domain, error)
	ListJoinableOrgsByDomain(ctx context.Context, email string) ([]string, error)
	Delete(ctx context.Context, id string) error
	Create(ctx context.Context, toCreate domain.Domain) (domain.Domain, error)
	VerifyDomain(ctx context.Context, id string) (domain.Domain, error)
	Join(ctx context.Context, orgID string, userID string) error
}

type EntitlementService interface {
	Check(ctx context.Context, customerID, featureID string) (bool, error)
	CheckPlanEligibility(ctx context.Context, customerID string) error
}

type OrgBillingService interface {
	Search(ctx context.Context, query *rql.Query) (orgbilling.OrgBilling, error)
	Export(ctx context.Context) ([]byte, string, error)
}

type MetaSchemaService interface {
	Get(ctx context.Context, id string) (metaschema.MetaSchema, error)
	Create(ctx context.Context, toCreate metaschema.MetaSchema) (metaschema.MetaSchema, error)
	List(ctx context.Context) ([]metaschema.MetaSchema, error)
	Update(ctx context.Context, id string, toUpdate metaschema.MetaSchema) (metaschema.MetaSchema, error)
	Delete(ctx context.Context, id string) error
	Validate(schema metadata.Metadata, data string) error
}

type ProductService interface {
	GetByID(ctx context.Context, id string) (product.Product, error)
	Create(ctx context.Context, product product.Product) (product.Product, error)
	Update(ctx context.Context, product product.Product) (product.Product, error)
	List(ctx context.Context, filter product.Filter) ([]product.Product, error)
	UpsertFeature(ctx context.Context, feature product.Feature) (product.Feature, error)
	GetFeatureByID(ctx context.Context, id string) (product.Feature, error)
	ListFeatures(ctx context.Context, filter product.Filter) ([]product.Feature, error)
}

type OrganizationService interface {
	Get(ctx context.Context, idOrSlug string) (organization.Organization, error)
	GetRaw(ctx context.Context, idOrSlug string) (organization.Organization, error)
	Create(ctx context.Context, org organization.Organization) (organization.Organization, error)
	AdminCreate(ctx context.Context, org organization.Organization, ownerEmail string) (organization.Organization, error)
	List(ctx context.Context, f organization.Filter) ([]organization.Organization, error)
	Update(ctx context.Context, toUpdate organization.Organization) (organization.Organization, error)
	ListByUser(ctx context.Context, principal authenticate.Principal, flt organization.Filter) ([]organization.Organization, error)
	AddUsers(ctx context.Context, orgID string, userID []string) error
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
}

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error)
	GetByEmail(ctx context.Context, email string) (user.User, error)
	Create(ctx context.Context, user user.User) (user.User, error)
	List(ctx context.Context, flt user.Filter) ([]user.User, error)
	ListByOrg(ctx context.Context, orgID string, roleFilter string) ([]user.User, error)
	ListByGroup(ctx context.Context, groupID string, roleFilter string) ([]user.User, error)
	Update(ctx context.Context, toUpdate user.User) (user.User, error)
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
	IsSudo(ctx context.Context, id string, permissionName string) (bool, error)
	Sudo(ctx context.Context, id string, relationName string) error
	UnSudo(ctx context.Context, id string) error
	Search(ctx context.Context, rql *rql.Query) (user.SearchUserResponse, error)
	Export(ctx context.Context) ([]byte, string, error)
}

type RelationService interface {
	Get(ctx context.Context, id string) (relation.Relation, error)
	Create(ctx context.Context, rel relation.Relation) (relation.Relation, error)
	List(ctx context.Context, f relation.Filter) ([]relation.Relation, error)
	Delete(ctx context.Context, rel relation.Relation) error
}

type CheckoutService interface {
	Create(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
	GetByID(ctx context.Context, id string) (checkout.Checkout, error)
	List(ctx context.Context, filter checkout.Filter) ([]checkout.Checkout, error)
	Apply(ctx context.Context, ch checkout.Checkout) (*subscription.Subscription, *product.Product, error)
	CreateSessionForPaymentMethod(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
	CreateSessionForCustomerPortal(ctx context.Context, ch checkout.Checkout) (checkout.Checkout, error)
}

type ProspectService interface {
	Create(ctx context.Context, prospect prospect.Prospect) (prospect.Prospect, error)
	List(ctx context.Context, query *rql.Query) (prospect.ListProspects, error)
	Get(ctx context.Context, prospectId string) (prospect.Prospect, error)
	Update(ctx context.Context, prospect prospect.Prospect) (prospect.Prospect, error)
	Delete(ctx context.Context, prospectId string) error
}

type SubscriptionService interface {
	GetByID(ctx context.Context, id string) (subscription.Subscription, error)
	List(ctx context.Context, filter subscription.Filter) ([]subscription.Subscription, error)
	Cancel(ctx context.Context, id string, immediate bool) (subscription.Subscription, error)
	ChangePlan(ctx context.Context, id string, change subscription.ChangeRequest) (subscription.Phase, error)
	HasUserSubscribedBefore(ctx context.Context, customerID string, planID string) (bool, error)
}

type RoleService interface {
	Get(ctx context.Context, id string) (role.Role, error)
	Upsert(ctx context.Context, toCreate role.Role) (role.Role, error)
	List(ctx context.Context, f role.Filter) ([]role.Role, error)
	Update(ctx context.Context, toUpdate role.Role) (role.Role, error)
	Delete(ctx context.Context, id string) error
}

type OrgServiceUserCredentialsService interface {
	Search(ctx context.Context, id string, query *rql.Query) (svc.OrganizationServiceUserCredentials, error)
}

type PlanService interface {
	GetByID(ctx context.Context, id string) (plan.Plan, error)
	Create(ctx context.Context, plan plan.Plan) (plan.Plan, error)
	List(ctx context.Context, filter plan.Filter) ([]plan.Plan, error)
	UpsertPlans(ctx context.Context, planFile plan.File) error
}

type PreferenceService interface {
	Create(ctx context.Context, preference preference.Preference) (preference.Preference, error)
	Describe(ctx context.Context) []preference.Trait
	List(ctx context.Context, filter preference.Filter) ([]preference.Preference, error)
	LoadPlatformPreferences(ctx context.Context) (map[string]string, error)
}

type PolicyService interface {
	Get(ctx context.Context, id string) (policy.Policy, error)
	List(ctx context.Context, f policy.Filter) ([]policy.Policy, error)
	Create(ctx context.Context, pol policy.Policy) (policy.Policy, error)
	Delete(ctx context.Context, id string) error
	ListRoles(ctx context.Context, principalType, principalID, objectNamespace, objectID string) ([]role.Role, error)
}

type ResourceService interface {
	Get(ctx context.Context, id string) (resource.Resource, error)
	List(ctx context.Context, flt resource.Filter) ([]resource.Resource, error)
	Create(ctx context.Context, resource resource.Resource) (resource.Resource, error)
	Update(ctx context.Context, resource resource.Resource) (resource.Resource, error)
	Delete(ctx context.Context, namespace, id string) error
	CheckAuthz(ctx context.Context, check resource.Check) (bool, error)
	BatchCheck(ctx context.Context, checks []resource.Check) ([]relation.CheckPair, error)
}

type UserProjectsService interface {
	Search(ctx context.Context, userId string, orgId string, query *rql.Query) (userprojects.UserProjects, error)
}

type CustomerService interface {
	GetByID(ctx context.Context, id string) (customer.Customer, error)
	GetByOrgID(ctx context.Context, orgID string) (customer.Customer, error)
	Create(ctx context.Context, customer customer.Customer, offline bool) (customer.Customer, error)
	List(ctx context.Context, filter customer.Filter) ([]customer.Customer, error)
	Delete(ctx context.Context, id string) error
	ListPaymentMethods(ctx context.Context, id string) ([]customer.PaymentMethod, error)
	Update(ctx context.Context, customer customer.Customer) (customer.Customer, error)
	RegisterToProviderIfRequired(ctx context.Context, customerID string) (customer.Customer, error)
	Disable(ctx context.Context, id string) error
	Enable(ctx context.Context, id string) error
	UpdateCreditMinByID(ctx context.Context, customerID string, limit int64) (customer.Details, error)
	GetDetails(ctx context.Context, customerID string) (customer.Details, error)
	UpdateDetails(ctx context.Context, customerID string, details customer.Details) (customer.Details, error)
}

type InvoiceService interface {
	List(ctx context.Context, filter invoice.Filter) ([]invoice.Invoice, error)
	ListAll(ctx context.Context, filter invoice.Filter) ([]invoice.Invoice, error)
	GetUpcoming(ctx context.Context, customerID string) (invoice.Invoice, error)
	TriggerCreditOverdraftInvoices(ctx context.Context) error
	SearchInvoices(ctx context.Context, rqlQuery *rql.Query) ([]invoice.InvoiceWithOrganization, error)
}

type OrgInvoicesService interface {
	Search(ctx context.Context, id string, query *rql.Query) (orginvoices.OrganizationInvoices, error)
}

type OrgProjectsService interface {
	Search(ctx context.Context, id string, query *rql.Query) (orgprojects.OrgProjects, error)
	Export(ctx context.Context, orgID string) ([]byte, string, error)
}

type ServiceUserService interface {
	List(ctx context.Context, flt serviceuser.Filter) ([]serviceuser.ServiceUser, error)
	ListAll(ctx context.Context) ([]serviceuser.ServiceUser, error)
	Create(ctx context.Context, serviceUser serviceuser.ServiceUser) (serviceuser.ServiceUser, error)
	Get(ctx context.Context, id string) (serviceuser.ServiceUser, error)
	Delete(ctx context.Context, id string) error
	ListKeys(ctx context.Context, serviceUserID string) ([]serviceuser.Credential, error)
	CreateKey(ctx context.Context, cred serviceuser.Credential) (serviceuser.Credential, error)
	GetKey(ctx context.Context, credID string) (serviceuser.Credential, error)
	DeleteKey(ctx context.Context, credID string) error
	CreateSecret(ctx context.Context, credential serviceuser.Credential) (serviceuser.Secret, error)
	ListSecret(ctx context.Context, serviceUserID string) ([]serviceuser.Credential, error)
	DeleteSecret(ctx context.Context, credID string) error
	CreateToken(ctx context.Context, credential serviceuser.Credential) (serviceuser.Token, error)
	ListToken(ctx context.Context, serviceUserID string) ([]serviceuser.Credential, error)
	DeleteToken(ctx context.Context, credID string) error
	ListByOrg(ctx context.Context, orgID string) ([]serviceuser.ServiceUser, error)
	IsSudo(ctx context.Context, id string, permissionName string) (bool, error)
	Sudo(ctx context.Context, id string, relationName string) error
	UnSudo(ctx context.Context, id string) error
	GetByIDs(ctx context.Context, ids []string) ([]serviceuser.ServiceUser, error)
}

type KycService interface {
	GetKyc(context.Context, string) (kyc.KYC, error)
	SetKyc(context.Context, kyc.KYC) (kyc.KYC, error)
	ListKycs(context.Context) ([]kyc.KYC, error)
}

type InvitationService interface {
	Get(ctx context.Context, id uuid.UUID) (invitation.Invitation, error)
	List(ctx context.Context, filter invitation.Filter) ([]invitation.Invitation, error)
	ListByUser(ctx context.Context, userID string) ([]invitation.Invitation, error)
	Create(ctx context.Context, inv invitation.Invitation) (invitation.Invitation, error)
	Accept(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type GroupService interface {
	Create(ctx context.Context, grp group.Group) (group.Group, error)
	Get(ctx context.Context, id string) (group.Group, error)
	List(ctx context.Context, flt group.Filter) ([]group.Group, error)
	Update(ctx context.Context, grp group.Group) (group.Group, error)
	ListByUser(ctx context.Context, principalId, principalType string, flt group.Filter) ([]group.Group, error)
	AddUsers(ctx context.Context, groupID string, userID []string) error
	RemoveUsers(ctx context.Context, groupID string, userID []string) error
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

type EventService interface {
	BillingWebhook(ctx context.Context, event event.ProviderWebhookEvent) error
}

type OrgTokensService interface {
	Search(ctx context.Context, id string, query *rql.Query) (orgtokens.OrganizationTokens, error)
	Export(ctx context.Context, orgID string) ([]byte, string, error)
}

type OrgServiceUserService interface {
	Search(ctx context.Context, orgID string, query *rql.Query) (orgserviceuser.OrganizationServiceUsers, error)
}

type ProjectUsersService interface {
	Search(ctx context.Context, id string, query *rql.Query) (projectusers.ProjectUsers, error)
}

type CreditService interface {
	List(ctx context.Context, filter credit.Filter) ([]credit.Transaction, error)
	GetBalance(ctx context.Context, accountID string) (int64, error)
	GetTotalDebitedAmount(ctx context.Context, accountID string) (int64, error)
}

type UsageService interface {
	Report(ctx context.Context, usages []usage.Usage) error
	Revert(ctx context.Context, accountID, usageID string, amount int64) error
}

type ProjectService interface {
	Get(ctx context.Context, idOrName string) (project.Project, error)
	Create(ctx context.Context, prj project.Project) (project.Project, error)
	List(ctx context.Context, f project.Filter) ([]project.Project, error)
	ListByUser(ctx context.Context, principalID, principalType string, flt project.Filter) ([]project.Project, error)
	Update(ctx context.Context, toUpdate project.Project) (project.Project, error)
	ListUsers(ctx context.Context, id string, permissionFilter string) ([]user.User, error)
	ListServiceUsers(ctx context.Context, id string, permissionFilter string) ([]serviceuser.ServiceUser, error)
	ListGroups(ctx context.Context, id string) ([]group.Group, error)
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
}

type OrgUsersService interface {
	Search(ctx context.Context, id string, query *rql.Query) (orgusers.OrgUsers, error)
	Export(ctx context.Context, orgID string) ([]byte, string, error)
}

type AuthnService interface {
	StartFlow(ctx context.Context, request authenticate.RegistrationStartRequest) (*authenticate.RegistrationStartResponse, error)
	FinishFlow(ctx context.Context, request authenticate.RegistrationFinishRequest) (*authenticate.RegistrationFinishResponse, error)
	BuildToken(ctx context.Context, principal authenticate.Principal, metadata map[string]string) ([]byte, error)
	JWKs(ctx context.Context) jwk.Set
	GetPrincipal(ctx context.Context, via ...authenticate.ClientAssertion) (authenticate.Principal, error)
	SupportedStrategies() []string
	InitFlows(ctx context.Context) error
	SanitizeReturnToURL(url string) string
	SanitizeCallbackURL(url string) string
}

type SessionService interface {
	ExtractFromContext(ctx context.Context) (*frontiersession.Session, error)
	Create(ctx context.Context, userID string, metadata frontiersession.SessionMetadata) (*frontiersession.Session, error)
	GetByID(ctx context.Context, sessionID uuid.UUID) (*frontiersession.Session, error)
	Refresh(ctx context.Context, sessionID uuid.UUID) error
	List(ctx context.Context, userID string) ([]*frontiersession.Session, error)
	Delete(ctx context.Context, sessionID uuid.UUID) error
	Ping(ctx context.Context, sessionID uuid.UUID, metadata frontiersession.SessionMetadata) error
}

type NamespaceService interface {
	Get(ctx context.Context, id string) (namespace.Namespace, error)
	List(ctx context.Context) ([]namespace.Namespace, error)
	Upsert(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
	Update(ctx context.Context, ns namespace.Namespace) (namespace.Namespace, error)
}

type CascadeDeleter interface {
	DeleteProject(ctx context.Context, id string) error
	DeleteOrganization(ctx context.Context, id string) error
	RemoveUsersFromOrg(ctx context.Context, orgID string, userIDs []string) error
	DeleteUser(ctx context.Context, userID string) error
}

type AuditRecordService interface {
	Create(ctx context.Context, record auditrecord.AuditRecord) (auditrecord.AuditRecord, bool, error)
	List(ctx context.Context, query *rql.Query) (auditrecord.AuditRecordsList, error)
	Export(ctx context.Context, query *rql.Query) (io.Reader, string, error)
}
