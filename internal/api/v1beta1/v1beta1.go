package v1beta1

import (
	"github.com/raystack/shield/internal/api"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	"google.golang.org/grpc"
)

type Handler struct {
	shieldv1beta1.UnimplementedShieldServiceServer
	shieldv1beta1.UnimplementedAdminServiceServer

	DisableOrgsListing  bool
	DisableUsersListing bool
	orgService          OrganizationService
	projectService      ProjectService
	groupService        GroupService
	roleService         RoleService
	policyService       PolicyService
	userService         UserService
	namespaceService    NamespaceService
	permissionService   PermissionService
	relationService     RelationService
	resourceService     ResourceService
	ruleService         RuleService
	sessionService      SessionService
	authnService        AuthnService
	deleterService      CascadeDeleter
	metaSchemaService   MetaSchemaService
	bootstrapService    BootstrapService
	invitationService   InvitationService
	serviceUserService  ServiceUserService
	auditService        AuditService
}

func Register(s *grpc.Server, deps api.Deps) error {
	handler := &Handler{
		DisableOrgsListing:  deps.DisableOrgsListing,
		DisableUsersListing: deps.DisableUsersListing,
		orgService:          deps.OrgService,
		projectService:      deps.ProjectService,
		groupService:        deps.GroupService,
		roleService:         deps.RoleService,
		policyService:       deps.PolicyService,
		userService:         deps.UserService,
		namespaceService:    deps.NamespaceService,
		permissionService:   deps.PermissionService,
		relationService:     deps.RelationService,
		resourceService:     deps.ResourceService,
		ruleService:         deps.RuleService,
		sessionService:      deps.SessionService,
		authnService:        deps.AuthnService,
		deleterService:      deps.DeleterService,
		metaSchemaService:   deps.MetaSchemaService,
		bootstrapService:    deps.BootstrapService,
		invitationService:   deps.InvitationService,
		serviceUserService:  deps.ServiceUserService,
		auditService:        deps.AuditService,
	}
	s.RegisterService(&shieldv1beta1.ShieldService_ServiceDesc, handler)
	s.RegisterService(&shieldv1beta1.AdminService_ServiceDesc, handler)
	return nil
}
