package api

import (
	"github.com/raystack/shield/core/action"
	"github.com/raystack/shield/core/authenticate"
	"github.com/raystack/shield/core/authenticate/session"
	"github.com/raystack/shield/core/group"
	"github.com/raystack/shield/core/namespace"
	"github.com/raystack/shield/core/organization"
	"github.com/raystack/shield/core/policy"
	"github.com/raystack/shield/core/project"
	"github.com/raystack/shield/core/relation"
	"github.com/raystack/shield/core/resource"
	"github.com/raystack/shield/core/role"
	"github.com/raystack/shield/core/rule"
	"github.com/raystack/shield/core/user"
)

type Deps struct {
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
}
