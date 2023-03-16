package api

import (
	"github.com/goto/shield/core/action"
	"github.com/goto/shield/core/group"
	"github.com/goto/shield/core/namespace"
	"github.com/goto/shield/core/organization"
	"github.com/goto/shield/core/policy"
	"github.com/goto/shield/core/project"
	"github.com/goto/shield/core/relation"
	"github.com/goto/shield/core/resource"
	"github.com/goto/shield/core/role"
	"github.com/goto/shield/core/rule"
	"github.com/goto/shield/core/user"
)

type Deps struct {
	OrgService       *organization.Service
	ProjectService   *project.Service
	GroupService     *group.Service
	RoleService      *role.Service
	PolicyService    *policy.Service
	UserService      *user.Service
	NamespaceService *namespace.Service
	ActionService    *action.Service
	RelationService  *relation.Service
	ResourceService  *resource.Service
	RuleService      *rule.Service
}
