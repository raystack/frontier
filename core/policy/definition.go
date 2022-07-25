package policy

import (
	"github.com/odpf/shield/core/action"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/role"
)

var DefinitionOrganizationManage = Policy{
	NamespaceID: namespace.DefinitionOrg.ID,
	RoleID:      role.DefinitionOrganizationAdmin.ID,
	ActionID:    action.DefinitionManageOrganization.ID,
}

var DefinitionCreateProject = Policy{
	NamespaceID: namespace.DefinitionOrg.ID,
	RoleID:      role.DefinitionOrganizationAdmin.ID,
	ActionID:    action.DefinitionCreateProject.ID,
}

var DefinitionCreateTeam = Policy{
	NamespaceID: namespace.DefinitionOrg.ID,
	RoleID:      role.DefinitionOrganizationAdmin.ID,
	ActionID:    action.DefinitionCreateTeam.ID,
}

var DefinitionManageTeam = Policy{
	NamespaceID: namespace.DefinitionTeam.ID,
	RoleID:      role.DefinitionTeamAdmin.ID,
	ActionID:    action.DefinitionManageTeam.ID,
}

var DefinitionViewTeamAdmin = Policy{
	NamespaceID: namespace.DefinitionTeam.ID,
	RoleID:      role.DefinitionTeamAdmin.ID,
	ActionID:    action.DefinitionViewTeam.ID,
}

var DefinitionViewTeamMember = Policy{
	NamespaceID: namespace.DefinitionTeam.ID,
	RoleID:      role.DefinitionTeamMember.ID,
	ActionID:    action.DefinitionViewTeam.ID,
}

var DefinitionManageProject = Policy{
	NamespaceID: namespace.DefinitionProject.ID,
	RoleID:      role.DefinitionProjectAdmin.ID,
	ActionID:    action.DefinitionManageProject.ID,
}

var DefinitionManageProjectOrg = Policy{
	NamespaceID: namespace.DefinitionProject.ID,
	RoleID:      role.DefinitionOrganizationAdmin.ID,
	ActionID:    action.DefinitionManageProject.ID,
}

var DefinitionTeamOrgAdmin = Policy{
	NamespaceID: namespace.DefinitionTeam.ID,
	RoleID:      role.DefinitionOrganizationAdmin.ID,
	ActionID:    action.DefinitionTeamAll.ID,
}

var DefinitionManageTeamOrgAdmin = Policy{
	NamespaceID: namespace.DefinitionTeam.ID,
	RoleID:      role.DefinitionOrganizationAdmin.ID,
	ActionID:    action.DefinitionManageTeam.ID,
}

var DefinitionProjectOrgAdmin = Policy{
	NamespaceID: namespace.DefinitionTeam.ID,
	RoleID:      role.DefinitionProjectAdmin.ID,
	ActionID:    action.DefinitionTeamAll.ID,
}
