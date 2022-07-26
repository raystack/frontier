package policy

import (
	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
)

var DefinitionOrganizationManage = Policy{
	NamespaceId: namespace.DefinitionOrg.Id,
	RoleId:      role.DefinitionOrganizationAdmin.Id,
	ActionId:    action.DefinitionManageOrganization.Id,
}

var DefinitionCreateProject = Policy{
	NamespaceId: namespace.DefinitionOrg.Id,
	RoleId:      role.DefinitionOrganizationAdmin.Id,
	ActionId:    action.DefinitionCreateProject.Id,
}

var DefinitionCreateTeam = Policy{
	NamespaceId: namespace.DefinitionOrg.Id,
	RoleId:      role.DefinitionOrganizationAdmin.Id,
	ActionId:    action.DefinitionCreateTeam.Id,
}

var DefinitionManageTeam = Policy{
	NamespaceId: namespace.DefinitionTeam.Id,
	RoleId:      role.DefinitionTeamAdmin.Id,
	ActionId:    action.DefinitionManageTeam.Id,
}

var DefinitionViewTeamAdmin = Policy{
	NamespaceId: namespace.DefinitionTeam.Id,
	RoleId:      role.DefinitionTeamAdmin.Id,
	ActionId:    action.DefinitionViewTeam.Id,
}

var DefinitionViewTeamMember = Policy{
	NamespaceId: namespace.DefinitionTeam.Id,
	RoleId:      role.DefinitionTeamMember.Id,
	ActionId:    action.DefinitionViewTeam.Id,
}

var DefinitionManageProject = Policy{
	NamespaceId: namespace.DefinitionProject.Id,
	RoleId:      role.DefinitionProjectAdmin.Id,
	ActionId:    action.DefinitionManageProject.Id,
}

var DefinitionManageProjectOrg = Policy{
	NamespaceId: namespace.DefinitionProject.Id,
	RoleId:      role.DefinitionOrganizationAdmin.Id,
	ActionId:    action.DefinitionManageProject.Id,
}

var DefinitionTeamOrgAdmin = Policy{
	NamespaceId: namespace.DefinitionTeam.Id,
	RoleId:      role.DefinitionOrganizationAdmin.Id,
	ActionId:    action.DefinitionTeamAll.Id,
}

var DefinitionManageTeamOrgAdmin = Policy{
	NamespaceId: namespace.DefinitionTeam.Id,
	RoleId:      role.DefinitionOrganizationAdmin.Id,
	ActionId:    action.DefinitionManageTeam.Id,
}

var DefinitionProjectOrgAdmin = Policy{
	NamespaceId: namespace.DefinitionTeam.Id,
	RoleId:      role.DefinitionProjectAdmin.Id,
	ActionId:    action.DefinitionTeamAll.Id,
}
