package api

import (
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/authenticate/session"
	"github.com/raystack/frontier/core/deleter"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/metaschema"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/permission"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/preference"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/rule"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap"
)

type Deps struct {
	DisableOrgsListing  bool
	DisableUsersListing bool
	OrgService          *organization.Service
	ProjectService      *project.Service
	GroupService        *group.Service
	RoleService         *role.Service
	PolicyService       *policy.Service
	UserService         *user.Service
	NamespaceService    *namespace.Service
	PermissionService   *permission.Service
	RelationService     *relation.Service
	ResourceService     *resource.Service
	RuleService         *rule.Service
	SessionService      *session.Service
	AuthnService        *authenticate.Service
	DeleterService      *deleter.Service
	MetaSchemaService   *metaschema.Service
	BootstrapService    *bootstrap.Service
	InvitationService   *invitation.Service
	ServiceUserService  *serviceuser.Service
	AuditService        *audit.Service
	DomainService       *domain.Service
	PreferenceService   *preference.Service
}
