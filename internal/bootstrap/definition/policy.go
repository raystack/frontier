package definition

import "github.com/odpf/shield/model"

var OrganizationManagePolicy = model.Policy{
	NamespaceId: OrgNamespace.Id,
	RoleId:      OrganizationAdminRole.Id,
	ActionId:    ManageOrganizationAction.Id,
}

var CreateProjectPolicy = model.Policy{
	NamespaceId: OrgNamespace.Id,
	RoleId:      OrganizationAdminRole.Id,
	ActionId:    CreateProjectAction.Id,
}

var CreateTeamPolicy = model.Policy{
	NamespaceId: OrgNamespace.Id,
	RoleId:      OrganizationAdminRole.Id,
	ActionId:    CreateTeamAction.Id,
}

var ManageTeamPolicy = model.Policy{
	NamespaceId: TeamNamespace.Id,
	RoleId:      TeamAdminRole.Id,
	ActionId:    ManageTeamAction.Id,
}

var ManageTeamOrgAdminPolicy = model.Policy{
	NamespaceId: TeamNamespace.Id,
	RoleId:      OrganizationAdminRole.Id,
	ActionId:    ManageTeamAction.Id,
}

var ViewTeamAdminPolicy = model.Policy{
	NamespaceId: TeamNamespace.Id,
	RoleId:      TeamAdminRole.Id,
	ActionId:    ViewTeamAction.Id,
}

var ViewTeamMemberPolicy = model.Policy{
	NamespaceId: TeamNamespace.Id,
	RoleId:      TeamMemberRole.Id,
	ActionId:    ViewTeamAction.Id,
}

var ManageProjectPolicy = model.Policy{
	NamespaceId: ProjectNamespace.Id,
	RoleId:      ProjectAdminRole.Id,
	ActionId:    ManageProjectAction.Id,
}

var ManageProjectOrgPolicy = model.Policy{
	NamespaceId: ProjectNamespace.Id,
	RoleId:      OrganizationAdminRole.Id,
	ActionId:    ManageProjectAction.Id,
}
