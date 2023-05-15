package api

import (
	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/authenticate"
	"github.com/odpf/shield/core/authenticate/session"
	"github.com/odpf/shield/core/deleter"
	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/metaschema"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/policy"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/core/rule"
	"github.com/odpf/shield/core/user"
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
	ActionService       *action.Service
	RelationService     *relation.Service
	ResourceService     *resource.Service
	RuleService         *rule.Service
	SessionService      *session.Service
	RegistrationService *authenticate.RegistrationService
	DeleterService      *deleter.Service
	MetaSchemaService   *metaschema.Service
}
