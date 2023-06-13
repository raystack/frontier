package api

import (
	"github.com/raystack/shield/core/authenticate"
	"github.com/raystack/shield/core/authenticate/session"
	"github.com/raystack/shield/core/deleter"
	"github.com/raystack/shield/core/group"
	"github.com/raystack/shield/core/invitation"
	"github.com/raystack/shield/core/metaschema"
	"github.com/raystack/shield/core/namespace"
	"github.com/raystack/shield/core/organization"
	"github.com/raystack/shield/core/permission"
	"github.com/raystack/shield/core/policy"
	"github.com/raystack/shield/core/project"
	"github.com/raystack/shield/core/relation"
	"github.com/raystack/shield/core/resource"
	"github.com/raystack/shield/core/role"
	"github.com/raystack/shield/core/rule"
	"github.com/raystack/shield/core/user"
	"github.com/raystack/shield/internal/bootstrap"
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
	RegistrationService *authenticate.RegistrationService
	DeleterService      *deleter.Service
	MetaSchemaService   *metaschema.Service
	BootstrapService    *bootstrap.Service
	InvitationService   *invitation.Service
}
